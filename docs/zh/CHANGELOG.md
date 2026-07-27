# 变更日志 (Changelog)

本项目的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并且本项目遵循 [语义化版本控制](https://semver.org/lang/zh-CN/spec/v2.0.0.html)。

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
