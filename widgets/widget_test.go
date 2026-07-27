package widgets

import (
	"testing"

	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/style"
)

// ---------------------------------------------------------------------------
// NewButton — option application
// ---------------------------------------------------------------------------

func TestNewButtonOptions(t *testing.T) {
	t.Parallel()

	var called bool
	btn := NewButton("OK",
		WithWidth(10),
		WithHeight(3),
		WithForeground(0xFF0000),
		WithBackground(0x00FF00),
		WithBold(),
		WithOnPress(func() { called = true }),
	)

	if btn.text != "OK" {
		t.Errorf("text: got %q, want %q", btn.text, "OK")
	}
	if btn.config.Width != 10 {
		t.Errorf("Width: got %d, want 10", btn.config.Width)
	}
	if btn.config.Height != 3 {
		t.Errorf("Height: got %d, want 3", btn.config.Height)
	}
	if btn.config.style.FG() != 0xFF0000 {
		t.Errorf("Fg: got 0x%06X, want 0xFF0000", btn.config.style.FG())
	}
	if btn.config.style.BG() != 0x00FF00 {
		t.Errorf("Bg: got 0x%06X, want 0x00FF00", btn.config.style.BG())
	}
	if btn.config.style.Attrs()&buffer.Bold == 0 {
		t.Error("Bold attribute not set")
	}
	if btn.config.onPress == nil {
		t.Fatal("onPress is nil")
	}
	btn.Press()
	if !called {
		t.Error("Press() did not invoke onPress")
	}
}

func TestNewButtonWithStyle(t *testing.T) {
	t.Parallel()

	s := style.New().Foreground(style.Red).Background(style.Blue).Bold().Italic()
	btn := NewButton("Styled", WithStyle(s))

	if btn.config.style.FG() != style.Red {
		t.Errorf("FG: got 0x%06X, want Red", btn.config.style.FG())
	}
	if btn.config.style.BG() != style.Blue {
		t.Errorf("BG: got 0x%06X, want Blue", btn.config.style.BG())
	}
	if btn.config.style.Attrs()&buffer.Bold == 0 {
		t.Error("Bold not set")
	}
	if btn.config.style.Attrs()&buffer.Italic == 0 {
		t.Error("Italic not set")
	}
}

func TestNewButtonDefaults(t *testing.T) {
	t.Parallel()

	btn := NewButton("X")
	if btn.config.style.FG() != style.Default {
		t.Errorf("default FG: got 0x%06X, want Default", btn.config.style.FG())
	}
	if btn.config.style.BG() != style.Default {
		t.Errorf("default BG: got 0x%06X, want Default", btn.config.style.BG())
	}
	if btn.config.style.Attrs() != 0 {
		t.Errorf("default Attrs: got %d, want 0", btn.config.style.Attrs())
	}
	if btn.config.onPress != nil {
		t.Error("default onPress should be nil")
	}
	btn.Press() // should not panic
}

// ---------------------------------------------------------------------------
// Text — Render
// ---------------------------------------------------------------------------

func TestTextRender(t *testing.T) {
	t.Parallel()

	buf := buffer.NewBuffer(10, 1)
	txt := NewText("Hello",
		WithForeground(0xAABBCC),
		WithBackground(0x112233),
		WithBold(),
	)

	txt.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 10, Height: 1})

	want := []struct {
		x    int
		r    rune
		fg   uint32
		bg   uint32
		attr buffer.Attrs
	}{
		{0, 'H', 0xAABBCC, 0x112233, buffer.Bold},
		{1, 'e', 0xAABBCC, 0x112233, buffer.Bold},
		{2, 'l', 0xAABBCC, 0x112233, buffer.Bold},
		{3, 'l', 0xAABBCC, 0x112233, buffer.Bold},
		{4, 'o', 0xAABBCC, 0x112233, buffer.Bold},
	}
	for _, w := range want {
		cell := buf.GetCell(w.x, 0)
		if cell.Rune != w.r || cell.Fg != w.fg || cell.Bg != w.bg || cell.Attrs != w.attr {
			t.Errorf("cell(%d,0): got {Rune:%q Fg:0x%06X Bg:0x%06X Attrs:%d}, want {Rune:%q Fg:0x%06X Bg:0x%06X Attrs:%d}",
				w.x, cell.Rune, cell.Fg, cell.Bg, cell.Attrs, w.r, w.fg, w.bg, w.attr)
		}
	}

	// Cells beyond the text should remain zero.
	if c := buf.GetCell(5, 0); c.Rune != 0 {
		t.Errorf("cell(5,0): got rune %q, want zero", c.Rune)
	}
}

