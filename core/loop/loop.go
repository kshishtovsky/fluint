// Package loop implements the engine's frame scheduler and event demultiplexer.
//
// The Loop owns the synchronous frame pipeline:
//
//	tty (raw bytes) ──► Parser ──► typed event channels
//	  │
//	  └──► ticker @ FPS ──► Flush() ──► Diff ──► Render ──► tty.Write
//
// Loop is the single owner of terminal I/O while it is running. It coordinates
// three goroutines:
//
//   - the I/O reader, which parses raw bytes from the terminal into typed
//     event channels and never blocks the render path (drops events on
//     full channels via non-blocking selects);
//   - the resize watcher (Unix only), which forwards SIGWINCH into the
//     resize channel;
//   - the frame scheduler, which fires Flush on a fixed-cadence ticker.
//
// Hot-path invariants (steady state, no input, no changes):
//
//   - Flush performs zero heap allocations;
//   - the scheduler ticker never allocates;
//   - event-channel dispatches are non-blocking and allocation-free.
package loop

import (
	"errors"
	"sync"
	"time"

	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/diff"
	"github.com/kshishtovsky/fluint/internal/platform"
	"github.com/kshishtovsky/fluint/internal/term"
	"github.com/kshishtovsky/fluint/render/ansi"
)

// Event type discriminator constants for the public Event struct.
//
// The internal/term package already defines EventKey/EventMouse for the
// parser's tagged union. Here we expose the same discriminator values plus
// EventResize and EventError for the loop-level event surface. The values
// are kept in sync with internal/term so callers can reuse a single
// discriminator if they only need input events.
const (
	// EventKey identifies a keyboard event.
	EventKey = term.EventKey
	// EventMouse identifies a mouse event.
	EventMouse = term.EventMouse
	// EventResize identifies a terminal-resize notification. Resize events
	// carry no payload — callers query the terminal dimensions via Term.
	EventResize = 3
	// EventError identifies an asynchronous error forwarded from I/O or
	// render. Inspect Err for the underlying cause.
	EventError = 4
)

// Event is a tagged-union value emitted on typed channels. It mirrors the
// discriminator of internal/term.InputEvent and adds Resize and Error
// variants for loop-level signals.
type Event struct {
	// Type discriminates the payload. See Event* constants.
	Type int
	// Key holds the keyboard payload when Type == EventKey.
	Key term.KeyEvent
	// Mouse holds the mouse payload when Type == EventMouse.
	Mouse term.MouseEvent
	// Err holds the underlying error when Type == EventError.
	Err error
}

// DefaultFPS is the steady-state target frame rate for the scheduler.
// 60 FPS keeps the frame budget at ~16.6ms which is sufficient for smooth
// animations while leaving headroom for input parsing.
const DefaultFPS = 60

// Default channel capacities.
//
// KeyEvents is sized to absorb a small burst (e.g. paste) without
// dropping. MouseEvents is smaller because mouse streams are dense.
// ResizeEvents / Errors are tiny because they are rare and informational.
const (
	defaultKeyCap   = 64
	defaultMouseCap = 32
	defaultResizeOp = 1
	defaultErrorCap = 8
)

// Channel buffer size for raw byte reads from the terminal. 4 KiB matches
// the typical pipe/tty buffer and keeps parsing latency bounded.
const readBufSize = 4096

// ioTerm is the minimal terminal surface the loop requires. It mirrors
// the read/write surface of *platform.Terminal so production code can use
// the real handle, while tests can plug in a fake.
type ioTerm interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
}

