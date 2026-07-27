package term

// ringBufSize is the capacity of the input ring buffer in bytes.
// PERF: must be a power of 2 for bitwise modulo (& ringBufMask).
const ringBufSize = 4096
const ringBufMask = ringBufSize - 1

// RingBuf is a fixed-size, single-goroutine ring buffer for terminal
// stdin input. It avoids all allocations after construction — the
// backing array is embedded in the struct.
//
// Not safe for concurrent use. Designed for the stdin reader goroutine
// where read and parse happen sequentially in the same goroutine.
type RingBuf struct {
	buf   [ringBufSize]byte
	head  int // Next write position.
	tail  int // Next read position.
	count int // Number of readable bytes.
}

// Len returns the number of bytes available for reading.
func (rb *RingBuf) Len() int { return rb.count }

// Free returns the number of bytes available for writing.
func (rb *RingBuf) Free() int { return ringBufSize - rb.count }

// WriteFrom copies bytes from p into the buffer, returning the number
// of bytes actually written. If the buffer is full, it writes as many
// bytes as space allows.
func (rb *RingBuf) WriteFrom(p []byte) int {
	free := rb.Free()
	n := len(p)
	if n > free {
		n = free
	}
	if n == 0 {
		return 0
	}

	// PERF: bulk copy in up to two segments (before and after wrap).
	end := rb.head + n
	if end <= ringBufSize {
		copy(rb.buf[rb.head:end], p[:n])
	} else {
		first := ringBufSize - rb.head
		copy(rb.buf[rb.head:], p[:first])
		copy(rb.buf[:n-first], p[first:n])
	}

	rb.head = (rb.head + n) & ringBufMask
	rb.count += n
	return n
}

// WritableSlice returns a contiguous slice of free space for direct I/O
// (e.g., passing to syscall.Read). After writing, call CommitWrite with
// the number of bytes written.
//
// If the free space wraps around the end of the buffer, only the first
// contiguous segment is returned. Call again after CommitWrite to get
// the wrapped segment.
func (rb *RingBuf) WritableSlice() []byte {
	if rb.count == ringBufSize {
		return nil
	}
	end := rb.head + rb.Free()
	if end > ringBufSize {
		end = ringBufSize
	}
	return rb.buf[rb.head:end]
}

// CommitWrite advances the write pointer by n bytes after a direct
// write via WritableSlice.
func (rb *RingBuf) CommitWrite(n int) {
	rb.head = (rb.head + n) & ringBufMask
	rb.count += n
}

// Peek returns a contiguous slice of readable bytes without consuming
// them. The returned slice is valid until the next write or consume.
//
// If readable data wraps around the buffer boundary, only the first
// contiguous chunk is returned. After consuming it with Consume, call
// Peek again to get the rest.
func (rb *RingBuf) Peek() []byte {
	if rb.count == 0 {
		return nil
	}
	end := rb.tail + rb.count
	if end > ringBufSize {
		end = ringBufSize
	}
	return rb.buf[rb.tail:end]
}

// Consume marks n bytes as read, advancing the read pointer.
func (rb *RingBuf) Consume(n int) {
	if n > rb.count {
		n = rb.count
	}
	rb.tail = (rb.tail + n) & ringBufMask
	rb.count -= n
}

// Reset discards all data and resets the buffer to empty state.
func (rb *RingBuf) Reset() {
	rb.head = 0
	rb.tail = 0
	rb.count = 0
}
