//go:build windows

package term

import "errors"

// Flush is not supported on Windows.
func (w *Writer) Flush(_ uintptr) error {
	return errors.ErrUnsupported
}
