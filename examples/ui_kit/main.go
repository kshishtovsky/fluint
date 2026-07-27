// ui_kit is an interactive terminal UI demo. Run it, then:
//
//   Tab        — cycle focus between widgets
//   Up/Down    — navigate the list
//   Enter      — select list item / press button
//   Left/Right — move cursor in text input
//   Type       — edit text input
//   Backspace  — delete char in text input
//   Escape     — quit
//
// Run:  go run ./examples/ui_kit
package main

import (
	"fmt"
	"os"

	"github.com/kshishtovsky/fluint/core/loop"
	"github.com/kshishtovsky/fluint/core/router"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/internal/platform"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/style"
	"github.com/kshishtovsky/fluint/widgets"
)

func main() {
	tty, err := platform.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open tty: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = tty.Close() }()

	if err := tty.EnterRawMode(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to enter raw mode: %v\n", err)
		os.Exit(1)
	}

	w, h, _ := tty.GetSize()
	if w < 1 || h < 1 {
		w, h = 80, 24
	}

	// Enable mouse, clear screen, hide cursor.
	_, _ = tty.Write([]byte("\x1b[?1000h\x1b[?1006h\x1b[2J\x1b[?25l"))
	defer func() { _, _ = tty.Write([]byte("\x1b[?1000l\x1b[?1006l\x1b[?25h\x1b[2J\x1b[H")) }()

	l := loop.NewLoop(tty, w, h)
	l.Start()
	defer l.Stop()

	// ── Build UI ────────────────────────────────────────────────────
	ui := buildUI()
	r := setupRouter(ui)

	ctx := viewport.RenderCtx{Buf: l.BackBuf}

	for {
		// Drain all pending events before rendering.
		drain := true
		for drain {
			select {
			case key := <-l.KeyEvents:
				if widgets.KeyCode(key.Code) == widgets.KeyEscape {
					return
				}
				r.DispatchKey(widgets.KeyEvent{
					Rune: key.Rune,
					Code: widgets.KeyCode(key.Code),
					Mod:  widgets.KeyMod(key.Mod),
				})
			case mouse := <-l.MouseEvents:
				r.DispatchMouse(widgets.MouseEvent{
					Button: widgets.MouseButton(mouse.Button),
					Action: widgets.MouseAction(mouse.Action),
					X:      int(mouse.X),
					Y:      int(mouse.Y),
					Mod:    widgets.KeyMod(mouse.Mod),
				})
			case <-l.ResizeEvents:
				nw, nh, _ := tty.GetSize()
				if nw > 0 && nh > 0 {
					l.BackBuf.Resize(nw, nh)
				}
			case err := <-l.Errors:
				if err != nil {
					return
				}
			default:
				drain = false
			}
		}

		// Update status line.
		ui.status = newStatus(r, ui)

		// Render.
		renderUI(ctx, ui, l.BackBuf.Width, l.BackBuf.Height)
	}
}

// ui holds all widget references.
type uiState struct {
	title   *widgets.Text
	btnA    *widgets.Button
	btnB    *widgets.Button
	list    *widgets.List
	input   *widgets.TextInput
	status  *widgets.Text
	header1 *widgets.Text
	header2 *widgets.Text
	header3 *widgets.Text
}

