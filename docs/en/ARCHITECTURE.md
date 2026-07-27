# Architecture Overview

## Layers

```
internal/term       — Raw I/O: termios, ioctl, SIGWINCH, stdin reader,
                      escape parser (keys, mouse, bracketed paste)
internal/platform   — OS-specific code behind build tags
core/buffer         — Cell grid (rune + style + attrs), front/back buffer
core/diff           — Back vs front → minimal changeset
core/loop           — Frame scheduler, delta-time, event bus
core/router         — Event routing, focus management, hit-testing
render/ansi         — Changeset → ANSI bytes, synchronized output (mode 2026)
anim                — Easing, tween/timeline, delta-time interpolation
layout              — Flexbox-like layout system (ADR-0001)
style               — Palettes, themes, text attributes
widgets             — UI kit: functional options API (ADR-0004)
examples            — Demos, VFX showcases
```

## Dependency DAG

```
examples → widgets → layout → core/buffer ← core/diff ← render/ansi
                ↘ anim ↗          ↑
              core/router      core/loop
                  ↑               ↑
                  └─────── internal/term ← internal/platform
```

Cyclic imports are **forbidden**.

## Invariants

1. `render/ansi` does NOT import `widgets`. `widgets` does NOT import `render/ansi`. Coupling is only through `core/buffer`.
2. `core/*` contains NO ANSI escape sequences.
3. `anim` knows nothing about the terminal. It interpolates numbers/colors/vectors.
4. Every layer is testable in isolation: snapshot tests against `Buffer`, not raw bytes.
5. Public API (`widgets`, `anim`, `layout`) is stable; internal optimizations (`internal/`, `core/diff`) may change without notice.

## Architecture Decision Records (ADR)

| ADR | Decision | Link |
|-----|----------|------|
| 0001 | Flexbox-like layout | [ADR-0001](adr/0001-layout-flexbox.md) |
| 0002 | Combo terminal detection | [ADR-0002](adr/0002-terminal-capabilities-combo.md) |
| 0003 | Graceful error degradation | [ADR-0003](adr/0003-error-handling-graceful.md) |
| 0004 | Functional options widget API | [ADR-0004](adr/0004-widget-api-functional-options.md) |
| 0005 | Typed channels event bus | [ADR-0005](adr/0005-event-bus-typed-channels.md) |
