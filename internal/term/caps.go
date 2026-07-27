package term

import (
	"os"
	"strings"
)

// ColorDepth represents the color capability level of the terminal.
type ColorDepth uint8

const (
	ColorMono ColorDepth = iota // Monochrome — no color support.
	Color16                     // 16 standard ANSI colors.
	Color256                    // 256-color palette (xterm).
	ColorTrue                   // 24-bit true color (16M colors).
)

// Capabilities describes the features supported by the terminal.
// Fields are populated by Detect (environment-only) or DetectActive
// (environment + ANSI queries).
type Capabilities struct {
	ColorDepth    ColorDepth // Color capability level.
	HasSyncOutput bool       // DEC private mode 2026 (synchronized output).
	HasKittyKbd   bool       // Kitty keyboard protocol support.
	TermProgram   string     // Terminal program name from $TERM_PROGRAM.
}

// Detect determines terminal capabilities from environment variables
// only. It performs no I/O and returns immediately.
//
// This is the fast path used at startup. For more accurate detection
// of advanced features, follow up with DetectActive.
func Detect() Capabilities {
	caps := Capabilities{
		ColorDepth: Color16, // Conservative default.
	}

	caps.TermProgram = os.Getenv("TERM_PROGRAM")
	colorTerm := os.Getenv("COLORTERM")
	term := os.Getenv("TERM")

	// Color depth from explicit env.
	switch {
	case colorTerm == "truecolor" || colorTerm == "24bit":
		caps.ColorDepth = ColorTrue
	case strings.Contains(term, "256color"):
		caps.ColorDepth = Color256
	case term == "dumb" || term == "":
		caps.ColorDepth = ColorMono
	}

	// Override / augment from known terminal programs.
	switch strings.ToLower(caps.TermProgram) {
	case "iterm.app":
		caps.ColorDepth = maxColorDepth(caps.ColorDepth, ColorTrue)
	case "wezterm":
		caps.ColorDepth = maxColorDepth(caps.ColorDepth, ColorTrue)
		caps.HasSyncOutput = true
	case "kitty":
		caps.ColorDepth = maxColorDepth(caps.ColorDepth, ColorTrue)
		caps.HasKittyKbd = true
		caps.HasSyncOutput = true
	case "ghostty":
		caps.ColorDepth = maxColorDepth(caps.ColorDepth, ColorTrue)
		caps.HasSyncOutput = true
	}

	// WezTerm, foot, Contour, etc. often set COLORTERM=truecolor,
	// already handled above.

	return caps
}

func maxColorDepth(a, b ColorDepth) ColorDepth {
	if a > b {
		return a
	}
	return b
}
