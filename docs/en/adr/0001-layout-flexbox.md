# ADR-0001: Flexbox-like Layout System

## Status
Accepted

## Context
The TUI engine needs a layout system to position and size widgets within the terminal grid. Three options were evaluated: Flexbox-like, constraint-based (Cassowary), and a hybrid.

## Decision
**Flexbox-like** — one-dimensional distribution along main/cross axis with grow, shrink, and basis properties. Two-dimensional layouts are composed by nesting row and column containers.

## Rationale
- Predictable and familiar to web developers.
- Covers ≥90% of TUI layout scenarios.
- Faster to compute than a constraint solver — important for zero-alloc render path.
- Simpler implementation, easier to debug.

## Consequences
- Complex grid layouts require nesting, increasing widget tree depth.
- No declarative cross-widget constraints.
- Implementation lives in `layout/` package.
