package layout

import (
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// Layout math correctness
// ---------------------------------------------------------------------------

// measureAll runs Measure on a tree and returns the rects produced.
//
// For tests that want to inspect the *container's* children rects
// independently of the recursive leaf rects, use childRects below.
func measureAll(n Node, aw, ah int) []Rect {
	buf := make([]Rect, 0, 64)
	return n.Measure(aw, ah, buf)
}

// childRects walks the rects a container produced for its direct
// children and returns them in order, skipping any recursive rects
// emitted by the children themselves.
//
// The interleaving pattern for a container of N children is:
//
//	[child[0], recurse(child[0]), child[1], recurse(child[1]), ...]
//
// The recursion's contribution depends on the subtree. For a leaf it
// is exactly 1 rect. For a sub-container it is 1 + sub-tree rects.
//
// Tests use childRects when they want to assert only the positions
// the container itself decided for its direct children, not the
// layout of the children.
func childRects(t *testing.T, c *Container, results []Rect) []Rect {
	t.Helper()
	out := make([]Rect, 0, len(c.Children))
	pos := 0
	for {
		if pos >= len(results) {
			break
		}
		out = append(out, results[pos])
		pos++
		// Skip the recursion contribution. A Leaf contributes 1 rect;
		// a Container contributes its own len(Children) + sum of its
		// own recursion contributions. We approximate by walking the
		// underlying Children slice.
		if pos-1 < len(c.Children) {
			ch := c.Children[pos-1].Node
			pos += contribution(ch)
		}
	}
	return out
}

// contribution returns how many rects the subtree rooted at n will
// emit when measured against any positive available size. It is a
// closed-form count so tests can compute childRects without running
// Measure twice.
//
// contribution(Leaf) = 1
// contribution(Container with k children) = k + sum contribution(child)
// contribution(nil) = 0
func contribution(n Node) int {
	switch v := n.(type) {
	case Leaf:
		return 1
	case *Container:
		s := len(v.Children)
		for _, ch := range v.Children {
			s += contribution(ch.Node)
		}
		return s
	}
	return 0
}

// ---------------------------------------------------------------------------
// Row tests
// ---------------------------------------------------------------------------

func TestRow_BasisOnly(t *testing.T) {
	t.Parallel()

	c := &Container{
		Dir: Row,
		Children: []Child{
			{Node: Leaf{}, Basis: 10},
			{Node: Leaf{}, Basis: 20},
			{Node: Leaf{}, Basis: 30},
		},
	}
	results := measureAll(c, 60, 5)

	// Total rects: 3 container + 3 leaves = 6.
	if got, want := len(results), 6; got != want {
		t.Fatalf("len(results) = %d, want %d", got, want)
	}
	children := childRects(t, c, results)
	if got, want := len(children), 3; got != want {
		t.Fatalf("len(children) = %d, want %d", got, want)
	}

	// Children sum must equal available width.
	var sum int
	for _, r := range children {
		sum += r.Width
	}
	if sum != 60 {
		t.Errorf("sum widths of children = %d, want 60", sum)
	}

	want := []Rect{
		{X: 0, Y: 0, Width: 10, Height: 5},
		{X: 10, Y: 0, Width: 20, Height: 5},
		{X: 30, Y: 0, Width: 30, Height: 5},
	}
	for i := range want {
		if children[i] != want[i] {
			t.Errorf("child[%d] = %+v, want %+v", i, children[i], want[i])
		}
	}
}

func TestRow_GrowEqual(t *testing.T) {
	t.Parallel()

	c := &Container{
		Dir: Row,
		Children: []Child{
			{Node: Leaf{}, Grow: 1},
			{Node: Leaf{}, Grow: 1},
			{Node: Leaf{}, Grow: 1},
		},
	}
	results := measureAll(c, 30, 4)

		if got, want := len(results), 6; got != want {
			t.Fatalf("len(results) = %d, want %d", got, want)
		}
		children := childRects(t, c, results)

		var sum int
		for _, r := range children {
			sum += r.Width
		}
		if sum != 30 {
			t.Errorf("sum widths = %d, want 30", sum)
		}
		for i := range int32(len(children)) {
				if children[i].Width != 10 {
					t.Errorf("child[%d].Width = %d, want 10", i, children[i].Width)
				}
			}
		}
	t.Parallel()

	c := &Container{
		Dir: Row,
		Children: []Child{
			{Node: Leaf{}, Grow: 1},
			{Node: Leaf{}, Grow: 2},
			{Node: Leaf{}, Grow: 1},
		},
	}
	results := measureAll(c, 40, 2)
	children := childRects(t, c, results)

	var sum int
	for _, r := range children {
		sum += r.Width
	}
	if sum != 40 {
		t.Fatalf("sum widths = %d, want 40", sum)
	}
	want := []int{10, 20, 10}
	for i, w := range want {
		if children[i].Width != w {
			t.Errorf("child[%d].Width = %d, want %d", i, children[i].Width, w)
		}
	}
}

