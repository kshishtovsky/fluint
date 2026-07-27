package widgets

import (
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/style"
)

// Button is a widget that renders a clickable label centred inside a
// filled rectangle. It is focusable and responds to Enter key and
// left mouse clicks.
type Button struct {
	text   string
	config Config
	rect   layout.Rect
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

// Style returns the button's current style.
func (b *Button) Style() style.Style {
	return b.config.style
}

// SetStyle replaces the button's style.
func (b *Button) SetStyle(s style.Style) {
	b.config.style = s
}

// Render draws the button into ctx.Buf within the given world-space rect.
// The entire rect is filled with the style background, then the label is
// centred. Widgets fully outside the viewport are culled; partial overlap
// is clipped per-cell.
func (b *Button) Render(ctx viewport.RenderCtx, rect layout.Rect) {
	b.rect = rect

	if !Visible(ctx.View, rect.X, rect.Y, rect.Width, rect.Height) {
		return
	}

	sx, sy := Screen(ctx.View, rect.X, rect.Y)

	bgCell := b.config.style.Apply(buffer.Cell{Rune: ' '})
	for y := sy; y < sy+rect.Height; y++ {
		for x := sx; x < sx+rect.Width; x++ {
			ctx.Buf.SetCell(x, y, bgCell)
		}
	}

	runeCount := 0
	for range b.text {
		runeCount++
	}

	startX := sx + (rect.Width-runeCount)/2
	if startX < sx {
		startX = sx
	}

	startY := sy + rect.Height/2
	if startY >= sy+rect.Height {
		startY = sy
	}

	labelCell := b.config.style.Apply(buffer.Cell{})
	x := startX
	for _, r := range b.text {
		if x >= sx+rect.Width {
			break
		}
		labelCell.Rune = r
		ctx.Buf.SetCell(x, startY, labelCell)
		x++
	}
}

// Geometry returns the widget's current position and size.
func (b *Button) Geometry() layout.Rect { return b.rect }

// SetGeometry updates the widget's position and size.
func (b *Button) SetGeometry(rect layout.Rect) { b.rect = rect }

// OnKey handles keyboard events. Pressing Enter triggers the button.
// Returns true if the event was consumed.
func (b *Button) OnKey(key KeyEvent) bool {
	if key.Code == KeyEnter {
		b.Press()
		return true
	}
	return false
}

// OnMouse handles mouse events. A left-button press inside the button
// triggers it. Returns true if the event was consumed.
func (b *Button) OnMouse(mouse MouseEvent) bool {
	if mouse.Button == MouseLeft && mouse.Action == MousePress {
		if HitTest(b.rect, mouse.X, mouse.Y) {
			b.Press()
			return true
		}
	}
	return false
}

// Focusable returns true — Button can receive keyboard focus.
func (b *Button) Focusable() bool { return true }
