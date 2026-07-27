package term

import "testing"

func TestParserPrintableASCII(t *testing.T) {
	var p Parser
	for b := byte(0x20); b <= 0x7E; b++ {
		ev, consumed, ok := p.Next([]byte{b})
		if !ok {
			t.Fatalf("byte 0x%02X: ok = false", b)
		}
		if consumed != 1 {
			t.Fatalf("byte 0x%02X: consumed = %d, want 1", b, consumed)
		}
		if ev.Type != EventKey {
			t.Fatalf("byte 0x%02X: type = %d, want EventKey", b, ev.Type)
		}
		if ev.Key.Rune != rune(b) {
			t.Fatalf("byte 0x%02X: rune = %q, want %q", b, ev.Key.Rune, rune(b))
		}
		if ev.Key.Code != KeyNone {
			t.Fatalf("byte 0x%02X: code = %d, want KeyNone", b, ev.Key.Code)
		}
		if ev.Key.Mod != 0 {
			t.Fatalf("byte 0x%02X: mod = %d, want 0", b, ev.Key.Mod)
		}
	}
}

func TestParserSpecialKeys(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		code  KeyCode
		mod   KeyMod
	}{
		{"Enter CR", []byte{0x0D}, KeyEnter, 0},
		{"Enter LF", []byte{0x0A}, KeyEnter, 0},
		{"Tab", []byte{0x09}, KeyTab, 0},
		{"Backspace 7F", []byte{0x7F}, KeyBackspace, 0},
		{"Backspace 08", []byte{0x08}, KeyBackspace, 0},
		{"Ctrl+A", []byte{0x01}, KeyNone, ModCtrl},
		{"Ctrl+C", []byte{0x03}, KeyNone, ModCtrl},
		{"Ctrl+Z", []byte{0x1A}, KeyNone, ModCtrl},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			ev, consumed, ok := p.Next(tt.input)
			if !ok {
				t.Fatal("ok = false")
			}
			if consumed != 1 {
				t.Fatalf("consumed = %d, want 1", consumed)
			}
			if ev.Type != EventKey {
				t.Fatalf("type = %d, want EventKey", ev.Type)
			}
			if ev.Key.Code != tt.code {
				t.Fatalf("code = %d, want %d", ev.Key.Code, tt.code)
			}
			if ev.Key.Mod != tt.mod {
				t.Fatalf("mod = %d, want %d", ev.Key.Mod, tt.mod)
			}
		})
	}
}

func TestParserCtrlLetters(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		wantRune rune
	}{
		{"Ctrl+A", 0x01, 'a'},
		{"Ctrl+B", 0x02, 'b'},
		{"Ctrl+C", 0x03, 'c'},
		{"Ctrl+L", 0x0C, 'l'},
		{"Ctrl+Z", 0x1A, 'z'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			ev, _, ok := p.Next([]byte{tt.input})
			if !ok {
				t.Fatal("ok = false")
			}
			if ev.Key.Rune != tt.wantRune {
				t.Fatalf("rune = %q, want %q", ev.Key.Rune, tt.wantRune)
			}
			if ev.Key.Mod != ModCtrl {
				t.Fatalf("mod = %d, want ModCtrl", ev.Key.Mod)
			}
		})
	}
}

func TestParserCSIArrows(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		code  KeyCode
	}{
		{"Up", []byte("\x1b[A"), KeyUp},
		{"Down", []byte("\x1b[B"), KeyDown},
		{"Right", []byte("\x1b[C"), KeyRight},
		{"Left", []byte("\x1b[D"), KeyLeft},
		{"Home", []byte("\x1b[H"), KeyHome},
		{"End", []byte("\x1b[F"), KeyEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			ev, consumed, ok := p.Next(tt.input)
			if !ok {
				t.Fatal("ok = false")
			}
			if consumed != len(tt.input) {
				t.Fatalf("consumed = %d, want %d", consumed, len(tt.input))
			}
			if ev.Type != EventKey {
				t.Fatalf("type = %d, want EventKey", ev.Type)
			}
			if ev.Key.Code != tt.code {
				t.Fatalf("code = %d, want %d", ev.Key.Code, tt.code)
			}
		})
	}
}

