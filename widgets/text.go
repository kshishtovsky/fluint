package widgets

import (
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/layout"
)

// Text is a widget that renders a single line of text.
// It is not focusable and ignores all events.
type Text struct {
	text   string
	config Config
	rect   layout.Rect
}

// NewText creates a Text widget. The text argument is the mandatory content;
// optional configuration is supplied via WithXxx options.
func NewText(text string, opts ...Option) *Text {
	return &Text{
		text:   text,
		config: newConfig(opts),
	}
}

// Render writes the text runes into buf starting at (rect.X, rect.Y).
// If the text exceeds rect.Width it is clipped. Each cell is written
// with the configured style. The widget's geometry is updated to rect.
func (t *Text) Render(buf *buffer.Buffer, rect layout.Rect) {
	t.rect = rect
	var cell buffer.Cell
	x := rect.X
	for _, r := range t.text {
		if x >= rect.X+rect.Width {
			break
		}
		cell.Rune = r
		buf.SetCell(x, rect.Y, t.config.style.Apply(cell))
		x++
	}
}

// Geometry returns the widget's current position and size.
func (t *Text) Geometry() layout.Rect { return t.rect }

// SetGeometry updates the widget's position and size.
func (t *Text) SetGeometry(rect layout.Rect) { t.rect = rect }

// OnKey is a no-op for Text — it never consumes key events.
func (t *Text) OnKey(key KeyEvent) bool { return false }

// OnMouse is a no-op for Text — it never consumes mouse events.
func (t *Text) OnMouse(mouse MouseEvent) bool { return false }

// Focusable returns false — Text cannot receive keyboard focus.
func (t *Text) Focusable() bool { return false }
