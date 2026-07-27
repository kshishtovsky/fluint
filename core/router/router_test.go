package router

import (
	"testing"

	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/widgets"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// focusableNode is a minimal focusable widget for testing.
type focusableNode struct {
	rect   layout.Rect
	keys   []widgets.KeyEvent
	mouse  []widgets.MouseEvent
	focus  bool
}

func (n *focusableNode) Render(_ *buffer.Buffer, _ layout.Rect) {}
func (n *focusableNode) Geometry() layout.Rect                  { return n.rect }
func (n *focusableNode) SetGeometry(rect layout.Rect)           { n.rect = rect }
func (n *focusableNode) Focusable() bool                        { return n.focus }
func (n *focusableNode) OnKey(key widgets.KeyEvent) bool        { n.keys = append(n.keys, key); return true }
func (n *focusableNode) OnMouse(mouse widgets.MouseEvent) bool  { n.mouse = append(n.mouse, mouse); return true }

func newFocusable(x, y, w, h int) *focusableNode {
	return &focusableNode{rect: layout.Rect{X: x, Y: y, Width: w, Height: h}, focus: true}
}

// ---------------------------------------------------------------------------
// Register / Unregister
// ---------------------------------------------------------------------------

func TestRegisterUnregister(t *testing.T) {
	t.Parallel()

	r := New()
	a := newFocusable(0, 0, 10, 1)
	b := newFocusable(0, 1, 10, 1)
	r.Register(a)
	r.Register(b)

	if len(r.Nodes()) != 2 {
		t.Fatalf("Nodes(): got %d, want 2", len(r.Nodes()))
	}

	// Duplicate register is ignored.
	r.Register(a)
	if len(r.Nodes()) != 2 {
		t.Fatalf("after dup register: got %d, want 2", len(r.Nodes()))
	}

	r.Unregister(a)
	if len(r.Nodes()) != 1 {
		t.Fatalf("after unregister: got %d, want 1", len(r.Nodes()))
	}
	if r.Nodes()[0] != widgets.Node(b) {
		t.Error("wrong node remaining")
	}
}

func TestUnregisterClearsFocus(t *testing.T) {
	t.Parallel()

	r := New()
	a := newFocusable(0, 0, 10, 1)
	r.Register(a)
	r.Focus(a)

	r.Unregister(a)
	if r.Focused() != nil {
		t.Error("focus should be cleared after unregister")
	}
}

// ---------------------------------------------------------------------------
// DispatchKey — focus routing
// ---------------------------------------------------------------------------

func TestDispatchKeyToFocused(t *testing.T) {
	t.Parallel()

	r := New()
	a := newFocusable(0, 0, 10, 1)
	b := newFocusable(0, 1, 10, 1)
	r.Register(a)
	r.Register(b)
	r.Focus(a)

	r.DispatchKey(widgets.KeyEvent{Rune: 'x'})
	if len(a.keys) != 1 {
		t.Errorf("a.keys: got %d, want 1", len(a.keys))
	}
	if len(b.keys) != 0 {
		t.Errorf("b.keys: got %d, want 0", len(b.keys))
	}
}

func TestDispatchKeyNoFocus(t *testing.T) {
	t.Parallel()

	r := New()
	a := newFocusable(0, 0, 10, 1)
	r.Register(a)

	consumed := r.DispatchKey(widgets.KeyEvent{Rune: 'x'})
	if consumed {
		t.Error("should not consume when nothing focused")
	}
}

// ---------------------------------------------------------------------------
// DispatchMouse — hit-testing and enter/leave
// ---------------------------------------------------------------------------

func TestDispatchMouseHitsTopmost(t *testing.T) {
	t.Parallel()

	r := New()
	bottom := newFocusable(0, 0, 10, 3)
	top := newFocusable(0, 0, 10, 3) // same rect, registered later = higher Z
	r.Register(bottom)
	r.Register(top)

	r.DispatchMouse(widgets.MouseEvent{Button: widgets.MouseLeft, Action: widgets.MousePress, X: 5, Y: 1})
	// top receives MouseEnter (hover transition) + MousePress = 2 events.
	if len(top.mouse) != 2 {
		t.Errorf("top.mouse: got %d, want 2 (Enter + Press)", len(top.mouse))
	}
	if len(bottom.mouse) != 0 {
		t.Errorf("bottom.mouse: got %d, want 0 (should not receive)", len(bottom.mouse))
	}
}

func TestDispatchMouseEnterLeave(t *testing.T) {
	t.Parallel()

	r := New()
	a := newFocusable(0, 0, 10, 3)
	b := newFocusable(15, 0, 10, 3)
	r.Register(a)
	r.Register(b)

	// Move over A.
	r.DispatchMouse(widgets.MouseEvent{Action: widgets.MouseMotion, X: 5, Y: 1})
	if r.Hovered() != a {
		t.Fatal("should be hovering A")
	}

	// Move over B — A gets Leave, B gets Enter.
	r.DispatchMouse(widgets.MouseEvent{Action: widgets.MouseMotion, X: 20, Y: 1})
	if r.Hovered() != b {
		t.Fatal("should be hovering B")
	}

	// Check A got MouseLeave.
	hasLeave := false
	for _, m := range a.mouse {
		if m.Action == widgets.MouseLeave {
			hasLeave = true
		}
	}
	if !hasLeave {
		t.Error("A did not receive MouseLeave")
	}

	// Check B got MouseEnter.
	hasEnter := false
	for _, m := range b.mouse {
		if m.Action == widgets.MouseEnter {
			hasEnter = true
		}
	}
	if !hasEnter {
		t.Error("B did not receive MouseEnter")
	}
}

func TestDispatchMouseMiss(t *testing.T) {
	t.Parallel()

	r := New()
	a := newFocusable(0, 0, 10, 3)
	r.Register(a)

	consumed := r.DispatchMouse(widgets.MouseEvent{Button: widgets.MouseLeft, Action: widgets.MousePress, X: 50, Y: 50})
	if consumed {
		t.Error("miss should not consume")
	}
}

// ---------------------------------------------------------------------------
// FocusNext / FocusPrev
// ---------------------------------------------------------------------------

func TestFocusNextPrev(t *testing.T) {
	t.Parallel()

	r := New()
	a := newFocusable(0, 0, 10, 1)
	b := newFocusable(0, 1, 10, 1)
	txt := &focusableNode{rect: layout.Rect{X: 0, Y: 2, Width: 10, Height: 1}, focus: false} // not focusable
	r.Register(a)
	r.Register(b)
	r.Register(txt)

	r.FocusNext()
	if r.Focused() != a {
		t.Fatalf("FocusNext #1: got %v, want a", r.Focused())
	}

	r.FocusNext()
	if r.Focused() != b {
		t.Fatalf("FocusNext #2: got %v, want b", r.Focused())
	}

	// Should skip non-focusable txt and wrap to a.
	r.FocusNext()
	if r.Focused() != a {
		t.Fatalf("FocusNext wrap: got %v, want a", r.Focused())
	}

	r.FocusPrev()
	if r.Focused() != b {
		t.Fatalf("FocusPrev: got %v, want b", r.Focused())
	}
}

func TestDispatchKeyTabCyclesFocus(t *testing.T) {
	t.Parallel()

	r := New()
	a := newFocusable(0, 0, 10, 1)
	b := newFocusable(0, 1, 10, 1)
	r.Register(a)
	r.Register(b)
	r.Focus(a)

	// Tab → b
	consumed := r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyTab})
	if !consumed {
		t.Error("Tab should be consumed by router")
	}
	if r.Focused() != b {
		t.Fatalf("after Tab: focused %v, want b", r.Focused())
	}

	// Backtab → a
	consumed = r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyBacktab})
	if !consumed {
		t.Error("Backtab should be consumed by router")
	}
	if r.Focused() != a {
		t.Fatalf("after Backtab: focused %v, want a", r.Focused())
	}
}

// ---------------------------------------------------------------------------
// Benchmark
// ---------------------------------------------------------------------------

func BenchmarkDispatchMouse(b *testing.B) {
	r := New()
	// 10 widgets in a row.
	for i := 0; i < 10; i++ {
		r.Register(newFocusable(i*10, 0, 8, 3))
	}
	mouse := widgets.MouseEvent{Button: widgets.MouseLeft, Action: widgets.MousePress, X: 25, Y: 1}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.DispatchMouse(mouse)
	}
}
