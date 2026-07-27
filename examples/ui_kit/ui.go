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
	title       *widgets.Text
	btnA        *widgets.Button
	btnB        *widgets.Button
	list        *widgets.List
	input       *widgets.TextInput
	status      *widgets.Text
	thinking    *widgets.Card // "thinking" block (dim, rounded)
	answer      *widgets.Card // "answer" block (bright, rounded)
	sectionBtns *widgets.Card // buttons wrapped in a section card
}

func buildUI() *uiState {
	ui := &uiState{}

	ui.title = widgets.NewText("  fluint ui kit  │  Esc quit  Tab focus", widgets.WithStyle(TitleStyle))

	// -- Buttons --
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

	// Wrap buttons in a section card.
	btnRow := widgets.NewText("  [Action]  [Reset]", widgets.WithStyle(BodyStyle))
	ui.sectionBtns = widgets.NewCard(
		widgets.NewText("  Click Me │ Reset", widgets.WithStyle(BodyStyle)),
		widgets.WithStyle(SectionCardStyle),
	)
	_ = btnRow

	// -- List --
	items := make([]widgets.ListItem, 50)
	for i := range items {
		items[i] = widgets.ListItem{Text: fmt.Sprintf("  Item %02d", i+1)}
	}
	ui.list = widgets.NewList(items,
		widgets.WithStyle(ListDefault),
		widgets.WithOnSelect(func(idx int, item widgets.ListItem) {
			ui.btnA.SetStyle(ButtonPressed)
		}),
	)

	// -- Text input --
	ui.input = widgets.NewTextInput("Hello, fluint!", widgets.WithStyle(InputDefault))

	// -- Chat demo cards --
	thinkingText := widgets.NewText("  Analyzing codebase structure...", widgets.WithStyle(
		LabelStyle,
	))
	ui.thinking = widgets.NewCard(thinkingText, widgets.WithStyle(ThinkingCardStyle))

	answerText := widgets.NewText("  Found 3 issues:\n  1. Missing error handling\n  2. Unused import\n  3. Race condition", widgets.WithStyle(
		BodyStyle,
	))
	ui.answer = widgets.NewCard(answerText, widgets.WithStyle(AnswerCardStyle))

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

	if focused == ui.btnA {
		if ui.btnA.Style().BG() == ButtonPressed.BG() {
			return
		}
		ui.btnA.SetStyle(ButtonFocused)
	} else {
		if ui.btnA.Style().BG() == ButtonPressed.BG() {
			return
		}
		ui.btnA.SetStyle(ButtonDefault)
	}

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
	// Layout: Column with rows for each section.
	root := &layout.Container{
		Dir: layout.Column,
		Children: []layout.Child{
			{Node: layout.Leaf{}, Basis: 1}, // title
			{Node: layout.Leaf{}, Basis: 4}, // thinking card
			{Node: layout.Leaf{}, Basis: 5}, // answer card
			{Node: layout.Leaf{}, Basis: 1}, // separator
			{Node: layout.Leaf{}, Basis: 1}, // buttons header
			{Node: layout.Leaf{}, Basis: 1}, // button row
			{Node: layout.Leaf{}, Grow: 3},  // list
			{Node: layout.Leaf{}, Basis: 1}, // input
			{Node: layout.Leaf{}, Basis: 1}, // status
		},
	}
	all := make([]layout.Rect, 0, 18)
	all = root.Measure(w, h, all)

	rTitle := all[0]
	rThinking := all[2]
	rAnswer := all[4]
	rSep := all[6]
	rBtnHeader := all[8]
	rBtnRow := all[10]
	rList := all[12]
	rInput := all[14]
	rStatus := all[16]

	// Title bar.
	ui.title.Render(ctx, rTitle)

	// Chat demo: Thinking card (dim) + Answer card (bright).
	ui.thinking.Render(ctx, rThinking)
	ui.answer.Render(ctx, rAnswer)

	// Separator.
	sepText := widgets.NewText("───────────── Controls ─────────────", widgets.WithStyle(LabelStyle))
	sepText.Render(ctx, rSep)

	// Buttons header + row.
	btnHeader := widgets.NewText(" Buttons ", widgets.WithStyle(HeaderStyle))
	btnHeader.Render(ctx, rBtnHeader)
	halfW := rBtnRow.Width / 2
	ui.btnA.Render(ctx, layout.Rect{X: rBtnRow.X, Y: rBtnRow.Y, Width: halfW, Height: rBtnRow.Height})
	ui.btnB.Render(ctx, layout.Rect{X: rBtnRow.X + halfW, Y: rBtnRow.Y, Width: rBtnRow.Width - halfW, Height: rBtnRow.Height})

	// List.
	listHeader := widgets.NewText(" List (↑↓ + Enter) ", widgets.WithStyle(HeaderStyle))
	listHeader.Render(ctx, rList)
	ui.list.Render(ctx, rList)

	// Input.
	inputHeader := widgets.NewText(" Input (type to edit) ", widgets.WithStyle(HeaderStyle))
	inputHeader.Render(ctx, rInput)
	ui.input.Render(ctx, rInput)

	// Status bar.
	ui.status.Render(ctx, rStatus)
}
