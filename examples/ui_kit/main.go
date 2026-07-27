// ui_kit demonstrates every fluint subsystem working together:
// layout, widgets (Text, Button, List, TextInput), style, animation,
// and event routing via the Router. Events are simulated to show the
// full interaction flow.
//
// Run:  go run ./examples/ui_kit
package main

import (
	"fmt"
	"time"

	"github.com/kshishtovsky/fluint/anim"
	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/router"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/style"
	"github.com/kshishtovsky/fluint/widgets"
)

const (
	cols = 60
	rows = 20
	fps  = 10 // slow enough to read output
)

func main() {
	// ── Styles ──────────────────────────────────────────────────────
	titleSt := style.New().Foreground(style.Cyan).Background(style.Black).Bold()
	headerSt := style.New().Foreground(style.Yellow).Bold()
	btnSt := style.New().Foreground(style.Black).Background(style.Green).Bold()
	listSt := style.New().Foreground(style.White).Background(style.Black)
	inputSt := style.New().Foreground(style.Black).Background(style.White)
	statusSt := style.New().Foreground(style.DarkGray)

	// ── Widgets ─────────────────────────────────────────────────────
	title := widgets.NewText("  fluint UI Kit — Full Demo  ", widgets.WithStyle(titleSt))
	header1 := widgets.NewText("[ Buttons ]", widgets.WithStyle(headerSt))
	btnA := widgets.NewButton("  Action  ", widgets.WithStyle(btnSt))
	btnB := widgets.NewButton("  Cancel  ", widgets.WithStyle(btnSt))
	header2 := widgets.NewText("[ List ]", widgets.WithStyle(headerSt))

	items := make([]widgets.ListItem, 20)
	for i := range items {
		items[i] = widgets.ListItem{Text: fmt.Sprintf("  Item %02d", i+1)}
	}
	list := widgets.NewList(items, widgets.WithStyle(listSt))
	list.SetSelected(3)

	header3 := widgets.NewText("[ Input ]", widgets.WithStyle(headerSt))
	input := widgets.NewTextInput("Hello, fluint!", widgets.WithStyle(inputSt))

	var status *widgets.Text

	// ── Layout ──────────────────────────────────────────────────────
	// Column layout: title(1) + btn-row(1) + header2(1) + list(6) + header3(1) + input(1) + status(1) + pad
	root := &layout.Container{
		Dir: layout.Column,
		Children: []layout.Child{
			{Node: layout.Leaf{}, Basis: 1}, // title
			{Node: layout.Leaf{}, Basis: 1}, // buttons header
			{Node: layout.Leaf{}, Basis: 1}, // btn row
			{Node: layout.Leaf{}, Basis: 1}, // list header
			{Node: layout.Leaf{}, Basis: 6}, // list
			{Node: layout.Leaf{}, Basis: 1}, // input header
			{Node: layout.Leaf{}, Basis: 1}, // input
			{Node: layout.Leaf{}, Grow: 1},  // status (fills remaining)
		},
	}
	rects := make([]layout.Rect, 0, 16)
	rects = root.Measure(cols, rows, rects)
	// Container produces child rects at even indices.
	rTitle := rects[0]
	rBtnHeader := rects[2]
	rBtnRow := rects[4]
	rListHeader := rects[6]
	rList := rects[8]
	rInputHeader := rects[10]
	rInput := rects[12]
	rStatus := rects[14]

	// Buttons share the row — use a sub-layout for horizontal distribution.
	btnLayout := &layout.Container{
		Dir: layout.Row,
		Children: []layout.Child{
			{Node: layout.Leaf{}, Grow: 1},
			{Node: layout.Leaf{}, Grow: 1},
		},
	}
	btnRects := make([]layout.Rect, 0, 4)
	btnRects = btnLayout.Measure(rBtnRow.Width, rBtnRow.Height, btnRects)
	// Offset to world coords.
	rBtnA := layout.Rect{
		X: rBtnRow.X + btnRects[0].X, Y: rBtnRow.Y + btnRects[0].Y,
		Width: btnRects[0].Width, Height: btnRects[0].Height,
	}
	rBtnB := layout.Rect{
		X: rBtnRow.X + btnRects[2].X, Y: rBtnRow.Y + btnRects[2].Y,
		Width: btnRects[2].Width, Height: btnRects[2].Height,
	}

	// ── Router ──────────────────────────────────────────────────────
	r := router.New()
	r.Register(btnA)
	r.Register(btnB)
	r.Register(list)
	r.Register(input)
	r.Focus(btnA)

	// ── Animation: pulse the focused widget's status indicator ──────
	tl := anim.NewTimeline()
	forward := true
	var newPulse func() *anim.Tween
	newPulse = func() *anim.Tween {
		s, e := 0.0, 1.0
		if !forward {
			s, e = 1.0, 0.0
		}
		return &anim.Tween{
			Start: s, End: e, Duration: 800 * time.Millisecond, Ease: anim.InOutQuad,
			OnUpdate: func(v float64) { _ = v },
			OnComplete: func() {
				forward = !forward
				tl.Add(newPulse())
			},
		}
	}
	tl.Add(newPulse())

	// ── Simulated event sequence ────────────────────────────────────
	type step struct {
		after time.Duration
		name  string
		fn    func()
	}
	start := time.Now()
	steps := []step{
		{0, "Focus btnA", func() { r.Focus(btnA) }},
		{500 * time.Millisecond, "Tab → btnB", func() { r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyTab}) }},
		{1 * time.Second, "Tab → list", func() { r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyTab}) }},
		{1500 * time.Millisecond, "List Down x3", func() {
			r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyDown})
			r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyDown})
			r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyDown})
		}},
		{2 * time.Second, "List Enter (select item)", func() {
			r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyEnter})
		}},
		{2500 * time.Millisecond, "Tab → input", func() { r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyTab}) }},
		{3 * time.Second, "Input: type '!'", func() {
			r.DispatchKey(widgets.KeyEvent{Rune: '!'})
		}},
		{3500 * time.Millisecond, "Input: Left x2", func() {
			r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyLeft})
			r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyLeft})
		}},
		{4 * time.Second, "Input: insert '~'", func() {
			r.DispatchKey(widgets.KeyEvent{Rune: '~'})
		}},
		{4500 * time.Millisecond, "Input: Backspace", func() {
			r.DispatchKey(widgets.KeyEvent{Code: widgets.KeyBackspace})
		}},
		{5 * time.Second, "Click btnA (mouse)", func() {
			r.DispatchMouse(widgets.MouseEvent{
				Button: widgets.MouseLeft, Action: widgets.MousePress,
				X: rBtnA.X + 2, Y: rBtnA.Y,
			})
		}},
		{5500 * time.Millisecond, "Done", func() {}},
	}

	stepIdx := 0

	// ── Render loop ─────────────────────────────────────────────────
	buf := buffer.NewBuffer(cols, rows)
	ctx := viewport.RenderCtx{Buf: buf}
	frameInterval := time.Second / fps
	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	frame := 0
	fmt.Println("ui_kit: 60x20 comprehensive demo")
	fmt.Println("========================================")

	for range ticker.C {
		frame++
		elapsed := time.Since(start)
		tl.Update(frameInterval)

		// Fire pending events.
		for stepIdx < len(steps) && elapsed >= steps[stepIdx].after {
			s := steps[stepIdx]
			fmt.Printf("  [%5.1fs] %s\n", elapsed.Seconds(), s.name)
			s.fn()
			stepIdx++
		}

		// Update status line.
		focused := "none"
		switch {
		case r.Focused() == btnA:
			focused = "btnA"
		case r.Focused() == btnB:
			focused = "btnB"
		case r.Focused() == list:
			focused = fmt.Sprintf("list[%d]", list.Selected())
		case r.Focused() == input:
			focused = "input"
		}
		status = widgets.NewText(
			fmt.Sprintf("  Focus: %-12s | Tab=cycle | Arrows=navigate", focused),
			widgets.WithStyle(statusSt),
		)

		// Render.
		buf.Clear()
		title.Render(ctx, rTitle)
		header1.Render(ctx, rBtnHeader)
		btnA.Render(ctx, rBtnA)
		btnB.Render(ctx, rBtnB)
		header2.Render(ctx, rListHeader)
		list.Render(ctx, rList)
		header3.Render(ctx, rInputHeader)
		input.Render(ctx, rInput)
		status.Render(ctx, rStatus)

		// Print buffer snapshot every 2 seconds.
		if frame%(fps*2) == 0 {
			fmt.Printf("\n  --- frame %d @ %s ---\n", frame, elapsed.Truncate(time.Millisecond))
			printBuffer(buf)
		}

		if stepIdx >= len(steps) {
			break
		}
	}

	// Final output.
	fmt.Println("\n========================================")
	fmt.Println("Final state:")
	fmt.Printf("  Input text: %q\n", input.Text())
	fmt.Printf("  List selected: %d (%q)\n", list.Selected(), list.Items()[list.Selected()].Text)
	printBuffer(buf)
	fmt.Println("========================================")
	fmt.Printf("ui_kit: %d frames in %v\n", frame, time.Since(start).Truncate(time.Millisecond))
}

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