func TestRow_OverflowWithShrink(t *testing.T) {
	t.Parallel()

	// Available = 10, sum(Basis) = 30, total overflow = 20.
	// Shrink weights are 1, 2, 1 → sum = 4.
	// Each child takes 1/4, 2/4, 1/4 of the overflow: 5, 10, 5.
	// Resulting sizes: 5, 0, 5 — note child 1 shrinks to 0.
	c := &Container{
		Dir: Row,
		Children: []Child{
			{Node: Leaf{}, Basis: 10, Shrink: 1},
			{Node: Leaf{}, Basis: 10, Shrink: 2},
			{Node: Leaf{}, Basis: 10, Shrink: 1},
		},
	}
	results := measureAll(c, 10, 3)
	children := childRects(t, c, results)

	want := []int{5, 0, 5}
	for i, w := range want {
		if children[i].Width != w {
			t.Errorf("child[%d].Width = %d, want %d", i, children[i].Width, w)
		}
	}
	var sum int
	for _, r := range children {
		sum += r.Width
	}
	if sum != 10 {
		t.Errorf("sum widths after overflow = %d, want 10", sum)
	}
}

func TestRow_OverflowNoShrink(t *testing.T) {
	t.Parallel()

	c := &Container{
		Dir: Row,
		Children: []Child{
			{Node: Leaf{}, Basis: 30},
			{Node: Leaf{}, Basis: 30},
		},
	}
	results := measureAll(c, 10, 5)
	children := childRects(t, c, results)

	for i, r := range children {
		if r.Width != 0 {
			t.Errorf("child[%d].Width = %d, want 0", i, r.Width)
		}
	}
}

func TestRow_ZeroWidth(t *testing.T) {
	t.Parallel()

	c := &Container{
		Dir: Row,
		Children: []Child{
			{Node: Leaf{}, Basis: 5},
			{Node: Leaf{}, Basis: 5},
		},
	}
	// Available = 0, sum(Basis) = 10, overflow = 10, no Shrink →
	// both children collapse to 0; recursion into size-0 children
	// is skipped.
	buf := make([]Rect, 0, 4)
	results := c.Measure(0, 5, buf)

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2 (no recursion into size-0 children)", len(results))
	}
	for i, r := range results {
		if r.Width != 0 {
			t.Errorf("rect[%d].Width = %d, want 0", i, r.Width)
		}
		if r.Height != 5 {
			t.Errorf("rect[%d].Height = %d, want 5", i, r.Height)
		}
	}
}

// TestRow_NoChildren returns an empty slice without allocating.
func TestRow_NoChildren(t *testing.T) {
	t.Parallel()

	c := &Container{Dir: Row}
	buf := make([]Rect, 0, 1)
	got := c.Measure(80, 24, buf)
	if len(got) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(got))
	}
	// Must not allocate beyond the supplied buffer.
	if cap(got) != 1 {
		t.Errorf("cap = %d, want 1 (Measure must not allocate)", cap(got))
	}
}

