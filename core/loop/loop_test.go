package loop

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/internal/platform"
	"github.com/kshishtovsky/fluint/internal/term"
)

// errFakeClosed is the sentinel Read error returned by fakeTerm after
// Close. It is intentionally distinct from ErrTTYLost so the loop's
// I/O goroutine treats it as a transient error and re-checks Quit
// instead of tearing down.
var errFakeClosed = errors.New("fake terminal closed")

// fakeTerm is a test double for the ioTerm interface. It feeds a
// scripted sequence of read payloads and records every write call so
// tests can assert on the rendered ANSI output.
type fakeTerm struct {
	mu sync.Mutex

	// done is closed by Close; Read returns errFakeClosed once it fires.
	// Tests should wire this to the loop's Quit channel so Stop()
	// interrupts any in-flight Read.
	done chan struct{}

	// reads is a FIFO of byte slices; each Read returns the next entry.
	reads [][]byte

	// writes records every Write payload.
	writes [][]byte

	// failNext, when > 0, causes the next Write to return writeErr and
	// decrements the counter. Useful for the error path test.
	failNext atomic.Int32

	// writeErr is the error returned by Write when failNext > 0.
	writeErr error

	// readErr, if non-nil, is returned by Read instead of dispatching
	// from the reads queue. Used by the TTY-loss test.
	readErr error
}

// newFakeTerm returns a fakeTerm that can be parked indefinitely in
// Read. Use wireFakeQuit to make the loop's Stop() unblock any in-flight
// Read before wg.Wait() deadlocks.
func newFakeTerm() *fakeTerm {
	return &fakeTerm{done: make(chan struct{})}
}

// closeDone closes the done channel to unblock any parked Read.
func (f *fakeTerm) closeDone() {
	select {
	case <-f.done:
	default:
		close(f.done)
	}
}

// push enqueues a byte payload that the next Read will return. The bytes
// are copied so callers may reuse their buffer.
func (f *fakeTerm) push(b ...byte) {
	cp := make([]byte, len(b))
	copy(cp, b)
	f.mu.Lock()
	f.reads = append(f.reads, cp)
	f.mu.Unlock()
}

func (f *fakeTerm) Read(p []byte) (int, error) {
	if f.readErr != nil {
		return 0, f.readErr
	}
	f.mu.Lock()
	empty := len(f.reads) == 0
	f.mu.Unlock()

	if empty {
		// Park until Close (i.e. f.done) is signalled. The loop's Stop()
		// closes Quit, which the test wires into done, so the I/O
		// goroutine unblocks and re-checks Quit at the top of ioLoop.
		select {
		case <-f.done:
			return 0, errFakeClosed
		}
	}

	f.mu.Lock()
	b := f.reads[0]
	f.reads = f.reads[1:]
	f.mu.Unlock()

	n := copy(p, b)
	return n, nil
}

func (f *fakeTerm) Write(p []byte) (int, error) {
	if f.failNext.Load() > 0 {
		f.failNext.Add(-1)
		err := f.writeErr
		if err == nil {
			err = fmt.Errorf("%w: simulated", platform.ErrWriteFailed)
		}
		return 0, err
	}
	cp := make([]byte, len(p))
	copy(cp, p)
	f.mu.Lock()
	f.writes = append(f.writes, cp)
	f.mu.Unlock()
	return len(p), nil
}

