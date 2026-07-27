//go:build windows

package term

import (
	"context"
	"errors"
	"time"
)

// DetectActive is not supported on Windows.
func DetectActive(_ context.Context, _ uintptr, _ time.Duration) (Capabilities, error) {
	return Detect(), errors.ErrUnsupported
}
