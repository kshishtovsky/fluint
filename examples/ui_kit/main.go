// ui_kit is an interactive agent chat demo. Run it, then type a
// message and press Enter. Escape quits.
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

	// Enable mouse, hide cursor, clear screen, set black background.
	_, _ = tty.Write([]byte(
		"\x1b[?1000h" + // enable mouse click
			"\x1b[?1006h" + // enable SGR mouse
			"\x1b[?25l" + // hide cursor
			"\x1b[2J" + // clear screen
			"\x1b[48;2;0;0;0m" + // set bg to black
			"\x1b[38;2;255;255;255m" + // set fg to white
			"\x1b[H", // cursor home
	))
	defer func() {
		_, _ = tty.Write([]byte(
			"\x1b[0m" + // reset attributes
				"\x1b[?25h" + // show cursor
				"\x1b[2J" + // clear screen
				"\x1b[H", // cursor home
		))
	}()

	l := loop.NewLoop(tty, w, h)
	l.Start()
	defer l.Stop()

	ui := buildUI()
	r := setupRouter(ui)
	ctx := viewport.RenderCtx{Buf: l.BackBuf}

	ui.updateFocusStyles(r)
	ui.status = newStatus(r, ui)
	renderUI(ctx, ui, l.BackBuf.Width, l.BackBuf.Height)

	for {
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

		// Drain remaining events.
		for {
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
				goto done
			}
		}
	done:

		ui.updateFocusStyles(r)
		ui.status = newStatus(r, ui)
		renderUI(ctx, ui, l.BackBuf.Width, l.BackBuf.Height)
	}
}