// Loop is the engine's main scheduler. All fields are zero-value usable
// after NewLoop returns; Start() drives the goroutines.
//
// Concurrency:
//
//   - Field reads after Start() are safe without locks for the typed
//     channels and the ticker.
//   - Flush() is the only method that may be called concurrently with
//     Start(); it acquires a mutex to serialise against the scheduler.
//   - Stop() is safe to call from any goroutine; it is idempotent.
type Loop struct {
	// Term is the platform terminal handle. It must outlive the Loop.
	Term *platform.Terminal

	// io is the indirection used by ioLoop and Flush. In production it
	// points to Term; in tests it is a fake.
	io ioTerm

	// FrontBuf is the cell grid previously presented to the user.
	FrontBuf *buffer.Buffer
	// BackBuf is the cell grid currently being mutated by the application.
	BackBuf *buffer.Buffer

	// Differ computes the minimal changeset between front and back buffers.
	Differ *diff.Differ
	// Renderer serialises a changeset into an ANSI byte payload.
	Renderer *ansi.Renderer

	// KeyEvents receives keyboard events. Buffered; full channels drop.
	KeyEvents chan term.KeyEvent
	// MouseEvents receives mouse events. Buffered; full channels drop.
	MouseEvents chan term.MouseEvent
	// ResizeEvents receives a struct{} on every terminal resize.
	ResizeEvents chan struct{}
	// Errors receives asynchronous errors from I/O and render.
	Errors chan error

	// Ticker is the frame scheduler. Fires every 1/FPS.
	Ticker *time.Ticker
	// Quit is closed by Stop() to signal every internal goroutine.
	Quit chan struct{}

	// fps is the cached frame rate; derived from Ticker period on Start.
	fps int

	// flushMu serialises Flush() between external callers and the scheduler.
	flushMu sync.Mutex

	// wg tracks the I/O and scheduler goroutines so Stop() can wait for
	// them to exit.
	wg sync.WaitGroup

	// started tracks whether Start() has been called. Guarded by sync.Once.
	once sync.Once
}

// NewLoop constructs a Loop with pre-allocated channels, buffers, differ
// and renderer. The terminal is not owned by the Loop and must be Closed
// by the caller after Stop() returns.
//
// width and height define the initial front/back buffer dimensions. The
// caller is expected to call Term.EnterRawMode before Start().
func NewLoop(t *platform.Terminal, width, height int) *Loop {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	front := buffer.NewBuffer(width, height)
	back := buffer.NewBuffer(width, height)

	// Pre-allocate the differ for the worst case (every cell changed).
	maxCells := width * height
	if maxCells < 1 {
		maxCells = 1
	}

	l := &Loop{
		Term:         t,
		FrontBuf:     front,
		BackBuf:      back,
		Differ:       diff.NewDiffer(maxCells),
		Renderer:     ansi.NewRenderer(),
		KeyEvents:    make(chan term.KeyEvent, defaultKeyCap),
		MouseEvents:  make(chan term.MouseEvent, defaultMouseCap),
		ResizeEvents: make(chan struct{}, defaultResizeOp),
		Errors:       make(chan error, defaultErrorCap),
		Ticker:       time.NewTicker(time.Second / DefaultFPS),
		Quit:         make(chan struct{}),
		fps:          DefaultFPS,
	}
	if t != nil {
		l.io = t
	}
	return l
}

// newLoopWithIO is the test-friendly constructor. Production code uses
// NewLoop; tests use this entry point to inject a fake ioTerm.
func newLoopWithIO(t ioTerm, width, height int) *Loop {
	l := NewLoop(nil, width, height)
	l.io = t
	return l
}

// Start launches the I/O and scheduler goroutines. It is safe to call
// exactly once; subsequent calls are no-ops.
func (l *Loop) Start() {
	l.once.Do(func() {
		l.wg.Add(2)
		go l.ioLoop()
		go l.schedulerLoop()
	})
}

// Stop signals all goroutines to exit, closes the ticker and waits for
// them to finish. Stop is idempotent and safe to call from any goroutine,
// including before Start has been called.
//
// Stop does NOT close the typed event channels: callers may still drain
// pending events after Stop returns.
func (l *Loop) Stop() {
	l.signalStop()
	l.wg.Wait()
}

// signalStop is the internal half of Stop. It signals every goroutine
// to exit (by closing Quit and stopping the ticker) without waiting
// for them. The I/O goroutine uses this on ErrTTYLost because it is
// itself one of the goroutines that wg tracks; calling the full Stop
// from inside the I/O goroutine would deadlock in wg.Wait.
func (l *Loop) signalStop() {
	select {
	case <-l.Quit:
		// Already stopped.
	default:
		close(l.Quit)
		if l.Ticker != nil {
			l.Ticker.Stop()
		}
	}
}

