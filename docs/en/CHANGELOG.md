# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.1] - 2026-07-27

### Added
- Package `core/buffer`: Data-Oriented 2D Cell grid (`Buffer`) with 16-byte cache-aligned `Cell` layout, zero-allocation `Clear()`, and safe coordinate clipping.
- Package `core/diff`: Minimal changeset diff engine (`Differ`) with zero-allocation steady state execution and 1D CPU cache-optimized cell comparisons.
- Package `render/ansi`: Zero-allocation ANSI Serializer (`Renderer`) with Mode 2026 synchronized output wrapping and allocation-free ASCII formatting.
- GitHub Actions CI/CD pipeline (`.github/workflows/ci.yml`) for Go testing with race detector, `golangci-lint`, and SonarCloud analysis on push/PR to `main`.
- Automated GitHub Release pipeline (`.github/workflows/release.yml`) triggered on version tag pushes (`v*`).
- Hybrid terminal capabilities detection (`Detect()` and `DetectActive()`) in `internal/term/caps.go` supporting DECRQM mode 2026 synchronized output and Kitty keyboard protocol according to ADR-0002.
- SonarCloud configuration file (`sonar-project.properties`).

### Fixed
- Updated Go directive in `go.mod` and CI pipeline to Go 1.24 to resolve `golangci-lint` release version incompatibility.

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
