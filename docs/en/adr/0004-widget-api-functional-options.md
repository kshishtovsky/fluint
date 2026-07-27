# ADR-0004: Functional Options for Widget API

## Status
Accepted

## Context
The public API for creating and configuring widgets must be extensible, type-safe, and idiomatic Go. Four styles were evaluated: builder, declarative (codegen), imperative setters, and functional options.

## Decision
**Functional options** — widgets are created via:
```go
btn := widgets.NewButton("OK",
    widgets.WithWidth(10),
    widgets.WithOnPress(handler),
)
```

Mandatory parameters are constructor arguments; optional parameters are `WithXxx` option functions.

## Rationale
- Most idiomatic Go pattern for extensible constructors.
- Adding new options is non-breaking (new `WithXxx` function).
- Compile-time safety for mandatory parameters.

## Consequences
- Each optional parameter requires a `WithXxx` function definition.
- All widgets in `widgets/` package follow this pattern.
