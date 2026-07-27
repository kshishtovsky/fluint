//go:build windows

package platform

import (
	"context"
	"errors"
	"time"
)

// MakeRaw is not supported on Windows.
func MakeRaw(fd uintptr) (restore func() error, err error) {
	return nil, errors.ErrUnsupported
}

// TerminalSize is not supported on Windows.
func TerminalSize(fd uintptr) (rows, cols int, err error) {
	return 0, 0, errors.ErrUnsupported
}

// WatchResize is not supported on Windows.
func WatchResize(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}

// PollReadable is not supported on Windows.
func PollReadable(fd uintptr, timeout time.Duration) (bool, error) {
	return false, errors.ErrUnsupported
}
