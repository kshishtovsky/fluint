package ansi

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/diff"
)

func TestWriteInt(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{123, "123"},
		{2026, "2026"},
		{-45, "-45"},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		writeInt(&buf, tt.input)
		if got := buf.String(); got != tt.want {
			t.Errorf("writeInt(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRender_SingleCellRed(t *testing.T) {
	renderer := NewRenderer()

	changes := []diff.Change{
		{
			X: 0,
			Y: 0,
			Cell: buffer.Cell{
				Rune: 'A',
				Fg:   0xFF0000, // Red (255, 0, 0)
			},
		},
	}

	out := string(renderer.Render(changes))

	if !strings.HasPrefix(out, Mode2026Start) {
		t.Fatalf("expected Mode 2026 start prefix, got %q", out)
	}
	if !strings.HasSuffix(out, Mode2026End) {
		t.Fatalf("expected Mode 2026 end suffix, got %q", out)
	}

	// Should contain cursor position 1;1H
	if !strings.Contains(out, "\x1b[1;1H") {
		t.Errorf("missing cursor move \\x1b[1;1H in output: %q", out)
	}

	// Should contain Red TrueColor SGR 38;2;255;0;0
	if !strings.Contains(out, "38;2;255;0;0") {
		t.Errorf("missing TrueColor red SGR in output: %q", out)
	}

	// Should contain rune 'A'
	if !strings.Contains(out, "A") {
		t.Errorf("missing rune 'A' in output: %q", out)
	}
}

func TestRender_SGROptimization(t *testing.T) {
	renderer := NewRenderer()

	redCell := buffer.Cell{Rune: 'X', Fg: 0xFF0000}
	changes := []diff.Change{
		{X: 0, Y: 0, Cell: redCell},
		{X: 1, Y: 0, Cell: redCell}, // Identical style
	}

	out := string(renderer.Render(changes))

	// SGR 38;2;255;0;0 should only appear ONCE in the output due to delta SGR optimization
	count := strings.Count(out, "38;2;255;0;0")
	if count != 1 {
		t.Fatalf("expected SGR to be emitted 1 time, got %d times in %q", count, out)
	}
}

func BenchmarkRender_1000Changes(b *testing.B) {
	renderer := NewRenderer()

	// Warm up internal buffer capacity
	changes := make([]diff.Change, 1000)
	for i := 0; i < 1000; i++ {
		changes[i] = diff.Change{
			X: i % 100,
			Y: i / 100,
			Cell: buffer.Cell{
				Rune:  rune('A' + i%26),
				Fg:    uint32(0xFF0000 | (i & 0xFF)),
				Bg:    0x001122,
				Attrs: buffer.Bold,
			},
		}
	}

	_ = renderer.Render(changes) // Warmup

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = renderer.Render(changes)
	}
}
