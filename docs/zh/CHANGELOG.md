# 变更日志 (Changelog)

本项目的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并且本项目遵循 [语义化版本控制](https://semver.org/lang/zh-CN/spec/v2.0.0.html)。

## [未发布]

### 新增
- GitHub Actions CI/CD 流水线 (`.github/workflows/ci.yml`)，在对 `main` 分支进行 push/PR 时自动运行 Go 竞态测试、`golangci-lint` 代码检查和 SonarQube 代码分析。
- 自动化 GitHub Release 流水线 (`.github/workflows/release.yml`)，在推送版本标签 (`v*`) 时自动触发。
- `internal/term/caps.go` 中的混合终端能力检测机制 (`Detect()` 和 `DetectActive()`)，遵循 ADR-0002 支持 DECRQM mode 2026 同步输出与 Kitty 键盘协议。
- SonarQube 配置文件 (`sonar-project.properties`)。

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
