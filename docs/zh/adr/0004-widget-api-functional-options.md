# ADR-0004: 函数选项模式组件 API

## 状态
已接受

## 上下文
组件创建与配置的公共 API 必须具备可扩展性、类型安全性与 Go 语言地道性。

## 决策
**函数选项 (Functional options)** — 组件通过以下方式创建：
```go
btn := widgets.NewButton("OK",
    widgets.WithWidth(10),
    widgets.WithOnPress(handler),
)
```
必需参数作为构造函数入参，可选参数通过 `WithXxx` 函数指定。

## 依据
- 地道的 Go 语言扩展构造函数模式。
- 新增选项无需破坏兼容性。
- 编译期保障必需参数。
