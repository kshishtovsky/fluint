# Обзор архитектуры

## Слои

```
internal/term       — Сырой I/O: termios, ioctl, SIGWINCH, чтение stdin,
                      escape-парсер (клавиши, мышь, bracketed paste)
internal/platform   — OS-специфичный код за интерфейсом (build tags)
core/buffer         — Клеточная сетка (rune + style + attrs), front/back буфер
core/diff           — Back vs front → минимальный changeset
core/loop           — Frame scheduler, delta-time, event bus
core/router         — Маршрутизация событий, управление фокусом, хит-тестинг
core/viewport       — Камера/вьюпорт, виртуальный холст, куллинг, трансляция координат
render/ansi         — Changeset → ANSI bytes, synchronized output (mode 2026)
anim                — Easing, tween/timeline, delta-time интерполяция
layout              — Flexbox-подобная система компоновки (ADR-0001)
style               — Палитры, темы, текстовые атрибуты
widgets             — UI kit: API функциональных опций (ADR-0004)
examples            — Демо, VFX-витрины
```

## Направление зависимостей (DAG)

```
examples → widgets → layout → core/buffer ← core/diff ← render/ansi
                ↘ anim ↗          ↑
              core/router      core/loop
                  ↑               ↑
                  └─────── internal/term ← internal/platform
```

Циклические импорты **запрещены**.

## Инварианты

1. `render/ansi` НЕ импортирует `widgets`. `widgets` НЕ импортирует `render/ansi`. Связь только через `core/buffer`.
2. `core/*` НЕ содержит ANSI escape-последовательностей.
3. `anim` НЕ знает о терминале. Он интерполирует числа/цвета/векторы.
4. Любой слой тестируется изолированно: snapshot-тесты по `Buffer`, а не по сырым байтам.
5. Публичный API (`widgets`, `anim`, `layout`) стабилен; внутренние оптимизации (`internal/`, `core/diff`) могут меняться без предупреждения.

## Принятые архитектурные решения (ADR)

| ADR | Решение | Ссылка |
|-----|---------|--------|
| 0001 | Flexbox-подобный layout | [ADR-0001](adr/0001-layout-flexbox.md) |
| 0002 | Комбо-определение возможностей терминала | [ADR-0002](adr/0002-terminal-capabilities-combo.md) |
| 0003 | Graceful degradation при ошибках | [ADR-0003](adr/0003-error-handling-graceful.md) |
| 0004 | Functional options для API виджетов | [ADR-0004](adr/0004-widget-api-functional-options.md) |
| 0005 | Типизированные каналы в event bus | [ADR-0005](adr/0005-event-bus-typed-channels.md) |
