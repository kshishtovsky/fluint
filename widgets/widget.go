// Package widgets provides UI primitives (Text, Button, etc.) that render
// directly into a buffer.Buffer. All widgets follow the Functional Options
// pattern per ADR-0004.
package widgets

import (
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/layout"
)

// Node is anything that can render itself into a Buffer within a given Rect.
type Node interface {
	Render(buf *buffer.Buffer, rect layout.Rect)
}

// Config holds optional configuration shared across all widget types.
type Config struct {
	Width   int
	Height  int
	Fg      uint32
	Bg      uint32
	Attrs   buffer.Attrs
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

// WithForeground sets the foreground color (packed 0x00RRGGBB).
func WithForeground(fg uint32) Option {
	return func(c *Config) { c.Fg = fg }
}

// WithBackground sets the background color (packed 0x00RRGGBB).
func WithBackground(bg uint32) Option {
	return func(c *Config) { c.Bg = bg }
}

// WithBold enables bold text.
func WithBold() Option {
	return func(c *Config) { c.Attrs |= buffer.Bold }
}

// WithItalic enables italic text.
func WithItalic() Option {
	return func(c *Config) { c.Attrs |= buffer.Italic }
}

// WithUnderline enables underline text.
func WithUnderline() Option {
	return func(c *Config) { c.Attrs |= buffer.Underline }
}

// WithDim enables dim text.
func WithDim() Option {
	return func(c *Config) { c.Attrs |= buffer.Dim }
}

// WithStrikethrough enables strikethrough text.
func WithStrikethrough() Option {
	return func(c *Config) { c.Attrs |= buffer.Strikethrough }
}

// WithReverse enables reverse video.
func WithReverse() Option {
	return func(c *Config) { c.Attrs |= buffer.Reverse }
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
