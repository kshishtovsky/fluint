package anim

import "time"

// Tween holds the state of a single interpolation between two values
// over a fixed duration. The caller supplies callback functions
// (OnUpdate, OnComplete) that the Timeline invokes during Update.
//
// A zero-value Tween is inactive; set Active = true after populating
// the fields to schedule it for the next Update tick.
//
// Memory contract: Tween is a plain struct with no pointers of its
// own (func values and EaseFunc are word-sized scalar descriptors).
// The Timeline owns a flat []*Tween slice; Update never re-allocates
// or appends to it, so the hot path is strictly 0 B/op.
type Tween struct {
	// Start is the initial interpolation value.
	Start float64

	// End is the target interpolation value.
	End float64

	// Duration is the wall-clock length of this tween.
	Duration time.Duration

	// Ease is the easing curve applied to normalised progress.
	// Use nil for linear interpolation.
	Ease EaseFunc

	// OnUpdate is called every frame while the tween is active.
	// The argument is the interpolated value after easing.
	OnUpdate func(value float64)

	// OnComplete is called once when the tween finishes. May be nil.
	OnComplete func()

	// --- internal state (managed by Timeline) ---

	// Elapsed accumulates the time spent so far.
	Elapsed time.Duration

	// Active is true while the tween is scheduled for updates.
	Active bool
}

// Timeline manages a set of Tweens and advances them by discrete time
// steps each frame.
//
// Memory contract: the Tweens slice is grown only by Add (setup path).
// Update never re-slices, appends, or allocates — it walks the
// existing slice and calls callbacks for active tweens. Completed
// tweens remain in the slice but are marked inactive and skipped.
// Call Compact before re-using a long-lived Timeline to trim dead
// entries, or Clear to reset completely.
type Timeline struct {
	Tweens []*Tween
}

// NewTimeline returns an empty Timeline.
func NewTimeline() *Timeline {
	return &Timeline{}
}

// Add schedules a Tween for updates. The tween's Elapsed is reset to
// zero and Active is set to true automatically.
//
// Add may allocate if the underlying slice needs to grow; it is a
// setup-time operation, not a per-frame hot path.
func (tl *Timeline) Add(tw *Tween) {
	if tw == nil {
		return
	}
	tw.Elapsed = 0
	tw.Active = true
	tl.Tweens = append(tl.Tweens, tw)
}

// Update advances every active tween by dt. For each active tween:
//
//  1. Elapsed += dt.
//  2. t = Elapsed / Duration, clamped to [0, 1].
//  3. easedT = Ease(t)  (or t when Ease is nil).
//  4. value = Start + (End - Start) * easedT.
//  5. OnUpdate(value) is invoked.
//  6. When Elapsed >= Duration, value is pinned to End, OnUpdate(End)
//     is called, OnComplete() fires (if non-nil), and Active is set
//     to false.
//
// The method touches only the existing Tweens slice — no allocation,
// no re-slicing, no map lookup.
func (tl *Timeline) Update(dt time.Duration) {
	for _, tw := range tl.Tweens {
		if !tw.Active {
			continue
		}

		tw.Elapsed += dt

		// Normalised progress, clamped to [0, 1].
		t := float64(tw.Elapsed) / float64(tw.Duration)
		if t > 1 {
			t = 1
		}

		// Apply easing curve.
		easedT := t
		if tw.Ease != nil {
			easedT = tw.Ease(t)
		}

		// Interpolate.
		value := tw.Start + (tw.End-tw.Start)*easedT

		// Check completion first to avoid a double OnUpdate call on
		// the final frame.
		if tw.Elapsed >= tw.Duration {
			// Pin to exact End value and fire callbacks once.
			if tw.OnUpdate != nil {
				tw.OnUpdate(tw.End)
			}
			tw.Active = false
			if tw.OnComplete != nil {
				tw.OnComplete()
			}
		} else {
			if tw.OnUpdate != nil {
				tw.OnUpdate(value)
			}
		}
	}
}

// Clear resets every tween to inactive and zeroes elapsed time. The
// underlying slice capacity is preserved to avoid re-allocation on
// the next Add cycle.
func (tl *Timeline) Clear() {
	for _, tw := range tl.Tweens {
		tw.Active = false
		tw.Elapsed = 0
	}
	tl.Tweens = tl.Tweens[:0]
}

// Compact removes inactive tweens from the slice without growing a
// new backing array (in-place filter). Call periodically on
// long-lived Timelines to keep iteration tight.
func (tl *Timeline) Compact() {
	dst := 0
	for _, tw := range tl.Tweens {
		if tw.Active {
			tl.Tweens[dst] = tw
			dst++
		}
	}
	// Zero the trailing pointers to allow GC of dead Tweens.
	for i := dst; i < len(tl.Tweens); i++ {
		tl.Tweens[i] = nil
	}
	tl.Tweens = tl.Tweens[:dst]
}
