// Package anim provides animation primitives for the fluint engine:
// easing functions, a zero-allocation Timeline, and Tween
// interpolation.
//
// All easing functions in this file satisfy the contract:
//
//	func(0.0) == 0.0
//	func(1.0) == 1.0
//
// and are designed to be allocation-free so they can run on the
// per-frame hot path.
package anim

import "math"

// EaseFunc maps a normalised time value t ∈ [0, 1] to an interpolated
// progress value. Implementations must be safe to call from the render
// hot path — allocation-free, panic-free for t in the closed interval
// [0, 1].
type EaseFunc func(t float64) float64

// Named easing curves. Each variable references a package-level
// function so the Go compiler can inline calls through the EaseFunc
// type.

// Linear — constant velocity, no acceleration.
var Linear EaseFunc = linearEase

// InQuad — accelerating from zero velocity. Quadratic ease-in.
var InQuad EaseFunc = inQuadEase

// OutQuad — decelerating to zero velocity. Quadratic ease-out.
var OutQuad EaseFunc = outQuadEase

// InOutQuad — quadratic ease-in/ease-out. Slow start, slow end.
var InOutQuad EaseFunc = inOutQuadEase

// InCubic — accelerating from zero velocity. Cubic ease-in.
var InCubic EaseFunc = inCubicEase

// OutCubic — decelerating to zero velocity. Cubic ease-out.
var OutCubic EaseFunc = outCubicEase

// InOutCubic — cubic ease-in/ease-out. Slow start, slow end.
var InOutCubic EaseFunc = inOutCubicEase

// OutBounce — ease-out with a decaying bounce at the end. Stays
// within [0, 1] except for the small overshoot at the very last
// bounce peak (clamp on read if required).
var OutBounce EaseFunc = outBounceEase

// OutElastic — ease-out with an elastic overshoot at the end. The
// curve briefly exceeds 1.0 and dips below 0.0 around the overshoot
// region — callers that need a bounded output should clamp the result.
var OutElastic EaseFunc = outElasticEase

// ---------------------------------------------------------------------------
// Implementations
//
// Kept as separate package-level functions so the Go compiler can
// inline calls through EaseFunc values. math.Pow is avoided wherever
// a plain multiply or call to math.Sqrt/Sin suffices.
// ---------------------------------------------------------------------------

func linearEase(t float64) float64 {
	return t
}

func inQuadEase(t float64) float64 {
	return t * t
}

func outQuadEase(t float64) float64 {
	// 1 - (1-t)²  →  2t - t²  (cheap, two multiplies).
	return 2*t - t*t
}

func inOutQuadEase(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	// 1 - (-2t + 2)² / 2  →  1 - (2-2t)² / 2.
	u := 2 - 2*t
	return 1 - u*u/2
}

func inCubicEase(t float64) float64 {
	return t * t * t
}

func outCubicEase(t float64) float64 {
	// 1 - (1-t)³ expanded → 3t - 3t² + t³. Multiplies only.
	u := 1 - t
	return 1 - u*u*u
}

func inOutCubicEase(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	// 1 - (-2t + 2)³ / 2.
	u := 2 - 2*t
	return 1 - u*u*u/2
}

// outBounceEase — Penner-style "bounce out" curve. The animation
// overshoots the target at the end in a series of decaying bounces.
//
// Coefficients derived from Penner's easing cheat sheet:
//
//	n1 = 7.5625, d1 = 2.75
//
// Phases:
//
//	t < 1/d1                 → n1 * t * t
//	t < 2/d1                 → n1 * (t -= 1.5/d1) * t + 0.75
//	t < 2.5/d1               → n1 * (t -= 2.25/d1) * t + 0.9375
//	else                     → n1 * (t -= 2.625/d1) * t + 0.984375
func outBounceEase(t float64) float64 {
	const (
		n1 = 7.5625
		d1 = 2.75
	)
	switch {
	case t < 1/d1:
		return n1 * t * t
	case t < 2/d1:
		t -= 1.5 / d1
		return n1*t*t + 0.75
	case t < 2.5/d1:
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	default:
		t -= 2.625 / d1
		return n1*t*t + 0.984375
	}
}

// outElasticEase — Penner "elastic out". Uses sin/cos with a tuned
// period. The curve overshoots 1.0 and undershoots 0.0 near the end.
//
// math.Pow(2, -10t) is the standard form for the elastic envelope and
// has no closed-form arithmetic alternative, so it is the single
// allowed math.Pow call in the package.
func outElasticEase(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	const (
		p = 0.3                 // elastic period
		a = 1.0                 // amplitude
		s = 0.3 / (4 * math.Pi) // p / (4π); since asin(1/a) = π/2 when a=1.
	)
	return a * math.Pow(2, -10*t) * math.Sin((t-s)*(2*math.Pi)/p)
}
