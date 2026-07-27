package widgets

import (
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/style"
)

// Card is a container widget that draws a border (solid or rounded),
// optional shadow, and interior padding around a child widget.
//
// Visual structure:
//
//	┌─────────────────┐
//	│  ┌───────────┐  │
//	│  │   child    │  │
//	│  └───────────┘  │
//	└─────────────────┘
//	  ░░░░░░░░░░░░░░░
//
// The outer border is 1 cell thick. Padding adds space between the
// border and the child. The shadow is drawn at (OffsetX, OffsetY)
// behind the card.
type Card struct {
	child  Node
	config Config
	rect   layout.Rect
}

// NewCard creates a Card wrapping the given child widget.
func NewCard(child Node, opts ...Option) *Card {
	return &Card{
		child:  child,
		config: newConfig(opts),
	}
}

// Render draws the card: shadow → background fill → border → child.
func (c *Card) Render(ctx viewport.RenderCtx, rect layout.Rect) {
	c.rect = rect

	if !Visible(ctx.View, rect.X, rect.Y, rect.Width, rect.Height) {
		return
	}

	sx, sy := Screen(ctx.View, rect.X, rect.Y)

	// ── Background fill ─────────────────────────────────────────────
	bgCell := c.config.style.Apply(buffer.Cell{Rune: ' '})
	for y := sy; y < sy+rect.Height; y++ {
		for x := sx; x < sx+rect.Width; x++ {
			ctx.Buf.SetCell(x, y, bgCell)
		}
	}

	// ── Shadow (only in areas NOT covered by the card) ──────────────
	// Uses half-block runes for a thinner, subtler shadow:
	//   ▀ (upper half block) — bottom shadow: colored top half, bg bottom half
	//   ▌ (left half block)  — right shadow: colored left half, bg right half
	if c.config.style.HasShadow() {
		sh := c.config.style.ShadowCfg()

		// Bottom strip: ▀ with shadow Fg, page Bg.
		if sh.OffsetY > 0 {
			bottomCell := buffer.Cell{Rune: style.ShadowBottom, Fg: uint32(sh.Color)}
			for y := sy + rect.Height; y < sy+rect.Height+sh.OffsetY; y++ {
				for x := sx + sh.OffsetX; x < sx+rect.Width+sh.OffsetX; x++ {
					ctx.Buf.SetCell(x, y, bottomCell)
				}
			}
		}

		// Right strip: ▌ with shadow Fg, page Bg.
		if sh.OffsetX > 0 {
			rightCell := buffer.Cell{Rune: style.ShadowRight, Fg: uint32(sh.Color)}
			for y := sy + sh.OffsetY; y < sy+rect.Height+sh.OffsetY; y++ {
				for x := sx + rect.Width; x < sx+rect.Width+sh.OffsetX; x++ {
					ctx.Buf.SetCell(x, y, rightCell)
				}
			}
		}
	}

	// ── Border ──────────────────────────────────────────────────────
	if c.config.style.HasBorder() {
		c.drawBorder(ctx, sx, sy, rect.Width, rect.Height)
	}

	// ── Child ───────────────────────────────────────────────────────
	// Compute inner rect: subtract border (1 cell each side) + padding.
	innerX := rect.X + c.borderWidth()
	innerY := rect.Y + c.borderWidth()
	innerW := rect.Width - 2*c.borderWidth()
	innerH := rect.Height - 2*c.borderWidth()

	px := c.config.style.PaddingX()
	py := c.config.style.PaddingY()
	innerX += px
	innerY += py
	innerW -= 2 * px
	innerH -= 2 * py

	if innerW > 0 && innerH > 0 && c.child != nil {
		c.child.Render(ctx, layout.Rect{
			X:      innerX,
			Y:      innerY,
			Width:  innerW,
			Height: innerH,
		})
	}
}

// drawBorder draws the border characters around the card area.
func (c *Card) drawBorder(ctx viewport.RenderCtx, sx, sy, w, h int) {
	bs := c.config.style.Border()
	if bs == style.BorderNone || w < 2 || h < 2 {
		return
	}

	bc := c.config.style.BorderColor()
	borderStyle := c.config.style.Foreground(bc)
	_ = borderStyle

	var tl, tr, bl, br rune
	var hLine, vLine rune

	switch bs {
	case style.BorderRounded:
		tl, tr, bl, br = style.BorderRoundedTL, style.BorderRoundedTR, style.BorderRoundedBL, style.BorderRoundedBR
		hLine, vLine = style.BorderRoundedH, style.BorderRoundedV
	default: // BorderSolid
		tl, tr, bl, br = style.BorderSolidTL, style.BorderSolidTR, style.BorderSolidBL, style.BorderSolidBR
		hLine, vLine = style.BorderSolidH, style.BorderSolidV
	}

	borderCell := buffer.Cell{Fg: uint32(bc), Bg: uint32(c.config.style.BG())}

	// Top row.
	borderCell.Rune = tl
	ctx.Buf.SetCell(sx, sy, borderCell)
	for x := sx + 1; x < sx+w-1; x++ {
		borderCell.Rune = hLine
		ctx.Buf.SetCell(x, sy, borderCell)
	}
	borderCell.Rune = tr
	ctx.Buf.SetCell(sx+w-1, sy, borderCell)

	// Bottom row.
	borderCell.Rune = bl
	ctx.Buf.SetCell(sx, sy+h-1, borderCell)
	for x := sx + 1; x < sx+w-1; x++ {
		borderCell.Rune = hLine
		ctx.Buf.SetCell(x, sy+h-1, borderCell)
	}
	borderCell.Rune = br
	ctx.Buf.SetCell(sx+w-1, sy+h-1, borderCell)

	// Left and right columns (between corners).
	borderCell.Rune = vLine
	for y := sy + 1; y < sy+h-1; y++ {
		ctx.Buf.SetCell(sx, y, borderCell)
		ctx.Buf.SetCell(sx+w-1, y, borderCell)
	}
}

// borderWidth returns 1 if the card has a border, 0 otherwise.
func (c *Card) borderWidth() int {
	if c.config.style.HasBorder() {
		return 1
	}
	return 0
}

// Geometry returns the widget's current position and size.
func (c *Card) Geometry() layout.Rect { return c.rect }

// SetGeometry updates the widget's position and size.
func (c *Card) SetGeometry(rect layout.Rect) { c.rect = rect }

// OnKey forwards keyboard events to the child if one is set.
func (c *Card) OnKey(key KeyEvent) bool {
	if c.child != nil {
		return c.child.OnKey(key)
	}
	return false
}

// OnMouse forwards mouse events to the child if the click is inside
// the inner area (accounting for border and padding).
func (c *Card) OnMouse(mouse MouseEvent) bool {
	if c.child == nil {
		return false
	}
	innerX := c.rect.X + c.borderWidth() + c.config.style.PaddingX()
	innerY := c.rect.Y + c.borderWidth() + c.config.style.PaddingY()
	innerW := c.rect.Width - 2*c.borderWidth() - 2*c.config.style.PaddingX()
	innerH := c.rect.Height - 2*c.borderWidth() - 2*c.config.style.PaddingY()

	inner := layout.Rect{X: innerX, Y: innerY, Width: innerW, Height: innerH}
	if HitTest(inner, mouse.X, mouse.Y) {
		return c.child.OnMouse(mouse)
	}
	return false
}

// Focusable returns the child's focusability, or false if no child.
func (c *Card) Focusable() bool {
	if c.child != nil {
		return c.child.Focusable()
	}
	return false
}
