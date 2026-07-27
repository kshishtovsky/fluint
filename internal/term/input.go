package term

import "unicode/utf8"

// parserState represents the current state of the escape sequence
// parser state machine.
type parserState uint8

const (
	stateGround parserState = iota // Default: processing regular input.
	stateEscape                    // Received ESC (0x1B).
	stateCSI                       // Received CSI introducer (ESC [).
	stateSS3                       // Received SS3 introducer (ESC O).
	stateUTF8                      // Accumulating multi-byte UTF-8.
)

// Parser is a zero-allocation terminal escape sequence parser.
// It implements a state machine that converts raw terminal input bytes
// into InputEvent values.
//
// Call Next to feed bytes. The parser maintains internal state across
// calls to handle multi-byte sequences that span read boundaries.
// Call Flush to emit any pending partial sequence (e.g., standalone ESC
// after a timeout).
type Parser struct {
	state parserState

	// CSI parameter accumulation.
	params   [8]uint16 // Parsed CSI numeric parameters.
	nparams  int       // Number of parameters accumulated.
	curParam uint16    // Current parameter being built from digits.
	hasParam bool      // Whether curParam has received any digits.
	private  byte      // CSI private marker ('<', '?', '>', '=') or 0.

	// UTF-8 accumulation.
	utf8Buf [4]byte // Partial UTF-8 bytes.
	utf8Len int     // Expected total byte count for current sequence.
	utf8N   int     // Bytes accumulated so far.

	// Pending event from re-processed bytes (e.g., ESC + unknown byte
	// emits ESC and re-processes the byte, possibly producing a second
	// event).
	pending    InputEvent
	hasPending bool
}

// Next parses the next event from data. Returns the event, the number
// of bytes consumed, and whether a valid event was produced.
//
// If consumed > 0 and ok is false, the bytes were absorbed into the
// parser's internal state (partial sequence). Call Next again with
// more data.
//
// If consumed == 0 and ok is true, the event is a pending result from
// a previous call (re-processed byte). No new input was consumed.
func (p *Parser) Next(data []byte) (event InputEvent, consumed int, ok bool) {
	// Return pending event from a previous re-processing step.
	if p.hasPending {
		p.hasPending = false
		return p.pending, 0, true
	}

	for i := range data {
		if ev, done := p.step(data[i]); done {
			return ev, i + 1, true
		}
	}
	return InputEvent{}, len(data), false
}

// Flush forces the parser to emit any pending partial sequence.
// Call this after an input timeout to resolve ambiguous sequences
// (e.g., standalone ESC vs. start of escape sequence).
func (p *Parser) Flush() (InputEvent, bool) {
	switch p.state {
	case stateEscape:
		p.state = stateGround
		return makeKeyEvent(KeyEscape, 0, 0), true
	default:
		p.state = stateGround
		p.hasPending = false
		return InputEvent{}, false
	}
}

// InEscapeState reports whether the parser is mid-sequence and might
// need a timeout-based Flush to resolve ambiguity.
func (p *Parser) InEscapeState() bool {
	return p.state == stateEscape
}

// step processes a single byte and advances the state machine.
func (p *Parser) step(b byte) (InputEvent, bool) {
	switch p.state {
	case stateGround:
		return p.ground(b)
	case stateEscape:
		return p.escape(b)
	case stateCSI:
		return p.csi(b)
	case stateSS3:
		return p.ss3(b)
	case stateUTF8:
		return p.utf8(b)
	default:
		p.state = stateGround
		return InputEvent{}, false
	}
}

// ground handles bytes in the default state.
func (p *Parser) ground(b byte) (InputEvent, bool) {
	switch {
	case b == 0x1B:
		p.state = stateEscape
		return InputEvent{}, false

	case b == 0x0D, b == 0x0A:
		return makeKeyEvent(KeyEnter, 0, 0), true
	case b == 0x09:
		return makeKeyEvent(KeyTab, 0, 0), true
	case b == 0x7F:
		return makeKeyEvent(KeyBackspace, 0, 0), true
	case b == 0x08:
		return makeKeyEvent(KeyBackspace, 0, 0), true
	case b == 0x00:
		return makeKeyEvent(KeyNone, '@', ModCtrl), true

	// Ctrl+A (0x01) through Ctrl+Z (0x1A), excluding Tab (0x09)
	// and Enter (0x0D) handled above.
	case b >= 0x01 && b <= 0x1A:
		return makeKeyEvent(KeyNone, rune(b-1+'a'), ModCtrl), true

	// Printable ASCII.
	case b >= 0x20 && b <= 0x7E:
		return makeKeyEvent(KeyNone, rune(b), 0), true

	// UTF-8 lead bytes.
	case b >= 0xC0 && b <= 0xDF:
		p.startUTF8(b, 2)
		return InputEvent{}, false
	case b >= 0xE0 && b <= 0xEF:
		p.startUTF8(b, 3)
		return InputEvent{}, false
	case b >= 0xF0 && b <= 0xF7:
		p.startUTF8(b, 4)
		return InputEvent{}, false

	default:
		return InputEvent{}, false
	}
}

