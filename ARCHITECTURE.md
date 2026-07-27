# Architecture Overview

<p align="center">
  <strong>English</strong> | <a href="docs/ru/ARCHITECTURE.md">Русский</a> | <a href="docs/zh/ARCHITECTURE.md">中文</a>
</p>

---

> For the full mirrored architecture document, see [`docs/en/ARCHITECTURE.md`](docs/en/ARCHITECTURE.md).

## Layers Summary

```
internal/term       — Raw I/O: termios, ioctl, SIGWINCH, stdin reader, escape parser
internal/platform   — OS-specific code behind build tags
core/buffer         — Cell grid (rune + style + attrs), front/back buffer
core/diff           — Back vs front → minimal changeset
core/loop           — Frame scheduler, delta-time, event bus
render/ansi         — Changeset → ANSI bytes, synchronized output (mode 2026)
anim                — Easing, tween/timeline, delta-time interpolation
layout              — Flexbox-like layout system (ADR-0001)
style               — Palettes, themes, text attributes
widgets             — UI kit: functional options API (ADR-0004)
examples            — Demos, VFX showcases
```

## Documentation by Language

- 🇬🇧 [English Architecture Document](docs/en/ARCHITECTURE.md)
- 🇷🇺 [Русский документ архитектуры](docs/ru/ARCHITECTURE.md)
- 🇨🇳 [中文架构文档](docs/zh/ARCHITECTURE.md)