// ---------------------------------------------------------------------------
// Column tests
// ---------------------------------------------------------------------------

func TestColumn_StacksVertically(t *testing.T) {
	t.Parallel()

	c := &Container{
		Dir: Column,
		Children: []Child{
			{Node: Leaf{}, Basis: 5},
			{Node: Leaf{}, Basis: 10},
			{Node: Leaf{}, Basis: 15},
		},
	}
	results := measureAll(c, 80, 30)

	if len(results) != 6 {
		t.Fatalf("len(results) = %d, want 6", len(results))
	}
	children := childRects(t, c, results)

	var sum int
	for _, r := range children {
		sum += r.Height
	}
	if sum != 30 {
		t.Errorf("sum heights of children = %d, want 30", sum)
	}
	want := []Rect{
		{X: 0, Y: 0, Width: 80, Height: 5},
		{X: 0, Y: 5, Width: 80, Height: 10},
		{X: 0, Y: 15, Width: 80, Height: 15},
	}
	for i := range want {
		if children[i] != want[i] {
			t.Errorf("child[%d] = %+v, want %+v", i, children[i], want[i])
		}
	}
}

func TestColumn_GrowEqual(t *testing.T) {
	t.Parallel()

	c := &Container{
		Dir: Column,
		Children: []Child{
			{Node: Leaf{}, Grow: 1},
			{Node: Leaf{}, Grow: 1},
			{Node: Leaf{}, Grow: 1},
		},
	}
	results := measureAll(c, 40, 30)
	children := childRects(t, c, results)

	var sum int
	for _, r := range children {
		sum += r.Height
	}
	if sum != 30 {
		t.Errorf("sum heights = %d, want 30", sum)
	}
	for i, r := range children {
		if r.Height != 10 {
			t.Errorf("child[%d].Height = %d, want 10", i, r.Height)
		}
	}
}

// ---------------------------------------------------------------------------
// Mixed and nested tests
// ---------------------------------------------------------------------------

func TestMixed_BasisAndGrow(t *testing.T) {
	t.Parallel()

	c := &Container{
		Dir: Row,
		Children: []Child{
			{Node: Leaf{}, Basis: 20}, // takes 20 first
			{Node: Leaf{}, Grow: 1},   // shares remainder
			{Node: Leaf{}, Grow: 1},   // shares remainder
		},
	}
	results := measureAll(c, 80, 4)
	children := childRects(t, c, results)

	want := []int{20, 30, 30}
	for i, w := range want {
		if children[i].Width != w {
			t.Errorf("child[%d].Width = %d, want %d", i, children[i].Width, w)
		}
	}
	var sum int
	for _, r := range children {
		sum += r.Width
	}
	if sum != 80 {
		t.Errorf("sum widths = %d, want 80", sum)
	}
}

