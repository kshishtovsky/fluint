package widgets

import (
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
)

// ListItem is a single entry in a List.
type ListItem struct {
	Text string
}

// List is a scrollable, focusable list widget. Only items that fit
// within the rect are rendered (virtualisation). The selected item
// is drawn with inverted colors.
type List struct {
	items    []ListItem
	selected int
	offset   int // first visible row
	config   Config
	rect     layout.Rect
}

// NewList creates a List widget with the given items.
func NewList(items []ListItem, opts ...Option) *List {
	return &List{
		items:  items,
		config: newConfig(opts),
	}
}

// Selected returns the index of the currently selected item.
func (l *List) Selected() int { return l.selected }

// SetSelected sets the selected index and auto-scrolls to keep it visible.
func (l *List) SetSelected(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(l.items) {
		idx = len(l.items) - 1
	}
	l.selected = idx
	l.ensureVisible()
}

// Items returns the list items.
func (l *List) Items() []ListItem { return l.items }

// SetItems replaces the list items and clamps selection.
func (l *List) SetItems(items []ListItem) {
	l.items = items
	if l.selected >= len(l.items) {
		l.selected = len(l.items) - 1
	}
	if l.selected < 0 {
		l.selected = 0
	}
	l.ensureVisible()
}

// ensureVisible scrolls so the selected item is within the visible window.
func (l *List) ensureVisible() {
	h := l.rect.Height
	if h <= 0 {
		h = l.config.Height
	}
	if h <= 0 {
		h = 1
	}
	if l.selected < l.offset {
		l.offset = l.selected
	}
	if l.selected >= l.offset+h {
		l.offset = l.selected - h + 1
	}
	if l.offset < 0 {
		l.offset = 0
	}
}

// Render draws the visible portion of the list. Only items in the
// range [offset, offset+rect.Height) are rendered.
func (l *List) Render(ctx viewport.RenderCtx, rect layout.Rect) {
	l.rect = rect

	if !Visible(ctx.View, rect.X, rect.Y, rect.Width, rect.Height) {
		return
	}

	sx, sy := Screen(ctx.View, rect.X, rect.Y)
	l.ensureVisible()

	normalCell := l.config.style.Apply(buffer.Cell{})
	selCell := l.config.style.Apply(buffer.Cell{})
	selCell.Fg, selCell.Bg = selCell.Bg, selCell.Fg

	h := rect.Height
	for row := 0; row < h; row++ {
		idx := l.offset + row
		y := sy + row

		if idx >= len(l.items) {
			bgCell := l.config.style.Apply(buffer.Cell{Rune: ' '})
			for x := sx; x < sx+rect.Width; x++ {
				ctx.Buf.SetCell(x, y, bgCell)
			}
			continue
		}

		item := l.items[idx]
		isSel := idx == l.selected
		cell := normalCell
		if isSel {
			cell = selCell
		}

		x := sx
		for _, r := range item.Text {
			if x >= sx+rect.Width {
				break
			}
			cell.Rune = r
			ctx.Buf.SetCell(x, y, cell)
			x++
		}
		cell.Rune = ' '
		for x < sx+rect.Width {
			ctx.Buf.SetCell(x, y, cell)
			x++
		}
	}
}

// Geometry returns the widget's current position and size.
func (l *List) Geometry() layout.Rect { return l.rect }

// SetGeometry updates the widget's position and size.
func (l *List) SetGeometry(rect layout.Rect) { l.rect = rect }

// OnKey handles keyboard events. Up/Down move selection, Enter invokes
// the onSelect callback.
func (l *List) OnKey(key KeyEvent) bool {
	switch key.Code {
	case KeyUp:
		if l.selected > 0 {
			l.selected--
			l.ensureVisible()
		}
		return true
	case KeyDown:
		if l.selected < len(l.items)-1 {
			l.selected++
			l.ensureVisible()
		}
		return true
	case KeyEnter:
		if l.config.onSelect != nil && l.selected >= 0 && l.selected < len(l.items) {
			l.config.onSelect(l.selected, l.items[l.selected])
		}
		return true
	}
	return false
}

// OnMouse handles mouse events. A left-click selects the clicked row.
func (l *List) OnMouse(mouse MouseEvent) bool {
	if mouse.Button == MouseLeft && mouse.Action == MousePress {
		if HitTest(l.rect, mouse.X, mouse.Y) {
			row := mouse.Y - l.rect.Y
			idx := l.offset + row
			if idx >= 0 && idx < len(l.items) {
				l.selected = idx
				if l.config.onSelect != nil {
					l.config.onSelect(l.selected, l.items[l.selected])
				}
			}
			return true
		}
	}
	return false
}

// Focusable returns true — List can receive keyboard focus.
func (l *List) Focusable() bool { return true }
