// Package router provides event routing from the terminal input layer
// to individual widgets. It manages keyboard focus and mouse hit-testing
// so that events reach the correct widget.
//
// Usage:
//
//	r := router.New()
//	r.Register(buttonA)
//	r.Register(buttonB)
//	r.FocusNext()          // focus buttonA
//	r.DispatchKey(keyEvent) // delivered to buttonA
//	r.DispatchMouse(mouseEvent) // hit-tested against all widgets
package router

import (
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/widgets"
)

// Router dispatches keyboard and mouse events to registered widgets.
type Router struct {
	nodes   []widgets.Node
	focused widgets.Node
	hovered widgets.Node
}

// New creates an empty Router.
func New() *Router {
	return &Router{}
}

// Register adds a widget to the router. Duplicate registrations are
// ignored. Only focusable widgets participate in tab cycling.
func (r *Router) Register(node widgets.Node) {
	for _, n := range r.nodes {
		if n == node {
			return
		}
	}
	r.nodes = append(r.nodes, node)
}

// Unregister removes a widget from the router. If the widget is
// currently focused or hovered, that state is cleared.
func (r *Router) Unregister(node widgets.Node) {
	for i, n := range r.nodes {
		if n == node {
			r.nodes = append(r.nodes[:i], r.nodes[i+1:]...)
			break
		}
	}
	if r.focused == node {
		r.focused = nil
	}
	if r.hovered == node {
		r.hovered = nil
	}
}

// DispatchKey sends a keyboard event to the currently focused widget.
// Returns true if the event was consumed. Tab and Backtab cycle focus
// before dispatch.
func (r *Router) DispatchKey(key widgets.KeyEvent) bool {
	if key.Code == widgets.KeyTab {
		r.FocusNext()
		return true
	}
	if key.Code == widgets.KeyBacktab {
		r.FocusPrev()
		return true
	}
	if r.focused != nil {
		return r.focused.OnKey(key)
	}
	return false
}

// DispatchMouse sends a mouse event to the topmost widget whose rect
// contains the cursor. Widgets are checked in reverse registration
// order (last registered = highest Z). On enter/leave transitions the
// previous hovered widget receives MouseLeave and the new one receives
// MouseEnter before the actual event. Returns true if the event was
// consumed.
func (r *Router) DispatchMouse(mouse widgets.MouseEvent) bool {
	// Hit-test in reverse order (topmost first).
	var hit widgets.Node
	for i := len(r.nodes) - 1; i >= 0; i-- {
		if widgets.HitTest(r.nodes[i].Geometry(), mouse.X, mouse.Y) {
			hit = r.nodes[i]
			break
		}
	}

	// Handle enter/leave transitions.
	if hit != r.hovered {
		if r.hovered != nil {
			r.hovered.OnMouse(widgets.MouseEvent{
				Action: widgets.MouseLeave,
				X:      mouse.X,
				Y:      mouse.Y,
			})
		}
		if hit != nil {
			hit.OnMouse(widgets.MouseEvent{
				Action: widgets.MouseEnter,
				X:      mouse.X,
				Y:      mouse.Y,
			})
		}
		r.hovered = hit
	}

	// Dispatch the actual event.
	if hit != nil {
		return hit.OnMouse(mouse)
	}
	return false
}

// DispatchMouseViewport is like DispatchMouse but translates screen-space
// mouse coordinates to world-space using the viewport's offset before
// hit-testing. This allows widgets placed at world coordinates (e.g. 50,50)
// to receive clicks at screen position (0,0) when the viewport is scrolled.
func (r *Router) DispatchMouseViewport(mouse widgets.MouseEvent, view *viewport.Viewport) bool {
	world := mouse
	world.X += view.OffsetX
	world.Y += view.OffsetY
	return r.DispatchMouse(world)
}

// Focus sets focus on the given widget. Pass nil to clear focus.
func (r *Router) Focus(node widgets.Node) {
	r.focused = node
}

// Focused returns the currently focused widget, or nil.
func (r *Router) Focused() widgets.Node {
	return r.focused
}

// Hovered returns the widget the mouse is currently over, or nil.
func (r *Router) Hovered() widgets.Node {
	return r.hovered
}

// FocusNext moves focus to the next focusable widget in registration
// order. Wraps around to the first. If nothing is focused, focuses
// the first focusable widget.
func (r *Router) FocusNext() {
	r.focusStep(1)
}

// FocusPrev moves focus to the previous focusable widget in
// registration order. Wraps around to the last.
func (r *Router) FocusPrev() {
	r.focusStep(-1)
}

// Nodes returns the registered widgets (read-only snapshot).
func (r *Router) Nodes() []widgets.Node {
	return r.nodes
}

// focusStep moves focus by delta (+1 or -1) through focusable widgets.
func (r *Router) focusStep(delta int) {
	n := len(r.nodes)
	if n == 0 {
		return
	}

	// Build the list of focusable indices.
	focusable := make([]int, 0, n)
	for i, node := range r.nodes {
		if node.Focusable() {
			focusable = append(focusable, i)
		}
	}
	if len(focusable) == 0 {
		return
	}

	// Find current focus position in the focusable list.
	currentIdx := -1
	for fi, idx := range focusable {
		if r.nodes[idx] == r.focused {
			currentIdx = fi
			break
		}
	}

	// Advance.
	nextIdx := currentIdx + delta
	if nextIdx < 0 {
		nextIdx = len(focusable) - 1
	} else if nextIdx >= len(focusable) {
		nextIdx = 0
	}

	r.focused = r.nodes[focusable[nextIdx]]
}
