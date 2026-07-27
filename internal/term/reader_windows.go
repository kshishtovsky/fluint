//go:build windows

package term

import (
	"context"
	"errors"
)

// ReadInput is not supported on Windows.
func ReadInput(_ context.Context, _ uintptr, _ chan<- InputEvent) error {
	return errors.ErrUnsupported
}
