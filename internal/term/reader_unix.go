//go:build linux || darwin

package term

import (
	"context"
	"syscall"
	"time"

	"github.com/kshishtovsky/fluint/internal/platform"
)

const (
	// escTimeout is the duration to wait for additional bytes after
	// receiving ESC to disambiguate standalone ESC from the start of
	// an escape sequence.
	escTimeout = 50 * time.Millisecond

	// pollTimeout is the interval between context cancellation checks
	// while waiting for stdin data.
	pollTimeout = 100 * time.Millisecond
)

// ReadInput reads terminal input from fd and sends parsed events to
// the events channel. It blocks until ctx is cancelled or an
// unrecoverable error occurs.
//
// The channel is NOT closed by ReadInput — the caller is responsible
// for channel lifecycle.
//
// Concurrency: this function is designed to run in a dedicated
// goroutine. It is the sole reader of fd.
func ReadInput(ctx context.Context, fd uintptr, events chan<- InputEvent) error {
	var (
		rb     RingBuf
		parser Parser
	)

	for {
		// Check context before blocking on poll.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Wait for data with timeout to periodically check ctx.
		ready, err := platform.PollReadable(fd, pollTimeout)
		if err != nil {
			return err
		}

		if !ready {
			// No data within timeout. If the parser has a pending ESC,
			// flush it as a standalone Escape key.
			if parser.InEscapeState() {
				if ev, ok := parser.Flush(); ok {
					if err := sendEvent(ctx, events, ev); err != nil {
						return err
					}
				}
			}
			continue
		}

		// Read into ring buffer.
		space := rb.WritableSlice()
		if len(space) == 0 {
			// Buffer unexpectedly full — should not happen in practice
			// since the parser consumes all bytes. Reset as safety valve.
			rb.Reset()
			space = rb.WritableSlice()
		}

		n, err := syscall.Read(int(fd), space)
		if err != nil {
			if err == syscall.EINTR {
				continue
			}
			return err
		}
		if n == 0 {
			continue
		}
		rb.CommitWrite(n)

		// Parse events from ring buffer. Data may wrap around the
		// buffer boundary, so we loop on Peek until all data is consumed.
		if err := drainBuffer(ctx, &rb, &parser, events); err != nil {
			return err
		}

		// After parsing, check for ESC timeout.
		if parser.InEscapeState() {
			ok, err := platform.PollReadable(fd, escTimeout)
			if err != nil {
				return err
			}
			if !ok {
				// Timeout — emit standalone ESC.
				if ev, flushed := parser.Flush(); flushed {
					if err := sendEvent(ctx, events, ev); err != nil {
						return err
					}
				}
			}
			// If ok, the next loop iteration will read the data.
		}
	}
}

// drainBuffer feeds all available ring buffer data to the parser and
// sends resulting events to the channel.
func drainBuffer(ctx context.Context, rb *RingBuf, parser *Parser, events chan<- InputEvent) error {
	for rb.Len() > 0 {
		chunk := rb.Peek()
		consumed := 0

		for consumed < len(chunk) {
			ev, ate, ok := parser.Next(chunk[consumed:])
			consumed += ate

			if ok {
				if err := sendEvent(ctx, events, ev); err != nil {
					return err
				}
			}

			// Drain any pending events from re-processing.
			for {
				pev, pok := parser.Next(nil)
				if !pok {
					break
				}
				if err := sendEvent(ctx, events, pev); err != nil {
					return err
				}
			}

			if ate == 0 && !ok {
				break
			}
		}

		rb.Consume(consumed)
	}
	return nil
}

// sendEvent sends an event to the channel, respecting context cancellation.
func sendEvent(ctx context.Context, events chan<- InputEvent, ev InputEvent) error {
	select {
	case events <- ev:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
