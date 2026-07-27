![Fluint](../repo_image.jpg)

<p align="center">

# Fluint — High-Performance Go TUI Engine

🇬🇧 <strong>English</strong> | 🇷🇺 <a href="../ru/README.md">Русский</a> | 🇨🇳 <a href="../zh/README.md">中文</a>

</p>

---

An open-source Terminal User Interface (TUI) engine for Go built **from scratch** with an emphasis on smooth cell animations, advanced VFX (easing, sub-cell rendering, transitions), and an integrated UI kit.

## Key Features

- **Built from scratch:** No dependencies on `Bubble Tea`, `tcell`, `termbox`, or `lipgloss`.
- **Zero-Allocation Hot Path:** Buffer diffing, state rendering, event dispatching, and animation ticks allocate 0 bytes per frame in steady state.
- **Single Syscall Frame Render:** Serialized ANSI diff stream sent in exactly one `write(2)` syscall per frame for tear-free rendering.
- **Advanced VFX Capabilities:** Native easing curves, sub-cell rendering (half-blocks, braille, density characters), and smooth transitions.
- **Pure Dependency Model:** Uses standard library and `golang.org/x/sys` for termios/ioctl operations.

## Architecture & Documentation

- 🇬🇧 [Architecture Overview](ARCHITECTURE.md)
- 🇬🇧 [Changelog](CHANGELOG.md)
- 🇬🇧 [Architecture Decision Records (ADR)](adr/)
