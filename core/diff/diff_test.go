package diff

import (
	"math/rand"
	"testing"

	"github.com/kshishtovsky/fluint/core/buffer"
)

func TestDiff_Identical(t *testing.T) {
	front := buffer.NewBuffer(10, 5)
	back := buffer.NewBuffer(10, 5)

	differ := NewDiffer(50)
	changes := differ.Diff(front, back)

	if len(changes) != 0 {
		t.Fatalf("Diff returned %d changes for identical buffers, want 0", len(changes))
	}
}

func TestDiff_Changes(t *testing.T) {
	front := buffer.NewBuffer(10, 5)
	back := buffer.NewBuffer(10, 5)

	cellA := buffer.Cell{Rune: 'A', Fg: 0xFF0000}
	cellB := buffer.Cell{Rune: 'B', Fg: 0x00FF00}

	// Change cell at (3, 2) and (7, 4)
	back.SetCell(3, 2, cellA)
	back.SetCell(7, 4, cellB)

	differ := NewDiffer(50)
	changes := differ.Diff(front, back)

	if len(changes) != 2 {
		t.Fatalf("Diff returned %d changes, want 2", len(changes))
	}

	if changes[0].X != 3 || changes[0].Y != 2 || changes[0].Cell != cellA {
		t.Fatalf("changes[0] = %+v, want (3, 2, cellA)", changes[0])
	}
	if changes[1].X != 7 || changes[1].Y != 4 || changes[1].Cell != cellB {
		t.Fatalf("changes[1] = %+v, want (7, 4, cellB)", changes[1])
	}
}

func TestDiff_NilBuffers(t *testing.T) {
	differ := NewDiffer(50)

	if changes := differ.Diff(nil, nil); len(changes) != 0 {
		t.Fatalf("Diff(nil, nil) returned %d changes, want 0", len(changes))
	}

	buf := buffer.NewBuffer(10, 5)
	if changes := differ.Diff(buf, nil); len(changes) != 0 {
		t.Fatalf("Diff(buf, nil) returned %d changes, want 0", len(changes))
	}
}

func BenchmarkDiff_Identical(b *testing.B) {
	front := buffer.NewBuffer(100, 50)
	back := buffer.NewBuffer(100, 50)
	differ := NewDiffer(5000)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = differ.Diff(front, back)
	}
}

func BenchmarkDiff_100Changes(b *testing.B) {
	front := buffer.NewBuffer(100, 50)
	back := buffer.NewBuffer(100, 50)
	differ := NewDiffer(5000)

	// Modify 100 random cells in back buffer
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		x := rng.Intn(100)
		y := rng.Intn(50)
		back.SetCell(x, y, buffer.Cell{Rune: rune('A' + i%26), Fg: uint32(i * 1000)})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = differ.Diff(front, back)
	}
}
