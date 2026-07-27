//go:build linux || darwin

package term

import (
	"bytes"
	"context"
	"syscall"
	"time"

	"github.com/kshishtovsky/fluint/internal/platform"
)

// DetectActive queries the terminal for advanced capabilities using
// ANSI escape sequences. It starts from the environment-based Detect
// result and augments it with active queries.
//
// readTimeout controls how long to wait for each query response.
// A typical value is 100ms. Returns the environment-based result
// without error if the terminal doesn't respond within the timeout.
func DetectActive(ctx context.Context, fd uintptr, readTimeout time.Duration) (Capabilities, error) {
	caps := Detect()

	// Query synchronized output support: DECRQM for mode 2026.
	// Request:  CSI ? 2026 $ p
	// Response: CSI ? 2026 ; Ps $ y
	//   Ps=1: mode is set
	//   Ps=2: mode is reset (but recognized)
	//   Ps=0: mode not recognized
	syncQuery := []byte("\x1b[?2026$p")
	if _, err := syscall.Write(int(fd), syncQuery); err != nil {
		return caps, err
	}

	var buf [128]byte
	ready, err := platform.PollReadable(fd, readTimeout)
	if err != nil {
		return caps, err
	}
	if ready {
		n, readErr := syscall.Read(int(fd), buf[:])
		if readErr == nil && n > 0 {
			resp := buf[:n]
			// Look for "2026;" in the DECRQM response.
			if idx := bytes.Index(resp, []byte("2026;")); idx >= 0 && idx+6 <= len(resp) {
				ps := resp[idx+5]
				// '1' = set, '2' = reset — either means the terminal
				// recognizes mode 2026.
				if ps == '1' || ps == '2' {
					caps.HasSyncOutput = true
				}
			}
		}
	}

	_ = ctx // reserved for future multi-query sequences with cancellation
	return caps, nil
}
