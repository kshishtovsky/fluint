package main

import (
	"fmt"

	"github.com/kshishtovsky/fluint/core/router"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/widgets"
)

// uiState holds all widget references and state.
type uiState struct {
	title  *widgets.Text
	btnA   *widgets.Button
	btnB   *widgets.Button
	list   *widgets.List
	input  *widgets.TextInput
	status *widgets.Text
	h1     *widgets.Text // "Buttons" header
	h2     *widgets.Text // "List" header
	h3     *widgets.Text // "Input" header
}

func buildUI() *uiState {
	ui := &uiState{}

	// -- Title bar (uses theme tokens) --
	ui.title = widgets.NewText("  fluint ui kit  │  Esc quit  Tab focus", widgets.WithStyle(TitleStyle))

	// -- Section headers (Bold + Accent, Hallmark typography hierarchy) --
	ui.h1 = widgets.NewText(" Buttons ", widgets.WithStyle(HeaderStyle))
	ui.h2 = widgets.NewText(" List (↑↓ + Enter) ", widgets.WithStyle(HeaderStyle))
	ui.h3 = widgets.NewText(" Input (type to edit) ", widgets.WithStyle(HeaderStyle))

	// -- Buttons (5-state system: Default/Focused/Pressed/Disabled) --
	ui.btnA = widgets.NewButton("  Click Me  ", widgets.WithStyle(ButtonDefault),
		widgets.WithOnPress(func() {
			ui.btnA.SetStyle(ButtonPressed)
		}),
	)

	ui.btnB = widgets.NewButton("  Reset  ", widgets.WithStyle(ButtonGhost),
		widgets.WithOnPress(func() {
			ui.btnA.SetStyle(ButtonDefault)
			ui.input.SetText("Hello, fluint!")
			ui.list.SetSelected(0)
		}),
	)

	// -- List (virtualised, 50 items) --
	items := make([]widgets.ListItem, 50)
	for i := range items {
		items[i] = widgets.ListItem{Text: fmt.Sprintf("  Item %02d", i+1)}
	}
	ui.list = widgets.NewList(items,
		widgets.WithStyle(ListDefault),
		widgets.WithOnSelect(func(idx int, item widgets.ListItem) {
			// Visual feedback: flash the button.
			ui.btnA.SetStyle(ButtonPressed)
		}),
	)

	// -- Text input --
	ui.input = widgets.NewTextInput("Hello, fluint!", widgets.WithStyle(InputDefault))

	// -- Status bar --
	ui.status = widgets.NewText("", widgets.WithStyle(StatusStyle))

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

// updateFocusStyles applies theme-driven state styles based on which
// widget currently has focus. This is the "state discipline" from
// Hallmark's component-scope: every interactive element has a focused
// visual that differs from its default.
func (ui *uiState) updateFocusStyles(r *router.Router) {
	focused := r.Focused()

	// Button A: focused vs default.
	if focused == ui.btnA {
		if ui.btnA.Style().BG() == ButtonPressed.BG() {
			return // keep pressed style
		}
		ui.btnA.SetStyle(ButtonFocused)
	} else {
		if ui.btnA.Style().BG() == ButtonPressed.BG() {
			return // keep pressed until reset
		}
		ui.btnA.SetStyle(ButtonDefault)
	}

	// Button B: ghost style always (secondary action).
	if focused == ui.btnB {
		ui.btnB.SetStyle(ButtonFocused)
	} else {
		ui.btnB.SetStyle(ButtonGhost)
	}
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
	return widgets.NewText(
		fmt.Sprintf(" Focus: %-10s │ Tab=cycle │ Esc=quit", focused),
		widgets.WithStyle(StatusStyle),
	)
}

func renderUI(ctx viewport.RenderCtx, ui *uiState, w, h int) {
	// Layout: Column with 8 rows.
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

	rTitle := all[0]
	rBtnHeader := all[2]
	rBtnRow := all[4]
	rListHeader := all[6]
	rList := all[8]
	rInputHeader := all[10]
	rInput := all[12]
	rStatus := all[14]

	// Title bar.
	ui.title.Render(ctx, rTitle)

	// Section: Buttons
	ui.h1.Render(ctx, rBtnHeader)
	halfW := rBtnRow.Width / 2
	ui.btnA.Render(ctx, layout.Rect{X: rBtnRow.X, Y: rBtnRow.Y, Width: halfW, Height: rBtnRow.Height})
	ui.btnB.Render(ctx, layout.Rect{X: rBtnRow.X + halfW, Y: rBtnRow.Y, Width: rBtnRow.Width - halfW, Height: rBtnRow.Height})

	// Section: List
	ui.h2.Render(ctx, rListHeader)
	ui.list.Render(ctx, rList)

	// Section: Input
	ui.h3.Render(ctx, rInputHeader)
	ui.input.Render(ctx, rInput)

	// Status bar.
	ui.status.Render(ctx, rStatus)

	// Draw a separator line between sections using theme border style.
	sep := widgets.NewText("", widgets.WithStyle(BorderStyle))
	_ = sep
}
