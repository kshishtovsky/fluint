# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.9.0] - 2026-07-28

### Added
- Package `style`: `BorderStyle` type (`BorderNone`, `BorderSolid`, `BorderRounded`), `ShadowStyle` struct, border/shadow/padding fields on `Style`. Builder methods: `SolidBorder(c)`, `RoundedBorder(c)`, `NoBorder()`, `Padding(x, y)`, `Shadow(offsetX, offsetY, color)`, `NoShadow()`. Border rune constants (`BorderSolidTL`, `BorderRoundedTL`, etc.) as package-level constants for zero-alloc rendering.
- Package `widgets`: `Card` container widget — draws rounded/solid border, optional drop shadow (`░`), interior padding, and delegates rendering/events to a child `Node`. Shadow drawn behind card, border drawn 1-cell thick, child rendered inside border+padding area. `OnMouse` forwards only clicks inside the inner area. Zero-alloc render (~580 ns/op).
- Example `examples/ui_kit`: Chat UI concept with `ThinkingCardStyle` (dim border, muted bg) and `AnswerCardStyle` (cyan border, normal bg) demonstrating visual separation of content blocks.

## [v0.8.0] - 2026-07-28

### Added
- Package `widgets`: `List` widget — scrollable, virtualised list with auto-scroll, inverted selection highlight, Up/Down/Enter keyboard navigation, and left-click row selection. Only visible rows are rendered (virtualisation). `WithOnSelect` callback option. Zero-alloc render (~800 ns/op for 1000 items).
- Package `widgets`: `TextInput` widget — single-line text editor with `[]rune` storage, cursor positioning, horizontal scroll for long text, and keyboard editing (printable runes, Backspace, Delete, Left, Right, Home, End). `WithOnChange` callback option. Cursor rendered as inverted color. Zero-alloc render (~50 ns/op).
- Example `examples/ui_kit`: comprehensive demo showcasing all widgets (Text, Button, List, TextInput), layout engine, style system, animation, Router with Tab cycling, keyboard navigation, and simulated mouse clicks. Runs a scripted event sequence with real-time buffer snapshots.

## [v0.7.0] - 2026-07-28

### Added
- Package `core/viewport`: viewport/camera system for virtual canvas rendering. `Viewport` struct with `Width`/`Height` (physical dimensions) and `OffsetX`/`OffsetY` (camera scroll position). Methods: `Scroll(dx, dy)` (clamped to non-negative), `Center(x, y)` (centre camera on world point), `Resize`, `Visible(wx, wy, ww, wh)` (intersection test for culling), `ScreenX`/`ScreenY` (world-to-screen conversion). `RenderCtx` struct bundles `*buffer.Buffer` with optional `*Viewport`.
- Package `widgets`: `Node.Render` signature changed from `Render(buf, rect)` to `Render(ctx viewport.RenderCtx, rect)`. Widgets now cull (skip entirely) when fully outside viewport and clip per-cell via `buffer.SetCell`. Passing `RenderCtx{Buf: buf}` (nil viewport) preserves backward-compatible screen-space rendering.
- Package `widgets`: `Visible()` and `Screen()` helpers for viewport-aware rendering.
- Package `router`: `DispatchMouseViewport(mouse, view)` translates screen-space mouse coordinates to world-space via viewport offset before hit-testing.

### Changed
- `widgets.Node.Render` signature: `Render(buf *buffer.Buffer, rect layout.Rect)` → `Render(ctx viewport.RenderCtx, rect layout.Rect)`. All widgets, tests, and examples updated.

## [v0.6.0] - 2026-07-28

