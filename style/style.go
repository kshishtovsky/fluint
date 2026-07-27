// Package style provides a value-semantic styling API for terminal widgets.
//
// Style is a struct that carries foreground/background colors, text
// attributes, border configuration, padding, and shadow. All modification
// methods return a new Style by value, enabling chainable construction:
//
//	s := style.New().Foreground(style.Red).Background(style.Black).Bold()
//	s := style.New().RoundedBorder(style.White).Padding(1, 1)
package style

import "github.com/kshishtovsky/fluint/core/buffer"

// Color is a packed 24-bit RGB color (0x00RRGGBB).
type Color uint32

// Predefined colors — the default ANSI-inspired palette.
const (
	Default     Color = 0x00000000
	Black       Color = 0x00000000
	Red         Color = 0x00FF0000
	Green       Color = 0x0000FF00
	Yellow      Color = 0x00FFFF00
	Blue        Color = 0x000000FF
	Magenta     Color = 0x00FF00FF
	Cyan        Color = 0x0000FFFF
	White       Color = 0x00FFFFFF
	DarkGray    Color = 0x00555555
	LightGray   Color = 0x00AAAAAA
	Orange      Color = 0x00FF8800
	Pink        Color = 0x00FF69B4
	Teal        Color = 0x00008080
	Purple      Color = 0x00800080
	Brown       Color = 0x008B4513
	ShadowColor Color = 0x0A0A0A // default shadow — near-black, barely visible

	// Shadow runes — half-blocks for sub-cell shadows.
	ShadowBottom rune = '▄' // U+2584 lower half block — bottom shadow
	ShadowRight  rune = '▐' // U+2590 right half block — right shadow
	ShadowCorner rune = '▒' // U+2592 medium shade — bottom-right join (softer)
)

// BorderStyle selects the border drawing mode.
type BorderStyle uint8

const (
	BorderNone    BorderStyle = iota // no border
	BorderSolid                      // │ ─ ┌ ┐ └ ┘
	BorderRounded                    // ╭ ╮ ╰ ╯
)

// Border rune constants — declared as package-level constants for zero
// allocation in the render hot path.
const (
	BorderSolidH  rune = '─'
	BorderSolidV  rune = '│'
	BorderSolidTL rune = '┌'
	BorderSolidTR rune = '┐'
	BorderSolidBL rune = '└'
	BorderSolidBR rune = '┘'

	BorderRoundedH  rune = '─'
	BorderRoundedV  rune = '│'
	BorderRoundedTL rune = '╭'
	BorderRoundedTR rune = '╮'
	BorderRoundedBL rune = '╰'
	BorderRoundedBR rune = '╯'
)

// ShadowMode selects the shadow rendering algorithm.
type ShadowMode uint8

const (
	// ShadowModeSubCell uses half-block runes (▐, ▄, ▒) for a thin
	// 1-cell shadow. This is the default when Shadow() is used.
	ShadowModeSubCell ShadowMode = iota
	// ShadowModeDither uses ASCII density gradient for a soft,
	// multi-cell blur. Dense characters near the card, sparse at the edge.
	ShadowModeDither
)

// ShadowDensityRamp maps a normalised distance [0..1] to a rune.
// Index 0 = empty (far from card), index 9 = dense (close to card).
// Declared as a package-level var (not const) because rune arrays
// cannot be const in Go. Still zero-alloc at runtime.
var ShadowDensityRamp = [10]rune{
	' ', // 0 — invisible (beyond shadow)
	'.', // 1 — barely visible
	':', // 2
	'-', // 3
	'=', // 4
	'+', // 5
	'*', // 6
	'%', // 7
	'@', // 8
	'#', // 9 — densest (closest to card)
}

// ShadowStyle configures a drop shadow behind a widget.
type ShadowStyle struct {
	Enabled bool
	Mode    ShadowMode
	OffsetX int   // for SubCell: 1 = enable. For Dither: blur radius in cells.
	OffsetY int   // for SubCell: 1 = enable. For Dither: blur radius in cells.
	Color   Color // shadow foreground color
}

// Style carries foreground, background, text attributes, border, padding,
// and shadow. The zero value is usable and represents "no style override".
type Style struct {
	fg    Color
	bg    Color
	attrs buffer.Attrs

	border      BorderStyle
	borderColor Color

	paddingX int // horizontal padding (cells)
	paddingY int // vertical padding (rows)

	shadow ShadowStyle
}

// New returns an empty Style (zero values — no overrides).
func New() Style {
	return Style{}
}

