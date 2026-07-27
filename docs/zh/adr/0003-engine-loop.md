# ADR-0003: 帧调度器与事件循环

## 状态
已接受

## 上下文
TUI 引擎需要单一所有者来管理终端 I/O。三个并发源为引擎提供数据：

1. **原始输入** 来自 raw 模式下的 tty — 转义序列、鼠标事件、焦点
   报告。
2. **终端尺寸变化通知** 来自操作系统（Unix 上是 SIGWINCH；
   Windows 上是控制台事件）。
3. **帧节拍** 来自时钟调度器，触发 Diff → Render → Write。

如果没有显式调度器，渲染路径会被输入处理器随意调用并烧 CPU，
或者在慢写入上阻塞并丢失输入。

## 决策
`core/loop` 包拥有引擎的三个 goroutine：

- **I/O goroutine** 读取原始字节，通过 `internal/term.Parser`
  解析它们，并将类型化事件（`KeyEvent`/`MouseEvent`）分发到缓冲
  通道。通道写入是非阻塞的；过载的输入流会丢弃事件，而不是阻塞
  渲染器。
- **调度器 goroutine** 以 `time.Ticker` 频率（默认 60 Hz）触发
  `Flush`。
- **`Flush`** 串行化 Diff、Render 和单次 Write。它获取互斥锁，以
  防外部调用方与调度器竞争。写入后，它通过零分配的 `copy` 将
  `BackBuf` 提升到 `FrontBuf`。

`Stop()` 是幂等的，并拆分为 `signalStop()`（关闭 Quit，停止
ticker）和 `wg.Wait()`；I/O goroutine 在 `ErrTTYLost` 时使用
`signalStop()`，因为在 `wg` 自身跟踪的 goroutine 内部调用 `Stop()`
会导致死锁。

私有的 `ioTerm` 接口（仅读/写表面）使循环可在没有真实终端的情况
下测试。

## 依据

- **确定性** — 单个帧预算，每帧一次写入，可预测的输入延迟。
- **零分配** — 通道、缓冲区、differ 和 renderer 全部预分配；
  热路径是纯算术和 `copy`。
- **背压** — 非阻塞发送防止输入背压饿死渲染器。
- **可测试性** — `ioTerm` 接口允许测试端到端驱动循环，无需派生
  真实 tty。

## 影响

- 满载丢弃是负载下唯一安全的选择；无法容忍丢失的应用程序必须
  及时排空 `KeyEvents`/`MouseEvents`。
- `Flush` 是修改终端状态的唯一公开方法；任何绕过它的代码路径
  必须复制 back-to-front 提升。
- 尺寸变化尚未接入平台层；SIGWINCH 接入被作为后续工作跟踪。