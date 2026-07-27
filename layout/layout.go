// Package layout implements a 1D Flexbox-like layout engine per ADR-0001.
//
// The engine distributes a container's main-axis size among its children
// using three integer properties:
//
//   - Basis  — a fixed size on the main axis; if > 0, the child gets at
//     least this much before grow/shrink runs.
//   - Grow   — the share of *remaining* space the child claims when the
//     container is not full.
//   - Shrink — the share of *overflow* the child bears when the
//     container is over-subscribed.
//
// Two container directions are supported: Row (main axis = X, children
// stacked left-to-right) and Column (main axis = Y, children stacked
// top-to-bottom). Two-dimensional layouts are composed by nesting
// containers of alternating directions.
//
// # Zero-allocation contract
//
// The package is designed for the render hot path: Measure is called once
// per layout pass per node and must not allocate. To keep that promise,
// callers MUST pre-size the results slice with enough capacity to hold
// every Rect produced by the entire tree:
//
//	buf := make([]layout.Rect, 0, expectedLeafCount)
//	results := root.Measure(width, height, buf)
//
// With sufficient capacity, Measure reuses the backing array through
// append and never touches the allocator. The benchmarks in
// layout_test.go encode the expected capacity for a representative tree
// (1 Row → 3 Column → 5 leaves).
package layout

// Rect is an axis-aligned rectangle on the terminal cell grid.
// Coordinates are 0-based and inclusive on the cell grid.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Direction discriminates the main axis of a Container.
type Direction int

const (
	// Row stacks children left-to-right. Main axis = X.
	Row Direction = iota
	// Column stacks children top-to-bottom. Main axis = Y.
	Column
)

// Node is anything that can be measured against an available rect.
//
// Measure appends the rectangles produced for this node (and its
// descendants) into results and returns the extended slice. The caller
// is responsible for providing a results slice with enough capacity to
// avoid reallocations.
type Node interface {
	Measure(availableWidth, availableHeight int, results []Rect) []Rect
}

// Child binds a node to its layout properties.
//
// Properties are independent of the container direction: the same
// struct is used for both Row and Column.
type Child struct {
	Node Node

	// Grow is the share of remaining main-axis space claimed when the
	// container has unused capacity. Values <= 0 are treated as zero.
	Grow int

	// Shrink is the share of overflow absorbed when the container is
	// over-subscribed. Values <= 0 are treated as zero.
	Shrink int

	// Basis is a fixed main-axis size taken before grow/shrink runs.
	// If Basis > 0, the child receives at least this much on the main
	// axis. Basis <= 0 means the child starts at 0 on the main axis.
	Basis int
}

// Container is a Node that distributes an available rect among its
// children along a single axis (Row or Column).
type Container struct {
	// Dir selects the main axis. Required.
	Dir Direction
	// Children is laid out along the main axis in order.
	Children []Child
}

// Measure distributes the available rectangle across Children and
// recurses into each child with its computed rect.
//
// The cross-axis size is passed through unchanged: every child receives
// the full cross size of the container (Row children all get Height;
// Column children all get Width). This matches the CSS rule that a
// flex item stretches by default along the cross axis.
//
// Allocation behaviour: zero allocations when len(results) + total tree
// leaves <= cap(results). Callers that care should pre-size results.
func (c *Container) Measure(aw, ah int, results []Rect) []Rect {
	switch c.Dir {
	case Row:
		return c.measure(aw, ah, results, true)
	case Column:
		return c.measure(ah, aw, results, false)
	default:
		// Unspecified direction: nothing to do, hand back the buffer
		// unchanged.
		return results
	}
}

