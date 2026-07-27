package widgets

import (
	"strings"
	"testing"

	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/style"
)

// ---------------------------------------------------------------------------
// Card — border + shadow + padding
// ---------------------------------------------------------------------------

func TestCardRoundedBorderWithShadow(t *testing.T) {
	t.Parallel()

	// Card 10x5 with rounded border and shadow offset (1,1).
	// Child is a simple text widget.
	child := NewText("Hi", WithStyle(style.New().Foreground(style.White)))
	card := NewCard(child,
		WithStyle(
			style.New().
				Background(style.Black).
				RoundedBorder(style.Green).
				Shadow(1, 1, style.ShadowColor).
				Padding(1, 0),
		),
	)

	// Buffer is 12x7 to accommodate shadow overflow.
	buf := buffer.NewBuffer(12, 7)
	card.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 10, Height: 5})

	dump := dumpBuf(buf)

	// Verify corners at expected positions.
	// Row 0: ╭────────╮
	if !strings.Contains(dump[0], "╭") {
		t.Errorf("row 0 missing top-left corner: %q", dump[0])
	}
	if !strings.Contains(dump[0], "╮") {
		t.Errorf("row 0 missing top-right corner: %q", dump[0])
	}

	// Row 4: ╰────────╯
	if !strings.Contains(dump[4], "╰") {
		t.Errorf("row 4 missing bottom-left corner: %q", dump[4])
	}
	if !strings.Contains(dump[4], "╯") {
		t.Errorf("row 4 missing bottom-right corner: %q", dump[4])
	}

	// Shadow: row 5 should have half-block shadow at x+1.
	if len(dump) > 5 {
		// Shadow at (1,5) — offset (1,1) from card bottom.
		c := buf.GetCell(1, 5)
		if c.Rune != style.ShadowBottom {
			t.Errorf("shadow cell(1,5): got %q, want %c", c.Rune, style.ShadowBottom)
		}
		if c.Fg != uint32(style.ShadowColor) {
			t.Errorf("shadow cell(1,5) Fg: got 0x%06X, want 0x%06X", c.Fg, style.ShadowColor)
		}
	}

	// Child "Hi" should be inside the border (row 1, shifted by border+padding).
	// Border=1, PaddingX=1, so child starts at x=2.
	c := buf.GetCell(2, 1)
	if c.Rune != 'H' {
		t.Errorf("child cell(2,1): got %q, want H", c.Rune)
	}
}

func TestCardShadowDoesNotOverlapContent(t *testing.T) {
	t.Parallel()

	// Two cards stacked vertically. Shadow of top card must NOT
	// overwrite the bottom card's content.
	topCard := NewCard(
		NewText("TOP", WithStyle(style.New().Foreground(style.White))),
		WithStyle(
			style.New().Background(style.Black).RoundedBorder(style.Green).Shadow(1, 1, style.ShadowColor),
		),
	)
	bottomCard := NewCard(
		NewText("BOT", WithStyle(style.New().Foreground(style.White))),
		WithStyle(
			style.New().Background(style.Black).RoundedBorder(style.Cyan),
		),
	)

	buf := buffer.NewBuffer(12, 12)
	topRect := layout.Rect{X: 0, Y: 0, Width: 10, Height: 4}
	bottomRect := layout.Rect{X: 0, Y: 5, Width: 10, Height: 4}

	// Render bottom first, then top (top's shadow could overwrite bottom).
	bottomCard.Render(viewport.RenderCtx{Buf: buf}, bottomRect)
	topCard.Render(viewport.RenderCtx{Buf: buf}, topRect)

	// Bottom card's top-left corner at (0,5) must still be ╭.
	c := buf.GetCell(0, 5)
	if c.Rune != '╭' {
		t.Errorf("bottom card corner(0,5): got %q, want ╭ — shadow overwrote content", c.Rune)
	}

	// Shadow should be at (1,4) — below top card, right of its left edge.
	c = buf.GetCell(1, 4)
	if c.Rune != style.ShadowBottom {
		t.Errorf("shadow(1,4): got %q, want %c", c.Rune, style.ShadowBottom)
	}

	// Shadow must NOT be inside the top card rect (e.g. at (1,1) should be border/bg, not shadow).
	c = buf.GetCell(1, 1)
	if c.Rune == style.ShadowBottom {
		t.Errorf("shadow leaked inside card at (1,1): got ░")
	}
}

