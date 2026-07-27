package term

import "testing"

func TestWriterWriteAndReset(t *testing.T) {
	w := NewWriter(64)

	w.WriteString("hello")
	if w.Len() != 5 {
		t.Fatalf("Len() = %d, want 5", w.Len())
	}

	w.WriteByte(' ')
	w.Write([]byte("world"))
	if w.Len() != 11 {
		t.Fatalf("Len() = %d, want 11", w.Len())
	}

	got := string(w.Bytes())
	if got != "hello world" {
		t.Fatalf("Bytes() = %q, want %q", got, "hello world")
	}

	w.Reset()
	if w.Len() != 0 {
		t.Fatalf("after Reset: Len() = %d, want 0", w.Len())
	}
	if w.Cap() != 64 {
		t.Fatalf("after Reset: Cap() = %d, want 64", w.Cap())
	}
}

func TestWriterGrow(t *testing.T) {
	w := NewWriter(16)

	w.Grow(32)
	if w.Cap() < 32 {
		t.Fatalf("Cap() = %d after Grow(32), want >= 32", w.Cap())
	}

	// Grow to smaller should be no-op.
	oldCap := w.Cap()
	w.Grow(8)
	if w.Cap() != oldCap {
		t.Fatalf("Cap() changed after Grow(8): %d -> %d", oldCap, w.Cap())
	}
}

func TestWriterAutoGrow(t *testing.T) {
	w := NewWriter(4)

	// Write more than initial capacity.
	w.WriteString("hello world")
	if w.Len() != 11 {
		t.Fatalf("Len() = %d, want 11", w.Len())
	}
	if string(w.Bytes()) != "hello world" {
		t.Fatalf("Bytes() = %q", w.Bytes())
	}
}

func TestWriterWriteStringNoAlloc(t *testing.T) {
	w := NewWriter(256)
	s := "test string that should not allocate"

	allocs := testing.AllocsPerRun(100, func() {
		w.Reset()
		w.WriteString(s)
	})
	if allocs > 0 {
		t.Fatalf("WriteString: %.0f allocs/op, want 0", allocs)
	}
}

func TestWriterWriteNoAlloc(t *testing.T) {
	w := NewWriter(256)
	p := []byte("test bytes that should not allocate")

	allocs := testing.AllocsPerRun(100, func() {
		w.Reset()
		w.Write(p)
	})
	if allocs > 0 {
		t.Fatalf("Write: %.0f allocs/op, want 0", allocs)
	}
}

func BenchmarkWriterWrite(b *testing.B) {
	w := NewWriter(4096)
	data := []byte("\x1b[38;2;255;128;64m█\x1b[0m")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Reset()
		for j := 0; j < 100; j++ {
			w.Write(data)
		}
	}
}

func BenchmarkWriterWriteString(b *testing.B) {
	w := NewWriter(4096)
	s := "\x1b[38;2;255;128;64m█\x1b[0m"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Reset()
		for j := 0; j < 100; j++ {
			w.WriteString(s)
		}
	}
}