// waitForWrites polls until the number of writes seen reaches want, or
// the deadline expires. Used by tests that need to synchronise with the
// async scheduler.
func (f *fakeTerm) waitForWrites(t *testing.T, want int, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for {
		f.mu.Lock()
		n := len(f.writes)
		f.mu.Unlock()
		if n >= want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("waitForWrites(%d): got %d writes after %v", want, n, d)
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// cellFromRune builds a buffer.Cell carrying a single rune with default
// colors. Used by tests that only care about the rune payload.
func cellFromRune(r rune) buffer.Cell {
	return buffer.Cell{Rune: r}
}

// slowTicker replaces the default 60Hz ticker with one that fires at a
// configurable interval; tests use a slow cadence to keep their
// scheduler-driven writes from drowning out input events.
func slowTicker(l *Loop, d time.Duration) {
	l.Ticker.Stop()
	l.Ticker = time.NewTicker(d)
}

// wireFakeQuit closes fakeTerm.done when l.Quit is closed, so the I/O
// goroutine parked inside fakeTerm.Read can unblock and re-check Quit at
// the top of ioLoop. Must be called before Start.
//
// The helper goroutine itself exits immediately after closing done, so
// it does not leak across the lifetime of the test.
func wireFakeQuit(l *Loop, ft *fakeTerm) {
	go func() {
		<-l.Quit
		ft.closeDone()
	}()
}

// ---------------------------------------------------------------------------
// Event delivery tests
// ---------------------------------------------------------------------------

// TestEventDelivery_PushToTypedChannels feeds a few common terminal input
// sequences through the parser via a fake terminal and asserts that the
// loop forwards them to the correct typed channel.
//
//	"A"        → KeyEvent with Rune 'A'
//	ESC        → KeyEvent with Code KeyEscape
//	CSI A (↑)  → KeyEvent with Code KeyUp
//	SGR mouse  → MouseEvent at the reported coordinates
func TestEventDelivery_PushToTypedChannels(t *testing.T) {
	t.Parallel()

	ft := newFakeTerm()
	ft.push('A')            // printable
	ft.push(0x1B)           // standalone ESC
	ft.push(0x1B, '[', 'A') // arrow up
	ft.push(0x1B, '[', '<', // SGR mouse press, button 0, x=10, y=5
		'0', ';', '1', '1', ';', '6', 'M')

	l := newLoopWithIO(ft, 100, 50)
	slowTicker(l, 20*time.Millisecond)

	wireFakeQuit(l, ft)
	l.Start()
	defer l.Stop()

	// 1) Printable → KeyEvents.
	select {
	case k := <-l.KeyEvents:
		if k.Rune != 'A' {
			t.Fatalf("KeyEvent.Rune = %q, want %q", k.Rune, 'A')
		}
		if k.Code != term.KeyNone {
			t.Fatalf("KeyEvent.Code = %d, want KeyNone", k.Code)
		}
	case <-time.After(time.Second):
		t.Fatal("no printable key event delivered within 1s")
	}

	// 2) ESC → KeyEscape.
	select {
	case k := <-l.KeyEvents:
		if k.Code != term.KeyEscape {
			t.Fatalf("KeyEvent.Code = %d, want KeyEscape", k.Code)
		}
	case <-time.After(time.Second):
		t.Fatal("no ESC key event delivered within 1s")
	}

	// 3) Arrow up.
	select {
	case k := <-l.KeyEvents:
		if k.Code != term.KeyUp {
			t.Fatalf("KeyEvent.Code = %d, want KeyUp", k.Code)
		}
	case <-time.After(time.Second):
		t.Fatal("no arrow-up key event delivered within 1s")
	}

	// 4) Mouse press.
	select {
	case m := <-l.MouseEvents:
		if m.X != 10 || m.Y != 5 {
			t.Fatalf("MouseEvent position = (%d,%d), want (10,5)", m.X, m.Y)
		}
	case <-time.After(time.Second):
		t.Fatal("no mouse event delivered within 1s")
	}
}

// TestDropOnFullChannel verifies that the I/O goroutine drops events
// instead of blocking when the typed channel is full.
//
// We configure a tiny key-event buffer (1 slot) and push more input than
// fits. The reader must drain input without deadlocking, and Flush()
// must still succeed afterwards — proving the renderer has not been
// starved by the over-fed input goroutine.
func TestDropOnFullChannel(t *testing.T) {
	t.Parallel()

	ft := newFakeTerm()
	for i := 0; i < 1000; i++ {
		ft.push('x')
	}

	l := newLoopWithIO(ft, 10, 10)
	l.KeyEvents = make(chan term.KeyEvent, 1) // tiny on purpose
	slowTicker(l, 50*time.Millisecond)

	wireFakeQuit(l, ft)
	l.Start()

	// Drain whatever survived into the 1-slot buffer; the goroutine
	// continues in the background.
	select {
	case <-l.KeyEvents:
	case <-time.After(time.Second):
		t.Fatal("expected at least 1 event on full channel")
	}

	// Flush must still succeed.
	if err := l.Flush(); err != nil {
		t.Fatalf("Flush() after channel overflow = %v", err)
	}

	l.Stop()
}

// ---------------------------------------------------------------------------
// Error path tests
// ---------------------------------------------------------------------------

// TestWriteErrorForwarded verifies that ErrWriteFailed from the terminal
// surfaces on the Errors channel and is also returned by Flush().
func TestWriteErrorForwarded(t *testing.T) {
	t.Parallel()

	ft := newFakeTerm()
	ft.failNext.Store(1)

	l := newLoopWithIO(ft, 10, 5)
	slowTicker(l, 1*time.Hour) // disable scheduler; we drive Flush manually

	// Mutate BackBuf so Diff produces a non-empty changeset and
	// Render emits a non-empty payload that the fake will refuse.
	l.BackBuf.SetCell(2, 1, cellFromRune('Z'))

	err := l.Flush()
	if err == nil {
		t.Fatal("Flush() returned nil, want error from ErrWriteFailed")
	}
	if !errors.Is(err, platform.ErrWriteFailed) {
		t.Fatalf("Flush() error = %v, want wrapping ErrWriteFailed", err)
	}

	// The same error must have been pushed to Errors.
	select {
	case got := <-l.Errors:
		if !errors.Is(got, platform.ErrWriteFailed) {
			t.Fatalf("Errors <- %v, want wrapping ErrWriteFailed", got)
		}
	case <-time.After(time.Second):
		t.Fatal("Flush did not forward ErrWriteFailed to Errors channel")
	}
}

// TestTTYLostStopsLoop verifies that a terminal disconnection is fatal:
// the I/O goroutine exits, the error reaches Errors, and Quit closes.
func TestTTYLostStopsLoop(t *testing.T) {
	t.Parallel()

	// Build a fake that always returns ErrTTYLost from Read.
	ft := newFakeTerm()
	ft.readErr = platform.ErrTTYLost

	l := newLoopWithIO(ft, 10, 5)
	slowTicker(l, 50*time.Millisecond)

	wireFakeQuit(l, ft)
	l.Start()

	// Wait for the I/O goroutine to detect TTY loss, forward the error
	// and tear down. Quit must close.
	select {
	case <-l.Quit:
		// good — Stop already ran from the goroutine.
	case <-time.After(2 * time.Second):
		t.Fatal("Quit channel not closed after ErrTTYLost")
	}

	// The error must be on the Errors channel.
	select {
	case got := <-l.Errors:
		if !errors.Is(got, platform.ErrTTYLost) {
			t.Fatalf("Errors <- %v, want wrapping ErrTTYLost", got)
		}
	case <-time.After(time.Second):
		t.Fatal("ErrTTYLost not forwarded to Errors channel")
	}

	// Stop must be idempotent after the goroutine already closed Quit.
	l.Stop()
}

// ---------------------------------------------------------------------------
// Lifecycle / goroutine leak test
// ---------------------------------------------------------------------------

// TestStopJoinsGoroutines verifies that Stop() waits for every goroutine
// it spawned to exit, and that no goroutines leak across a Start/Stop
// cycle.
//
// We measure runtime.NumGoroutine() before and after; the test passes
// only if the count returns to baseline (with a small slack to account
// for the Go test runner itself).
func TestStopJoinsGoroutines(t *testing.T) {
	t.Parallel()

	// Warm up the runtime so transient goroutines (e.g. signal handling
	// in earlier tests) don't pollute our baseline.
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	base := runtime.NumGoroutine()

	ft := newFakeTerm()
	l := newLoopWithIO(ft, 10, 5)
	slowTicker(l, 5*time.Millisecond)

	wireFakeQuit(l, ft)
	l.Start()
	// Let the scheduler tick a few times.
	time.Sleep(40 * time.Millisecond)

	l.Stop()

	// Give the runtime a moment to reap finished goroutines.
	deadline := time.Now().Add(2 * time.Second)
	for {
		now := runtime.NumGoroutine()
		if now <= base+2 { // small slack for test-runner helpers
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("goroutine leak: baseline=%d, after Stop=%d", base, now)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// TestStopBeforeStartIsSafe ensures Stop() does not panic and does not
// block when invoked before Start().
func TestStopBeforeStartIsSafe(t *testing.T) {
	t.Parallel()
	ft := newFakeTerm()
	l := newLoopWithIO(ft, 10, 5)
	done := make(chan struct{})
	go func() {
		l.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop before Start blocked")
	}
}

// ---------------------------------------------------------------------------
// Flush() behaviour
// ---------------------------------------------------------------------------

// TestFlush_NoChangesReturnsNilAndNoWrite exercises the steady-state
// path: equal buffers, no payload, no Write call.
func TestFlush_NoChangesReturnsNilAndNoWrite(t *testing.T) {
	t.Parallel()

	ft := newFakeTerm()
	l := newLoopWithIO(ft, 20, 10)

	if err := l.Flush(); err != nil {
		t.Fatalf("Flush() = %v, want nil", err)
	}
	if got := len(ft.writes); got != 0 {
		t.Fatalf("Write called %d times on steady-state Flush, want 0", got)
	}
}

// TestFlush_PromotesBackToFront verifies that after Flush, FrontBuf
// reflects the latest BackBuf content.
func TestFlush_PromotesBackToFront(t *testing.T) {
	t.Parallel()

	ft := newFakeTerm()
	l := newLoopWithIO(ft, 10, 10)

	l.BackBuf.SetCell(3, 4, cellFromRune('Q'))
	if err := l.Flush(); err != nil {
		t.Fatalf("Flush() = %v", err)
	}

	want := cellFromRune('Q')
	if got := l.FrontBuf.GetCell(3, 4); got != want {
		t.Fatalf("FrontBuf[3,4] = %+v, want %+v", got, want)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

// BenchmarkFlush measures the steady-state cost of Flush() with a 100x50
// buffer and no mutations. The acceptance criterion is 0 allocs/op —
// every component on the hot path (Diff, Render, copy, Write) is
// pre-allocated.
//
// Run with: go test -bench=BenchmarkFlush -benchmem ./core/loop/
func BenchmarkFlush(b *testing.B) {
	ft := newFakeTerm()
	l := newLoopWithIO(ft, 100, 50)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = l.Flush()
	}
}

// BenchmarkFlush_10PctChanged measures Flush when 10% of cells differ —
// a representative scene with sparse updates.
func BenchmarkFlush_10PctChanged(b *testing.B) {
	ft := newFakeTerm()
	l := newLoopWithIO(ft, 100, 50)

	// Mutate 500 cells out of 5000 once.
	for i := 0; i < 500; i++ {
		x := i % 100
		y := i / 100
		l.BackBuf.SetCell(x, y, cellFromRune(rune('A'+(i%26))))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// After the odd iteration, Front == Back; reset Front to the
		// pristine state so the differ emits changes again.
		if i%2 == 1 {
			clear(l.FrontBuf.Cells)
		}
		_ = l.Flush()
	}
}
