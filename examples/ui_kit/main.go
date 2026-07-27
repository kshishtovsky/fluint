// ui_kit demonstrates all fluint subsystems working together:
// layout, widgets, style, and animation. A button pulses its background
// color via anim.Timeline while the layout engine distributes space.
//
// Run:  go run ./examples/ui_kit
package main

import (
	"fmt"
	"time"

	"github.com/kshishtovsky/fluint/anim"
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/style"
	"github.com/kshishtovsky/fluint/widgets"
)

const (
	width  = 40
	height = 12
	fps    = 60
)

func main() {
	// ── Style definitions ───────────────────────────────────────────
	titleStyle := style.New().Foreground(style.Cyan).Background(style.Black).Bold()
	textStyle := style.New().Foreground(style.White)
	btnStyle := style.New().Foreground(style.Black).Background(style.Green).Bold()
	btnPulseStyle := style.New().Foreground(style.Black).Background(style.Yellow).Bold()

	// ── Widget tree ─────────────────────────────────────────────────
	title := widgets.NewText("  fluint UI Kit Demo  ", widgets.WithStyle(titleStyle))
	body := widgets.NewText("Widgets + Layout + Anim", widgets.WithStyle(textStyle))
	button := widgets.NewButton("[ Pulse ]", widgets.WithStyle(btnStyle))

	// ── Layout (Column: title → body → button) ─────────────────────
	root := &layout.Container{
		Dir: layout.Column,
		Children: []layout.Child{
			{Node: layout.Leaf{}, Basis: 1},  // title row
			{Node: layout.Leaf{}, Basis: 1},  // body text
			{Node: layout.Leaf{}, Grow: 1},   // button fills remaining space
		},
	}

	rects := make([]layout.Rect, 0, 6)
	rects = root.Measure(width, height, rects)
	// rects[0,2,4] are the child positions (from Container); rects[1,3,5]
	// are the Leaf-local rects (always 0,0). Use the child-position rects.
	titleRect, bodyRect, btnRect := rects[0], rects[2], rects[4]

	// ── Animation: pulse button background green ↔ yellow ──────────
	tl := anim.NewTimeline()
	pulseValue := 0.0
	forward := true

	var newPulseTween func() *anim.Tween
	newPulseTween = func() *anim.Tween {
		startV, endV := 0.0, 1.0
		if !forward {
			startV, endV = 1.0, 0.0
		}
		return &anim.Tween{
			Start:    startV,
			End:      endV,
			Duration: 2 * time.Second,
			Ease:     anim.InOutQuad,
			OnUpdate: func(v float64) {
				pulseValue = v
				r := uint32(v * 255)
				bg := style.Color(0x0000FF00 | (r << 16))
				button.SetStyle(btnPulseStyle.Background(bg))
			},
			OnComplete: func() {
				forward = !forward
				tl.Add(newPulseTween())
			},
		}
	}
	tl.Add(newPulseTween())

	// ── Render context ───────────────────────────────────────────────
	buf := buffer.NewBuffer(width, height)
	ctx := viewport.RenderCtx{Buf: buf} // no viewport — screen-space
	frameInterval := time.Second / fps
	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	frame := 0
	start := time.Now()

	fmt.Println("ui_kit: 40x12 buffer, 60 fps, pulse animation")
	fmt.Println("------------------------------------------------")

	for range ticker.C {
		frame++
		tl.Update(frameInterval)

		// Clear and render.
		buf.Clear()
		title.Render(ctx, titleRect)
		body.Render(ctx, bodyRect)
		button.Render(ctx, btnRect)

		// Log every 30 frames (~0.5 s).
		if frame%30 == 0 {
			btnBg := button.Style().BG()
			fmt.Printf("  frame %3d  pulse=%.2f  btnBg=0x%06X  elapsed=%v\n",
				frame, pulseValue, btnBg, time.Since(start).Truncate(time.Millisecond))
		}

		// Stop after 4 seconds (2 full pulse cycles).
		if time.Since(start) > 4*time.Second {
			break
		}
	}

	// Print the final buffer content.
	fmt.Println("------------------------------------------------")
	fmt.Println("Final frame buffer:")
	printBuffer(buf)
	fmt.Printf("ui_kit: completed %d frames in %v\n",
		frame, time.Since(start).Truncate(time.Millisecond))
}

// printBuffer renders the buffer contents as ASCII art.
func printBuffer(buf *buffer.Buffer) {
	for y := 0; y < buf.Height; y++ {
		fmt.Print("  |")
		for x := 0; x < buf.Width; x++ {
			c := buf.GetCell(x, y)
			if c.Rune == 0 {
				fmt.Print(" ")
			} else {
				fmt.Printf("%c", c.Rune)
			}
		}
		fmt.Println("|")
	}
}
