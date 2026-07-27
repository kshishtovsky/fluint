package style

import (
	"testing"

	"github.com/kshishtovsky/fluint/core/buffer"
)

func TestNewReturnsZeroValue(t *testing.T) {
	t.Parallel()
	s := New()
	if s.fg != Default {
		t.Errorf("fg: got 0x%06X, want Default", s.fg)
	}
	if s.bg != Default {
		t.Errorf("bg: got 0x%06X, want Default", s.bg)
	}
	if s.attrs != 0 {
		t.Errorf("attrs: got %d, want 0", s.attrs)
	}
}

func TestChainingReturnsNewValues(t *testing.T) {
	t.Parallel()

	base := New()
	red := base.Foreground(Red)
	boldRed := red.Bold()

	// base unchanged
	if base.fg != Default {
		t.Error("base.Foreground mutated receiver")
	}
	// red has color but no bold
	if red.fg != Red {
		t.Errorf("red.fg: got 0x%06X, want Red", red.fg)
	}
	if red.attrs != 0 {
		t.Error("red.attrs should be 0")
	}
	// boldRed has both
	if boldRed.fg != Red {
		t.Errorf("boldRed.fg: got 0x%06X, want Red", boldRed.fg)
	}
	if boldRed.attrs&buffer.Bold == 0 {
		t.Error("boldRed should have Bold")
	}
}

func TestApplySetsColorsAndAttrs(t *testing.T) {
	t.Parallel()

	s := New().Foreground(White).Background(Blue).Bold().Underline()
	cell := s.Apply(buffer.Cell{Rune: 'X'})

	if cell.Fg != uint32(White) {
		t.Errorf("Fg: got 0x%06X, want 0x%06X", cell.Fg, White)
	}
	if cell.Bg != uint32(Blue) {
		t.Errorf("Bg: got 0x%06X, want 0x%06X", cell.Bg, Blue)
	}
	if cell.Attrs&buffer.Bold == 0 {
		t.Error("Bold not set")
	}
	if cell.Attrs&buffer.Underline == 0 {
		t.Error("Underline not set")
	}
	if cell.Rune != 'X' {
		t.Errorf("Rune changed: got %q, want 'X'", cell.Rune)
	}
}

func TestApplyDefaultPreservesCellColors(t *testing.T) {
	t.Parallel()

	s := New() // zero style
	cell := buffer.Cell{Rune: 'A', Fg: 0x111111, Bg: 0x222222}
	got := s.Apply(cell)

	if got.Fg != 0x111111 {
		t.Errorf("Fg overwritten: got 0x%06X, want 0x111111", got.Fg)
	}
	if got.Bg != 0x222222 {
		t.Errorf("Bg overwritten: got 0x%06X, want 0x222222", got.Bg)
	}
}

func TestApplyPartialStyle(t *testing.T) {
	t.Parallel()

	// Only set foreground, leave background as default.
	s := New().Foreground(Red)
	cell := buffer.Cell{Rune: 'Z', Bg: 0x333333}
	got := s.Apply(cell)

	if got.Fg != uint32(Red) {
		t.Errorf("Fg: got 0x%06X, want Red", got.Fg)
	}
	if got.Bg != 0x333333 {
		t.Errorf("Bg: got 0x%06X, want 0x333333 (preserved)", got.Bg)
	}
}

func TestAccessors(t *testing.T) {
	t.Parallel()

	s := New().Foreground(Cyan).Background(DarkGray).Italic().Dim()
	if s.FG() != Cyan {
		t.Errorf("FG(): got 0x%06X, want Cyan", s.FG())
	}
	if s.BG() != DarkGray {
		t.Errorf("BG(): got 0x%06X, want DarkGray", s.BG())
	}
	if s.Attrs()&buffer.Italic == 0 {
		t.Error("Italic not set")
	}
	if s.Attrs()&buffer.Dim == 0 {
		t.Error("Dim not set")
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkStyleChain(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New().Foreground(Red).Background(Black).Bold()
	}
}

func BenchmarkApply(b *testing.B) {
	s := New().Foreground(White).Background(Blue).Bold()
	cell := buffer.Cell{Rune: 'X'}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Apply(cell)
	}
}
