# 变更日志 (Changelog)

本项目的所有重要变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并且本项目遵循 [语义化版本控制](https://semver.org/lang/zh-CN/spec/v2.0.0.html)。

## [未发布]

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
