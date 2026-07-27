package widgets

import (
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/layout"
)

// Text is a widget that renders a single line of text.
type Text struct {
	text   string
	config Config
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
// with the configured style.
func (t *Text) Render(buf *buffer.Buffer, rect layout.Rect) {
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
