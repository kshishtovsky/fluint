//go:build linux || darwin

package platform

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// MakeRaw puts the terminal into raw mode, disabling canonical processing,
// echo, and signal generation. It returns a restore function that reverts
// the terminal to its original state.
func MakeRaw(fd uintptr) (restore func() error, err error) {
	termios, err := unix.IoctlGetTermios(int(fd), ioctlGetTermios)
	if err != nil {
		return nil, err
	}

	saved := *termios

	// Input flags: disable break, CR-to-NL, parity, strip, flow control.
	termios.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	// Output flags: disable post-processing.
	termios.Oflag &^= unix.OPOST
	// Control flags: set 8-bit chars.
	termios.Cflag |= unix.CS8
	// Local flags: disable echo, canonical mode, signals, extended input.
	termios.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG | unix.IEXTEN

	// Minimum bytes for read to return.
	termios.Cc[unix.VMIN] = 1
	// No timeout.
	termios.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(int(fd), ioctlSetTermios, termios); err != nil {
		return nil, err
	}

	return func() error {
		return unix.IoctlSetTermios(int(fd), ioctlSetTermios, &saved)
	}, nil
}

// TerminalSize returns the current terminal dimensions in rows and columns.
func TerminalSize(fd uintptr) (rows, cols int, err error) {
	ws, err := unix.IoctlGetWinsize(int(fd), unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}
	return int(ws.Row), int(ws.Col), nil
}

// WatchResize returns a channel that receives a signal each time the
// terminal is resized (SIGWINCH). The channel is closed when ctx is
// cancelled. The returned channel is buffered (cap 1) to avoid blocking
// the signal handler.
func WatchResize(ctx context.Context) <-chan struct{} {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	out := make(chan struct{}, 1)

	go func() {
		defer signal.Stop(sig)
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case <-sig:
				// Non-blocking send to avoid goroutine leak if consumer is slow.
				select {
				case out <- struct{}{}:
				default:
				}
			}
		}
	}()

	return out
}

// PollReadable waits for data to be available on fd for up to timeout.
// Returns true if data is available for reading, false on timeout.
// A negative timeout blocks indefinitely.
func PollReadable(fd uintptr, timeout time.Duration) (bool, error) {
	fds := []unix.PollFd{
		{Fd: int32(fd), Events: unix.POLLIN},
	}

	timeoutMs := int(timeout.Milliseconds())
	if timeout < 0 {
		timeoutMs = -1
	}

	n, err := unix.Poll(fds, timeoutMs)
	if err != nil {
		if err == syscall.EINTR {
			return false, nil
		}
		return false, err
	}
	return n > 0, nil
}
