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
// Border, Padding, Shadow
// ---------------------------------------------------------------------------

func TestBorder(t *testing.T) {
	t.Parallel()

	s := New().SolidBorder(Red)
	if s.Border() != BorderSolid {
		t.Errorf("Border: got %d, want BorderSolid", s.Border())
	}
	if s.BorderColor() != Red {
		t.Errorf("BorderColor: got 0x%06X, want Red", s.BorderColor())
	}
	if !s.HasBorder() {
		t.Error("HasBorder should be true")
	}

	s2 := s.RoundedBorder(Blue)
	if s2.Border() != BorderRounded {
		t.Errorf("RoundedBorder: got %d, want BorderRounded", s2.Border())
	}
	// Original unchanged.
	if s.Border() != BorderSolid {
		t.Error("RoundedBorder mutated receiver")
	}

	s3 := s2.NoBorder()
	if s3.HasBorder() {
		t.Error("NoBorder should clear border")
	}
}

func TestPadding(t *testing.T) {
	t.Parallel()

	s := New().Padding(2, 1)
	if s.PaddingX() != 2 {
		t.Errorf("PaddingX: got %d, want 2", s.PaddingX())
	}
	if s.PaddingY() != 1 {
		t.Errorf("PaddingY: got %d, want 1", s.PaddingY())
	}
}

func TestShadow(t *testing.T) {
	t.Parallel()

	s := New().Shadow(1, 2, ShadowColor)
	if !s.HasShadow() {
		t.Error("HasShadow should be true")
	}
	sh := s.ShadowCfg()
	if sh.OffsetX != 1 || sh.OffsetY != 2 || sh.Color != ShadowColor {
		t.Errorf("ShadowCfg: got %+v", sh)
	}

	s2 := s.NoShadow()
	if s2.HasShadow() {
		t.Error("NoShadow should clear shadow")
	}
}

func TestBorderRunes(t *testing.T) {
	t.Parallel()

	// Verify rune constants are correct.
	if BorderSolidTL != '┌' {
		t.Errorf("BorderSolidTL: got %c, want ┌", BorderSolidTL)
	}
	if BorderRoundedTL != '╭' {
		t.Errorf("BorderRoundedTL: got %c, want ╭", BorderRoundedTL)
	}
	if BorderRoundedBR != '╯' {
		t.Errorf("BorderRoundedBR: got %c, want ╯", BorderRoundedBR)
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

func BenchmarkStyleChainWithBorder(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = New().Foreground(White).Background(Black).RoundedBorder(Green).Padding(1, 0).Shadow(1, 1, ShadowColor)
	}
}