### Added
- Package `core/router`: event routing system with keyboard focus management and mouse hit-testing. `Router` dispatches key events to the focused widget and mouse events to the topmost widget under the cursor (reverse Z-order). Supports `Register`/`Unregister`, `FocusNext`/`FocusPrev` (tab cycling), and automatic `MouseEnter`/`MouseLeave` transitions. Zero-alloc `DispatchMouse` (10 widgets, ~190 ns/op).
- Package `widgets`: `Node` interface extended with `Geometry()`/`SetGeometry()`, `OnKey(KeyEvent) bool`, `OnMouse(MouseEvent) bool`, and `Focusable() bool`. Event types (`KeyEvent`, `MouseEvent`, `KeyCode`, `MouseButton`, `MouseAction`) defined in `widgets/events.go`.
- Package `widgets`: `Text` implements the full `Node` interface (not focusable, ignores all events). `Button` responds to Enter key and left-click mouse events. `HitTest(rect, x, y)` helper function for point-in-rect testing.

## [v0.5.0] - 2026-07-28

### Added
- Package `style`: value-semantic styling API with `Color` type (packed 0x00RRGGBB), 16 predefined palette colors, and `Style` struct. All modification methods (`Foreground`, `Background`, `Bold`, `Italic`, `Underline`, `Dim`, `Strikethrough`, `Reverse`) return `Style` by value — zero-allocation chainable construction. `Apply(buffer.Cell)` merges colors and attributes into a cell copy.
- Package `widgets`: `WithStyle(style.Style)` option for passing a complete style to widgets. Existing `WithForeground`/`WithBackground`/`WithBold` options are preserved for backward compatibility. `Button` gains `Style()` and `SetStyle()` accessors for animation-driven style mutation.
- Example `examples/ui_kit`: full subsystem showcase — layout (Column container with Basis/Grow), widgets (Text, Button), style, and animation (pulsing button background via `anim.Tween` with looping green ↔ yellow color interpolation at 60 fps).

## [v0.2.0] - 2026-07-28

### Added
- Package `anim`: Easing functions package (`EaseFunc`) with 9 named curves — `Linear`, `InQuad`, `OutQuad`, `InOutQuad`, `InCubic`, `OutCubic`, `InOutCubic`, `OutBounce`, `OutElastic` — backed by package-level functions (not closures) for zero-allocation hot-path usage. Only `OutElastic` uses `math.Pow`; all other curves use plain arithmetic.
- Package `core/buffer`: `Buffer.SetSubCellY(x, y, ySub int, color uint32)` method for vertical sub-cell rendering using Unicode half-block glyphs (`▀` U+2580, `▄` U+2584). Doubles vertical resolution with zero additional memory by repurposing existing `Cell.Fg`/`Cell.Bg` fields.
- ADR-0006: Sub-cell Rendering & Easing Functions architectural decision record.

## [v0.4.0] - 2026-07-28

### Added
- Package `widgets`: base widget system with Functional Options API (ADR-0004). Includes `Node` interface, shared `Config` struct, and `Option` functional option type with helpers (`WithWidth`, `WithHeight`, `WithForeground`, `WithBackground`, `WithBold`, `WithItalic`, `WithUnderline`, `WithDim`, `WithStrikethrough`, `WithReverse`, `WithOnPress`).
- Package `widgets`: `Text` widget for single-line text rendering with clipping, color, and attribute support. Zero-allocation render path.
- Package `widgets`: `Button` widget with centred label rendering, full-rect background fill, and `Press()` callback. Zero-allocation render path (0 allocs/op, ~193 ns/op).
- Git pre-commit hook (`.git/hooks/pre-commit`) running `go test -race ./...` and `golangci-lint run ./...` before every commit.

## [v0.3.0] - 2026-07-28

### Added
- Package `anim`: `Timeline` and `Tween` system for zero-allocation animation management. `Tween` holds interpolation state (`Start`, `End`, `Duration`, `EaseFunc`, `OnUpdate`, `OnComplete` callbacks). `Timeline.Add` schedules tweens; `Timeline.Update(dt)` advances all active tweens with 0 bytes/op, 0 allocs/op. Completed tweens are marked inactive and skipped. `Timeline.Compact` reclaims dead entries in-place; `Timeline.Clear` resets the timeline preserving slice capacity.
- Example `examples/vfx_demo`: real-time demo driving 10 concurrent tweens at 60 fps for 2.5 seconds, proving the system works without memory leaks.

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