// measure is the shared distribution loop. The transposed flag chooses
// which axis is the main axis and swaps X/Y accordingly.
//
// When transposed=true (Row), main=Width, cross=Height.
// When transposed=false (Column), main=Height, cross=Width.
func (c *Container) measure(main, cross int, results []Rect, transposed bool) []Rect {
	n := len(c.Children)
	if n == 0 {
		return results
	}

	// Negative sizes are not meaningful; clamp to zero so the math is
	// well-defined.
	if main < 0 {
		main = 0
	}
	if cross < 0 {
		cross = 0
	}

	// Phase 1: compute initial main-axis sizes from Basis and tally
	// totals. mainAlloc is reused as the per-child running main-axis
	// size; it is allocated once on the stack and reused.
	var mainAlloc [256]int
	if n > len(mainAlloc) {
		// Fall back to a heap slice for very wide containers. The hot
		// path of small UIs never hits this branch.
		heap := make([]int, n)
		return c.measureSlow(main, cross, results, transposed, heap)
	}
	sizes := mainAlloc[:n]

	var used, sumGrow, sumShrink int
	for i := 0; i < n; i++ {
		ch := &c.Children[i]
		if ch.Basis > 0 {
			sizes[i] = ch.Basis
			used += ch.Basis
		}
		if ch.Grow > 0 {
			sumGrow += ch.Grow
		}
		if ch.Shrink > 0 {
			sumShrink += ch.Shrink
		}
	}

	// Phase 2: distribute remaining space (underflow) or absorb
	// overflow (overflow).
	free := main - used
	switch {
	case free > 0 && sumGrow > 0:
		// Underflow: distribute proportional to Grow.
		// Allocate remainder = free % sumGrow to the first eligible
		// children so the totals line up exactly with main.
		rem := free % sumGrow
		share := free / sumGrow
		for i := 0; i < n; i++ {
			ch := &c.Children[i]
			if ch.Grow <= 0 {
				continue
			}
			give := share * ch.Grow
			if rem > 0 {
				// Hand out the rounding remainder one unit at a
				// time, top-down.
				give++
				rem--
			}
			sizes[i] += give
		}
	case free < 0 && sumShrink > 0:
		// Overflow: shrink proportional to Shrink.
		over := -free
		rem := over % sumShrink
		share := over / sumShrink
		for i := 0; i < n; i++ {
			ch := &c.Children[i]
			if ch.Shrink <= 0 {
				continue
			}
			take := share * ch.Shrink
			if rem > 0 {
				take++
				rem--
			}
			sizes[i] -= take
			if sizes[i] < 0 {
				sizes[i] = 0
			}
		}
	case free < 0:
		// Overflow with no shrink factors: clip every child to 0.
		for i := 0; i < n; i++ {
			sizes[i] = 0
		}
	}

	// Phase 3: lay out each child along the main axis and recurse.
	// We honour a single invariant: sum(sizes) == main when main > 0.
	// If integer rounding left a residual, we attach it to the last
	// child's size so totals match exactly.
	pos := 0
	for i := 0; i < n; i++ {
		ch := &c.Children[i]
		size := sizes[i]

		// Honour the invariant on the last child: give it whatever
		// slack is left so totals match exactly. We adjust the size in
		// place so the position stride stays consistent.
		if i == n-1 {
			residual := main - (pos + size)
			if residual > 0 {
				size += residual
			} else if residual < 0 && size+residual >= 0 {
				size += residual
			} else if size+residual < 0 {
				size = 0
			}
			sizes[i] = size
		}

		var rect Rect
		if transposed {
			rect = Rect{X: pos, Y: 0, Width: size, Height: cross}
		} else {
			rect = Rect{X: 0, Y: pos, Width: cross, Height: size}
		}
		pos += size

		results = results[:len(results)+1]
		results[len(results)-1] = rect

		if ch.Node != nil && size > 0 && cross > 0 {
			results = ch.Node.Measure(rect.Width, rect.Height, results)
		}
	}

	return results
}

// measureSlow is the heap-backed fallback for containers wider than the
// stack scratch buffer. It mirrors measure exactly.
func (c *Container) measureSlow(main, cross int, results []Rect, transposed bool, sizes []int) []Rect {
	n := len(c.Children)
	if n == 0 {
		return results
	}
	if main < 0 {
		main = 0
	}
	if cross < 0 {
		cross = 0
	}

	var used, sumGrow, sumShrink int
	for i := 0; i < n; i++ {
		ch := &c.Children[i]
		if ch.Basis > 0 {
			sizes[i] = ch.Basis
			used += ch.Basis
		}
		if ch.Grow > 0 {
			sumGrow += ch.Grow
		}
		if ch.Shrink > 0 {
			sumShrink += ch.Shrink
		}
	}

	free := main - used
	switch {
	case free > 0 && sumGrow > 0:
		rem := free % sumGrow
		share := free / sumGrow
		for i := 0; i < n; i++ {
			ch := &c.Children[i]
			if ch.Grow <= 0 {
				continue
			}
			give := share * ch.Grow
			if rem > 0 {
				give++
				rem--
			}
			sizes[i] += give
		}
	case free < 0 && sumShrink > 0:
		over := -free
		rem := over % sumShrink
		share := over / sumShrink
		for i := 0; i < n; i++ {
			ch := &c.Children[i]
			if ch.Shrink <= 0 {
				continue
			}
			take := share * ch.Shrink
			if rem > 0 {
				take++
				rem--
			}
			sizes[i] -= take
			if sizes[i] < 0 {
				sizes[i] = 0
			}
		}
	case free < 0:
		for i := 0; i < n; i++ {
			sizes[i] = 0
		}
	}

	pos := 0
	for i := 0; i < n; i++ {
		ch := &c.Children[i]
		size := sizes[i]
		if i == n-1 {
			residual := main - (pos + size)
			if residual > 0 {
				size += residual
			} else if residual < 0 && size+residual >= 0 {
				size += residual
			} else if size+residual < 0 {
				size = 0
			}
			sizes[i] = size
		}

		var rect Rect
		if transposed {
			rect = Rect{X: pos, Y: 0, Width: size, Height: cross}
		} else {
			rect = Rect{X: 0, Y: pos, Width: cross, Height: size}
		}
		pos += size

		results = results[:len(results)+1]
		results[len(results)-1] = rect

		if ch.Node != nil && size > 0 && cross > 0 {
			results = ch.Node.Measure(rect.Width, rect.Height, results)
		}
	}
	return results
}

// Leaf is a leaf node that occupies exactly its available rect.
// It produces exactly one Rect per Measure call.
type Leaf struct{}

// Measure appends a single Rect covering the full available space and
// returns the extended slice.
//
// A Leaf has no children, so it produces exactly one Rect.
func (Leaf) Measure(aw, ah int, results []Rect) []Rect {
	if aw < 0 {
		aw = 0
	}
	if ah < 0 {
		ah = 0
	}
	results = results[:len(results)+1]
	results[len(results)-1] = Rect{X: 0, Y: 0, Width: aw, Height: ah}
	return results
}
