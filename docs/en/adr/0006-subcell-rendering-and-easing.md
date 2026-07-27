# ADR-0006: Sub-cell Rendering & Easing Functions

## Status
Accepted

## Context
Fluint targets smooth discrete-cell animations and advanced VFX, but the
rendering pipeline operates at cell granularity (one glyph per terminal
cell). Vertical resolution is therefore limited to the terminal height in
rows. Easing functions are the second critical primitive: every animation
interpolation — tweens, transitions, particle fades — needs a reusable
`EaseFunc` abstraction that can be called on the per-frame hot path
without allocations.

## Decision

### Sub-cell rendering (`core/buffer/subcell.go`)
Add a `Buffer.SetSubCellY(x, y, ySub int, color uint32)` method that
paints half-block "subpixels" along the vertical axis using the Unicode
half-block glyphs `▀` (U+2580, upper) and `▄` (U+2584, lower).

- `ySub == 0` → top half: colour stored in `Cell.Fg`, glyph forced to `▀`.
- `ySub == 1` → bottom half: colour stored in `Cell.Bg`, glyph becomes
  `▀` when both halves are present, or `▄` when only the bottom half was
  painted.

This doubles vertical resolution with zero additional memory — only the
Cell's existing Fg/Bg fields are repurposed.

### Easing functions (`anim/easing.go`)
A new `anim` package exports:

```
type EaseFunc func(t float64) float64
```

Nine named curves backed by package-level functions (not closures):
`Linear`, `InQuad`, `OutQuad`, `InOutQuad`, `InCubic`, `OutCubic`,
`InOutCubic`, `OutBounce`, `OutElastic`.

## Rationale
- Sub-cell rendering follows the half-block technique used by Sixel and
  braille-based TUI renderers but keeps it inside the normal `Cell`
  model — no extra grid or framebuffer.
- Named package-level functions (vs. closures stored in `var`) let the
  Go compiler inline calls through `EaseFunc` values, keeping the
  animation tick at 0 allocs/op.
- `math.Pow` is used only for `OutElastic` (the `2^(-10t)` envelope has
  no arithmetic equivalent). All other curves use plain `*` and `+`.

## Consequences
- Any renderer that serialises cells to ANSI must respect the half-block
  glyph — `render/ansi` already writes arbitrary Unicode code-points, so
  no change required today.
- Future sub-cell horizontal (quarter-block / braille) extends this
  model naturally by adding `SetSubCellX` that manipulates `Cell.Attrs`
  or a dedicated subcell flag.
- `OutBounce` and `OutElastic` may briefly overshoot [0, 1]; callers
  that need bounded output should clamp explicitly.
