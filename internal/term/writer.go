package term

// Writer is a pre-allocated byte buffer for terminal output. It
// collects all escape sequences and content for a frame, then flushes
// them in a single write(2) syscall via Flush.
//
// The buffer grows only on terminal resize (via Grow). In steady-state
// rendering it performs zero allocations.
//
// Principle 5.3: enables single-syscall output for tear-free rendering.
type Writer struct {
	buf []byte // Pre-allocated output buffer.
	n   int    // Number of bytes written.
}

// NewWriter creates a Writer with the given initial capacity.
// Capacity should be sized for the worst case: rows × cols × maxEscapeLen.
func NewWriter(capacity int) *Writer {
	return &Writer{
		buf: make([]byte, capacity),
	}
}

// Len returns the number of bytes currently in the buffer.
func (w *Writer) Len() int { return w.n }

// Cap returns the total buffer capacity.
func (w *Writer) Cap() int { return len(w.buf) }

// Bytes returns the buffered content. The returned slice is valid until
// the next Reset, Flush, or write operation.
func (w *Writer) Bytes() []byte { return w.buf[:w.n] }

// Reset clears the buffer without releasing memory.
func (w *Writer) Reset() { w.n = 0 }

// Grow ensures the buffer has at least minCap bytes of capacity.
// Called on terminal resize. Does not shrink.
func (w *Writer) Grow(minCap int) {
	if minCap <= len(w.buf) {
		return
	}
	newBuf := make([]byte, minCap)
	copy(newBuf, w.buf[:w.n])
	w.buf = newBuf
}

// WriteByte appends a single byte to the buffer.
func (w *Writer) WriteByte(b byte) {
	w.ensureSpace(1)
	w.buf[w.n] = b
	w.n++
}

// Write appends p to the buffer.
func (w *Writer) Write(p []byte) {
	w.ensureSpace(len(p))
	w.n += copy(w.buf[w.n:], p)
}

// WriteString appends s to the buffer without allocating.
// Go's copy builtin accepts string as source for []byte destination.
func (w *Writer) WriteString(s string) {
	w.ensureSpace(len(s))
	w.n += copy(w.buf[w.n:], s)
}

// ensureSpace grows the buffer if needed. This should not fire during
// steady-state rendering — if it does, initial capacity was too small.
func (w *Writer) ensureSpace(n int) {
	required := w.n + n
	if required <= len(w.buf) {
		return
	}
	// Double until sufficient.
	newCap := len(w.buf) * 2
	if newCap < required {
		newCap = required
	}
	w.Grow(newCap)
}
