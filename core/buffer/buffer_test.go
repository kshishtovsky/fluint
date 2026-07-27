package buffer

import (
	"testing"
	"unsafe"
)

func TestCellSize(t *testing.T) {
	size := unsafe.Sizeof(Cell{})
	if size != 16 {
		t.Fatalf("Cell size = %d bytes, want exactly 16 bytes for cache line alignment", size)
	}
}

func TestBufferSetGet(t *testing.T) {
	buf := NewBuffer(10, 5)
	if buf.Width != 10 || buf.Height != 5 {
		t.Fatalf("Buffer dimensions = %dx%d, want 10x5", buf.Width, buf.Height)
	}

	cell := Cell{
		Rune:  'A',
		Fg:    0x00FF0000,
		Bg:    0x0000FF00,
		Attrs: Bold | Underline,
	}

	buf.SetCell(3, 2, cell)
	got := buf.GetCell(3, 2)

	if got != cell {
		t.Fatalf("GetCell(3, 2) = %+v, want %+v", got, cell)
	}
}

func TestBufferOutOfBounds(t *testing.T) {
	buf := NewBuffer(10, 5)
	cell := Cell{Rune: 'X', Fg: 0xFFFFFF}

	// Negative coordinates
	buf.SetCell(-1, 2, cell)
	buf.SetCell(3, -1, cell)
	// Out-of-bounds coordinates
	buf.SetCell(10, 2, cell)
	buf.SetCell(3, 5, cell)

	// Verify all returned cells are zero values
	if got := buf.GetCell(-1, 2); got != (Cell{}) {
		t.Fatalf("GetCell(-1, 2) = %+v, want zero Cell", got)
	}
	if got := buf.GetCell(10, 2); got != (Cell{}) {
		t.Fatalf("GetCell(10, 2) = %+v, want zero Cell", got)
	}
}

func TestBufferClear(t *testing.T) {
	buf := NewBuffer(10, 5)
	cell := Cell{Rune: 'Z', Fg: 0x123456}

	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			buf.SetCell(x, y, cell)
		}
	}

	buf.Clear()

	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			if got := buf.GetCell(x, y); got != (Cell{}) {
				t.Fatalf("GetCell(%d, %d) after Clear = %+v, want zero Cell", x, y, got)
			}
		}
	}
}

func TestBufferResize(t *testing.T) {
	buf := NewBuffer(10, 10)
	capBefore := cap(buf.Cells)

	// Downsize — should reuse capacity
	buf.Resize(5, 5)
	if buf.Width != 5 || buf.Height != 5 {
		t.Fatalf("Resized dimensions = %dx%d, want 5x5", buf.Width, buf.Height)
	}
	if cap(buf.Cells) != capBefore {
		t.Fatalf("Capacity changed after downsize = %d, want %d", cap(buf.Cells), capBefore)
	}

	// Upsize within capacity — should reuse capacity
	buf.Resize(8, 8)
	if cap(buf.Cells) != capBefore {
		t.Fatalf("Capacity changed after upsize within cap = %d, want %d", cap(buf.Cells), capBefore)
	}

	// Upsize exceeding capacity — reallocates
	buf.Resize(20, 20)
	if buf.Width != 20 || buf.Height != 20 {
		t.Fatalf("Resized dimensions = %dx%d, want 20x20", buf.Width, buf.Height)
	}
}

func BenchmarkBufferSetCell(b *testing.B) {
	buf := NewBuffer(100, 50)
	cell := Cell{Rune: 'A', Fg: 0x00FF0000, Bg: 0x0000FF00, Attrs: Bold}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.SetCell(i%100, (i/100)%50, cell)
	}
}

func BenchmarkBufferClear(b *testing.B) {
	buf := NewBuffer(100, 50)
	cell := Cell{Rune: 'A', Fg: 0x00FF0000, Bg: 0x0000FF00, Attrs: Bold}
	for y := 0; y < 50; y++ {
		for x := 0; x < 100; x++ {
			buf.SetCell(x, y, cell)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Clear()
	}
}
