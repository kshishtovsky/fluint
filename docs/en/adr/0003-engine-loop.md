# ADR-0003: Frame Scheduler and Event Loop

## Status
Accepted

## Context
The TUI engine needs a single owner for terminal I/O. Three concurrent
sources feed the engine:

1. **Raw input** from the tty in raw mode — escape sequences, mouse
   events, focus reports.
2. **Resize notifications** from the OS (SIGWINCH on Unix; console
   events on Windows).
3. **Frame ticks** from a wall-clock scheduler, which trigger Diff →
   Render → Write.

Without an explicit scheduler, the render path is invoked ad-hoc from
input handlers and burns CPU, or blocks on slow writes and drops input.

## Decision
A `core/loop` package owns the engine's three goroutines:

- **I/O goroutine** reads raw bytes, parses them through
  `internal/term.Parser`, and dispatches typed events
  (`KeyEvent`/`MouseEvent`) to buffered channels. Channel writes are
  non-blocking; an over-fed input stream drops events rather than
  stalling the renderer.
- **Scheduler goroutine** fires `Flush` on a `time.Ticker` cadence
  (60 Hz by default).
- **`Flush`** serialises Diff, Render, and a single Write. It
  acquires a mutex so external callers and the scheduler cannot race.
  After writing, it promotes `BackBuf` → `FrontBuf` with an
  allocation-free `copy`.

`Stop()` is idempotent and splits into `signalStop()` (close Quit,
stop ticker) and `wg.Wait()`; the I/O goroutine uses `signalStop()`
on `ErrTTYLost` because calling `Stop()` from inside a goroutine that
`wg` itself tracks would deadlock.

A private `ioTerm` interface (read/write surface only) keeps the loop
testable without a real terminal.

## Rationale

- **Determinism** — one frame budget, one write per frame, predictable
  input latency.
- **Zero allocations** — channels, buffers, differ, and renderer are
  all pre-allocated; the hot path is pure arithmetic and `copy`.
- **Backpressure** — non-blocking sends prevent input backpressure
  from starving the renderer.
- **Testability** — the `ioTerm` interface lets tests drive the loop
  end-to-end without forking a real tty.

## Consequences

- Drop-on-full input is the only safe choice under load; applications
  that cannot tolerate loss must drain `KeyEvents`/`MouseEvents`
  promptly.
- `Flush` is the only public method that mutates terminal state; any
  code path that bypasses it must replicate the back-to-front
  promotion.
- Resize is not yet wired into the platform layer; SIGWINCH plumbing
  is tracked as a follow-up.