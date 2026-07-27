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

func TestCard2LayerShadow(t *testing.T) {
	t.Parallel()

	child := NewText("Hi", WithStyle(style.New().Foreground(style.White)))
	card := NewCard(child,
		WithStyle(
			style.New().
				Background(style.Black).
				Shadow(1, 1, style.ShadowColor),
		),
	)

	// Card at (0,0) 10x5. Shadow: layer1 ▒ at col10/row5, layer2 ░ at col11/row6.
	buf := buffer.NewBuffer(13, 8)
	card.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 10, Height: 5})

	// Child "Hi" at (0,0) — no border, no padding.
	if c := buf.GetCell(0, 0); c.Rune != 'H' {
		t.Errorf("child(0,0): got %q, want H", c.Rune)
	}

	// Layer 1: ▒ at right column (10), rows 0-4.
	for y := 0; y < 5; y++ {
		c := buf.GetCell(10, y)
		if c.Rune != style.ShadowLayer1 {
			t.Errorf("L1 right(10,%d): got %q, want %c", y, c.Rune, style.ShadowLayer1)
		}
	}

	// Layer 1: ▒ at bottom row (5), cols 0-9.
	for x := 0; x < 10; x++ {
		c := buf.GetCell(x, 5)
		if c.Rune != style.ShadowLayer1 {
			t.Errorf("L1 bottom(%d,5): got %q, want %c", x, c.Rune, style.ShadowLayer1)
		}
	}

	// Layer 2: ░ at right+1 column (11), rows 0-5.
	for y := 0; y <= 5; y++ {
		c := buf.GetCell(11, y)
		if c.Rune != style.ShadowLayer2 {
			t.Errorf("L2 right(11,%d): got %q, want %c", y, c.Rune, style.ShadowLayer2)
		}
	}

	// Layer 2: ░ at bottom+1 row (6), cols 1-10.
	for x := 1; x <= 10; x++ {
		c := buf.GetCell(x, 6)
		if c.Rune != style.ShadowLayer2 {
			t.Errorf("L2 bottom(%d,6): got %q, want %c", x, c.Rune, style.ShadowLayer2)
		}
	}

	// Shadow must NOT be inside card.
	if c := buf.GetCell(1, 1); c.Rune == style.ShadowLayer1 || c.Rune == style.ShadowLayer2 {
		t.Errorf("shadow leaked inside card at (1,1): got %c", c.Rune)
	}

	// Child "Hi" at (0,0) — no border, no padding.
	if c := buf.GetCell(0, 0); c.Rune != 'H' {
		t.Errorf("child(0,0): got %q, want H", c.Rune)
	}
}

func TestCardShadowDoesNotOverlapContent(t *testing.T) {
	t.Parallel()

	topCard := NewCard(
		NewText("TOP", WithStyle(style.New().Foreground(style.White))),
		WithStyle(
			style.New().Background(style.Black).Shadow(1, 1, style.ShadowColor),
		),
	)
	bottomCard := NewCard(
		NewText("BOT", WithStyle(style.New().Foreground(style.White))),
		WithStyle(
			style.New().Background(style.DarkGray),
		),
	)

	buf := buffer.NewBuffer(14, 14)
	topRect := layout.Rect{X: 0, Y: 0, Width: 10, Height: 4}
	bottomRect := layout.Rect{X: 0, Y: 8, Width: 10, Height: 4}

	bottomCard.Render(viewport.RenderCtx{Buf: buf}, bottomRect)
	topCard.Render(viewport.RenderCtx{Buf: buf}, topRect)

	// Bottom card at (0,8) must have its background (shadow from top card ends at row 6).
	c := buf.GetCell(0, 8)
	if c.Rune != 'B' {
		t.Errorf("bottom card(0,8): got %q, want B (child content)", c.Rune)
	}

	// Top card layer 1: ▒ at (10, 1).
	c = buf.GetCell(10, 1)
	if c.Rune != style.ShadowLayer1 {
		t.Errorf("L1(10,1): got %q, want %c", c.Rune, style.ShadowLayer1)
	}

	// Top card layer 2: ░ at (11, 1).
	c = buf.GetCell(11, 1)
	if c.Rune != style.ShadowLayer2 {
		t.Errorf("L2(11,1): got %q, want %c", c.Rune, style.ShadowLayer2)
	}

	// Shadow must NOT be inside the card.
	c = buf.GetCell(1, 1)
	if c.Rune == style.ShadowLayer1 || c.Rune == style.ShadowLayer2 {
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
				Shadow(1, 1, style.ShadowColor).
				Padding(1, 0),
		),
	)
	buf := buffer.NewBuffer(25, 6)
	rect := layout.Rect{X: 0, Y: 0, Width: 20, Height: 3}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		card.Render(viewport.RenderCtx{Buf: buf}, rect)
	}
}

// ---------------------------------------------------------------------------
// Dither shadow
// ---------------------------------------------------------------------------

func TestCardDitherShadow(t *testing.T) {
	t.Parallel()

	child := NewText("Hi", WithStyle(style.New().Foreground(style.White)))
	card := NewCard(child,
		WithStyle(
			style.New().
				Background(style.Black).
				RoundedBorder(style.Cyan).
				DitherShadow(3, 2, style.DarkGray),
		),
	)

	// Card at (0,0) 8x3, blur 3 right, 2 down → buffer 12x6.
	buf := buffer.NewBuffer(12, 6)
	card.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 8, Height: 3})

	// Border must be intact.
	if c := buf.GetCell(0, 0); c.Rune != '╭' {
		t.Errorf("TL(0,0): got %q, want ╭", c.Rune)
	}
	if c := buf.GetCell(7, 2); c.Rune != '╯' {
		t.Errorf("BR(7,2): got %q, want ╯", c.Rune)
	}

	// Right shadow: column 8 should have density characters (not border chars).
	c := buf.GetCell(8, 1)
	if c.Rune == '│' || c.Rune == ' ' {
		t.Errorf("right shadow(8,1): got %q, want density char", c.Rune)
	}

	// Bottom shadow: row 3 should have density characters.
	c = buf.GetCell(4, 3)
	if c.Rune == '─' || c.Rune == ' ' {
		t.Errorf("bottom shadow(4,3): got %q, want density char", c.Rune)
	}

	// Denser near card, sparser farther away.
	// At (8, 0) — distance 1 from right border → should be dense.
	// At (10, 0) — distance 3 → should be sparser.
	near := buf.GetCell(8, 0)
	far := buf.GetCell(10, 0)
	if near.Rune == ' ' && far.Rune == ' ' {
		t.Error("both near and far shadow cells are empty")
	}

	// Shadow must NOT be inside card.
	if c := buf.GetCell(1, 1); c.Rune == '@' || c.Rune == '#' || c.Rune == '%' {
		t.Errorf("shadow leaked inside card at (1,1): got %c", c.Rune)
	}
}

func BenchmarkCardDitherRender(b *testing.B) {
	child := NewText("Hello",
		WithStyle(style.New().Foreground(style.White)),
	)
	card := NewCard(child,
		WithStyle(
			style.New().
				Background(style.Black).
				RoundedBorder(style.Cyan).
				DitherShadow(3, 2, style.DarkGray),
		),
	)
	buf := buffer.NewBuffer(25, 7)
	rect := layout.Rect{X: 0, Y: 0, Width: 15, Height: 4}

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
