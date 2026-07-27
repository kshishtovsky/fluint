package platform

import "errors"

var (
	// ErrUnsupportedPlatform is returned when an operation is executed on an unsupported operating system.
	ErrUnsupportedPlatform = errors.New("platform not supported")

	// ErrTTYLost is returned when the terminal device is lost (e.g. SSH disconnect or EOF).
	ErrTTYLost = errors.New("tty lost")

	// ErrWriteFailed is returned when writing to the terminal fails.
	ErrWriteFailed = errors.New("write failed")
)