func TestTextRenderWithStyle(t *testing.T) {
	t.Parallel()

	s := style.New().Foreground(style.Cyan).Background(style.DarkGray).Underline()
	buf := buffer.NewBuffer(5, 1)
	txt := NewText("Hi", WithStyle(s))
	txt.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 5, Height: 1})

	c0 := buf.GetCell(0, 0)
	if c0.Fg != uint32(style.Cyan) {
		t.Errorf("Fg: got 0x%06X, want Cyan", c0.Fg)
	}
	if c0.Bg != uint32(style.DarkGray) {
		t.Errorf("Bg: got 0x%06X, want DarkGray", c0.Bg)
	}
	if c0.Attrs&buffer.Underline == 0 {
		t.Error("Underline not set")
	}
}

func TestTextRenderClipsByWidth(t *testing.T) {
	t.Parallel()

	buf := buffer.NewBuffer(3, 1)
	txt := NewText("ABCDE")
	txt.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 3, Height: 1})

	for i, want := range []rune{'A', 'B', 'C'} {
		if got := buf.GetCell(i, 0).Rune; got != want {
			t.Errorf("cell(%d,0): got %q, want %q", i, got, want)
		}
	}
	// D and E should not be written (clipped).
}

// ---------------------------------------------------------------------------
// Button — Render
// ---------------------------------------------------------------------------

func TestButtonRender(t *testing.T) {
	t.Parallel()

	buf := buffer.NewBuffer(10, 3)
	btn := NewButton("OK",
		WithForeground(0xFFFFFF),
		WithBackground(0x0000FF),
	)

	btn.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 10, Height: 3})

	// Background fill: every cell should have Bg=0x0000FF and Rune=' '.
	for y := 0; y < 3; y++ {
		for x := 0; x < 10; x++ {
			cell := buf.GetCell(x, y)
			if cell.Bg != 0x0000FF {
				t.Errorf("cell(%d,%d) Bg: got 0x%06X, want 0x0000FF", x, y, cell.Bg)
			}
		}
	}

	// Label "OK" (2 runes) centred in 10-wide rect → startX = 0 + (10-2)/2 = 4.
	// Vertical centre in 3-tall rect → startY = 0 + 3/2 = 1.
	centreY := 1
	wantCells := []struct {
		x int
		r rune
	}{
		{4, 'O'},
		{5, 'K'},
	}
	for _, wc := range wantCells {
		cell := buf.GetCell(wc.x, centreY)
		if cell.Rune != wc.r {
			t.Errorf("cell(%d,%d) Rune: got %q, want %q", wc.x, centreY, cell.Rune, wc.r)
		}
		if cell.Fg != 0xFFFFFF {
			t.Errorf("cell(%d,%d) Fg: got 0x%06X, want 0xFFFFFF", wc.x, centreY, cell.Fg)
		}
	}

	// The cell before the label on the centre row should be background space.
	if c := buf.GetCell(3, centreY); c.Rune != ' ' {
		t.Errorf("cell(3,%d): got %q, want space", centreY, c.Rune)
	}
	// The cell after the label on the centre row should be background space.
	if c := buf.GetCell(6, centreY); c.Rune != ' ' {
		t.Errorf("cell(6,%d): got %q, want space", centreY, c.Rune)
	}
}

func TestButtonSetStyle(t *testing.T) {
	t.Parallel()

	btn := NewButton("X")
	btn.SetStyle(style.New().Foreground(style.Green).Background(style.Black))
	if btn.Style().FG() != style.Green {
		t.Errorf("FG after SetStyle: got 0x%06X, want Green", btn.Style().FG())
	}
}

func TestButtonPressNil(t *testing.T) {
	t.Parallel()

	btn := NewButton("X")
	btn.Press() // should not panic
}

func TestButtonPressCallsHandler(t *testing.T) {
	t.Parallel()

	var n int
	btn := NewButton("X", WithOnPress(func() { n++ }))
	btn.Press()
	btn.Press()
	if n != 2 {
		t.Errorf("Press count: got %d, want 2", n)
	}
}

// ---------------------------------------------------------------------------
// Node interface — geometry and events
// ---------------------------------------------------------------------------

func TestButtonGeometry(t *testing.T) {
	t.Parallel()

	btn := NewButton("X")
	btn.SetGeometry(layout.Rect{X: 5, Y: 3, Width: 10, Height: 2})
	got := btn.Geometry()
	if got.X != 5 || got.Y != 3 || got.Width != 10 || got.Height != 2 {
		t.Errorf("Geometry: got %+v, want {X:5 Y:3 Width:10 Height:2}", got)
	}
}