func TestCardSolidBorder(t *testing.T) {
	t.Parallel()

	child := NewText("OK", WithStyle(style.New().Foreground(style.White)))
	card := NewCard(child,
		WithStyle(
			style.New().Background(style.Black).SolidBorder(style.White),
		),
	)

	buf := buffer.NewBuffer(8, 3)
	card.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 8, Height: 3})

	// Top-left corner should be ┌
	if c := buf.GetCell(0, 0); c.Rune != '┌' {
		t.Errorf("TL: got %q, want ┌", c.Rune)
	}
	// Top-right corner should be ┐
	if c := buf.GetCell(7, 0); c.Rune != '┐' {
		t.Errorf("TR: got %q, want ┐", c.Rune)
	}
	// Bottom-left corner should be └
	if c := buf.GetCell(0, 2); c.Rune != '└' {
		t.Errorf("BL: got %q, want └", c.Rune)
	}
	// Bottom-right corner should be ┘
	if c := buf.GetCell(7, 2); c.Rune != '┘' {
		t.Errorf("BR: got %q, want ┘", c.Rune)
	}
}

func TestCardNoBorder(t *testing.T) {
	t.Parallel()

	child := NewText("X", WithStyle(style.New().Foreground(style.White)))
	card := NewCard(child, WithStyle(style.New().Background(style.Black)))

	buf := buffer.NewBuffer(5, 3)
	card.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 5, Height: 3})

	// No border — child should render at (0,0).
	if c := buf.GetCell(0, 0); c.Rune != 'X' {
		t.Errorf("cell(0,0): got %q, want X", c.Rune)
	}
}

func TestCardOnKeyForwards(t *testing.T) {
	t.Parallel()

	var pressed bool
	child := NewButton("X", WithOnPress(func() { pressed = true }))
	card := NewCard(child)

	consumed := card.OnKey(KeyEvent{Code: KeyEnter})
	if !consumed {
		t.Error("Card should forward Enter to child Button")
	}
	if !pressed {
		t.Error("child Button should have been pressed")
	}
}

func TestCardOnMouseForwardsToInner(t *testing.T) {
	t.Parallel()

	var clicked bool
	child := NewButton("X", WithOnPress(func() { clicked = true }))
	card := NewCard(child,
		WithStyle(
			style.New().Background(style.Black).SolidBorder(style.White).Padding(1, 0),
		),
	)

	// Card at (0,0) size 8x4. Border=1, PaddingX=1.
	// Inner area: x=2..6, y=1..2.
	card.SetGeometry(layout.Rect{X: 0, Y: 0, Width: 8, Height: 4})

	// Render once to set child geometry (Card computes inner rect during render).
	buf := buffer.NewBuffer(8, 4)
	card.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 8, Height: 4})

	// Click inside inner area.
	consumed := card.OnMouse(MouseEvent{Button: MouseLeft, Action: MousePress, X: 3, Y: 1})
	if !consumed {
		t.Error("click inside inner area should be consumed")
	}
	if !clicked {
		t.Error("child Button should have been clicked")
	}

	// Click on border — should NOT forward.
	clicked = false
	consumed = card.OnMouse(MouseEvent{Button: MouseLeft, Action: MousePress, X: 0, Y: 0})
	if consumed {
		t.Error("click on border should not be consumed by child")
	}
}

func TestCardFocusable(t *testing.T) {
	t.Parallel()

	btn := NewButton("X")
	card := NewCard(btn)
	if !card.Focusable() {
		t.Error("Card with focusable child should be focusable")
	}

	txt := NewText("X")
	card2 := NewCard(txt)
	if card2.Focusable() {
		t.Error("Card with non-focusable child should not be focusable")
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkCardRender(b *testing.B) {
	child := NewText("Hello, World!",
		WithStyle(style.New().Foreground(style.White)),
	)
	card := NewCard(child,
		WithStyle(
			style.New().
				Background(style.Black).
				RoundedBorder(style.Green).
				Shadow(1, 1, style.ShadowColor).
				Padding(1, 0),
		),
	)
	buf := buffer.NewBuffer(30, 5)
	rect := layout.Rect{X: 0, Y: 0, Width: 20, Height: 3}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		card.Render(viewport.RenderCtx{Buf: buf}, rect)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// dumpBuf returns the buffer contents as a slice of strings (one per row).
func dumpBuf(buf *buffer.Buffer) []string {
	rows := make([]string, buf.Height)
	for y := 0; y < buf.Height; y++ {
		var sb strings.Builder
		for x := 0; x < buf.Width; x++ {
			c := buf.GetCell(x, y)
			if c.Rune == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(c.Rune)
			}
		}
		rows[y] = sb.String()
	}
	return rows
}
