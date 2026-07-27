# ADR-0005: 1D Flexbox Layout Completion

## Status
Accepted

## Context
ADR-0001 selected a 1D flexbox layout engine but did not implement it.
We needed to deliver the package so applications can build component
trees immediately, while keeping the zero-allocation contract from
AGENTS.md §7.

## Decision
The `layout` package implements `Container`, `Leaf`, and `Child` types
per ADR-0001, with three integer layout properties (`Basis`, `Grow`,
`Shrink`) and two directions (`Row`, `Column`).

Three algorithmic rules govern distribution:

1. **Phase 1 — Basis**: every `Basis > 0` child receives exactly
   that much on the main axis; remaining sum is tallied.
2. **Phase 2 — Grow / Shrink**: free space is split by `Grow`, or
   overflow is absorbed by `Shrink`. Integer rounding remainder
   distributes one unit at a time to the first eligible children.
3. **Phase 3 — Layout**: each child is positioned along the main
   axis; the last child absorbs any rounding residual so
   `sum(child main sizes) == available`.

Zero allocation requires a stack-allocated per-container scratch
buffer for child sizes. The buffer holds 256 entries; deeper
containers fall back to a one-shot heap allocation in
`measureSlow`. Cross-axis size is passed through unchanged (Row
children all get the parent's height; Column children all get the
parent's width).

## Rationale

- The CSS flexbox mental model minimises the learning curve for new
  contributors.
- The three-phase algorithm is O(N) per container and trivially
  verifiable by hand.
- Pre-allocating the result slice and threading it through
  recursion keeps the hot path at 0 B/op, 0 allocs/op (verified by
  benchmark).

## Consequences

- Multi-axis grids require nesting (Row → Column → Row …) — the same
  trade-off the spec already accepted.
- `measureSlow` allocates once for containers wider than 256
  children; in practice this never triggers on a TUI tree.
- The `Child.Node` field is optional; `nil` produces a layout rect
  without recursion.