func TestButtonFocusable(t *testing.T) {
	t.Parallel()

	btn := NewButton("X")
	if !btn.Focusable() {
		t.Error("Button should be focusable")
	}
}

func TestButtonOnKeyEnter(t *testing.T) {
	t.Parallel()

	var called bool
	btn := NewButton("X", WithOnPress(func() { called = true }))
	consumed := btn.OnKey(KeyEvent{Code: KeyEnter})
	if !consumed {
		t.Error("Enter should be consumed")
	}
	if !called {
		t.Error("Enter should trigger onPress")
	}
}

func TestButtonOnKeyOther(t *testing.T) {
	t.Parallel()

	btn := NewButton("X")
	consumed := btn.OnKey(KeyEvent{Rune: 'a'})
	if consumed {
		t.Error("non-Enter key should not be consumed")
	}
}

func TestButtonOnMouseLeftClick(t *testing.T) {
	t.Parallel()

	var called bool
	btn := NewButton("X", WithOnPress(func() { called = true }))
	btn.SetGeometry(layout.Rect{X: 0, Y: 0, Width: 10, Height: 3})

	consumed := btn.OnMouse(MouseEvent{Button: MouseLeft, Action: MousePress, X: 5, Y: 1})
	if !consumed {
		t.Error("left click inside should be consumed")
	}
	if !called {
		t.Error("left click should trigger onPress")
	}
}

func TestButtonOnMouseOutside(t *testing.T) {
	t.Parallel()

	btn := NewButton("X")
	btn.SetGeometry(layout.Rect{X: 0, Y: 0, Width: 10, Height: 3})

	consumed := btn.OnMouse(MouseEvent{Button: MouseLeft, Action: MousePress, X: 15, Y: 5})
	if consumed {
		t.Error("click outside should not be consumed")
	}
}

func TestButtonOnMouseMotion(t *testing.T) {
	t.Parallel()

	btn := NewButton("X")
	btn.SetGeometry(layout.Rect{X: 0, Y: 0, Width: 10, Height: 3})

	consumed := btn.OnMouse(MouseEvent{Action: MouseMotion, X: 5, Y: 1})
	if consumed {
		t.Error("motion should not be consumed")
	}
}

func TestTextNotFocusable(t *testing.T) {
	t.Parallel()

	txt := NewText("X")
	if txt.Focusable() {
		t.Error("Text should not be focusable")
	}
	if txt.OnKey(KeyEvent{Code: KeyEnter}) {
		t.Error("Text should not consume keys")
	}
	if txt.OnMouse(MouseEvent{Button: MouseLeft, Action: MousePress, X: 0, Y: 0}) {
		t.Error("Text should not consume mouse")
	}
}

func TestHitTest(t *testing.T) {
	t.Parallel()

	rect := layout.Rect{X: 5, Y: 3, Width: 10, Height: 4}
	cases := []struct {
		x, y int
		want bool
	}{
		{5, 3, true},   // top-left corner
		{14, 6, true},  // bottom-right corner (exclusive: 5+10-1=14, 3+4-1=6)
		{4, 3, false},  // left of rect
		{15, 3, false}, // right of rect (5+10=15, out)
		{5, 2, false},  // above rect
		{5, 7, false},  // below rect (3+4=7, out)
		{10, 5, true},  // center
	}
	for _, tc := range cases {
		if got := HitTest(rect, tc.x, tc.y); got != tc.want {
			t.Errorf("HitTest(%d,%d): got %v, want %v", tc.x, tc.y, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Viewport — culling and clipping
// ---------------------------------------------------------------------------

func TestViewportCullingSkipsRender(t *testing.T) {
	t.Parallel()

	buf := buffer.NewBuffer(10, 10)
	view := viewport.New(10, 10) // visible: [0,10) x [0,10)
	btn := NewButton("X",
		WithForeground(0xFFFFFF),
		WithBackground(0x0000FF),
	)

	// Widget at world (100, 100) — completely outside viewport.
	btn.Render(viewport.RenderCtx{Buf: buf, View: view}, layout.Rect{X: 100, Y: 100, Width: 5, Height: 3})

	// Buffer should be untouched (all zeros).
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if c := buf.GetCell(x, y); c.Rune != 0 || c.Bg != 0 {
				t.Errorf("cell(%d,%d) was written but widget is off-screen: %+v", x, y, c)
			}
		}
	}
}

func TestViewportPartialClipping(t *testing.T) {
	t.Parallel()

	buf := buffer.NewBuffer(10, 10)
	view := viewport.New(10, 10) // visible: [0,10) x [0,10)
	btn := NewButton("X",
		WithForeground(0xFFFFFF),
		WithBackground(0x0000FF),
	)

	// Widget at (-2, -2) with size 5x5. Only cells [0,3) x [0,3) are
	// on screen (the rest are clipped by negative coords).
	btn.Render(viewport.RenderCtx{Buf: buf, View: view}, layout.Rect{X: -2, Y: -2, Width: 5, Height: 5})

	// Cells [0,3) x [0,3) should have the background color.
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			if c := buf.GetCell(x, y); c.Bg != 0x0000FF {
				t.Errorf("cell(%d,%d) Bg: got 0x%06X, want 0x0000FF", x, y, c.Bg)
			}
		}
	}

	// Cells outside [0,3) should be untouched.
	for y := 3; y < 10; y++ {
		for x := 3; x < 10; x++ {
			if c := buf.GetCell(x, y); c.Rune != 0 || c.Bg != 0 {
				t.Errorf("cell(%d,%d) should be untouched: %+v", x, y, c)
			}
		}
	}
}

