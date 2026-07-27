# ADR-0006: 子像素渲染与缓动函数

## 状态
已接受

## 上下文
Fluint 的目标是实现平滑的离散单元格动画和高级 VFX，但渲染管线
以单元格粒度运行（每个终端单元格一个字形）。因此垂直分辨率受限于
终端的行数。缓动函数是第二个关键原语：每次动画插值（补间、过渡、
粒子淡出）都需要一个可在每帧热路径上无分配调用的可复用 `EaseFunc`
抽象。

## 决策

### 子像素渲染 (`core/buffer/subcell.go`)
新增 `Buffer.SetSubCellY(x, y, ySub int, color uint32)` 方法，使用
Unicode 半块字形 `▀` (U+2580, 上半) 和 `▄` (U+2584, 下半) 沿
垂直轴绘制半块"子像素"。

- `ySub == 0` → 上半部分：颜色存入 `Cell.Fg`，字形设为 `▀`。
- `ySub == 1` → 下半部分：颜色存入 `Cell.Bg`，当两半均存在时字形
  为 `▀`，仅下半存在时为 `▄`。

无需额外内存即可实现垂直分辨率翻倍——仅复用单元格已有的 Fg/Bg 字段。

### 缓动函数 (`anim/easing.go`)
新包 `anim` 导出：

```
type EaseFunc func(t float64) float64
```

九个命名曲线，基于包级函数（非闭包）：
`Linear`、`InQuad`、`OutQuad`、`InOutQuad`、`InCubic`、`OutCubic`、
`InOutCubic`、`OutBounce`、`OutElastic`。

## 依据
- 子像素渲染遵循 Sixel 和 braille-TUI 渲染器使用的半块技术，但
  保持在普通 `Cell` 模型内——无需额外的帧缓冲区。
- 命名包级函数（而非存储在 `var` 中的闭包）让 Go 编译器能够通过
  `EaseFunc` 值内联调用，使动画 tick 保持在 0 allocs/op。
- `math.Pow` 仅用于 `OutElastic`（`2^(-10t)` 包络无算术等价形式）。
  其他所有曲线仅使用 `*` 和 `+`。

## 影响
- 任何将单元格序列化为 ANSI 的渲染器必须正确处理半块字形——
  `render/ansi` 已支持任意 Unicode 码位，因此当前无需修改。
- 水平子像素（四分块 / braille）可通过添加 `SetSubCellX` 自然扩展。
- `OutBounce` 和 `OutElastic` 可能短暂超出 [0, 1]；需要有界输出的
  调用方应显式钳位。
