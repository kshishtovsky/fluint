// Package widgets provides UI primitives (Text, Button, etc.) that render
// directly into a buffer.Buffer. All widgets follow the Functional Options
// pattern per ADR-0004.
package widgets

import (
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/style"
)

// Node is the interface implemented by all widgets. It covers rendering,
// geometry tracking, and event handling.
type Node interface {
	Render(buf *buffer.Buffer, rect layout.Rect)

	// Geometry returns the widget's current position and size.
	Geometry() layout.Rect
	// SetGeometry updates the widget's position and size.
	SetGeometry(rect layout.Rect)

	// OnKey handles a keyboard event. Returns true if the event was consumed.
	OnKey(key KeyEvent) bool
	// OnMouse handles a mouse event. Returns true if the event was consumed.
	OnMouse(mouse MouseEvent) bool
	// Focusable reports whether this widget can receive keyboard focus.
	Focusable() bool
}

// Config holds optional configuration shared across all widget types.
type Config struct {
	Width   int
	Height  int
	style   style.Style
	onPress func()
}

// Option is a functional option that modifies Config.
type Option func(*Config)

// WithWidth sets the preferred width.
func WithWidth(w int) Option {
	return func(c *Config) { c.Width = w }
}

// WithHeight sets the preferred height.
func WithHeight(h int) Option {
	return func(c *Config) { c.Height = h }
}

// WithStyle sets the style (colors + attributes) from a style.Style value.
func WithStyle(s style.Style) Option {
	return func(c *Config) { c.style = s }
}

// WithForeground sets the foreground color (packed 0x00RRGGBB).
func WithForeground(fg uint32) Option {
	return func(c *Config) { c.style = c.style.Foreground(style.Color(fg)) }
}

// WithBackground sets the background color (packed 0x00RRGGBB).
func WithBackground(bg uint32) Option {
	return func(c *Config) { c.style = c.style.Background(style.Color(bg)) }
}

// WithBold enables bold text.
func WithBold() Option {
	return func(c *Config) { c.style = c.style.Bold() }
}

// WithItalic enables italic text.
func WithItalic() Option {
	return func(c *Config) { c.style = c.style.Italic() }
}

// WithUnderline enables underline text.
func WithUnderline() Option {
	return func(c *Config) { c.style = c.style.Underline() }
}

// WithDim enables dim text.
func WithDim() Option {
	return func(c *Config) { c.style = c.style.Dim() }
}

// WithStrikethrough enables strikethrough text.
func WithStrikethrough() Option {
	return func(c *Config) { c.style = c.style.Strikethrough() }
}

// WithReverse enables reverse video.
func WithReverse() Option {
	return func(c *Config) { c.style = c.style.Reverse() }
}

// WithOnPress sets the callback invoked when a Button is pressed.
func WithOnPress(fn func()) Option {
	return func(c *Config) { c.onPress = fn }
}

// newConfig applies options to a zero Config and returns it.
func newConfig(opts []Option) Config {
	var cfg Config
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// HitTest reports whether point (x, y) falls inside rect.
func HitTest(rect layout.Rect, x, y int) bool {
	return x >= rect.X && x < rect.X+rect.Width &&
		y >= rect.Y && y < rect.Y+rect.Height
}