// Foreground returns a copy with the foreground color set.
func (s Style) Foreground(c Color) Style {
	s.fg = c
	return s
}

// Background returns a copy with the background color set.
func (s Style) Background(c Color) Style {
	s.bg = c
	return s
}

// Bold returns a copy with the Bold attribute enabled.
func (s Style) Bold() Style {
	s.attrs |= buffer.Bold
	return s
}

// Italic returns a copy with the Italic attribute enabled.
func (s Style) Italic() Style {
	s.attrs |= buffer.Italic
	return s
}

// Underline returns a copy with the Underline attribute enabled.
func (s Style) Underline() Style {
	s.attrs |= buffer.Underline
	return s
}

// Dim returns a copy with the Dim attribute enabled.
func (s Style) Dim() Style {
	s.attrs |= buffer.Dim
	return s
}

// Strikethrough returns a copy with the Strikethrough attribute enabled.
func (s Style) Strikethrough() Style {
	s.attrs |= buffer.Strikethrough
	return s
}

// Reverse returns a copy with the Reverse attribute enabled.
func (s Style) Reverse() Style {
	s.attrs |= buffer.Reverse
	return s
}

// ── Border ──────────────────────────────────────────────────────────

// SolidBorder returns a copy with a solid border in the given color.
func (s Style) SolidBorder(c Color) Style {
	s.border = BorderSolid
	s.borderColor = c
	return s
}

// RoundedBorder returns a copy with a rounded border in the given color.
func (s Style) RoundedBorder(c Color) Style {
	s.border = BorderRounded
	s.borderColor = c
	return s
}

// NoBorder returns a copy with the border removed.
func (s Style) NoBorder() Style {
	s.border = BorderNone
	s.borderColor = Default
	return s
}

// ── Padding ─────────────────────────────────────────────────────────

// Padding returns a copy with the given horizontal and vertical padding.
func (s Style) Padding(x, y int) Style {
	s.paddingX = x
	s.paddingY = y
	return s
}

// ── Shadow ──────────────────────────────────────────────────────────

// Shadow returns a copy with a sub-cell drop shadow (half-block runes).
func (s Style) Shadow(offsetX, offsetY int, color Color) Style {
	s.shadow = ShadowStyle{
		Enabled: true,
		Mode:    ShadowModeSubCell,
		OffsetX: offsetX,
		OffsetY: offsetY,
		Color:   color,
	}
	return s
}

// DitherShadow returns a copy with an ASCII density gradient shadow.
// blurX/blurY set the blur radius in cells. Characters fade from dense
// (#, @, %) near the card to sparse (., :) at the edge.
func (s Style) DitherShadow(blurX, blurY int, color Color) Style {
	s.shadow = ShadowStyle{
		Enabled: true,
		Mode:    ShadowModeDither,
		OffsetX: blurX,
		OffsetY: blurY,
		Color:   color,
	}
	return s
}

// NoShadow returns a copy with the shadow disabled.
func (s Style) NoShadow() Style {
	s.shadow = ShadowStyle{}
	return s
}

// ── Accessors ───────────────────────────────────────────────────────

// FG returns the foreground color.
func (s Style) FG() Color { return s.fg }

// BG returns the background color.
func (s Style) BG() Color { return s.bg }

// Attrs returns the text attributes bitmask.
func (s Style) Attrs() buffer.Attrs { return s.attrs }

// Border returns the border style.
func (s Style) Border() BorderStyle { return s.border }

// BorderColor returns the border color.
func (s Style) BorderColor() Color { return s.borderColor }

// PaddingX returns the horizontal padding.
func (s Style) PaddingX() int { return s.paddingX }

// PaddingY returns the vertical padding.
func (s Style) PaddingY() int { return s.paddingY }

// Shadow returns the shadow configuration.
func (s Style) ShadowCfg() ShadowStyle { return s.shadow }

// HasBorder reports whether a border is configured.
func (s Style) HasBorder() bool { return s.border != BorderNone }

// HasShadow reports whether a shadow is configured.
func (s Style) HasShadow() bool { return s.shadow.Enabled }

// Apply returns a copy of cell with this style's colors and attributes
// merged in. Zero-value colors (Default) are not applied, preserving
// the cell's existing color.
func (s Style) Apply(cell buffer.Cell) buffer.Cell {
	if s.fg != Default {
		cell.Fg = uint32(s.fg)
	}
	if s.bg != Default {
		cell.Bg = uint32(s.bg)
	}
	cell.Attrs |= s.attrs
	return cell
}