func TestParserCSIModifiedArrows(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		code  KeyCode
		mod   KeyMod
	}{
		{"Shift+Up", []byte("\x1b[1;2A"), KeyUp, ModShift},
		{"Alt+Down", []byte("\x1b[1;3B"), KeyDown, ModAlt},
		{"Ctrl+Right", []byte("\x1b[1;5C"), KeyRight, ModCtrl},
		{"Shift+Ctrl+Left", []byte("\x1b[1;6D"), KeyLeft, ModShift | ModCtrl},
		{"Alt+Ctrl+Up", []byte("\x1b[1;7A"), KeyUp, ModAlt | ModCtrl},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			ev, _, ok := p.Next(tt.input)
			if !ok {
				t.Fatal("ok = false")
			}
			if ev.Key.Code != tt.code {
				t.Fatalf("code = %d, want %d", ev.Key.Code, tt.code)
			}
			if ev.Key.Mod != tt.mod {
				t.Fatalf("mod = %d, want %d", ev.Key.Mod, tt.mod)
			}
		})
	}
}

func TestParserCSITilde(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		code  KeyCode
	}{
		{"Insert", []byte("\x1b[2~"), KeyInsert},
		{"Delete", []byte("\x1b[3~"), KeyDelete},
		{"PageUp", []byte("\x1b[5~"), KeyPageUp},
		{"PageDown", []byte("\x1b[6~"), KeyPageDown},
		{"Home", []byte("\x1b[1~"), KeyHome},
		{"End", []byte("\x1b[4~"), KeyEnd},
		{"F1", []byte("\x1b[11~"), KeyF1},
		{"F5", []byte("\x1b[15~"), KeyF5},
		{"F6", []byte("\x1b[17~"), KeyF6},
		{"F12", []byte("\x1b[24~"), KeyF12},
		{"PasteStart", []byte("\x1b[200~"), KeyPasteStart},
		{"PasteEnd", []byte("\x1b[201~"), KeyPasteEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			ev, _, ok := p.Next(tt.input)
			if !ok {
				t.Fatal("ok = false")
			}
			if ev.Key.Code != tt.code {
				t.Fatalf("code = %d, want %d", ev.Key.Code, tt.code)
			}
		})
	}
}

func TestParserSS3(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		code  KeyCode
	}{
		{"F1", []byte("\x1bOP"), KeyF1},
		{"F2", []byte("\x1bOQ"), KeyF2},
		{"F3", []byte("\x1bOR"), KeyF3},
		{"F4", []byte("\x1bOS"), KeyF4},
		{"Up", []byte("\x1bOA"), KeyUp},
		{"Down", []byte("\x1bOB"), KeyDown},
		{"Right", []byte("\x1bOC"), KeyRight},
		{"Left", []byte("\x1bOD"), KeyLeft},
		{"Home", []byte("\x1bOH"), KeyHome},
		{"End", []byte("\x1bOF"), KeyEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			ev, _, ok := p.Next(tt.input)
			if !ok {
				t.Fatal("ok = false")
			}
			if ev.Key.Code != tt.code {
				t.Fatalf("code = %d, want %d", ev.Key.Code, tt.code)
			}
		})
	}
}

func TestParserBacktab(t *testing.T) {
	var p Parser
	ev, _, ok := p.Next([]byte("\x1b[Z"))
	if !ok {
		t.Fatal("ok = false")
	}
	if ev.Key.Code != KeyBacktab {
		t.Fatalf("code = %d, want KeyBacktab", ev.Key.Code)
	}
	if ev.Key.Mod != ModShift {
		t.Fatalf("mod = %d, want ModShift", ev.Key.Mod)
	}
}

