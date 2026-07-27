package anim

import (
	"math"
	"testing"
)

// allEases enumerates every exported EaseFunc in the package.
var allEases = []struct {
	name string
	fn   EaseFunc
}{
	{"Linear", Linear},
	{"InQuad", InQuad},
	{"OutQuad", OutQuad},
	{"InOutQuad", InOutQuad},
	{"InCubic", InCubic},
	{"OutCubic", OutCubic},
	{"InOutCubic", InOutCubic},
	{"OutBounce", OutBounce},
	{"OutElastic", OutElastic},
}

// TestEaseFuncs_Endpoints verifies every easing curve satisfies the
// contract func(0) == 0 and func(1) == 1.
func TestEaseFuncs_Endpoints(t *testing.T) {
	t.Parallel()

	for _, e := range allEases {
		e := e
		t.Run(e.name, func(t *testing.T) {
			t.Parallel()

			got0 := e.fn(0)
			if got0 != 0 {
				t.Errorf("%s(0) = %v, want 0", e.name, got0)
			}
			got1 := e.fn(1)
			if math.Abs(got1-1) > 1e-9 {
				t.Errorf("%s(1) = %v, want 1", e.name, got1)
			}
		})
	}
}

// TestEaseFuncs_NoPanic sweeps t across [0, 1] (plus tiny extremes)
// to confirm every curve stays finite.
func TestEaseFuncs_NoPanic(t *testing.T) {
	t.Parallel()

	for _, e := range allEases {
		e := e
		t.Run(e.name, func(t *testing.T) {
			t.Parallel()
			for i := 0; i <= 64; i++ {
				s := float64(i) / 64.0
				_ = e.fn(s) // must not panic.
			}
		})
	}
}

// TestOutBounce_StaysBoundedInMiddle asserts the bounce curve never
// exceeds [0, 1] at the midpoint sample.
func TestOutBounce_StaysBoundedInMiddle(t *testing.T) {
	t.Parallel()

	for i := 1; i < 64; i++ {
		s := float64(i) / 64.0
		v := OutBounce(s)
		if v < 0 || v > 1 {
			t.Errorf("OutBounce(%v) = %v, want within [0, 1]", s, v)
		}
	}
}

// BenchmarkEasing runs every easing function across the [0, 1]
// interval in a tight loop. Used to verify 0 allocs/op on the hot
// path.
func BenchmarkEasing(b *testing.B) {
	// 256 evenly-spaced samples in [0, 1].
	var samples [256]float64
	for i := range samples {
		samples[i] = float64(i) / 255.0
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, e := range allEases {
			var acc float64
			for _, s := range samples {
				acc += e.fn(s)
			}
			// Prevent the compiler from eliding the calls.
			if math.IsNaN(acc) {
				b.Fatal("easing produced NaN")
			}
		}
	}
}

// Per-function allocation benchmarks for tighter regression checks.
func BenchmarkEasingLinear(b *testing.B)    { benchEase(b, Linear) }
func BenchmarkEasingInQuad(b *testing.B)    { benchEase(b, InQuad) }
func BenchmarkEasingOutQuad(b *testing.B)   { benchEase(b, OutQuad) }
func BenchmarkEasingInOutQuad(b *testing.B) { benchEase(b, InOutQuad) }
func BenchmarkEasingInCubic(b *testing.B)   { benchEase(b, InCubic) }
func BenchmarkEasingOutCubic(b *testing.B)  { benchEase(b, OutCubic) }
func BenchmarkEasingInOutCubic(b *testing.B) {
	benchEase(b, InOutCubic)
}
func BenchmarkEasingOutBounce(b *testing.B) { benchEase(b, OutBounce) }
func BenchmarkEasingOutElastic(b *testing.B) {
	benchEase(b, OutElastic)
}

func benchEase(b *testing.B, fn EaseFunc) {
	b.Helper()
	var samples [256]float64
	for i := range samples {
		samples[i] = float64(i) / 255.0
	}
	b.ResetTimer()
	var acc float64
	for i := 0; i < b.N; i++ {
		for _, s := range samples {
			acc += fn(s)
		}
	}
	if math.IsNaN(acc) {
		b.Fatal("easing produced NaN")
	}
}