// escape handles the byte following ESC.
func (p *Parser) escape(b byte) (InputEvent, bool) {
	switch {
	case b == '[':
		p.state = stateCSI
		p.resetCSI()
		return InputEvent{}, false

	case b == 'O':
		p.state = stateSS3
		return InputEvent{}, false

	case b == 0x1B:
		// Double ESC: emit first ESC, stay in escape state for the second.
		return makeKeyEvent(KeyEscape, 0, 0), true

	case b >= 0x20 && b <= 0x7E:
		// ESC + printable = Alt+key.
		p.state = stateGround
		return makeKeyEvent(KeyNone, rune(b), ModAlt), true

	default:
		// ESC + unexpected byte: emit ESC, re-process byte in ground.
		p.state = stateGround
		if ev, done := p.ground(b); done {
			p.pending = ev
			p.hasPending = true
		}
		return makeKeyEvent(KeyEscape, 0, 0), true
	}
}

// csi handles bytes within a CSI sequence (ESC [ ...).
func (p *Parser) csi(b byte) (InputEvent, bool) {
	switch {
	case b >= '0' && b <= '9':
		// Parameter digit.
		p.curParam = p.curParam*10 + uint16(b-'0')
		p.hasParam = true
		return InputEvent{}, false

	case b == ';':
		// Parameter separator.
		p.pushParam()
		return InputEvent{}, false

	case b == '<' || b == '?' || b == '>' || b == '=':
		// Private marker — must appear before any parameters.
		if p.nparams == 0 && !p.hasParam {
			p.private = b
		}
		return InputEvent{}, false

	case b >= 0x40 && b <= 0x7E:
		// Final byte — dispatch.
		p.pushParam()
		return p.dispatchCSI(b)

	default:
		// Unexpected byte — abort sequence.
		p.state = stateGround
		return InputEvent{}, false
	}
}

// ss3 handles the byte following ESC O (SS3 sequence).
func (p *Parser) ss3(b byte) (InputEvent, bool) {
	p.state = stateGround
	switch b {
	case 'P':
		return makeKeyEvent(KeyF1, 0, 0), true
	case 'Q':
		return makeKeyEvent(KeyF2, 0, 0), true
	case 'R':
		return makeKeyEvent(KeyF3, 0, 0), true
	case 'S':
		return makeKeyEvent(KeyF4, 0, 0), true
	case 'A':
		return makeKeyEvent(KeyUp, 0, 0), true
	case 'B':
		return makeKeyEvent(KeyDown, 0, 0), true
	case 'C':
		return makeKeyEvent(KeyRight, 0, 0), true
	case 'D':
		return makeKeyEvent(KeyLeft, 0, 0), true
	case 'H':
		return makeKeyEvent(KeyHome, 0, 0), true
	case 'F':
		return makeKeyEvent(KeyEnd, 0, 0), true
	default:
		return InputEvent{}, false
	}
}

// utf8 accumulates continuation bytes for a multi-byte UTF-8 sequence.
func (p *Parser) utf8(b byte) (InputEvent, bool) {
	// Validate continuation byte (10xxxxxx).
	if b < 0x80 || b > 0xBF {
		// Invalid continuation — re-process in ground state.
		p.state = stateGround
		return p.ground(b)
	}

	p.utf8Buf[p.utf8N] = b
	p.utf8N++

	if p.utf8N < p.utf8Len {
		return InputEvent{}, false
	}

	// Complete sequence — decode.
	p.state = stateGround
	r, size := utf8.DecodeRune(p.utf8Buf[:p.utf8Len])
	if r == utf8.RuneError && size <= 1 {
		return InputEvent{}, false
	}
	return makeKeyEvent(KeyNone, r, 0), true
}

// dispatchCSI interprets a complete CSI sequence based on its final byte.
func (p *Parser) dispatchCSI(final byte) (InputEvent, bool) {
	p.state = stateGround

	// Mouse SGR: CSI < button ; x ; y M/m
	if p.private == '<' && (final == 'M' || final == 'm') {
		return p.parseMouseSGR(final)
	}

	mod := p.extractMod()

	switch final {
	case 'A':
		return makeKeyEvent(KeyUp, 0, mod), true
	case 'B':
		return makeKeyEvent(KeyDown, 0, mod), true
	case 'C':
		return makeKeyEvent(KeyRight, 0, mod), true
	case 'D':
		return makeKeyEvent(KeyLeft, 0, mod), true
	case 'H':
		return makeKeyEvent(KeyHome, 0, mod), true
	case 'F':
		return makeKeyEvent(KeyEnd, 0, mod), true
	case 'Z':
		return makeKeyEvent(KeyBacktab, 0, ModShift), true
	case '~':
		return p.dispatchTilde(mod)
	default:
		return InputEvent{}, false
	}
}