func TestViewportNoViewBackwardCompat(t *testing.T) {
	t.Parallel()

	buf := buffer.NewBuffer(10, 3)
	btn := NewButton("OK",
		WithForeground(0xFFFFFF),
		WithBackground(0x0000FF),
	)

	// nil viewport — legacy mode, no offset.
	btn.Render(viewport.RenderCtx{Buf: buf}, layout.Rect{X: 0, Y: 0, Width: 10, Height: 3})

	if c := buf.GetCell(0, 0); c.Bg != 0x0000FF {
		t.Errorf("nil viewport should render normally: Bg=0x%06X", c.Bg)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkButtonRender(b *testing.B) {
	btn := NewButton("Submit",
		WithForeground(0xFFFFFF),
		WithBackground(0x0000FF),
		WithBold(),
	)
	buf := buffer.NewBuffer(20, 3)
	rect := layout.Rect{X: 0, Y: 0, Width: 20, Height: 3}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		btn.Render(viewport.RenderCtx{Buf: buf}, rect)
	}
}

func BenchmarkTextRender(b *testing.B) {
	txt := NewText("Hello, World!",
		WithForeground(0xCCCCCC),
		WithBackground(0x333333),
	)
	buf := buffer.NewBuffer(40, 1)
	rect := layout.Rect{X: 0, Y: 0, Width: 40, Height: 1}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		txt.Render(viewport.RenderCtx{Buf: buf}, rect)
	}
}

// BenchmarkRenderWithViewport proves culling: 1000 widgets off-screen,
// only 10 visible. Render time should equal rendering 10 widgets.
func BenchmarkRenderWithViewport(b *testing.B) {
	const total = 1000
	const visible = 10

	view := viewport.New(40, 10) // visible: [0,40) x [0,10)
	buf := buffer.NewBuffer(40, 10)

	// Create 1000 widgets: 10 visible at (0..9, 0), 990 off-screen at (100+).
	widgets := make([]*Button, total)
	for i := 0; i < total; i++ {
		x := i
		if i >= visible {
			x = 100 + i // off-screen
		}
		widgets[i] = NewButton("X",
			WithForeground(0xFFFFFF),
			WithBackground(0x0000FF),
		)
		widgets[i].SetGeometry(layout.Rect{X: x, Y: 0, Width: 3, Height: 1})
	}

	ctx := viewport.RenderCtx{Buf: buf, View: view}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		for _, w := range widgets {
			w.Render(ctx, w.Geometry())
		}
	}
}

// BenchmarkRenderNoViewportBaseline renders 10 widgets without viewport
// for comparison with BenchmarkRenderWithViewport.
func BenchmarkRenderNoViewportBaseline(b *testing.B) {
	buf := buffer.NewBuffer(40, 10)
	widgets := make([]*Button, 10)
	for i := range widgets {
		widgets[i] = NewButton("X",
			WithForeground(0xFFFFFF),
			WithBackground(0x0000FF),
		)
		widgets[i].SetGeometry(layout.Rect{X: i * 3, Y: 0, Width: 3, Height: 1})
	}

	ctx := viewport.RenderCtx{Buf: buf}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
		for _, w := range widgets {
			w.Render(ctx, w.Geometry())
		}
	}
}