func buildUI() *uiState {
	titleSt := style.New().Foreground(style.Cyan).Background(style.Black).Bold()
	headerSt := style.New().Foreground(style.Yellow).Bold()
	btnSt := style.New().Foreground(style.Black).Background(style.Green).Bold()
	listSt := style.New().Foreground(style.White).Background(style.Black)
	inputSt := style.New().Foreground(style.Black).Background(style.White)
	statusSt := style.New().Foreground(style.DarkGray)

	ui := &uiState{
		title:   widgets.NewText("  fluint interactive demo  |  Esc=quit  Tab=focus", widgets.WithStyle(titleSt)),
		header1: widgets.NewText(" Buttons ", widgets.WithStyle(headerSt)),
		header2: widgets.NewText(" List (Up/Down + Enter) ", widgets.WithStyle(headerSt)),
		header3: widgets.NewText(" Input (type to edit) ", widgets.WithStyle(headerSt)),
		status:  widgets.NewText("", widgets.WithStyle(statusSt)),
	}

	ui.btnA = widgets.NewButton("  Click Me  ", widgets.WithStyle(btnSt),
		widgets.WithOnPress(func() {
			ui.btnA.SetStyle(style.New().Foreground(style.White).Background(style.Red).Bold())
		}),
	)

	ui.btnB = widgets.NewButton("  Reset  ", widgets.WithStyle(btnSt),
		widgets.WithOnPress(func() {
			ui.btnA.SetStyle(btnSt)
			ui.input.SetText("Hello, fluint!")
			ui.list.SetSelected(0)
		}),
	)

	items := make([]widgets.ListItem, 50)
	for i := range items {
		items[i] = widgets.ListItem{Text: fmt.Sprintf("  Item %02d", i+1)}
	}
	ui.list = widgets.NewList(items,
		widgets.WithStyle(listSt),
		widgets.WithOnSelect(func(idx int, item widgets.ListItem) {
			ui.btnA.SetStyle(style.New().Foreground(style.Black).Background(style.Magenta).Bold())
		}),
	)

	ui.input = widgets.NewTextInput("Hello, fluint!", widgets.WithStyle(inputSt))

	return ui
}

func setupRouter(ui *uiState) *router.Router {
	r := router.New()
	r.Register(ui.btnA)
	r.Register(ui.btnB)
	r.Register(ui.list)
	r.Register(ui.input)
	r.Focus(ui.btnA)
	return r
}

func newStatus(r *router.Router, ui *uiState) *widgets.Text {
	focused := "none"
	switch r.Focused() {
	case ui.btnA:
		focused = "btnA"
	case ui.btnB:
		focused = "btnB"
	case ui.list:
		focused = fmt.Sprintf("list[%d]", ui.list.Selected())
	case ui.input:
		focused = "input"
	}
	st := style.New().Foreground(style.DarkGray)
	return widgets.NewText(
		fmt.Sprintf(" Focus: %-10s | Tab=cycle | Esc=quit", focused),
		widgets.WithStyle(st),
	)
}

func renderUI(ctx viewport.RenderCtx, ui *uiState, w, h int) {
	root := &layout.Container{
		Dir: layout.Column,
		Children: []layout.Child{
			{Node: layout.Leaf{}, Basis: 1}, // title
			{Node: layout.Leaf{}, Basis: 1}, // btn header
			{Node: layout.Leaf{}, Basis: 1}, // btn row
			{Node: layout.Leaf{}, Basis: 1}, // list header
			{Node: layout.Leaf{}, Grow: 3},  // list
			{Node: layout.Leaf{}, Basis: 1}, // input header
			{Node: layout.Leaf{}, Basis: 1}, // input
			{Node: layout.Leaf{}, Basis: 1}, // status
		},
	}
	all := make([]layout.Rect, 0, 16)
	all = root.Measure(w, h, all)

	// all[0]=title, [2]=btnRow, [4]=listHeader(actually unused), [6]=list, [8]=inputHeader, [10]=input, [12]=status
	// But Container produces child rects at even indices.
	rTitle := all[0]
	rBtnHeader := all[2]
	rBtnRow := all[4]
	rListHeader := all[6]
	rList := all[8]
	rInputHeader := all[10]
	rInput := all[12]
	rStatus := all[14]

	ui.title.Render(ctx, rTitle)
	ui.header1.Render(ctx, rBtnHeader)

	// Buttons share the row.
	halfW := rBtnRow.Width / 2
	ui.btnA.Render(ctx, layout.Rect{X: rBtnRow.X, Y: rBtnRow.Y, Width: halfW, Height: rBtnRow.Height})
	ui.btnB.Render(ctx, layout.Rect{X: rBtnRow.X + halfW, Y: rBtnRow.Y, Width: rBtnRow.Width - halfW, Height: rBtnRow.Height})

	ui.header2.Render(ctx, rListHeader)
	ui.list.Render(ctx, rList)
	ui.header3.Render(ctx, rInputHeader)
	ui.input.Render(ctx, rInput)
	ui.status.Render(ctx, rStatus)
}