func TestNested_RowOfColumnsOfLeaves(t *testing.T) {
	t.Parallel()

	cols := make([]*Container, 3)
	for c := 0; c < 3; c++ {
		col := &Container{
			Dir: Column,
			Children: []Child{
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 1},
			},
		}
		cols[c] = col
	}

	root := &Container{
		Dir: Row,
		Children: []Child{
			{Node: cols[0], Grow: 1},
			{Node: cols[1], Grow: 1},
			{Node: cols[2], Grow: 1},
		},
	}
	results := measureAll(root, 90, 60)

	// Total rects: 3 (row children) + 3*5 (column children) + 3*5
	// (leaf children) = 33.
	if got, want := len(results), 33; got != want {
		t.Fatalf("len(results) = %d, want %d", got, want)
	}

	// Root row invariant: its 3 column children widths sum to 90.
	rowChildren := childRects(t, root, results)
	if len(rowChildren) != 3 {
		t.Fatalf("len(rowChildren) = %d, want 3", len(rowChildren))
	}
	var rootSum int
	for _, r := range rowChildren {
		rootSum += r.Width
	}
	if rootSum != 90 {
		t.Errorf("root row sum widths = %d, want 90", rootSum)
	}

	// Each column rect at row-children level: width 30, height 60.
	for i, r := range rowChildren {
		if r.Width != 30 {
			t.Errorf("column[%d].Width = %d, want 30", i, r.Width)
		}
		if r.Height != 60 {
			t.Errorf("column[%d].Height = %d, want 60", i, r.Height)
		}
	}

	// Each column invariant: its 5 leaf children heights sum to 60.
	for i, col := range cols {
		// Walk into the column's emitted rects. They start right
		// after each row-child rect. The column rects are
		// interleaved at indices: rowChildren[i] is at results
		// index 2*i (since each preceding column emits 1 +
		// 5 leaves = 6 rects, except the first).
		//
		// Easier: walk the slice manually using contribution.
		// root contributes (3 row children + 3*5 column children +
		// 3*5 leaves) = 33. Each row child contributes (5 + 5
		// leaves) = 10 except the last column has fewer.
		//
		// We'll directly inspect known indices: the layout walks in
		// DFS order, so the column's child rects sit immediately
		// after the row's child rect for that column.
		colResults := extractSubtree(t, results, root, col)
		if len(colResults) != 10 { // 5 column children + 5 leaves
			t.Fatalf("column[%d] emitted %d rects, want 10", i, len(colResults))
		}
		// First 5 are the column's children rects (the leaves).
		var colSum int
		for j := 0; j < 5; j++ {
			colSum += colResults[j].Height
		}
		if colSum != 60 {
			t.Errorf("column[%d] sum heights = %d, want 60", i, colSum)
		}
	}
}

// extractSubtree finds the rects that a sub-container emitted. It
// relies on the layout's DFS order: the sub-container's children rects
// and their recursive contributions are contiguous.
//
// For test use only; the algorithm is O(n) over results and O(d) over
// the tree depth.
func extractSubtree(t *testing.T, results []Rect, root Node, target Node) []Rect {
	t.Helper()
	// Walk root's child sequence and find target.
	rootC, ok := root.(*Container)
	if !ok {
		t.Fatalf("root is not a Container")
	}

	idx := 0
	for _, ch := range rootC.Children {
		if ch.Node == target {
			idx++ // skip target's own rect (the row-child rect)
			contrib := contribution(target)
			return results[idx : idx+contrib]
		}
		idx += 1 + contribution(ch.Node)
	}
	t.Fatalf("target not found in root's children")
	return nil
}

func TestLeaf_ProducesOneRect(t *testing.T) {
	t.Parallel()

	var l Leaf
	results := measureAll(l, 7, 11)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0] != (Rect{X: 0, Y: 0, Width: 7, Height: 11}) {
		t.Fatalf("rect = %+v, want full available", results[0])
	}
}

