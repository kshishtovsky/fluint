package buffer

import "testing"

// TestSetSubCellY_TopOnly paints the upper half of an empty cell.
// Expected: glyph = ▀, Fg = colour, Bg = 0.
func TestSetSubCellY_TopOnly(t *testing.T) {
	t.Parallel()

	b := NewBuffer(2, 2)
	const fg = 0x00FF8800

	b.SetSubCellY(1, 0, 0, fg)

	got := b.GetCell(1, 0)
	if got.Rune != '\u2580' {
		t.Errorf("Rune = %q (U+%04X), want %q (U+2580)", got.Rune, got.Rune, '\u2580')
	}
	if got.Fg != fg {
		t.Errorf("Fg = %#X, want %#X", got.Fg, fg)
	}
	if got.Bg != 0 {
		t.Errorf("Bg = %#X, want 0", got.Bg)
	}
}

// TestSetSubCellY_BottomOnly paints the lower half first.
// Expected: glyph = ▄, Fg = 0, Bg = colour.
func TestSetSubCellY_BottomOnly(t *testing.T) {
	t.Parallel()

	b := NewBuffer(2, 2)
	const bg = 0x0000AAFF

	b.SetSubCellY(0, 0, 1, bg)

	got := b.GetCell(0, 0)
	if got.Rune != '\u2584' {
		t.Errorf("Rune = %q (U+%04X), want %q (U+2584)", got.Rune, got.Rune, '\u2584')
	}
	if got.Fg != 0 {
		t.Errorf("Fg = %#X, want 0", got.Fg)
	}
	if got.Bg != bg {
		t.Errorf("Bg = %#X, want %#X", got.Bg, bg)
	}
}

// TestSetSubCellY_BothHalves verifies that painting top then bottom
// promotes the cell to ▀ with the correct Fg/Bg split.
func TestSetSubCellY_BothHalves(t *testing.T) {
	t.Parallel()

	b := NewBuffer(1, 1)
	const (
		fg = 0x00FF0000
		bg = 0x000000FF
	)

	b.SetSubCellY(0, 0, 0, fg) // top
	b.SetSubCellY(0, 0, 1, bg) // bottom

	got := b.GetCell(0, 0)
	if got.Rune != '\u2580' {
		t.Errorf("Rune = %q (U+%04X), want %q (U+2580)", got.Rune, got.Rune, '\u2580')
	}
	if got.Fg != fg {
		t.Errorf("Fg = %#X, want %#X", got.Fg, fg)
	}
	if got.Bg != bg {
		t.Errorf("Bg = %#X, want %#X", got.Bg, bg)
	}
}

// TestSetSubCellY_BottomThenTop verifies the reverse order: bottom
// first, then top. The cell should still end up as ▀.
func TestSetSubCellY_BottomThenTop(t *testing.T) {
	t.Parallel()

	b := NewBuffer(1, 1)
	const (
		fg = 0x00123456
		bg = 0x00ABCDEF
	)

	b.SetSubCellY(0, 0, 1, bg)
	b.SetSubCellY(0, 0, 0, fg)

	got := b.GetCell(0, 0)
	if got.Rune != '\u2580' {
		t.Errorf("Rune = %q (U+%04X), want %q (U+2580)", got.Rune, got.Rune, '\u2580')
	}
	if got.Fg != fg {
		t.Errorf("Fg = %#X, want %#X", got.Fg, fg)
	}
	if got.Bg != bg {
		t.Errorf("Bg = %#X, want %#X", got.Bg, bg)
	}
}

// TestSetSubCellY_OutOfRange confirms the safe-clipping contract:
// coordinates outside the buffer and an invalid ySub are no-ops.
func TestSetSubCellY_OutOfRange(t *testing.T) {
	t.Parallel()

	b := NewBuffer(1, 1)
	original := b.GetCell(0, 0)

	cases := []struct {
		name       string
		x, y, ySub int
	}{
		{"x-negative", -1, 0, 0},
		{"x-too-large", 1, 0, 0},
		{"y-negative", 0, -1, 0},
		{"y-too-large", 0, 1, 0},
		{"ySub-invalid", 0, 0, 2},
		{"ySub-negative", 0, 0, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b.SetSubCellY(tc.x, tc.y, tc.ySub, 0x00DEADBE)
			got := b.GetCell(0, 0)
			if got != original {
				t.Errorf("cell mutated after out-of-range SetSubCellY: got %+v, want %+v", got, original)
			}
		})
	}
}

// TestSetSubCellY_NilReceiver confirms a nil Buffer is a safe no-op.
func TestSetSubCellY_NilReceiver(t *testing.T) {
	t.Parallel()

	var b *Buffer
	// Should not panic.
	b.SetSubCellY(0, 0, 0, 0x00FF00FF)
	b.SetSubCellY(0, 0, 1, 0x00FF00FF)
}

// BenchmarkSetSubCellY_Hot measures the per-frame cost of painting
// two half-block rows into a full-screen-sized buffer.
func BenchmarkSetSubCellY_Hot(b *testing.B) {
	buf := NewBuffer(80, 24)
	var color uint32 = 0x00C0FFEE
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := 0; y < buf.Height; y++ {
			for x := 0; x < buf.Width; x++ {
				buf.SetSubCellY(x, y, y&1, color)
			}
		}
	}
}
