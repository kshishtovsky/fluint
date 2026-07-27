# ADR-0003: Graceful Degradation for Terminal Errors

## Status
Accepted

## Context
Terminal errors (resize mid-frame, tty loss, write failure) must be handled without crashing the application. The Performance & Safety Doctrine (§3.4) prohibits panic in the render loop.

## Decision
**Graceful degradation** — errors are caught and handled internally:
- Resize mid-frame → skip current frame, re-layout next frame.
- Write error → report via error channel, attempt recovery on next frame.
- TTY loss (e.g., SSH disconnect) → cleanup terminal state, exit with error.

Errors are delivered to user code through a **typed error channel** (`chan error`).

## Rationale
- No panic in render loop (§3.4 compliance).
- Error channel is a non-blocking send (one atomic op in hot path).
- User code can handle errors asynchronously in its own goroutine.
- Terminal cleanup (restore cursor, alternate screen) is deterministic.

## Consequences
- Render loop includes error checking on every write.
- Error channel must be drained by user code to avoid blocking.
