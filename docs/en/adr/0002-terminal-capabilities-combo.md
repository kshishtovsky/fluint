# ADR-0002: Combo Terminal Capabilities Detection

## Status
Accepted

## Context
The engine must determine what features the terminal supports (color depth, synchronized output, kitty keyboard protocol). Detection strategy affects startup latency, reliability, and feature coverage.

## Decision
**Combo** — environment variables (`$TERM`, `$COLORTERM`, `$TERM_PROGRAM`) for a reliable baseline; optional active ANSI queries (DECRQM for mode 2026, DA1) for advanced feature detection with timeouts and conservative fallback.

## Rationale
- Env vars provide zero-I/O, zero-latency baseline.
- Active queries give accurate information about real capabilities that env vars often miss (e.g., synchronized output support).
- Timeout + fallback ensures robustness in pipe scenarios or unresponsive terminals.

## Consequences
- Two code paths for detection (env-based and query-based).
- Active detection adds startup latency (bounded by timeout, ~100ms).
- Implementation: `Detect()` (env-only) and `DetectActive()` in `internal/term/caps.go`.