func TestParserAltKey(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantRune rune
	}{
		{"Alt+a", []byte("\x1ba"), 'a'},
		{"Alt+z", []byte("\x1bz"), 'z'},
		{"Alt+A", []byte("\x1bA"), 'A'},
		{"Alt+1", []byte("\x1b1"), '1'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			ev, _, ok := p.Next(tt.input)
			if !ok {
				t.Fatal("ok = false")
			}
			if ev.Key.Rune != tt.wantRune {
				t.Fatalf("rune = %q, want %q", ev.Key.Rune, tt.wantRune)
			}
			if ev.Key.Mod != ModAlt {
				t.Fatalf("mod = %d, want ModAlt", ev.Key.Mod)
			}
		})
	}
}

func TestParserMouseSGR(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		button MouseButton
		action MouseAction
		x, y   uint16
	}{
		{
			"left press at 10,20",
			[]byte("\x1b[<0;11;21M"),
			MouseLeft, MousePress, 10, 20,
		},
		{
			"right press at 0,0",
			[]byte("\x1b[<2;1;1M"),
			MouseRight, MousePress, 0, 0,
		},
		{
			"left release at 5,3",
			[]byte("\x1b[<0;6;4m"),
			MouseLeft, MouseRelease, 5, 3,
		},
		{
			"wheel up",
			[]byte("\x1b[<64;10;10M"),
			MouseWheelUp, MousePress, 9, 9,
		},
		{
			"wheel down",
			[]byte("\x1b[<65;10;10M"),
			MouseWheelDown, MousePress, 9, 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			ev, _, ok := p.Next(tt.input)
			if !ok {
				t.Fatal("ok = false")
			}
			if ev.Type != EventMouse {
				t.Fatalf("type = %d, want EventMouse", ev.Type)
			}
			if ev.Mouse.Button != tt.button {
				t.Fatalf("button = %d, want %d", ev.Mouse.Button, tt.button)
			}
			if ev.Mouse.Action != tt.action {
				t.Fatalf("action = %d, want %d", ev.Mouse.Action, tt.action)
			}
			if ev.Mouse.X != tt.x {
				t.Fatalf("x = %d, want %d", ev.Mouse.X, tt.x)
			}
			if ev.Mouse.Y != tt.y {
				t.Fatalf("y = %d, want %d", ev.Mouse.Y, tt.y)
			}
		})
	}
}

func TestParserUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantRune rune
	}{
		{"2-byte: ä", []byte{0xC3, 0xA4}, 'ä'},
		{"3-byte: ★", []byte{0xE2, 0x98, 0x85}, '★'},
		{"4-byte: 🎉", []byte{0xF0, 0x9F, 0x8E, 0x89}, '🎉'},
		{"2-byte: ñ", []byte{0xC3, 0xB1}, 'ñ'},
		{"3-byte: 日", []byte{0xE6, 0x97, 0xA5}, '日'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			ev, consumed, ok := p.Next(tt.input)
			if !ok {
				t.Fatal("ok = false")
			}
			if consumed != len(tt.input) {
				t.Fatalf("consumed = %d, want %d", consumed, len(tt.input))
			}
			if ev.Key.Rune != tt.wantRune {
				t.Fatalf("rune = %q, want %q", ev.Key.Rune, tt.wantRune)
			}
		})
	}
}

func TestParserPartialSequence(t *testing.T) {
	// Feed CSI in two parts: ESC then [A.
	var p Parser

	// First byte: ESC — absorbed into state, no event.
	_, consumed1, ok1 := p.Next([]byte{0x1B})
	if ok1 {
		t.Fatal("ESC alone should not produce an event")
	}
	if consumed1 != 1 {
		t.Fatalf("consumed = %d, want 1", consumed1)
	}

	// Remaining bytes: [A — should complete the sequence.
	ev, consumed2, ok2 := p.Next([]byte("[A"))
	if !ok2 {
		t.Fatal("[A should complete Up arrow")
	}
	if consumed2 != 2 {
		t.Fatalf("consumed = %d, want 2", consumed2)
	}
	if ev.Key.Code != KeyUp {
		t.Fatalf("code = %d, want KeyUp", ev.Key.Code)
	}
}

