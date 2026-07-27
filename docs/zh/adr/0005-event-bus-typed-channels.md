# ADR-0005: 类型化通道事件总线

## 状态
已接受

## 上下文
连接输入源（键盘、鼠标、resize、定时器）与消费端。

## 决策
**类型化通道** — 每种事件类型一个 `chan T` (`chan KeyEvent`, `chan MouseEvent`, `chan struct{}`)。内部通过 `chan InputEvent` (tagged union) 传输。

## 依据
- 编译期类型安全，无接口派发或类型断言。
- 结构体按值传递，零内存分配。
- 3–5 个通道上的 `select` 开销极低。
