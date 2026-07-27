package buffer

import "unsafe"

// Attrs represents text styling attributes (bold, italic, underline, etc.)
// encoded as a bitmask (1 byte).
type Attrs uint8

const (
	Bold Attrs = 1 << iota
	Italic
	Underline
	Reverse
	Blink
	Dim
	Strikethrough
)

// Cell represents a single terminal display cell occupying exactly 16 bytes in memory
// for optimal L1/L2 CPU cache line alignment.
type Cell struct {
	Rune  rune   // 4 bytes: Unicode code point
	Fg    uint32 // 4 bytes: Packed RGB foreground color (0x00RRGGBB)
	Bg    uint32 // 4 bytes: Packed RGB background color (0x00RRGGBB)
	Attrs Attrs  // 1 byte:  Bitmask of text styling attributes
	// Go compiler automatically pads 3 bytes to align struct size to 16 bytes.
}

// Buffer represents a flat 2D grid of display cells (Data-Oriented Design).
type Buffer struct {
	Width, Height int
	Cells         []Cell // Flat slice indexed by y*Width + x
}

// NewBuffer creates a new Buffer with the specified width and height.
func NewBuffer(width, height int) *Buffer {
	if width < 0 {
		width = 0
	}
	// NewBuffer allocates a Buffer of width×height cells, all set to the
	// zero-value Cell (space rune, default colors).
	func NewBuffer(width, height int) *Buffer {
		return &Buffer{
			Width:  width,
			Height: height,
			Cells:   make([]Cell, width*height),
		}
	}

// SetCell sets the Cell at coordinate (x, y).
//
// Out-of-bounds coordinates (x < 0, x >= Width, y < 0, y >= Height) are safely ignored (clipped).
// Rationale: Safe clipping prevents runtime panics during dynamic UI animations, sub-cell VFX,
// and window resizing events while maintaining a zero-allocation rendering pipeline.
func (b *Buffer) SetCell(x, y int, cell Cell) {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return
	}
	b.Cells[y*b.Width+x] = cell
}

// GetCell returns the Cell at coordinate (x, y) by value (Value semantics).
//
// Out-of-bounds coordinates return a zero-value Cell.
func (b *Buffer) GetCell(x, y int) Cell {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return Cell{}
	}
	return b.Cells[y*b.Width+x]
}

// Clear resets all cells in the buffer to zero values without reallocating
// underlying slice memory, preserving capacity.
func (b *Buffer) Clear() {
	clear(b.Cells)
}

// Resize resizes the buffer dimensions. It reuses existing slice capacity
// if the requested total size fits, avoiding heap allocations during window resizes.
func (b *Buffer) Resize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	if b.Width == width && b.Height == height {
		return
	}

	b.Width = width
	b.Height = height
	n := width * height

	if cap(b.Cells) >= n {
		b.Cells = b.Cells[:n]
		clear(b.Cells)
	} else {
		b.Cells = make([]Cell, n)
	}
}

// SizeofCell returns the unsafe byte size of Cell struct.
func SizeofCell() uintptr {
	return unsafe.Sizeof(Cell{})
}
