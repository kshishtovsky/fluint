package ansi

import (
	"bytes"
	"unicode/utf8"

	"github.com/kshishtovsky/fluint/core/buffer"
	"github.com/kshishtovsky/fluint/core/diff"
)

// Mode 2026 constants for synchronized output rendering.
const (
	Mode2026Start = "\x1b[?2026h"
	Mode2026End   = "\x1b[?2026l"
	CSI           = "\x1b["
)

// Renderer converts buffer changesets into a single pre-formatted byte payload
// suitable for single-syscall write(2) terminal rendering.
type Renderer struct {
	buf       *bytes.Buffer
	lastFg    uint32
	lastBg    uint32
	lastAttrs buffer.Attrs
	stateInit bool
}

// NewRenderer creates a new Renderer with a pre-allocated internal buffer.
func NewRenderer() *Renderer {
	return &Renderer{
		buf: bytes.NewBuffer(make([]byte, 0, 4096)),
	}
}

// Render converts changes into an ANSI byte stream wrapped in Mode 2026.
// In steady state, Render produces 0 heap allocations.
func (r *Renderer) Render(changes []diff.Change) []byte {
	r.buf.Reset()
	r.stateInit = false
	r.lastFg = 0
	r.lastBg = 0
	r.lastAttrs = 0

	r.buf.WriteString(Mode2026Start)

	var runeBuf [utf8.UTFMax]byte

	for i := range changes {
		ch := &changes[i]
		writeCursorMove(r.buf, ch.X, ch.Y)
		r.writeSGR(ch.Cell)

		ru := ch.Cell.Rune
		if ru == 0 {
			ru = ' '
		}
		n := utf8.EncodeRune(runeBuf[:], ru)
		r.buf.Write(runeBuf[:n])
	}

	r.buf.WriteString(Mode2026End)

	return r.buf.Bytes()
}

// writeCursorMove appends ANSI 1-indexed cursor position sequence (\x1b[y+1;x+1H).
func writeCursorMove(b *bytes.Buffer, x, y int) {
	b.WriteString(CSI)
	writeInt(b, y+1)
	b.WriteByte(';')
	writeInt(b, x+1)
	b.WriteByte('H')
}

// writeSGR emits SGR attribute and color escape sequences if cell styling differs
// from previous state.
func (r *Renderer) writeSGR(cell buffer.Cell) {
	if r.stateInit && cell.Fg == r.lastFg && cell.Bg == r.lastBg && cell.Attrs == r.lastAttrs {
		return
	}

	r.buf.WriteString(CSI)

	needReset := r.stateInit && (cell.Attrs != r.lastAttrs || (r.lastFg != 0 && cell.Fg == 0) || (r.lastBg != 0 && cell.Bg == 0))

	first := true
	if needReset || !r.stateInit {
		r.buf.WriteByte('0')
		first = false
	}

	if cell.Attrs != 0 {
		if cell.Attrs&buffer.Bold != 0 {
			r.appendSGRParam(1, &first)
		}
		if cell.Attrs&buffer.Dim != 0 {
			r.appendSGRParam(2, &first)
		}
		if cell.Attrs&buffer.Italic != 0 {
			r.appendSGRParam(3, &first)
		}
		if cell.Attrs&buffer.Underline != 0 {
			r.appendSGRParam(4, &first)
		}
		if cell.Attrs&buffer.Blink != 0 {
			r.appendSGRParam(5, &first)
		}
		if cell.Attrs&buffer.Reverse != 0 {
			r.appendSGRParam(7, &first)
		}
		if cell.Attrs&buffer.Strikethrough != 0 {
			r.appendSGRParam(9, &first)
		}
	}

	if cell.Fg != 0 {
		if !first {
			r.buf.WriteByte(';')
		}
		r.buf.WriteString("38;2;")
		writeInt(r.buf, int((cell.Fg>>16)&0xFF))
		r.buf.WriteByte(';')
		writeInt(r.buf, int((cell.Fg>>8)&0xFF))
		r.buf.WriteByte(';')
		writeInt(r.buf, int(cell.Fg&0xFF))
		first = false
	}

	if cell.Bg != 0 {
		if !first {
			r.buf.WriteByte(';')
		}
		r.buf.WriteString("48;2;")
		writeInt(r.buf, int((cell.Bg>>16)&0xFF))
		r.buf.WriteByte(';')
		writeInt(r.buf, int((cell.Bg>>8)&0xFF))
		r.buf.WriteByte(';')
		writeInt(r.buf, int(cell.Bg&0xFF))
		first = false
	}

	r.buf.WriteByte('m')

	r.lastFg = cell.Fg
	r.lastBg = cell.Bg
	r.lastAttrs = cell.Attrs
	r.stateInit = true
}

func (r *Renderer) appendSGRParam(param int, first *bool) {
	if !*first {
		r.buf.WriteByte(';')
	}
	writeInt(r.buf, param)
	*first = false
}

// writeInt formats an integer as ASCII bytes directly into b without heap allocations.
func writeInt(b *bytes.Buffer, v int) {
	if v == 0 {
		b.WriteByte('0')
		return
	}
	if v < 0 {
		b.WriteByte('-')
		v = -v
	}
	var buf [10]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	b.Write(buf[i:])
}
