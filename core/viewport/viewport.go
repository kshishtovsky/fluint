// Package viewport provides a camera/viewport for rendering content that
// exceeds the physical terminal size. It translates between world
// coordinates (the virtual canvas) and screen coordinates (the terminal).
//
// Widgets are placed in world coordinates. The viewport determines which
// portion of the world is visible on screen. Widgets outside the viewport
// are culled (skipped entirely). Widgets partially inside are clipped.
package viewport

import "github.com/kshishtovsky/fluint/core/buffer"

// Viewport defines the visible region of a virtual canvas.
type Viewport struct {
	Width   int // Physical width of the visible area (terminal columns).
	Height  int // Physical height of the visible area (terminal rows).
	OffsetX int // World X coordinate of the top-left visible cell.
	OffsetY int // World Y coordinate of the top-left visible cell.
}

// New creates a Viewport with the given physical dimensions and zero offset.
func New(width, height int) *Viewport {
	return &Viewport{Width: width, Height: height}
}

// Resize updates the physical dimensions of the viewport.
func (v *Viewport) Resize(width, height int) {
	v.Width = width
	v.Height = height
}

// Scroll shifts the viewport offset by (dx, dy). Negative offsets are
// clamped to zero.
func (v *Viewport) Scroll(dx, dy int) {
	v.OffsetX += dx
	v.OffsetY += dy
	if v.OffsetX < 0 {
		v.OffsetX = 0
	}
	if v.OffsetY < 0 {
		v.OffsetY = 0
	}
}

// Center positions the viewport so that world coordinate (x, y) is at
// the centre of the visible area. Negative offsets are clamped to zero.
func (v *Viewport) Center(x, y int) {
	v.OffsetX = x - v.Width/2
	v.OffsetY = y - v.Height/2
	if v.OffsetX < 0 {
		v.OffsetX = 0
	}
	if v.OffsetY < 0 {
		v.OffsetY = 0
	}
}

// Visible reports whether a world-space rectangle intersects the
// viewport's visible area. Used for culling.
func (v *Viewport) Visible(wx, wy, ww, wh int) bool {
	return wx < v.OffsetX+v.Width &&
		wx+ww > v.OffsetX &&
		wy < v.OffsetY+v.Height &&
		wy+wh > v.OffsetY
}

// ScreenX converts a world X coordinate to a screen X coordinate.
func (v *Viewport) ScreenX(wx int) int {
	return wx - v.OffsetX
}

// ScreenY converts a world Y coordinate to a screen Y coordinate.
func (v *Viewport) ScreenY(wy int) int {
	return wy - v.OffsetY
}

// RenderCtx is the rendering context passed to widget Render methods.
// It bundles the buffer with an optional viewport. When View is nil,
// rendering operates in screen-space with no offset (legacy mode).
type RenderCtx struct {
	Buf  *buffer.Buffer
	View *Viewport
}
