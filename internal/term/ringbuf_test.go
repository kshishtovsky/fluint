package term

import "testing"

func TestRingBufWriteRead(t *testing.T) {
	tests := []struct {
		name    string
		writes  []string
		wantLen int
		wantOut string
	}{
		{
			name:    "single write",
			writes:  []string{"hello"},
			wantLen: 5,
			wantOut: "hello",
		},
		{
			name:    "multiple writes",
			writes:  []string{"he", "llo"},
			wantLen: 5,
			wantOut: "hello",
		},
		{
			name:    "empty write",
			writes:  []string{""},
			wantLen: 0,
			wantOut: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rb RingBuf

			for _, w := range tt.writes {
				n := rb.WriteFrom([]byte(w))
				if n != len(w) {
					t.Fatalf("WriteFrom(%q): wrote %d, want %d", w, n, len(w))
				}
			}

			if rb.Len() != tt.wantLen {
				t.Fatalf("Len() = %d, want %d", rb.Len(), tt.wantLen)
			}

			// Read back via Peek + Consume.
			var got []byte
			for rb.Len() > 0 {
				chunk := rb.Peek()
				got = append(got, chunk...)
				rb.Consume(len(chunk))
			}

			if string(got) != tt.wantOut {
				t.Fatalf("read back %q, want %q", got, tt.wantOut)
			}
		})
	}
}

func TestRingBufWrapAround(t *testing.T) {
	var rb RingBuf

	// Fill most of the buffer.
	fill := make([]byte, ringBufSize-10)
	for i := range fill {
		fill[i] = 'A'
	}
	rb.WriteFrom(fill)
	rb.Consume(ringBufSize - 10) // consume all

	// Now head is near the end, tail is near the end.
	// Write data that wraps around.
	wrap := []byte("0123456789ABCDEF") // 16 bytes
	n := rb.WriteFrom(wrap)
	if n != 16 {
		t.Fatalf("WriteFrom(wrap): wrote %d, want 16", n)
	}

	if rb.Len() != 16 {
		t.Fatalf("Len() = %d, want 16", rb.Len())
	}

	// Peek should return the first contiguous chunk (before wrap).
	// Then a second Peek after Consume should return the rest.
	var got []byte
	for rb.Len() > 0 {
		chunk := rb.Peek()
		if len(chunk) == 0 {
			t.Fatal("Peek() returned empty slice with Len() > 0")
		}
		got = append(got, chunk...)
		rb.Consume(len(chunk))
	}

	if string(got) != string(wrap) {
		t.Fatalf("wrapped read: got %q, want %q", got, wrap)
	}
}

func TestRingBufFull(t *testing.T) {
	var rb RingBuf

	full := make([]byte, ringBufSize)
	for i := range full {
		full[i] = byte(i & 0xFF)
	}

	n := rb.WriteFrom(full)
	if n != ringBufSize {
		t.Fatalf("WriteFrom(full): wrote %d, want %d", n, ringBufSize)
	}

	if rb.Free() != 0 {
		t.Fatalf("Free() = %d, want 0", rb.Free())
	}

	// Overflow: writing more should return 0.
	n = rb.WriteFrom([]byte("extra"))
	if n != 0 {
		t.Fatalf("overflow write: wrote %d, want 0", n)
	}
}

func TestRingBufWritableSlice(t *testing.T) {
	var rb RingBuf

	ws := rb.WritableSlice()
	if len(ws) != ringBufSize {
		t.Fatalf("empty WritableSlice len = %d, want %d", len(ws), ringBufSize)
	}

	copy(ws[:5], "hello")
	rb.CommitWrite(5)

	if rb.Len() != 5 {
		t.Fatalf("after CommitWrite(5): Len() = %d, want 5", rb.Len())
	}
}

func TestRingBufReset(t *testing.T) {
	var rb RingBuf
	rb.WriteFrom([]byte("data"))
	rb.Reset()

	if rb.Len() != 0 {
		t.Fatalf("after Reset: Len() = %d, want 0", rb.Len())
	}
	if rb.Free() != ringBufSize {
		t.Fatalf("after Reset: Free() = %d, want %d", rb.Free(), ringBufSize)
	}
}

func BenchmarkRingBufWriteConsume(b *testing.B) {
	var rb RingBuf
	data := []byte("\x1b[1;2A\x1b[<0;10;20M") // typical escape sequences

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.WriteFrom(data)
		rb.Consume(rb.Len())
	}
}

func BenchmarkRingBufWritableSlice(b *testing.B) {
	var rb RingBuf

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws := rb.WritableSlice()
		copy(ws[:16], "0123456789ABCDEF")
		rb.CommitWrite(16)
		rb.Consume(16)
	}
}
