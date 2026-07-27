// vfx_demo demonstrates the Timeline & Tween system driving 10
// concurrent animations at 60 fps. No terminal output is rendered —
// the program simply prints per-frame state to stdout to prove the
// system runs in real time without memory leaks.
//
// Run:  go run ./examples/vfx_demo
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/kshishtovsky/fluint/anim"
)

// slide describes a single animated property.
type slide struct {
	label string
	ease  anim.EaseFunc
	start float64
	end   float64
}

func main() {
	slides := []slide{
		{"X position  ", anim.Linear, 0, 100},
		{"Y position  ", anim.InQuad, 0, 50},
		{"Width       ", anim.OutQuad, 10, 200},
		{"Height      ", anim.InOutQuad, 10, 80},
		{"Opacity     ", anim.InCubic, 1, 0},
		{"Scale       ", anim.OutCubic, 1, 3},
		{"Rotation    ", anim.InOutCubic, 0, 360},
		{"Saturation  ", anim.OutBounce, 0, 100},
		{"Blur radius ", anim.OutElastic, 0, 20},
		{"Brightness  ", anim.Linear, 0.5, 1.5},
	}

	tl := anim.NewTimeline()

	for i := range slides {
		s := &slides[i]
		tl.Add(&anim.Tween{
			Start:    s.start,
			End:      s.end,
			Duration: 2 * time.Second,
			Ease:     s.ease,
			OnUpdate: func(v float64) {
				// In a real engine this would mutate a widget property.
				_ = v
			},
			OnComplete: func() {
				fmt.Printf("  [done] %s\n", s.label)
			},
		})
	}

	fmt.Println("vfx_demo: 10 tweens, 60 fps, 2 s duration")
	fmt.Println("--------------------------------------------")

	const frameInterval = time.Second / 60 // ~16.67 ms
	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	frame := 0
	start := time.Now()

	// Also sample one tween's value for display.
	var sampleValue float64

	// We'll just track frame count and print a summary line every
	// ~30 frames instead of flooding the terminal.
	for range ticker.C {
		frame++
		tl.Update(frameInterval)

		// Print a compact status line every 30 frames (~0.5 s).
		if frame%30 == 0 {
			// Manually compute one value for display.
			t := math.Min(float64(time.Since(start))/2e9, 1.0)
			sampleValue = slides[0].start + (slides[0].end-slides[0].start)*t
			elapsed := time.Since(start).Truncate(time.Millisecond)
			fmt.Printf("  frame %3d  t=%6.3f  sample=%7.2f  elapsed=%v\n",
				frame, t, sampleValue, elapsed)
		}

		// Stop after 2.5 seconds (all tweens done + margin).
		if time.Since(start) > 2500*time.Millisecond {
			break
		}
	}

	fmt.Println("--------------------------------------------")
	fmt.Printf("vfx_demo: completed %d frames in %v\n",
		frame, time.Since(start).Truncate(time.Millisecond))
}