// dispatchTilde handles CSI <param> ~ sequences (Insert, Delete, PgUp, etc.).
func (p *Parser) dispatchTilde(mod KeyMod) (InputEvent, bool) {
	if p.nparams == 0 {
		return InputEvent{}, false
	}

	switch p.params[0] {
	case 1:
		return makeKeyEvent(KeyHome, 0, mod), true
	case 2:
		return makeKeyEvent(KeyInsert, 0, mod), true
	case 3:
		return makeKeyEvent(KeyDelete, 0, mod), true
	case 4:
		return makeKeyEvent(KeyEnd, 0, mod), true
	case 5:
		return makeKeyEvent(KeyPageUp, 0, mod), true
	case 6:
		return makeKeyEvent(KeyPageDown, 0, mod), true
	case 11:
		return makeKeyEvent(KeyF1, 0, mod), true
	case 12:
		return makeKeyEvent(KeyF2, 0, mod), true
	case 13:
		return makeKeyEvent(KeyF3, 0, mod), true
	case 14:
		return makeKeyEvent(KeyF4, 0, mod), true
	case 15:
		return makeKeyEvent(KeyF5, 0, mod), true
	case 17:
		return makeKeyEvent(KeyF6, 0, mod), true
	case 18:
		return makeKeyEvent(KeyF7, 0, mod), true
	case 19:
		return makeKeyEvent(KeyF8, 0, mod), true
	case 20:
		return makeKeyEvent(KeyF9, 0, mod), true
	case 21:
		return makeKeyEvent(KeyF10, 0, mod), true
	case 23:
		return makeKeyEvent(KeyF11, 0, mod), true
	case 24:
		return makeKeyEvent(KeyF12, 0, mod), true
	case 200:
		return makeKeyEvent(KeyPasteStart, 0, 0), true
	case 201:
		return makeKeyEvent(KeyPasteEnd, 0, 0), true
	default:
		return InputEvent{}, false
	}
}

// parseMouseSGR parses a CSI < button;x;y M/m mouse event (SGR mode 1006).
func (p *Parser) parseMouseSGR(final byte) (InputEvent, bool) {
	if p.nparams < 3 {
		return InputEvent{}, false
	}

	cb := p.params[0]
	x := p.params[1]
	y := p.params[2]

	// SGR coordinates are 1-based; convert to 0-based.
	if x > 0 {
		x--
	}
	if y > 0 {
		y--
	}

	// Extract modifier flags from button byte.
	var mod KeyMod
	if cb&4 != 0 {
		mod |= ModShift
	}
	if cb&8 != 0 {
		mod |= ModMeta
	}
	if cb&16 != 0 {
		mod |= ModCtrl
	}

	isMotion := cb&32 != 0

	var button MouseButton
	var action MouseAction

	btn := cb & 3
	if cb&64 != 0 {
		// Wheel events.
		switch btn {
		case 0:
			button = MouseWheelUp
		case 1:
			button = MouseWheelDown
		default:
			button = MouseNone
		}
		action = MousePress
	} else {
		switch btn {
		case 0:
			button = MouseLeft
		case 1:
			button = MouseMiddle
		case 2:
			button = MouseRight
		default:
			button = MouseNone
		}

		switch {
		case isMotion:
			action = MouseMotion
		case final == 'M':
			action = MousePress
		default:
			action = MouseRelease
		}
	}

	return InputEvent{
		Type: EventMouse,
		Mouse: MouseEvent{
			Button: button,
			Action: action,
			X:      x,
			Y:      y,
			Mod:    mod,
		},
	}, true
}

// --- helpers --------------------------------------------------------

// extractMod decodes the modifier parameter from a CSI sequence.
// In xterm-style sequences, params[1] encodes (modifier + 1).
func (p *Parser) extractMod() KeyMod {
	if p.nparams >= 2 && p.params[1] > 1 {
		return decodeModifier(p.params[1])
	}
	return 0
}

// decodeModifier converts an xterm modifier parameter to KeyMod flags.
// The parameter encoding is: value = modifier_bits + 1.
// Bits: 0=Shift, 1=Alt, 2=Ctrl, 3=Meta.
func decodeModifier(param uint16) KeyMod {
	m := param - 1
	var mod KeyMod
	if m&1 != 0 {
		mod |= ModShift
	}
	if m&2 != 0 {
		mod |= ModAlt
	}
	if m&4 != 0 {
		mod |= ModCtrl
	}
	if m&8 != 0 {
		mod |= ModMeta
	}
	return mod
}

func (p *Parser) resetCSI() {
	p.nparams = 0
	p.curParam = 0
	p.hasParam = false
	p.private = 0
}

func (p *Parser) pushParam() {
	if p.nparams < len(p.params) {
		if p.hasParam {
			p.params[p.nparams] = p.curParam
		} else {
			p.params[p.nparams] = 0
		}
		p.nparams++
	}
	p.curParam = 0
	p.hasParam = false
}

func (p *Parser) startUTF8(lead byte, length int) {
	p.state = stateUTF8
	p.utf8Buf[0] = lead
	p.utf8Len = length
	p.utf8N = 1
}

func makeKeyEvent(code KeyCode, r rune, mod KeyMod) InputEvent {
	return InputEvent{
		Type: EventKey,
		Key: KeyEvent{
			Rune: r,
			Code: code,
			Mod:  mod,
		},
	}
}
