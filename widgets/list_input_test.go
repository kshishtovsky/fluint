package widgets

import (
	"testing"

	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
)

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestListNavigation(t *testing.T) {
	t.Parallel()

	items := make([]ListItem, 100)
	for i := range items {
		items[i] = ListItem{Text: "item"}
	}
	l := NewList(items)

	l.OnKey(KeyEvent{Code: KeyDown})
	if l.Selected() != 1 {
		t.Errorf("after Down: got %d, want 1", l.Selected())
	}
	l.OnKey(KeyEvent{Code: KeyUp})
	if l.Selected() != 0 {
		t.Errorf("after Up: got %d, want 0", l.Selected())
	}
	// Can't go above 0.
	l.OnKey(KeyEvent{Code: KeyUp})
	if l.Selected() != 0 {
		t.Errorf("Up at 0: got %d, want 0", l.Selected())
	}
}

func TestListAutoScroll(t *testing.T) {
	t.Parallel()

	items := make([]ListItem, 100)
	for i := range items {
		items[i] = ListItem{Text: "item"}
	}
	l := NewList(items)
	l.SetGeometry(layout.Rect{X: 0, Y: 0, Width: 10, Height: 5})

	l.SetSelected(50)
	// Selected=50 with height=5 should auto-scroll: offset should be 50-5+1=46.
	if l.offset != 46 {
		t.Errorf("offset after SetSelected(50): got %d, want 46", l.offset)
	}
}

func TestListRenderVirtualisation(t *testing.T) {
	t.Parallel()

	items := make([]ListItem, 100)
	for i := range items {
		items[i] = ListItem{Text: "X"}
	}
	l := NewList(items,
		WithForeground(0xFFFFFF),
		WithBackground(0x000000),
	)
	l.SetGeometry(layout.Rect{X: 0, Y: 0, Width: 10, Height: 5})
	l.SetSelected(50) // offset=46

	buf := buffer.NewBuffer(10, 5)
	l.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 10, Height: 5})

	// Items 46..50 should be rendered. Item 50 is selected (inverted).
	// Row 0 = item 46 (not selected): normal style.
	c0 := buf.GetCell(0, 0)
	if c0.Fg != 0xFFFFFF {
		t.Errorf("row 0 fg: got 0x%06X, want 0xFFFFFF (normal)", c0.Fg)
	}

	// Row 4 = item 50 (selected): inverted (Fg↔Bg).
	c4 := buf.GetCell(0, 4)
	if c4.Fg != 0x000000 || c4.Bg != 0xFFFFFF {
		t.Errorf("row 4 (selected): got fg=0x%06X bg=0x%06X, want fg=0x000000 bg=0xFFFFFF", c4.Fg, c4.Bg)
	}
}

func TestListOnSelect(t *testing.T) {
	t.Parallel()

	var selIdx int
	var selText string
	items := []ListItem{{Text: "a"}, {Text: "b"}, {Text: "c"}}
	l := NewList(items, WithOnSelect(func(idx int, item ListItem) {
		selIdx = idx
		selText = item.Text
	}))
	l.SetSelected(1)
	l.OnKey(KeyEvent{Code: KeyEnter})
	if selIdx != 1 || selText != "b" {
		t.Errorf("onSelect: got (%d, %q), want (1, \"b\")", selIdx, selText)
	}
}

func TestListFocusable(t *testing.T) {
	t.Parallel()
	l := NewList([]ListItem{{Text: "x"}})
	if !l.Focusable() {
		t.Error("List should be focusable")
	}
}

// ---------------------------------------------------------------------------
// TextInput
// ---------------------------------------------------------------------------

func TestTextInputInsertAndCursor(t *testing.T) {
	t.Parallel()

	ti := NewTextInput("")

	// Type "Hello"
	for _, r := range "Hello" {
		ti.OnKey(KeyEvent{Rune: r})
	}
	if ti.Text() != "Hello" {
		t.Errorf("after typing Hello: got %q, want %q", ti.Text(), "Hello")
	}
	if ti.pos != 5 {
		t.Errorf("cursor: got %d, want 5", ti.pos)
	}
}

func TestTextInputCursorMoveAndInsert(t *testing.T) {
	t.Parallel()

	ti := NewTextInput("Hello")
	ti.pos = 5

	// Move left twice.
	ti.OnKey(KeyEvent{Code: KeyLeft})
	ti.OnKey(KeyEvent{Code: KeyLeft})
	if ti.pos != 3 {
		t.Errorf("after 2xLeft: got %d, want 3", ti.pos)
	}

	// Insert 'X'.
	ti.OnKey(KeyEvent{Rune: 'X'})
	if ti.Text() != "HelXlo" {
		t.Errorf("after insert X: got %q, want %q", ti.Text(), "HelXlo")
	}
}

