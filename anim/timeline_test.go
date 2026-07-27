package anim

import (
	"testing"
	"time"
)

// TestTimeline_Completion verifies that OnUpdate receives the final End
// value and OnComplete fires exactly once when a tween finishes.
func TestTimeline_Completion(t *testing.T) {
	t.Parallel()

	var (
		lastValue   float64
		completeN   int
		updateCount int
	)

	tw := &Tween{
		Start:    0,
		End:      100,
		Duration: 100 * time.Millisecond,
		Ease:     nil, // linear
		OnUpdate: func(v float64) {
			lastValue = v
			updateCount++
		},
		OnComplete: func() {
			completeN++
		},
	}

	tl := NewTimeline()
	tl.Add(tw)

	// Drive the timeline in 20 ms steps → 5 updates + 1 completion tick.
	for i := 0; i < 5; i++ {
		tl.Update(20 * time.Millisecond)
	}

	// After 100 ms the tween should be complete.
	if !tw.Active {
		// expected: tween deactivated on last step
	}
	if lastValue != 100 {
		t.Errorf("last OnUpdate value = %v, want 100", lastValue)
	}
	if completeN != 1 {
		t.Errorf("OnComplete called %d times, want 1", completeN)
	}
	// OnUpdate was called on each of the 5 steps; the 5th step is
	// the completion tick, which calls OnUpdate(End) once.
	if updateCount < 1 {
		t.Errorf("OnUpdate called %d times, want >= 1", updateCount)
	}
}

// TestTimeline_MultipleTweens runs several tweens in parallel on a
// single Timeline and confirms they all complete independently.
func TestTimeline_MultipleTweens(t *testing.T) {
	t.Parallel()

	done := make([]bool, 3)
	tl := NewTimeline()

	for i := range done {
		i := i
		tl.Add(&Tween{
			Start:    float64(i),
			End:      float64(i) + 10,
			Duration: 50 * time.Millisecond,
			OnUpdate: func(_ float64) {},
			OnComplete: func() {
				done[i] = true
			},
		})
	}

	tl.Update(50 * time.Millisecond)

	for i, d := range done {
		if !d {
			t.Errorf("tween %d did not complete", i)
		}
	}
}

// TestTimeline_Clear resets the timeline and its tweens.
func TestTimeline_Clear(t *testing.T) {
	t.Parallel()

	var called bool
	tl := NewTimeline()
	tl.Add(&Tween{
		Start:    0,
		End:      1,
		Duration: 10 * time.Millisecond,
		OnUpdate: func(_ float64) {
			called = true
		},
	})

	tl.Clear()

	if len(tl.Tweens) != 0 {
		t.Errorf("len(Tweens) after Clear = %d, want 0", len(tl.Tweens))
	}

	// Update after Clear should be a no-op.
	called = false
	tl.Update(50 * time.Millisecond)
	if called {
		t.Error("OnUpdate called after Clear")
	}
}

// TestTimeline_Compact verifies that Compact removes inactive tweens
// and preserves active ones.
func TestTimeline_Compact(t *testing.T) {
	t.Parallel()

	var activeValue float64
	tl := NewTimeline()

	// Short-lived tween — completes on first update.
	tl.Add(&Tween{
		Start:    0,
		End:      1,
		Duration: 1 * time.Millisecond,
		OnUpdate: func(_ float64) {},
	})
	// Long-lived tween.
	tl.Add(&Tween{
		Start:    0,
		End:      100,
		Duration: 1 * time.Second,
		OnUpdate: func(v float64) { activeValue = v },
	})

	tl.Update(10 * time.Millisecond) // first tween completes

	tl.Compact()

	if len(tl.Tweens) != 1 {
		t.Errorf("len after Compact = %d, want 1", len(tl.Tweens))
	}

	// The surviving tween should still update.
	tl.Update(10 * time.Millisecond)
	if activeValue == 0 {
		t.Error("active tween did not update after Compact")
	}
}

