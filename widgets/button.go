package widgets

import (
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/style"
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

// Style returns the button's current style, enabling external mutation
// (e.g. animation of background color via anim.Tween).
func (b *Button) Style() style.Style {
	return b.config.style
}

// SetStyle replaces the button's style. Intended for animation use cases
// where a Tween callback updates the style each frame.
func (b *Button) SetStyle(s style.Style) {
	b.config.style = s
}

// Render draws the button into buf within rect. The entire rect is
// filled with the style background, then the label is centred
// horizontally and vertically.
func (b *Button) Render(buf *buffer.Buffer, rect layout.Rect) {
	// Fill the entire rect with background.
	bgCell := b.config.style.Apply(buffer.Cell{Rune: ' '})
	for y := rect.Y; y < rect.Y+rect.Height; y++ {
		for x := rect.X; x < rect.X+rect.Width; x++ {
			buf.SetCell(x, y, bgCell)
		}
	}

	// Count runes for correct centering.
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

	// Write the label.
	labelCell := b.config.style.Apply(buffer.Cell{})
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
