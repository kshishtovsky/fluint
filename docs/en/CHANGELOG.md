# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - 2026-07-27

### Added
- Project bootstrap and infrastructure setup.
- Package `internal/platform`: OS abstractions for raw terminal mode, window size, SIGWINCH signals, and I/O polling.
- Package `internal/term`:
  - Tagged union `InputEvent` for allocation-free event passing.
  - Zero-allocation state-machine ANSI escape sequence parser (CSI, SS3, SGR mouse, UTF-8, bracketed paste).
  - Power-of-2 ring buffer (`RingBuf`) for zero-allocation stdin reading.
  - Terminal capability detection (`Caps`) supporting environment checks and active DECRQM queries.
  - Single-syscall pre-allocated frame writer (`Writer`).
  - Stdin reader goroutine with ESC sequence timeout handling.
- `AGENTS.md` guidelines for AI coding agents.
- Multilingual documentation mirrored structure (`docs/en/`, `docs/ru/`, `docs/zh/`).
