//go:build unix

package platform

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"golang.org/x/sys/unix"
)

// Terminal represents a Unix terminal interface for raw mode I/O,
// size querying, and SIGWINCH handling.
type Terminal struct {
	tty          *os.File
	fd           uintptr
	mu           sync.Mutex
	savedTermios *unix.Termios
	rawActive    bool

	width  atomic.Int32
	height atomic.Int32

	sigCh  chan os.Signal
	doneCh chan struct{}
}

// New opens the controlling terminal (/dev/tty) and initializes a Terminal instance.
func New() (*Terminal, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to open /dev/tty: %v", ErrTTYLost, err)
	}
	return NewFromFile(tty)
}

// NewFromFile initializes a Terminal instance using the provided TTY file.
func NewFromFile(tty *os.File) (*Terminal, error) {
	if tty == nil {
		return nil, ErrTTYLost
	}

	fd := tty.Fd()

	t := &Terminal{
		tty:    tty,
		fd:     fd,
		sigCh:  make(chan os.Signal, 1),
		doneCh: make(chan struct{}),
	}

	t.updateSize()

	signal.Notify(t.sigCh, syscall.SIGWINCH)
	go t.watchResize()

	return t, nil
}

// EnterRawMode switches the terminal to raw mode.
func (t *Terminal) EnterRawMode() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.rawActive {
		return nil
	}

	termios, err := unix.IoctlGetTermios(int(t.fd), ioctlGetTermios)
	if err != nil {
		return fmt.Errorf("%w: failed to get termios: %v", ErrTTYLost, err)
	}

	saved := *termios

	termios.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Cflag |= unix.CS8
	termios.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN

	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(int(t.fd), ioctlSetTermios, termios); err != nil {
		return fmt.Errorf("%w: failed to set raw termios: %v", ErrTTYLost, err)
	}

	t.savedTermios = &saved
	t.rawActive = true
	return nil
}

// ExitRawMode restores the terminal to its state before raw mode was entered.
func (t *Terminal) ExitRawMode() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.exitRawModeLocked()
}

func (t *Terminal) exitRawModeLocked() error {
	if !t.rawActive || t.savedTermios == nil {
		return nil
	}

	if err := unix.IoctlSetTermios(int(t.fd), ioctlSetTermios, t.savedTermios); err != nil {
		return fmt.Errorf("%w: failed to restore termios: %v", ErrTTYLost, err)
	}

	t.rawActive = false
	t.savedTermios = nil
	return nil
}

// GetSize returns the cached terminal dimensions (width, height).
// It performs no ioctl calls on the hot path.
func (t *Terminal) GetSize() (width, height int, err error) {
	w := int(t.width.Load())
	h := int(t.height.Load())
	if w <= 0 || h <= 0 {
		return 0, 0, ErrTTYLost
	}
	return w, h, nil
}

// Read reads input bytes directly into p without allocation or buffering.
func (t *Terminal) Read(p []byte) (int, error) {
	n, err := t.tty.Read(p)
	if err != nil {
		if errors.Is(err, os.ErrClosed) || errors.Is(err, syscall.EIO) || errors.Is(err, syscall.EBADF) {
			return n, fmt.Errorf("%w: %v", ErrTTYLost, err)
		}
		return n, err
	}
	return n, nil
}

// Write writes bytes directly from p to the terminal output.
func (t *Terminal) Write(p []byte) (int, error) {
	n, err := t.tty.Write(p)
	if err != nil {
		return n, fmt.Errorf("%w: %v", ErrWriteFailed, err)
	}
	return n, nil
}

// Close closes the terminal resources and signal handling.
func (t *Terminal) Close() error {
	t.mu.Lock()
	if t.rawActive {
		_ = t.exitRawModeLocked()
	}
	t.mu.Unlock()

	signal.Stop(t.sigCh)

	select {
	case <-t.doneCh:
	default:
		close(t.doneCh)
	}

	if t.tty != nil {
		return t.tty.Close()
	}
	return nil
}

func (t *Terminal) watchResize() {
	for {
		select {
		case <-t.doneCh:
			return
		case sig, ok := <-t.sigCh:
			if !ok || sig == nil {
				return
			}
			t.updateSize()
		}
	}
}

func (t *Terminal) updateSize() {
	ws, err := unix.IoctlGetWinsize(int(t.fd), unix.TIOCGWINSZ)
	if err == nil && ws.Col > 0 && ws.Row > 0 {
		t.width.Store(int32(ws.Col))
		t.height.Store(int32(ws.Row))
	}
}
