package widgets

import (
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/layout"
)

// Button is a widget that renders a clickable label centred inside a
// filled rectangle.
type Button struct {
	text   string
	config Config
}

// NewButton creates a Button widget. The text argument is the mandatory
// label; optional configuration is supplied via WithXxx options.
func NewButton(text string, opts ...Option) *Button {
	return &Button{
		text:   text,
		config: newConfig(opts),
	}
}

// Press invokes the onPress callback, if one was set.
func (b *Button) Press() {
	if b.config.onPress != nil {
		b.config.onPress()
	}
}

// Render draws the button into buf within rect. The entire rect is
// filled with config.Bg, then the label is centred horizontally and
// vertically.
func (b *Button) Render(buf *buffer.Buffer, rect layout.Rect) {
	bgCell := buffer.Cell{
		Rune: ' ',
		Bg:   b.config.Bg,
	}
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			buf.SetCell(x, y, bgCell)
		}
	}

	runeCount := 0
	for range b.text {
		runeCount++
	}

	startX := rect.X + (rect.Width-runeCount)/2
	if startX < rect.X {
		startX = rect.X
	}

	startY := rect.Y + rect.Height/2
	if startY >= rect.Y+rect.Height {
		startY = rect.Y
	}

	labelCell := buffer.Cell{
		Fg:    b.config.Fg,
		Bg:    b.config.Bg,
		Attrs: b.config.Attrs,
	}
	x := startX
	for _, r := range b.text {
		if x >= rect.X+rect.Width {
			break
		}
		labelCell.Rune = r
		buf.SetCell(x, startY, labelCell)
		x++
	}
}
