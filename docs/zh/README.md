![Fluint](../repo_image.png)

# Fluint — 高性能 Go 语言 TUI 引擎

<p align="center">
  🇬🇧 <a href="../en/README.md">English</a> | 🇷🇺 <a href="../ru/README.md">Русский</a> | 🇨🇳 <strong>中文</strong>
</p>

---

从零开始构建的 Go 语言开源终端用户界面 (TUI) 引擎，专注于流畅的网格动画、高级 VFX 特效（缓动、亚单元格渲染、渐变过渡）以及内置 UI 组件库。

## 核心特性

- **从零构建：** 零依赖 `Bubble Tea`、`tcell`、`termbox` 或 `lipgloss`。
- **零内存分配热路径 (Zero-Allocation Hot Path)：** 缓冲区 Diff 计算、状态渲染、事件分发和动画 Tick 在稳态下每帧 0 字节内存分配。
- **单系统调用帧渲染：** 序列化的 ANSI Diff 流在每帧中严格通过一次 `write(2)` 系统调用发送，实现无撕裂 (tear-free) 渲染。
- **高级 VFX 工具包：** 内置缓动曲线 (easing curves)、亚单元格渲染 (half-blocks, braille, 密度字符) 和平滑过渡。
- **纯净依赖模型：** 仅使用 Go 标准库及 `golang.org/x/sys` 处理 termios/ioctl。

## 架构与文档

- 🇨🇳 [架构概览](ARCHITECTURE.md)
- 🇨🇳 [变更日志 (Changelog)](CHANGELOG.md)
- 🇨🇳 [架构决策记录 (ADR)](adr/)
