package widgets

import (
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
)

// TextInput is a single-line text editing widget. It maintains a rune
// slice for O(1) cursor movement and supports horizontal scrolling
// when the text exceeds the visible width.
type TextInput struct {
	text   []rune
	pos    int // cursor position (index into text)
	scroll int // horizontal scroll offset
	config Config
	rect   layout.Rect
}

// NewTextInput creates a TextInput with optional initial text.
func NewTextInput(initial string, opts ...Option) *TextInput {
	return &TextInput{
		text:   []rune(initial),
		pos:    len([]rune(initial)),
		config: newConfig(opts),
	}
}

// Text returns the current text content as a string.
func (ti *TextInput) Text() string { return string(ti.text) }

// SetText replaces the entire text and moves the cursor to the end.
func (ti *TextInput) SetText(s string) {
	ti.text = []rune(s)
	ti.pos = len(ti.text)
	ti.scroll = 0
	ti.adjustScroll()
}

// adjustScroll ensures the cursor is visible within the widget width.
func (ti *TextInput) adjustScroll() {
	w := ti.rect.Width
	if w <= 0 {
		w = ti.config.Width
	}
	if w <= 0 {
		w = 1
	}
	if ti.pos < ti.scroll {
		ti.scroll = ti.pos
	}
	if ti.pos >= ti.scroll+w {
		ti.scroll = ti.pos - w + 1
	}
	if ti.scroll < 0 {
		ti.scroll = 0
	}
}

// Render draws the text and cursor. If the text is longer than the
// rect width, only a window around the cursor is shown.
func (ti *TextInput) Render(ctx viewport.RenderCtx, rect layout.Rect) {
	ti.rect = rect

	if !Visible(ctx.View, rect.X, rect.Y, rect.Width, rect.Height) {
		return
	}

	sx, sy := Screen(ctx.View, rect.X, rect.Y)
	ti.adjustScroll()

	w := rect.Width
	normalCell := ti.config.style.Apply(buffer.Cell{})
	cursorCell := ti.config.style.Apply(buffer.Cell{})
	cursorCell.Fg, cursorCell.Bg = cursorCell.Bg, cursorCell.Fg

	x := sx
	for i := 0; i < w; i++ {
		idx := ti.scroll + i
		cell := normalCell
		if idx == ti.pos {
			cell = cursorCell
		}
		if idx < len(ti.text) {
			cell.Rune = ti.text[idx]
		} else if idx == ti.pos {
			cell.Rune = '_'
		} else {
			cell.Rune = ' '
		}
		ctx.Buf.SetCell(x, sy, cell)
		x++
	}
}

// Geometry returns the widget's current position and size.
func (ti *TextInput) Geometry() layout.Rect { return ti.rect }

// SetGeometry updates the widget's position and size.
func (ti *TextInput) SetGeometry(rect layout.Rect) { ti.rect = rect }

// OnKey handles keyboard events for text editing.
func (ti *TextInput) OnKey(key KeyEvent) bool {
	switch {
	case key.Code == KeyLeft:
		if ti.pos > 0 {
			ti.pos--
		}
		ti.adjustScroll()
		return true
	case key.Code == KeyRight:
		if ti.pos < len(ti.text) {
			ti.pos++
		}
		ti.adjustScroll()
		return true
	case key.Code == KeyHome:
		ti.pos = 0
		ti.adjustScroll()
		return true
	case key.Code == KeyEnd:
		ti.pos = len(ti.text)
		ti.adjustScroll()
		return true
	case key.Code == KeyBackspace:
		if ti.pos > 0 {
			ti.text = append(ti.text[:ti.pos-1], ti.text[ti.pos:]...)
			ti.pos--
		}
		ti.adjustScroll()
		ti.fireChange()
		return true
	case key.Code == KeyDelete:
		if ti.pos < len(ti.text) {
			ti.text = append(ti.text[:ti.pos], ti.text[ti.pos+1:]...)
		}
		ti.fireChange()
		return true
	case key.Rune != 0 && key.Code == KeyNone:
		ti.text = append(ti.text[:ti.pos], append([]rune{key.Rune}, ti.text[ti.pos:]...)...)
		ti.pos++
		ti.adjustScroll()
		ti.fireChange()
		return true
	}
	return false
}

// fireChange invokes the onChange callback if set.
func (ti *TextInput) fireChange() {
	if ti.config.onChange != nil {
		ti.config.onChange(string(ti.text))
	}
}

// OnMouse handles mouse events. A left-click moves the cursor.
func (ti *TextInput) OnMouse(mouse MouseEvent) bool {
	if mouse.Button == MouseLeft && mouse.Action == MousePress {
		if HitTest(ti.rect, mouse.X, mouse.Y) {
			clickX := mouse.X - ti.rect.X + ti.scroll
			if clickX > len(ti.text) {
				clickX = len(ti.text)
			}
			if clickX < 0 {
				clickX = 0
			}
			ti.pos = clickX
			ti.adjustScroll()
			return true
		}
	}
	return false
}

// Focusable returns true — TextInput can receive keyboard focus.
func (ti *TextInput) Focusable() bool { return true }