// Flush executes a single frame: compute the diff, render changes, write
// them to the terminal and update the front buffer. It is safe to call
// concurrently with the scheduler; concurrent calls are serialised.
//
// Flush returns the first error from the underlying terminal Write, or
// nil on success. Errors from Flush are also forwarded to the Errors
// channel.
//
// Hot-path: when BackBuf equals FrontBuf (no application mutations),
// Flush performs zero heap allocations.
func (l *Loop) Flush() error {
	l.flushMu.Lock()
	defer l.flushMu.Unlock()

	changes := l.Differ.Diff(l.FrontBuf, l.BackBuf)

	var payload []byte
	if len(changes) > 0 {
		payload = l.Renderer.Render(changes)
	}

	if len(payload) > 0 {
		if _, err := l.io.Write(payload); err != nil {
			l.sendError(err)
			return err
		}
	}

	// Promote back → front for the next frame. copy on []Cell is alloc-free.
	if len(l.BackBuf.Cells) > 0 {
		copy(l.FrontBuf.Cells, l.BackBuf.Cells)
	}

	return nil
}

// sendError pushes err to the Errors channel without blocking. If the
// channel is full the error is dropped — losing an error is preferable
// to blocking the renderer.
func (l *Loop) sendError(err error) {
	if err == nil {
		return
	}
	select {
	case l.Errors <- err:
	default:
	}
}

// schedulerLoop is the frame scheduler goroutine. It fires Flush on every
// ticker beat until Quit is closed.
func (l *Loop) schedulerLoop() {
	defer l.wg.Done()

	for {
		select {
		case <-l.Quit:
			return
		case <-l.Ticker.C:
			_ = l.Flush()
		}
	}
}

// ioLoop reads raw bytes from the terminal, parses escape sequences and
// forwards typed events to the appropriate channel. It exits when Quit is
// closed or when the terminal returns ErrTTYLost, in which case the error
// is forwarded before exit.
func (l *Loop) ioLoop() {
	defer l.wg.Done()

	// 4 KiB stack-allocated scratch — no heap traffic on the hot path.
	var buf [readBufSize]byte

	var parser term.Parser

	for {
		select {
		case <-l.Quit:
			return
		default:
		}

		n, err := l.io.Read(buf[:])
		if err != nil {
			if isTTYLost(err) {
				l.sendError(err)
				// Tear down the loop so callers waiting on Quit unblock.
				// We cannot call the full Stop here — that would wg.Wait
				// from inside this very goroutine.
				l.signalStop()
				return
			}
			// Transient read error — surface and continue. Reads are
			// cheap; we don't want to drop input on a single bad syscall.
			l.sendError(err)
			continue
		}
		if n == 0 {
			continue
		}

		l.parse(&parser, buf[:n])
	}
}

// parse feeds data through the parser and dispatches typed events.
// All dispatches are non-blocking: a full channel drops the event rather
// than stalling the read loop.
func (l *Loop) parse(p *term.Parser, data []byte) {
	// Handle a timed-out standalone ESC before draining new data.
	if p.InEscapeState() {
		if ev, ok := p.Flush(); ok {
			l.dispatchInput(ev)
		}
	}

	for len(data) > 0 {
		ev, consumed, ok := p.Next(data)
		if ok {
			l.dispatchInput(ev)
		}
		if consumed <= 0 {
			// No progress — avoid an infinite loop on a degenerate parser
			// state. Bail out; the parser will resync on the next read.
			return
		}
		data = data[consumed:]
	}
}

// dispatchInput forwards a parsed InputEvent to the appropriate typed
// channel using a non-blocking send. Channels are sized for typical
// bursts; an overflow drops the event to keep the renderer alive.
func (l *Loop) dispatchInput(ev term.InputEvent) {
	switch ev.Type {
	case term.EventKey:
		select {
		case l.KeyEvents <- ev.Key:
		default:
		}
	case term.EventMouse:
		select {
		case l.MouseEvents <- ev.Mouse:
		default:
		}
	}
}

// isTTYLost reports whether err originates from a terminal disconnection.
// errors.Is unwraps the platform wrapper added in platform.Terminal.Read.
func isTTYLost(err error) bool {
	return errors.Is(err, platform.ErrTTYLost)
}
