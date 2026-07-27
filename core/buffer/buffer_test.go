package buffer

import "testing"

func TestNewBuffer_Dimensions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		w, h    int
		wantLen int
	}{
		{"normal", 10, 5, 50},
		{"wide", 80, 24, 1920},
		{"unit", 1, 1, 1},
		{"empty-zero", 0, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := NewBuffer(tc.w, tc.h)
			if buf == nil {
				t.Fatal("NewBuffer returned nil")
			}
			if buf.Width != tc.w {
				t.Errorf("Width = %d, want %d", buf.Width, tc.w)
			}
			if buf.Height != tc.h {
				t.Errorf("Height = %d, want %d", buf.Height, tc.h)
			}
			if len(buf.Cells) != tc.wantLen {
				t.Errorf("len(Cells) = %d, want %d", len(buf.Cells), tc.wantLen)
			}
		})
	}
}

func TestBufferClone_DeepCopy(t *testing.T) {
	t.Parallel()

	src := NewBuffer(4, 3)
	src.SetCell(2, 1, Cell{Rune: 'Q'})

	clone := src.Clone()
	if clone == nil {
		t.Fatal("Clone returned nil")
	}
	if clone == src {
		t.Fatal("Clone returned the same pointer")
	}
	if len(clone.Cells) > 0 && len(src.Cells) > 0 &&
		&clone.Cells[0] == &src.Cells[0] {
		t.Fatal("Clone shares the Cells backing array")
	}

	src.SetCell(0, 0, Cell{Rune: 'X'})
	if got := clone.GetCell(0, 0); got != (Cell{}) {
		t.Errorf("clone picked up source mutation: clone[0,0] = %+v, want zero", got)
	}

	clone.SetCell(3, 2, Cell{Rune: 'Y'})
	if got := src.GetCell(3, 2); got != (Cell{}) {
		t.Errorf("original picked up clone mutation: src[3,2] = %+v, want zero", got)
	}

	if clone.Width != src.Width || clone.Height != src.Height {
		t.Errorf("clone dims = (%d,%d), want (%d,%d)",
			clone.Width, clone.Height, src.Width, src.Height)
	}

	if got := (*Buffer)(nil).Clone(); got != nil {
		t.Errorf("Clone of nil = %+v, want nil", got)
	}
}

func BenchmarkBufferSetCell(b *testing.B) {
	buf := NewBuffer(80, 24)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.SetCell(i%80, (i/80)%24, Cell{Rune: 'X'})
	}
}