func TestParserFlush(t *testing.T) {
	var p Parser

	// Feed ESC, then flush (simulating timeout).
	p.Next([]byte{0x1B})
	ev, ok := p.Flush()
	if !ok {
		t.Fatal("Flush after ESC should produce event")
	}
	if ev.Key.Code != KeyEscape {
		t.Fatalf("code = %d, want KeyEscape", ev.Key.Code)
	}
}

func TestParserMultipleEvents(t *testing.T) {
	// Feed multiple events in one buffer.
	var p Parser
	input := []byte("abc")

	var events []InputEvent
	data := input
	for len(data) > 0 {
		ev, consumed, ok := p.Next(data)
		data = data[consumed:]
		if ok {
			events = append(events, ev)
		}
		if consumed == 0 && !ok {
			break
		}
	}

	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	for i, r := range []rune{'a', 'b', 'c'} {
		if events[i].Key.Rune != r {
			t.Fatalf("event[%d].rune = %q, want %q", i, events[i].Key.Rune, r)
		}
	}
}

func TestParserDoubleESC(t *testing.T) {
	var p Parser

	// Double ESC: first ESC emits KeyEscape, second stays in escape state.
	ev, consumed, ok := p.Next([]byte{0x1B, 0x1B})
	if !ok {
		t.Fatal("double ESC should produce an event")
	}
	if consumed != 2 {
		t.Fatalf("consumed = %d, want 2", consumed)
	}
	if ev.Key.Code != KeyEscape {
		t.Fatalf("code = %d, want KeyEscape", ev.Key.Code)
	}
	// Parser should still be in escape state.
	if !p.InEscapeState() {
		t.Fatal("parser should be in escape state after double ESC")
	}
}

func BenchmarkParserPrintable(b *testing.B) {
	var p Parser
	input := []byte("Hello, World! This is a test of printable ASCII input.")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := input
		for len(data) > 0 {
			_, consumed, _ := p.Next(data)
			data = data[consumed:]
			if consumed == 0 {
				break
			}
		}
	}
}

func BenchmarkParserCSI(b *testing.B) {
	var p Parser
	input := []byte("\x1b[A\x1b[B\x1b[1;5C\x1b[3~\x1b[<0;10;20M")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := input
		for len(data) > 0 {
			_, consumed, _ := p.Next(data)
			data = data[consumed:]
			if consumed == 0 {
				break
			}
		}
	}
}

func BenchmarkParserUTF8(b *testing.B) {
	var p Parser
	// Mix of 2, 3, and 4-byte UTF-8 sequences.
	input := []byte("日本語テスト🎉✨")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := input
		for len(data) > 0 {
			_, consumed, _ := p.Next(data)
			data = data[consumed:]
			if consumed == 0 {
				break
			}
		}
	}
}

func TestParserZeroAllocs(t *testing.T) {
	// Verify zero allocations for all common input types.
	tests := []struct {
		name  string
		input []byte
	}{
		{"printable", []byte("a")},
		{"CSI arrow", []byte("\x1b[A")},
		{"CSI tilde", []byte("\x1b[3~")},
		{"CSI modified", []byte("\x1b[1;5C")},
		{"mouse SGR", []byte("\x1b[<0;10;20M")},
		{"UTF-8 2byte", []byte{0xC3, 0xA4}},
		{"UTF-8 3byte", []byte{0xE2, 0x98, 0x85}},
		{"Alt+key", []byte("\x1ba")},
		{"SS3 F1", []byte("\x1bOP")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Parser
			allocs := testing.AllocsPerRun(100, func() {
				data := tt.input
				for len(data) > 0 {
					_, consumed, _ := p.Next(data)
					data = data[consumed:]
					if consumed == 0 {
						break
					}
				}
			})
			if allocs > 0 {
				t.Fatalf("%.0f allocs/op, want 0", allocs)
			}
		})
	}
}
