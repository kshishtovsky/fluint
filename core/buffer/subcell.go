package buffer

// Half-block glyphs used for sub-cell vertical resolution.
//
//	▀  U+2580  Upper half block — top half painted (Fg), bottom is Bg.
//	▄  U+2584  Lower half block — bottom half painted (Fg), top is Bg.
//
// SetSubCellY paints "subpixels" along the Y axis by stacking two
// logical rows into one terminal row using these glyphs, giving the
// renderer a 2× vertical resolution at zero allocation cost.
const (
	runeUpperHalfBlock = '\u2580' // ▀  : top half filled, bottom is Bg.
	runeLowerHalfBlock = '\u2584' // ▄  : bottom half filled, top is Bg.
)

// SetSubCellY paints a single sub-row of the cell at (x, y).
//
//	ySub == 0 paints the upper half — the colour is stored in Cell.Fg
//	          and the glyph is forced to ▀.
//	ySub == 1 paints the lower half — the colour is stored in Cell.Bg
//	          and the glyph is forced to ▀ (or ▄ when Fg is unset).
//
// Combination semantics:
//   - If only the top half was painted: Fg holds the colour, glyph is ▀,
//     Bg stays at its zero value (terminal default background).
//   - If only the bottom half was painted: Bg holds the colour, glyph is
//     ▄ so the painted half ends up on the bottom.
//   - If both halves are painted (regardless of order): glyph is ▀ with
//     Fg = top colour, Bg = bottom colour.
//
// Coordinates outside the buffer and ySub values other than 0 or 1 are
// safely clipped. A nil receiver is a no-op so callers do not need to
// guard before drawing.
func (b *Buffer) SetSubCellY(x, y, ySub int, color uint32) {
	if b == nil {
		return
	}
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}
	if ySub != 0 && ySub != 1 {
		return
	}

	cell := &b.Cells[y*b.Width+x]

	switch ySub {
	case 0:
		// Top half → Cell.Fg. The glyph is always ▀ regardless of
		// whether the bottom half was painted before or after.
		cell.Fg = color
		cell.Rune = runeUpperHalfBlock
	case 1:
		// Bottom half → Cell.Bg.
		cell.Bg = color
		// If no top half is painted yet (Fg still default), the cell
		// represents the bottom half only → use ▄. Otherwise keep ▀.
		if cell.Fg == 0 {
			cell.Rune = runeLowerHalfBlock
		} else {
			cell.Rune = runeUpperHalfBlock
		}
	}
}
