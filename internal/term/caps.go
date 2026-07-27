package term

import (
	"bytes"
	"os"
	"strings"
	"time"
)

// ColorDepth constants representing terminal color capabilities.
const (
	ColorMono = 0        // Monochrome — no color support.
	Color8    = 8        // 8 standard ANSI colors.
	Color16   = 16       // 16 standard/bright ANSI colors.
	Color256  = 256      // 256-color palette (xterm).
	ColorTrue = 16777216 // 24-bit true color (16M colors).
)

// Escape sequence constants for terminal capability detection.
const (
	DECRQMSyncOutput     = "\x1b[?2026$p"
	DECRQMSyncOutputResp = "\x1b[?2026;"
)

// TerminalReaderWriter defines the minimal interface for terminal I/O
// required by active capability detection.
type TerminalReaderWriter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

// Capabilities describes the features supported by the terminal.
type Capabilities struct {
	ColorDepth    int  // 0, 8, 16, 256, 16777216 (TrueColor)
	SyncOutput    bool // Mode 2026 (synchronized output)
	KittyKeyboard bool // Kitty keyboard protocol support
}

// Detect determines terminal capabilities from environment variables
// only ($COLORTERM, $TERM, $TERM_PROGRAM). It performs zero I/O operations
// and completes in < 1µs.
func Detect() Capabilities {
	return DetectEnv(os.Getenv("COLORTERM"), os.Getenv("TERM"), os.Getenv("TERM_PROGRAM"))
}

// DetectEnv determines terminal capabilities from pre-fetched environment variable values.
// This pure function runs with zero allocations in < 1µs.
func DetectEnv(colorTerm, term, termProgram string) Capabilities {
	caps := Capabilities{
		ColorDepth: Color16, // Conservative default baseline.
	}

	// Color depth from explicit environment variables.
	switch {
	case colorTerm == "truecolor" || colorTerm == "24bit":
		caps.ColorDepth = ColorTrue
	case strings.Contains(term, "256color"):
		caps.ColorDepth = Color256
	case term == "dumb" || term == "":
		caps.ColorDepth = ColorMono
	}

	// Augment / override based on known terminal programs.
	switch {
	case strings.EqualFold(termProgram, "iterm.app"):
		caps.ColorDepth = maxColorDepth(caps.ColorDepth, ColorTrue)
	case strings.EqualFold(termProgram, "wezterm"):
		caps.ColorDepth = maxColorDepth(caps.ColorDepth, ColorTrue)
		caps.SyncOutput = true
		caps.KittyKeyboard = true
	case strings.EqualFold(termProgram, "kitty"):
		caps.ColorDepth = maxColorDepth(caps.ColorDepth, ColorTrue)
		caps.SyncOutput = true
		caps.KittyKeyboard = true
	case strings.EqualFold(termProgram, "ghostty"):
		caps.ColorDepth = maxColorDepth(caps.ColorDepth, ColorTrue)
		caps.SyncOutput = true
	}

	return caps
}

// DetectActive queries the terminal for advanced capabilities using ANSI escape
// sequences, starting from environment baseline Detect() and waiting up to 100ms
// for responses before falling back to baseline.
func DetectActive(t TerminalReaderWriter) Capabilities {
	return DetectActiveTimeout(t, 100*time.Millisecond)
}

// DetectActiveTimeout performs active capability detection with a custom timeout.
func DetectActiveTimeout(t TerminalReaderWriter, timeout time.Duration) Capabilities {
	caps := Detect()
	if t == nil {
		return caps
	}

	// Send DECRQM for mode 2026.
	if _, err := t.Write([]byte(DECRQMSyncOutput)); err != nil {
		return caps
	}

	type readResult struct {
		n   int
		err error
		buf [128]byte
	}

	resCh := make(chan readResult, 1)

	go func() {
		var res readResult
		res.n, res.err = t.Read(res.buf[:])
		resCh <- res
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case res := <-resCh:
		if res.err == nil && res.n > 0 {
			if parseDECRQMSyncOutput(res.buf[:res.n]) {
				caps.SyncOutput = true
			}
		}
	case <-timer.C:
		// Timeout expired — return baseline capabilities.
	}

	return caps
}

// parseDECRQMSyncOutput parses terminal responses for DECRQM mode 2026
// using direct byte-by-byte inspection without regex or allocations.
func parseDECRQMSyncOutput(resp []byte) bool {
	prefix := []byte(DECRQMSyncOutputResp)
	idx := bytes.Index(resp, prefix)
	if idx >= 0 && idx+len(prefix)+1 <= len(resp) {
		ps := resp[idx+len(prefix)]
		// '1' = set, '2' = reset -> recognized by terminal
		if ps == '1' || ps == '2' {
			return true
		}
	}
	return false
}

func maxColorDepth(a, b int) int {
	if a > b {
		return a
	}
	return b
}
