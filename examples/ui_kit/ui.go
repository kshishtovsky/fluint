package main

import (
	"github.com/kshishtovsky/fluint/core/router"
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/layout"
	"github.com/kshishtovsky/fluint/widgets"
)

// uiState holds all widget references.
type uiState struct {
	title  *widgets.Text
	input  *widgets.TextInput
	status *widgets.Text
	messages []widgets.Node // rendered top to bottom
}

func buildUI() *uiState {
	ui := &uiState{}

	ui.title = widgets.NewText("  fluint agent chat  │  Type to message  │  Esc quit", widgets.WithStyle(TitleStyle))

	// ── Chat messages (static demo content) ─────────────────────────

	// User message.
	userMsg := widgets.NewCard(
		widgets.NewText("  You: Explain the diff engine in fluint", widgets.WithStyle(UserStyle)),
		widgets.WithStyle(UserCardStyle),
	)

	// Agent thinking block.
	thinking := widgets.NewCard(
		widgets.NewText("  Thinking...", widgets.WithStyle(LabelStyle)),
		widgets.WithStyle(ThinkingCardStyle),
	)

	// Agent answer.
	answer := widgets.NewCard(
		widgets.NewText("  The diff engine compares front and back buffers cell\n  by cell. Only changed cells are sent to the terminal\n  via ANSI escape sequences. This means static content\n  costs zero I/O per frame.", widgets.WithStyle(BodyStyle)),
		widgets.WithStyle(AnswerCardStyle),
	)

	// Info block — code analysis.
	info := widgets.NewCard(
		widgets.NewText("  core/diff/diff.go\n  42 lines · 0 allocs/op · 12 ns/cell", widgets.WithStyle(LabelStyle)),
		widgets.WithStyle(InfoCardStyle),
	)

	// Another user message.
	userMsg2 := widgets.NewCard(
		widgets.NewText("  You: How does the Card widget handle shadows?", widgets.WithStyle(UserStyle)),
		widgets.WithStyle(UserCardStyle),
	)

	// Agent answer 2.
	answer2 := widgets.NewCard(
		widgets.NewText("  Shadows are drawn only in the bottom and right strips\n  outside the card rect — never under the card itself.\n  This prevents overlap with content below.", widgets.WithStyle(BodyStyle)),
		widgets.WithStyle(AnswerCardStyle),
	)

	ui.messages = []widgets.Node{userMsg, thinking, answer, info, userMsg2, answer2}

	// ── Input ───────────────────────────────────────────────────────
	ui.input = widgets.NewTextInput("", widgets.WithStyle(InputDefault))

	// ── Status ──────────────────────────────────────────────────────
	ui.status = widgets.NewText("", widgets.WithStyle(StatusStyle))

	return ui
}

func setupRouter(ui *uiState) *router.Router {
	r := router.New()
	r.Register(ui.input)
	r.Focus(ui.input)
	return r
}

func (ui *uiState) updateFocusStyles(r *router.Router) {
	if r.Focused() == ui.input {
		ui.input.SetStyle(InputFocused)
	} else {
		ui.input.SetStyle(InputDefault)
	}
}

func newStatus(r *router.Router, ui *uiState) *widgets.Text {
	focused := "none"
	if r.Focused() == ui.input {
		focused = "input"
	}
	return widgets.NewText(
		" Focus: "+focused+" │ Esc quit",
		widgets.WithStyle(StatusStyle),
	)
}

func renderUI(ctx viewport.RenderCtx, ui *uiState, w, h int) {
	// Layout: title + messages area (scrollable) + input + status.
	// Messages get all remaining space. Input is fixed 3 rows (border + text).
	root := &layout.Container{
		Dir: layout.Column,
		Children: []layout.Child{
			{Node: layout.Leaf{}, Basis: 1}, // title
			{Node: layout.Leaf{}, Grow: 1},  // messages (fills remaining)
			{Node: layout.Leaf{}, Basis: 3}, // input card
			{Node: layout.Leaf{}, Basis: 1}, // status
		},
	}
	all := make([]layout.Rect, 0, 8)
	all = root.Measure(w, h, all)

	rTitle := all[0]
	rMessages := all[2]
	rInput := all[4]
	rStatus := all[6]

	// Title.
	ui.title.Render(ctx, rTitle)

	// Messages — stack vertically inside the messages area.
	msgY := rMessages.Y
	for _, msg := range ui.messages {
		msgH := 4 // card height: 1 border + 2 content + 1 border
		if msgY+msgH > rMessages.Y+rMessages.Height {
			break // clip: no room
		}
		msg.Render(ctx, layout.Rect{
			X:      rMessages.X + 1,
			Y:      msgY,
			Width:  rMessages.Width - 2,
			Height: msgH,
		})
		msgY += msgH + 1 // +1 gap between messages
	}

	// Input card — full width at bottom.
	ui.input.Render(ctx, rInput)

	// Status.
	ui.status.Render(ctx, rStatus)
}
