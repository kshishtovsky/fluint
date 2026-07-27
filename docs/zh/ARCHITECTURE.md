# 架构说明

## 分层架构

```
internal/term       — 原始 I/O: termios, ioctl, SIGWINCH, stdin 读取,
                      转义序列解析器 (按键, 鼠标, bracketed paste)
internal/platform   — 构建标签背后的 OS 特性代码
core/buffer         — 单元格网格 (rune + style + attrs), front/back 缓冲区
core/diff           — Back vs front → 最小变更集
core/loop           — 帧调度器, delta-time, 事件总线
core/router         — 事件路由, 焦点管理, 命中测试
core/viewport       — 摄像机/视口, 虚拟画布, 剔除, 坐标转换
render/ansi         — 变更集 → ANSI 字节, 同步输出 (mode 2026)
anim                — 缓动, 补间/时间轴, delta-time 插值
layout              — 类 Flexbox 布局系统 (ADR-0001)
style               — 调色板, 主题, 文本属性
widgets             — UI 组件库: 函数选项 API (ADR-0004)
examples            — 示例与 VFX 演示
```

## 依赖有向无环图 (DAG)

```
examples → widgets → layout → core/buffer ← core/diff ← render/ansi
                ↘ anim ↗          ↑
              core/router      core/loop
                  ↑               ↑
                  └─────── internal/term ← internal/platform
```

严禁**循环导入**。

## 不变性规则 (Invariants)

1. `render/ansi` 不得导入 `widgets`；`widgets` 不得导入 `render/ansi`。仅通过 `core/buffer` 关联。
2. `core/*` 不得包含 ANSI 转义序列。
3. `anim` 不感知终端，仅插值数值/颜色/向量。
4. 各层独立可测：针对 `Buffer` 进行快照测试，而非原始字节。
5. 公共 API (`widgets`, `anim`, `layout`) 保持稳定；内部优化 (`internal/`, `core/diff`) 可随时变更。

## 已确定的架构决策 (ADR)

| ADR | 决策 | 链接 |
|-----|------|------|
| 0001 | 类 Flexbox 布局系统 | [ADR-0001](adr/0001-layout-flexbox.md) |
| 0002 | 组合式终端能力检测 | [ADR-0002](adr/0002-terminal-capabilities-combo.md) |
| 0003 | 优雅降级的错误处理 | [ADR-0003](adr/0003-error-handling-graceful.md) |
| 0004 | 函数选项模式组件 API | [ADR-0004](adr/0004-widget-api-functional-options.md) |
| 0005 | 类型化通道事件总线 | [ADR-0005](adr/0005-event-bus-typed-channels.md) |
