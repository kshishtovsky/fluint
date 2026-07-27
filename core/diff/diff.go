package diff

import "github.com/kshishtovsky/fluint/core/buffer"

// Change represents a modified cell and its target 2D coordinates.
type Change struct {
	X, Y int
	Cell buffer.Cell
}

// Differ compares front and back buffer grids to compute a minimal changeset.
type Differ struct {
	changes []Change // Pre-allocated slice reused across Diff calls
}

// NewDiffer creates a new Differ with a pre-allocated capacity for maxCells.
func NewDiffer(maxCells int) *Differ {
	if maxCells < 0 {
		maxCells = 0
	}
	return &Differ{
		changes: make([]Change, maxCells),
	}
}

// Diff compares front and back buffers element-wise in a 1D loop and returns
// the slice of detected changes.
//
// In steady state, Diff produces 0 heap allocations. The returned slice
// aliases internal buffer capacity up to the number of detected changes.
func (d *Differ) Diff(front, back *buffer.Buffer) []Change {
	if front == nil || back == nil || front.Width <= 0 || front.Height <= 0 {
		return d.changes[:0]
	}

	width := front.Width
	minLen := len(front.Cells)
	if len(back.Cells) < minLen {
		minLen = len(back.Cells)
	}

	maxCap := cap(d.changes)
	count := 0

	for i := 0; i < minLen; i++ {
		if front.Cells[i] != back.Cells[i] {
			if count >= maxCap {
				break
			}
			d.changes[count] = Change{
				X:    i % width,
				Y:    i / width,
				Cell: back.Cells[i],
			}
			count++
		}
	}

	return d.changes[:count]
}
