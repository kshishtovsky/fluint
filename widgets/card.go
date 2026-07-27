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

	// ── Shadow (outside rect, after everything else) ─────────────
	// Drawn last so it never interferes with border or content.
	// Half-blocks sit in cells adjacent to the border:
	//   ▐ at X+W — left half colored, adjacent to │ at X+W-1
	//   ▄ at Y+H — bottom half colored, adjacent to ─ at Y+H-1
	//   ▒ at (X+W, Y+H) — corner join
	if c.config.style.HasShadow() {
		c.drawShadow(ctx, sx, sy, rect.Width, rect.Height)
	}
}

// drawShadow draws shadow OUTSIDE the card rect in cells adjacent to
// the border. Half-block runes create a sub-cell illusion — ▐'s left
// half is colored (touching the border), right half is transparent.
// The border is never touched.
//
// Visual:
//
//	╭───────────────╮▐  ← ▐ at col X+W, adjacent to ╮ at X+W-1
//	│    CARD       │▐
//	╰───────────────╯▐
//	 ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▒  ← ▄ at row Y+H, adjacent to ─ at Y+H-1; ▒ corner
func (c *Card) drawShadow(ctx viewport.RenderCtx, sx, sy, w, h int) {
	sh := c.config.style.ShadowCfg()
	switch sh.Mode {
	case style.ShadowModeDither:
		c.drawDitherShadow(ctx, sx, sy, w, h, sh)
	default:
		c.drawSubCellShadow(ctx, sx, sy, w, h, sh)
	}
}

// drawSubCellShadow draws a 1-cell shadow using half-block runes (▐, ▄, ▒).
func (c *Card) drawSubCellShadow(ctx viewport.RenderCtx, sx, sy, w, h int, sh style.ShadowStyle) {
	fg := uint32(sh.Color)
	bufW := ctx.Buf.Width
	bufH := ctx.Buf.Height

	if sh.OffsetX > 0 {
		x := sx + w
		if x >= 0 && x < bufW {
			for y := sy; y < sy+h; y++ {
				if y < 0 || y >= bufH {
					continue
				}
				old := ctx.Buf.GetCell(x, y)
				ctx.Buf.SetCell(x, y, buffer.Cell{
					Rune: style.ShadowRight,
					Fg:   fg,
					Bg:   old.Bg,
				})
			}
		}
	}

	if sh.OffsetY > 0 {
		y := sy + h
		if y >= 0 && y < bufH {
			for x := sx; x < sx+w; x++ {
				if x < 0 || x >= bufW {
					continue
				}
				old := ctx.Buf.GetCell(x, y)
				ctx.Buf.SetCell(x, y, buffer.Cell{
					Rune: style.ShadowBottom,
					Fg:   fg,
					Bg:   old.Bg,
				})
			}
		}
	}

	if sh.OffsetX > 0 && sh.OffsetY > 0 {
		cx, cy := sx+w, sy+h
		if cx >= 0 && cx < bufW && cy >= 0 && cy < bufH {
			old := ctx.Buf.GetCell(cx, cy)
			ctx.Buf.SetCell(cx, cy, buffer.Cell{
				Rune: style.ShadowCorner,
				Fg:   fg,
				Bg:   old.Bg,
			})
		}
	}
}

// drawDitherShadow draws a multi-cell gradient shadow using ASCII
// density characters. Dense near the card (#, @), sparse at the edge
// (., :). Distance is Manhattan for corners, axis-aligned for sides.
//
// Visual (blur=3):
//
//	╭──────────╮%=:.
//	│   CARD   │%=:.
//	╰──────────╯%=:.
//	 %*+=-:.         ← density fades left-to-right
//	  =-:.
//	   :.
func (c *Card) drawDitherShadow(ctx viewport.RenderCtx, sx, sy, w, h int, sh style.ShadowStyle) {
	fg := uint32(sh.Color)
	bufW := ctx.Buf.Width
	bufH := ctx.Buf.Height
	ramp := style.ShadowDensityRamp
	rampLen := len(ramp)

	blurX := sh.OffsetX
	blurY := sh.OffsetY

	cardRight := sx + w
	cardBottom := sy + h

	for dy := 0; dy < blurY; dy++ {
		for dx := 0; dx < blurX; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			// Distance from card edge. Manhattan for smooth corner fade.
			dist := dx + dy
			maxDist := blurX + blurY - 2
			if maxDist < 1 {
				maxDist = 1
			}

			// Normalise to ramp index (closer = higher index).
			idx := rampLen - 1 - (dist*rampLen)/maxDist
			if idx < 0 {
				idx = 0
			}
			if idx >= rampLen {
				idx = rampLen - 1
			}
			ch := ramp[idx]
			if ch == ' ' {
				continue
			}

			cell := buffer.Cell{Rune: ch, Fg: fg}

			// Bottom-right corner area.
			if dx > 0 && dy > 0 {
				cx, cy := cardRight+dx-1, cardBottom+dy-1
				if cx >= 0 && cx < bufW && cy >= 0 && cy < bufH {
					old := ctx.Buf.GetCell(cx, cy)
					cell.Bg = old.Bg
					ctx.Buf.SetCell(cx, cy, cell)
				}
			}

			// Right strip: column cardRight+dx-1, rows cardBottom-1 upward.
			if dx > 0 && dy == 0 {
				x := cardRight + dx - 1
				if x >= 0 && x < bufW {
					for y := sy; y < cardBottom; y++ {
						if y < 0 || y >= bufH {
							continue
						}
						old := ctx.Buf.GetCell(x, y)
						cell.Bg = old.Bg
						ctx.Buf.SetCell(x, y, cell)
					}
				}
			}

			// Bottom strip: row cardBottom+dy-1, columns cardRight-1 leftward.
			if dy > 0 && dx == 0 {
				y := cardBottom + dy - 1
				if y >= 0 && y < bufH {
					for x := sx; x < cardRight; x++ {
						if x < 0 || x >= bufW {
							continue
						}
						old := ctx.Buf.GetCell(x, y)
						cell.Bg = old.Bg
						ctx.Buf.SetCell(x, y, cell)
					}
				}
			}
		}
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
