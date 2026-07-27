# ADR-0005: Typed Channels for Event Bus

## Status
Accepted

## Context
The event bus connects input sources (keyboard, mouse, resize, timers) to consumers (render loop, widgets).

## Decision
**Typed channels** — one `chan T` per event type:
```go
chan KeyEvent
chan MouseEvent
chan struct{} // resize notification
```

Internal transport from `internal/term` uses a single `chan InputEvent` (tagged union struct). `core/loop` demultiplexes into the typed public channels.

## Rationale
- Compile-time type safety — no interface dispatch, no type assertions.
- Zero allocation — structs passed by value through channels.
- `select` on 3–5 channels has negligible overhead.

## Consequences
- Adding a new event type requires a new channel and `select` case.
- Implementation in `core/loop`.
