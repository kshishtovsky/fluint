# 变更日志 (Changelog)

本项目的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并且本项目遵循 [语义化版本控制](https://semver.org/lang/zh-CN/spec/v2.0.0.html)。

## [v0.8.0] - 2026-07-28

### 新增
- `widgets` 包：`List` 组件 — 可滚动虚拟化列表，支持自动滚动、反色选中高亮、Up/Down/Enter 键盘导航和鼠标左键行选择。仅渲染可见行。`WithOnSelect` 回调选项。渲染零分配（1000 项约 800 ns/op）。
- `widgets` 包：`TextInput` 组件 — 单行文本编辑器，使用 `[]rune` 存储、光标定位、长文本水平滚动和键盘编辑（可打印字符、Backspace、Delete、Left、Right、Home、End）。`WithOnChange` 回调选项。光标以反色显示。渲染零分配（约 50 ns/op）。
- 示例 `examples/ui_kit`：全面演示所有组件（Text、Button、List、TextInput）、布局引擎、样式系统、动画、Router 的 Tab 切换、键盘导航和模拟鼠标点击。运行脚本化事件序列并实时显示缓冲区快照。

## [v0.7.0] - 2026-07-28

### 新增
- `core/viewport` 包：虚拟画布渲染的视口/摄像机系统。`Viewport` 结构体包含 `Width`/`Height`（物理尺寸）和 `OffsetX`/`OffsetY`（摄像机滚动偏移）。方法：`Scroll(dx, dy)`（限制为非负值）、`Center(x, y)`（将摄像机居中于世界坐标点）、`Resize`、`Visible(wx, wy, ww, wh)`（用于剔除的相交测试）、`ScreenX`/`ScreenY`（世界坐标到屏幕坐标转换）。`RenderCtx` 结构体将 `*buffer.Buffer` 与可选的 `*Viewport` 绑定。
- `widgets` 包：`Node.Render` 签名从 `Render(buf, rect)` 改为 `Render(ctx viewport.RenderCtx, rect)`。组件在完全超出视口时跳过渲染（剔除），并通过 `buffer.SetCell` 进行逐单元格裁剪。传入 `RenderCtx{Buf: buf}`（nil viewport）保持向后兼容的屏幕空间渲染。
- `widgets` 包：`Visible()` 和 `Screen()` 辅助函数用于视口感知渲染。
- `router` 包：`DispatchMouseViewport(mouse, view)` 在命中测试前通过视口偏移将屏幕坐标转换为世界坐标。

### 变更
- `widgets.Node.Render` 签名：`Render(buf *buffer.Buffer, rect layout.Rect)` → `Render(ctx viewport.RenderCtx, rect layout.Rect)`。所有组件、测试和示例已更新。

## [v0.5.0] - 2026-07-28

### 新增
- `style` 包：值语义样式 API，包含 `Color` 类型（packed 0x00RRGGBB）、16 个预定义调色板颜色及 `Style` 结构体。所有修改方法（`Foreground`、`Background`、`Bold`、`Italic`、`Underline`、`Dim`、`Strikethrough`、`Reverse`）按值返回 `Style`，支持零分配链式调用。`Apply(buffer.Cell)` 将颜色和属性合并到单元格副本中。
- `widgets` 包：`WithStyle(style.Style)` 选项用于向组件传递完整样式。保留现有 `WithForeground`/`WithBackground`/`WithBold` 选项以确保向后兼容。`Button` 新增 `Style()` 和 `SetStyle()` 访问器，支持动画驱动的样式变更。
- 示例 `examples/ui_kit`：完整子系统演示 — layout（含 Basis/Grow 的 Column 容器）、组件（Text、Button）、样式与动画（通过 `anim.Tween` 实现按钮背景脉冲效果，以 60 fps 循环插值 green ↔ yellow 颜色）。

## [v0.2.0] - 2026-07-28

### 新增
- `anim` 包：缓动函数包 (`EaseFunc`)，提供 9 个命名曲线 — `Linear`、`InQuad`、`OutQuad`、`InOutQuad`、`InCubic`、`OutCubic`、`InOutCubic`、`OutBounce`、`OutElastic`，基于包级函数（非闭包）实现热路径零分配调用。仅 `OutElastic` 使用 `math.Pow`；其余均为纯算术运算。
- `core/buffer` 包：`Buffer.SetSubCellY(x, y, ySub int, color uint32)` 方法，使用 Unicode 半块字形（`▀` U+2580、`▄` U+2584）实现垂直子像素渲染。通过复用现有 `Cell.Fg`/`Cell.Bg` 字段，在无需额外内存的情况下将垂直分辨率翻倍。
- ADR-0006：子像素渲染与缓动函数架构决策记录。

## [v0.4.0] - 2026-07-28

### 新增
- `widgets` 包：基于函数选项模式的基础组件系统 (ADR-0004)。包含 `Node` 接口、共享 `Config` 结构体及 `Option` 函数选项类型，附带辅助函数（`WithWidth`、`WithHeight`、`WithForeground`、`WithBackground`、`WithBold`、`WithItalic`、`WithUnderline`、`WithDim`、`WithStrikethrough`、`WithReverse`、`WithOnPress`）。
- `widgets` 包：`Text` 组件，支持单行文本渲染、裁剪、颜色与属性设置。渲染路径零分配。
- `widgets` 包：`Button` 组件，支持标签居中渲染、全矩形背景填充及 `Press()` 回调。渲染路径零分配（0 allocs/op，约 193 ns/op）。
- Git pre-commit 钩子 (`.git/hooks/pre-commit`)，在每次提交前运行 `go test -race ./...` 和 `golangci-lint run ./...`。

## [v0.3.0] - 2026-07-28

### 新增
- `anim` 包：`Timeline` 和 `Tween` 零分配动画管理系统。`Tween` 保存插值状态（`Start`、`End`、`Duration`、`EaseFunc`、`OnUpdate`/`OnComplete` 回调）。`Timeline.Add` 调度补间；`Timeline.Update(dt)` 推进所有活动补间，实现 0 bytes/op、0 allocs/op。已完成的补间标记为非活动并跳过。`Timeline.Compact` 原地回收死条目；`Timeline.Clear` 重置时间线并保留切片容量。
- 示例 `examples/vfx_demo`：实时演示，以 60 fps 驱动 10 个并发补间运行 2.5 秒，证明系统无内存泄漏。

## [v0.1.1] - 2026-07-27

### 新增
- `core/buffer` 包：基于面向数据设计 (Data-Oriented Design) 的扁平终端单元格网格 (`Buffer`)，具备 16 字节对齐的 `Cell` 内存布局、零内存分配 `Clear()` 与安全的坐标裁剪。
- `core/diff` 包：最小变更集比较引擎 (`Differ`)，通过 1D CPU 缓存优化对比实现零内存分配增量渲染。
- `render/ansi` 包：零内存分配 ANSI 序列化器 (`Renderer`)，具备 Mode 2026 同步输出封装与无分配 ASCII 格式化。
- GitHub Actions CI/CD 流水线 (`.github/workflows/ci.yml`)，在对 `main` 分支进行 push/PR 时自动运行 Go 竞态测试、`golangci-lint` 代码检查和 SonarCloud 代码分析。
- 自动化 GitHub Release 流水线 (`.github/workflows/release.yml`)，在推送版本标签 (`v*`) 时自动触发。
- `internal/term/caps.go` 中的混合终端能力检测机制 (`Detect()` 和 `DetectActive()`)，遵循 ADR-0002 支持 DECRQM mode 2026 同步输出与 Kitty 键盘协议。
- SonarCloud 配置文件 (`sonar-project.properties`)。

### 修复
- 将 `go.mod` 及 CI 流水线中的 Go 版本更新为 1.24，以解决 `golangci-lint` 的版本兼容性冲突。

## [v0.1.0] - 2026-07-27

### 新增
- 项目初始化与基础架构搭建。
- `internal/platform` 包：操作系统抽象，包含终端 Raw 模式、窗口尺寸获取、SIGWINCH 信号处理与 I/O 轮询。
- `internal/term` 包：
  - 零分配事件传递的 `InputEvent` 联合体类型。
  - 基于状态机的零分配 ANSI 转义序列解析器 (CSI, SS3, SGR 鼠标, UTF-8, bracketed paste)。
  - 2的幂次容量 `RingBuf` 环形缓冲区，实现零分配 stdin 读取。
  - 终端能力检测 (`Caps`)，支持环境变量检测与主动 DECRQM 查询。
  - 单次 `write(2)` 系统调用的预分配帧写入器 (`Writer`)。
  - 带有 ESC 序列超时处理的 Stdin 读取 Goroutine。
- `AGENTS.md` AI 代理操作指南。
- 多语言镜像文档结构 (`docs/en/`, `docs/ru/`, `docs/zh/`)。
