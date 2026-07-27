package viewport

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()
	v := New(80, 24)
	if v.Width != 80 || v.Height != 24 {
		t.Errorf("New(80,24): got %dx%d, want 80x24", v.Width, v.Height)
	}
	if v.OffsetX != 0 || v.OffsetY != 0 {
		t.Errorf("offset: got (%d,%d), want (0,0)", v.OffsetX, v.OffsetY)
	}
}

func TestScroll(t *testing.T) {
	t.Parallel()
	v := New(80, 24)
	v.Scroll(10, 20)
	if v.OffsetX != 10 || v.OffsetY != 20 {
		t.Errorf("after Scroll(10,20): got (%d,%d), want (10,20)", v.OffsetX, v.OffsetY)
	}
	v.Scroll(-5, -10)
	if v.OffsetX != 5 || v.OffsetY != 10 {
		t.Errorf("after Scroll(-5,-10): got (%d,%d), want (5,10)", v.OffsetX, v.OffsetY)
	}
	// Negative clamp.
	v.Scroll(-100, -100)
	if v.OffsetX != 0 || v.OffsetY != 0 {
		t.Errorf("after Scroll(-100,-100): got (%d,%d), want (0,0)", v.OffsetX, v.OffsetY)
	}
}

func TestCenter(t *testing.T) {
	t.Parallel()
	v := New(10, 10)
	v.Center(50, 50)
	if v.OffsetX != 45 || v.OffsetY != 45 {
		t.Errorf("Center(50,50): got (%d,%d), want (45,45)", v.OffsetX, v.OffsetY)
	}
	// Negative clamp.
	v.Center(2, 2)
	if v.OffsetX != 0 || v.OffsetY != 0 {
		t.Errorf("Center(2,2): got (%d,%d), want (0,0)", v.OffsetX, v.OffsetY)
	}
}

func TestResize(t *testing.T) {
	t.Parallel()
	v := New(80, 24)
	v.Resize(120, 40)
	if v.Width != 120 || v.Height != 40 {
		t.Errorf("Resize(120,40): got %dx%d, want 120x40", v.Width, v.Height)
	}
}

func TestVisible(t *testing.T) {
	t.Parallel()
	v := New(10, 10) // visible: [0,10) x [0,10)
	v.OffsetX = 5
	v.OffsetY = 5 // visible: [5,15) x [5,15)

	cases := []struct {
		wx, wy, ww, wh int
		want            bool
	}{
		{0, 0, 3, 3, false},  // fully outside (top-left)
		{20, 20, 3, 3, false}, // fully outside (bottom-right)
		{5, 5, 3, 3, true},   // fully inside
		{3, 3, 5, 5, true},   // partially overlapping top-left
		{12, 12, 5, 5, true}, // partially overlapping bottom-right
		{0, 0, 100, 100, true}, // huge rect covers viewport
	}
	for _, tc := range cases {
		got := v.Visible(tc.wx, tc.wy, tc.ww, tc.wh)
		if got != tc.want {
			t.Errorf("Visible(%d,%d,%d,%d): got %v, want %v",
				tc.wx, tc.wy, tc.ww, tc.wh, got, tc.want)
		}
	}
}

func TestScreenXY(t *testing.T) {
	t.Parallel()
	v := New(10, 10)
	v.OffsetX = 50
	v.OffsetY = 30

	if got := v.ScreenX(55); got != 5 {
		t.Errorf("ScreenX(55): got %d, want 5", got)
	}
	if got := v.ScreenY(33); got != 3 {
		t.Errorf("ScreenY(33): got %d, want 3", got)
	}
}