func TestTextInputBackspace(t *testing.T) {
	t.Parallel()

	ti := NewTextInput("HelXlo")
	ti.pos = 4 // after 'X'

	ti.OnKey(KeyEvent{Code: KeyBackspace})
	if ti.Text() != "Hello" {
		t.Errorf("after backspace: got %q, want %q", ti.Text(), "Hello")
	}
	if ti.pos != 3 {
		t.Errorf("cursor: got %d, want 3", ti.pos)
	}
}

func TestTextInputDelete(t *testing.T) {
	t.Parallel()

	ti := NewTextInput("Hello")
	ti.pos = 2 // on 'l'

	ti.OnKey(KeyEvent{Code: KeyDelete})
	if ti.Text() != "Helo" {
		t.Errorf("after delete: got %q, want %q", ti.Text(), "Helo")
	}
}

func TestTextInputHomeEnd(t *testing.T) {
	t.Parallel()

	ti := NewTextInput("Hello")
	ti.OnKey(KeyEvent{Code: KeyHome})
	if ti.pos != 0 {
		t.Errorf("Home: got %d, want 0", ti.pos)
	}
	ti.OnKey(KeyEvent{Code: KeyEnd})
	if ti.pos != 5 {
		t.Errorf("End: got %d, want 5", ti.pos)
	}
}

func TestTextInputOnChange(t *testing.T) {
	t.Parallel()

	var last string
	ti := NewTextInput("", WithOnChange(func(s string) { last = s }))
	ti.OnKey(KeyEvent{Rune: 'A'})
	if last != "A" {
		t.Errorf("onChange: got %q, want %q", last, "A")
	}
	ti.OnKey(KeyEvent{Code: KeyBackspace})
	if last != "" {
		t.Errorf("onChange after backspace: got %q, want %q", last, "")
	}
}

func TestTextInputHorizontalScroll(t *testing.T) {
	t.Parallel()

	ti := NewTextInput("ABCDEFGHIJ") // 10 chars
	ti.SetGeometry(layout.Rect{X: 0, Y: 0, Width: 5, Height: 1})

	// Move cursor to end.
	ti.OnKey(KeyEvent{Code: KeyEnd})
	if ti.pos != 10 {
		t.Fatalf("pos: got %d, want 10", ti.pos)
	}
	// With width=5 and pos=10, scroll should be 10-5+1=6.
	if ti.scroll != 6 {
		t.Errorf("scroll: got %d, want 6", ti.scroll)
	}
}

func TestTextInputRender(t *testing.T) {
	t.Parallel()

	ti := NewTextInput("Hi")
	ti.SetGeometry(layout.Rect{X: 0, Y: 0, Width: 5, Height: 1})

	buf := buffer.NewBuffer(5, 1)
	ti.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 5, Height: 1})

	// "Hi" + cursor '_' + 2 spaces
	if c := buf.GetCell(0, 0); c.Rune != 'H' {
		t.Errorf("cell 0: got %q, want H", c.Rune)
	}
	if c := buf.GetCell(1, 0); c.Rune != 'i' {
		t.Errorf("cell 1: got %q, want i", c.Rune)
	}
	// Cursor at pos=2, should be '_' with inverted colors.
	if c := buf.GetCell(2, 0); c.Rune != '_' {
		t.Errorf("cell 2 (cursor): got %q, want _", c.Rune)
	}
}

func TestTextInputFocusable(t *testing.T) {
	t.Parallel()
	ti := NewTextInput("")
	if !ti.Focusable() {
		t.Error("TextInput should be focusable")
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkListRender(b *testing.B) {
	items := make([]ListItem, 1000)
	for i := range items {
		items[i] = ListItem{Text: "item"}
	}
	l := NewList(items,
		WithForeground(0xFFFFFF),
		WithBackground(0x000000),
	)
	l.SetGeometry(layout.Rect{X: 0, Y: 0, Width: 10, Height: 20})
	l.SetSelected(500)
	buf := buffer.NewBuffer(10, 20)
	rect := layout.Rect{X: 0, Y: 0, Width: 10, Height: 20}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		l.Render(viewport.RenderCtx{Buf: buf}, rect)
	}
}

func BenchmarkTextInputRender(b *testing.B) {
	ti := NewTextInput("Hello, World! This is a benchmark.")
	ti.SetGeometry(layout.Rect{X: 0, Y: 0, Width: 20, Height: 1})
	buf := buffer.NewBuffer(20, 1)
	rect := layout.Rect{X: 0, Y: 0, Width: 20, Height: 1}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		ti.Render(viewport.RenderCtx{Buf: buf}, rect)
	}
}
