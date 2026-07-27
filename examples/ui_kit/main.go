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
	"github.com/kshishtovsky/fluint/core/viewport"
	"github.com/kshishtovsky/fluint/internal/platform"
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

	ui := buildUI()
	r := setupRouter(ui)
	ctx := viewport.RenderCtx{Buf: l.BackBuf}

	// Initial render.
	ui.updateFocusStyles(r)
	ui.status = newStatus(r, ui)
	renderUI(ctx, ui, l.BackBuf.Width, l.BackBuf.Height)

	for {
		// Block until an event arrives. No idle spinning — render only
		// when something actually changed. This eliminates the flicker
		// caused by Clear() + diff on every frame.
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
		case <-l.Quit:
			return
		}

		// Drain any remaining queued events before rendering.
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
			default:
				drain = false
			}
		}

		ui.updateFocusStyles(r)
		ui.status = newStatus(r, ui)
		renderUI(ctx, ui, l.BackBuf.Width, l.BackBuf.Height)
	}
}
