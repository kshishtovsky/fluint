# ADR-0004: MIT 许可证

## 状态
已接受

## 上下文
项目在明确许可证下发布源代码。选择影响每个下游使用者：宽松许可
证鼓励采用，限制性许可证保留商业可选性。

我们评估了 MIT、BSD-3-Clause、Apache-2.0 和 GPL-3.0。

## 决策
**MIT 许可证** — Copyright (c) 2026 Vitaly &lt;mihaylovvsjob@gmail.com&gt;。

完整文本位于 `LICENSE.md`（英文），并镜像到 `docs/ru/LICENSE.md`
和 `docs/zh/LICENSE.md` 供非英语读者阅读。英文版本具有法律约束
力；翻译仅供参考。

## 依据

- MIT 是 Go 生态系统中接受度最广的宽松许可证，可最大程度减少
  下游用户的摩擦。
- 两段式文本简短、明确，且易于通过现有工具验证（GitHub 许可证
  检测、licensecheck）。
- 与项目从零开始的开放源码理念一致（AGENTS.md §1）— 最大化重用，
  最小化采用摩擦。

## 影响

- 下游用户无 copyleft 义务。
- 项目未附带专利授权（MIT 不包含专利授权）。如果专利保护成为
  问题，我们可以在未来的 ADR 中迁移到 Apache-2.0。
- 法律文本为英文 `LICENSE.md`；翻译仅供参考，在冲突时不能取代
  英文版本。