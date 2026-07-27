package widgets

import (
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

	// Buffer 12x7: card at (0,0) 10x5, shadow at col 10 and row 5.
	buf := buffer.NewBuffer(12, 7)
	card.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 10, Height: 5})

	// All four border corners must be intact (shadow is outside).
	if c := buf.GetCell(0, 0); c.Rune != '╭' {
		t.Errorf("TL(0,0): got %q, want ╭", c.Rune)
	}
	if c := buf.GetCell(9, 0); c.Rune != '╮' {
		t.Errorf("TR(9,0): got %q, want ╮", c.Rune)
	}
	if c := buf.GetCell(0, 4); c.Rune != '╰' {
		t.Errorf("BL(0,4): got %q, want ╰", c.Rune)
	}
	if c := buf.GetCell(9, 4); c.Rune != '╯' {
		t.Errorf("BR(9,4): got %q, want ╯", c.Rune)
	}

	// Top border ─ intact.
	if c := buf.GetCell(5, 0); c.Rune != '─' {
		t.Errorf("top(5,0): got %q, want ─", c.Rune)
	}
	// Left border │ intact.
	if c := buf.GetCell(0, 2); c.Rune != '│' {
		t.Errorf("left(0,2): got %q, want │", c.Rune)
	}
	// Right border │ intact.
	if c := buf.GetCell(9, 2); c.Rune != '│' {
		t.Errorf("right(9,2): got %q, want │", c.Rune)
	}
	// Bottom border ─ intact.
	if c := buf.GetCell(5, 4); c.Rune != '─' {
		t.Errorf("bottom(5,4): got %q, want ─", c.Rune)
	}

	// Shadow outside: ▐ at column 10, rows 0-4.
	for y := 0; y < 5; y++ {
		c := buf.GetCell(10, y)
		if c.Rune != style.ShadowRight {
			t.Errorf("right(10,%d): got %q, want %c", y, c.Rune, style.ShadowRight)
		}
		if c.Fg != uint32(style.ShadowColor) {
			t.Errorf("right(10,%d) Fg: 0x%06X", y, c.Fg)
		}
	}

	// Shadow outside: ▄ at row 5, columns 0-9.
	for x := 0; x < 10; x++ {
		c := buf.GetCell(x, 5)
		if c.Rune != style.ShadowBottom {
			t.Errorf("bottom(%d,5): got %q, want %c", x, c.Rune, style.ShadowBottom)
		}
	}

	// Corner: ▒ at (10, 5).
	if c := buf.GetCell(10, 5); c.Rune != style.ShadowCorner {
		t.Errorf("corner(10,5): got %q, want %c", c.Rune, style.ShadowCorner)
	}

	// Child inside.
	if c := buf.GetCell(2, 1); c.Rune != 'H' {
		t.Errorf("child(2,1): got %q, want H", c.Rune)
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

	// Top card at (0,0) 10x4. Shadow outside: col 10, row 4.
	// Bottom card at (0,5) 10x4. Its top-left corner at (0,5) must be intact.
	c := buf.GetCell(0, 5)
	if c.Rune != '╭' {
		t.Errorf("bottom card corner(0,5): got %q, want ╭", c.Rune)
	}

	// Top card's border intact.
	if c := buf.GetCell(9, 3); c.Rune != '╯' {
		t.Errorf("top card BR(9,3): got %q, want ╯", c.Rune)
	}

	// Top card's bottom shadow: ▄ at row 4 (outside rect).
	c = buf.GetCell(1, 4)
	if c.Rune != style.ShadowBottom {
		t.Errorf("shadow(1,4): got %q, want %c", c.Rune, style.ShadowBottom)
	}

	// Top card's right shadow: ▐ at column 10 (outside rect).
	c = buf.GetCell(10, 1)
	if c.Rune != style.ShadowRight {
		t.Errorf("shadow(10,1): got %q, want %c", c.Rune, style.ShadowRight)
	}

	// Shadow must NOT be inside the card.
	c = buf.GetCell(1, 1)
	if c.Rune == style.ShadowBottom || c.Rune == style.ShadowRight {
		t.Errorf("shadow leaked inside card at (1,1): got %c", c.Rune)
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