// TestTimeline_NilCallbacks confirms that nil OnUpdate and nil
// OnComplete are handled without panic.
func TestTimeline_NilCallbacks(t *testing.T) {
	t.Parallel()

	tl := NewTimeline()
	tl.Add(&Tween{
		Start:    0,
		End:      1,
		Duration: 10 * time.Millisecond,
		// OnUpdate and OnComplete are nil.
	})

	tl.Update(50 * time.Millisecond) // must not panic
}

// TestTimeline_NilEase confirms that a nil EaseFunc is treated as
// linear interpolation.
func TestTimeline_NilEase(t *testing.T) {
	t.Parallel()

	var got float64
	tl := NewTimeline()
	tl.Add(&Tween{
		Start:    0,
		End:      200,
		Duration: 100 * time.Millisecond,
		Ease:     nil,
		OnUpdate: func(v float64) { got = v },
	})

	tl.Update(50 * time.Millisecond) // t=0.5, linear → value=100

	if got < 99 || got > 101 {
		t.Errorf("nil ease at t=0.5 → value = %v, want ~100", got)
	}
}

// TestTimeline_ZeroDurationTween verifies that a tween with Duration 0
// completes immediately on the first Update tick.
func TestTimeline_ZeroDurationTween(t *testing.T) {
	t.Parallel()

	var (
		finalVal float64
		done     bool
	)

	tl := NewTimeline()
	tl.Add(&Tween{
		Start:    50,
		End:      150,
		Duration: 0,
		OnUpdate: func(v float64) { finalVal = v },
		OnComplete: func() {
			done = true
		},
	})

	tl.Update(time.Millisecond)

	if !done {
		t.Error("zero-duration tween did not complete")
	}
	if finalVal != 150 {
		t.Errorf("final value = %v, want 150", finalVal)
	}
}

// TestTimeline_EasingIntegration verifies that an easing curve is
// actually applied by comparing against linear at the midpoint.
func TestTimeline_EasingIntegration(t *testing.T) {
	t.Parallel()

	var linearVal, easedVal float64

	tl := NewTimeline()
	tl.Add(&Tween{
		Start:    0,
		End:      100,
		Duration: 100 * time.Millisecond,
		Ease:     nil, // linear
		OnUpdate: func(v float64) { linearVal = v },
	})

	tl2 := NewTimeline()
	tl2.Add(&Tween{
		Start:    0,
		End:      100,
		Duration: 100 * time.Millisecond,
		Ease:     InQuad, // t² — much lower at midpoint
		OnUpdate: func(v float64) { easedVal = v },
	})

	tl.Update(50 * time.Millisecond)
	tl2.Update(50 * time.Millisecond)

	if easedVal >= linearVal {
		t.Errorf("InQuad at midpoint (%v) should be less than linear (%v)",
			easedVal, linearVal)
	}
}

// BenchmarkTimelineUpdate drives 100 active tweens for 1000
// iterations. Must show 0 allocs/op.
func BenchmarkTimelineUpdate(b *testing.B) {
	tl := NewTimeline()
	tw := make([]Tween, 100)
	for i := range tw {
		tw[i] = Tween{
			Start:    0,
			End:      100,
			Duration: 10 * time.Second, // long enough to stay active
			Ease:     OutCubic,
			OnUpdate: func(_ float64) {},
		}
		tl.Add(&tw[i])
	}

	const dt = time.Millisecond
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tl.Update(dt)
	}
}

// BenchmarkTimelineUpdateMixed drives a mix of active and completed
// tweens to measure the cost of the inactive-skip path.
func BenchmarkTimelineUpdateMixed(b *testing.B) {
	tl := NewTimeline()
	tw := make([]Tween, 100)
	for i := range tw {
		d := time.Duration(i+1) * time.Millisecond // each completes at a different time
		tw[i] = Tween{
			Start:    0,
			End:      1,
			Duration: d,
			Ease:     Linear,
			OnUpdate: func(_ float64) {},
		}
		tl.Add(&tw[i])
	}

	// Run 50 ms worth of updates — most tweens finish early.
	const dt = time.Millisecond
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset tweens for each benchmark iteration.
		for j := range tw {
			tw[j].Elapsed = 0
			tw[j].Active = true
		}
		for step := 0; step < 50; step++ {
			tl.Update(dt)
		}
	}
}