// TestRow_ContainerInvariantSumsCorrectly is a property-style check: for
// any combination of Basis/Grow values that fully covers the available
// space, sum of child widths == available.
func TestRow_ContainerInvariantSumsCorrectly(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		avail int
		props []Child
	}{
		{
			name:  "Grow 1:1:1, avail=100",
			avail: 100,
			props: []Child{
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 1},
			},
		},
		{
			name:  "Grow 1:2:3, avail=120",
			avail: 120,
			props: []Child{
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 2},
				{Node: Leaf{}, Grow: 3},
			},
		},
		{
			name:  "Basis 30 + Grow 1:2, avail=100",
			avail: 100,
			props: []Child{
				{Node: Leaf{}, Basis: 30},
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 2},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Container{Dir: Row, Children: tc.props}
			results := measureAll(c, tc.avail, 5)
			children := childRects(t, c, results)
			var sum int
			for _, r := range children {
				sum += r.Width
			}
			if sum != tc.avail {
				t.Errorf("sum widths = %d, want %d", sum, tc.avail)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestSanity_Contribution exercises the helper itself: print the
// contribution of a known tree so any drift in the layout logic is
// caught early by humans running -v.
// ---------------------------------------------------------------------------

func TestSanity_Contribution(t *testing.T) {
	t.Parallel()

	cols := []*Container{}
	for i := 0; i < 3; i++ {
		cols = append(cols, &Container{
			Dir:      Column,
			Children: []Child{{Node: Leaf{}}, {Node: Leaf{}}, {Node: Leaf{}}, {Node: Leaf{}}, {Node: Leaf{}}},
		})
	}
	root := &Container{
		Dir:      Row,
		Children: []Child{{Node: cols[0]}, {Node: cols[1]}, {Node: cols[2]}},
	}

	want := 3 + 3*(5+5) // row + 3 × (5 column children + 5 leaves)
	if got := contribution(root); got != want {
		t.Fatalf("contribution(root) = %d, want %d", got, want)
	}
	t.Logf("contribution(root) = %d (3 row + 3*10 column + 15 leaves = 33)", contribution(root))
}

// quietFmt silences fmt usage when the file isn't compiled with tests
// that use it; we keep the import explicit so go vet stays happy.
var _ = fmt.Sprintf

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

// buildTree constructs the spec's reference tree:
//
//	Row → 3 × Column → 5 × Leaf
//
// Total rects produced:
//
//	3 (row's column children) +
//	3 × 5 (each column's leaf children) +
//	3 × 5 (each leaf itself) = 33
//
// The tree is built once outside the timed loop so the benchmark
// measures layout math only.
func buildTree() (Node, int) {
	cols := make([]Node, 3)
	for i := 0; i < 3; i++ {
		c := &Container{
			Dir: Column,
			Children: []Child{
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 1},
				{Node: Leaf{}, Grow: 1},
			},
		}
		cols[i] = c
	}

	root := &Container{
		Dir: Row,
		Children: []Child{
			{Node: cols[0], Grow: 1},
			{Node: cols[1], Grow: 1},
			{Node: cols[2], Grow: 1},
		},
	}
	return root, 33
}

// BenchmarkLayout measures Measure on the spec's reference tree with
// sufficient results capacity for 0 allocs/op.
//
// Run with: go test -bench=BenchmarkLayout -benchmem ./layout/
func BenchmarkLayout(b *testing.B) {
	tree, total := buildTree()

	buf := make([]Rect, 0, total)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tree.Measure(90, 60, buf[:0])
	}
}

// BenchmarkLayout_Deep is a stress benchmark that measures a deeper
// tree to ensure the recursion path stays alloc-free.
func BenchmarkLayout_Deep(b *testing.B) {
	// 5 levels of nesting, each level with 3 children. The pattern
	// alternates Row and Column at each depth to exercise both
	// directions in the recursion.
	const levels = 5
	const fanout = 3

	var build func(depth int) Node
	build = func(depth int) Node {
		children := make([]Child, fanout)
		for i := 0; i < fanout; i++ {
			if depth == 0 {
				children[i] = Child{Node: Leaf{}, Grow: 1}
			} else {
				children[i] = Child{Node: build(depth - 1), Grow: 1}
			}
		}
		if depth%2 == 0 {
			return &Container{Dir: Row, Children: children}
		}
		return &Container{Dir: Column, Children: children}
	}

	tree := build(levels - 1)

	// Compute exact contribution for the bench's capacity hint.
	// 5 levels, fanout=3: contribution at depth d is sum of geometric
	// series. Bottom level (depth 0) = 3 (children) + 3*1 (leaves) =
	// 6. Each higher level multiplies by 3 and adds 3:
	//   depth 0: 6
	//   depth 1: 3 + 3*6 = 21
	//   depth 2: 3 + 3*21 = 66
	//   depth 3: 3 + 3*66 = 201
	//   depth 4: 3 + 3*201 = 606
	const total = 606
	buf := make([]Rect, 0, total)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tree.Measure(120, 40, buf[:0])
	}
}
