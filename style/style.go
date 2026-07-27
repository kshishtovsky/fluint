// Package style provides a value-semantic styling API for terminal widgets.
//
// Style is a small struct (12 bytes) that carries foreground/background
// colors and text attributes. All modification methods return a new
// Style by value, enabling chainable construction:
//
//	s := style.New().Foreground(style.Red).Background(style.Black).Bold()
//
// The resulting Style is applied to a buffer.Cell via the Apply method,
// which returns a copy with the style's colors and attributes merged in.
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
)

// Style carries foreground, background, and text attributes.
// The zero value is usable and represents "no style override".
type Style struct {
	fg    Color
	bg    Color
	attrs buffer.Attrs
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

// FG returns the foreground color.
func (s Style) FG() Color { return s.fg }

// BG returns the background color.
func (s Style) BG() Color { return s.bg }

// Attrs returns the text attributes bitmask.
func (s Style) Attrs() buffer.Attrs { return s.attrs }

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
