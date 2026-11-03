---
type: Backlog
id: BACKLOG
title: Backlog
description: Ideas deliberately left out of the current phase, each with the reason and the decision that parked it, to be planned in isolation later.
tags: [planning]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# How to use this

Each item becomes its own spec when the maintainer asks for it (see the
[feature workflow](/process/feature-workflow.md)). Items marked **rejected**
were decided against and need a new decision to revive.

# Later phases

| Idea | Notes | From |
|------|-------|------|
| Slots / children for component tags | Maintainer's idea: `map[string]func(io.Writer) error`, default slot `""` as children; may need templates shaped as `func(props, slots)` | [DEC-007](/decisions/007-component-tags.md) |
| Attribute spread | `<div {attrs}>`; needs a value type that generated code understands without importing golpe | Interview round 9 |
| `json` keyword | Maintainer's idea: a special keyword returning a structured type rendered with `encoding/json` (v1 or v2) | Interview round 7 |
| Scoped `<style>` | Needs CSS parsing; a project of its own | Interview round 7 |
| Emit script/style once per page | Needs shared render state (a context), which DEC-008 rejects for now | Interview round 7 |
| URL and JavaScript safety | Sanitize dynamic `href`/`src`/`action` values and `on*` handlers; its own security spec | [DEC-009](/decisions/009-escape-by-default.md) |
| Type-aware lowering | Emit type-specific writes using go/types; must produce identical bytes | [DEC-005](/decisions/005-standard-library-only-output.md) |
| Generic components as tags | Type arguments cannot be inferred from composite literals | [SPEC-005](/specs/005-component-tags.md) |
| `golpe fmt` | Formatter for `.golpe` files | [SPEC-003](/specs/003-cli.md) |
| Watch mode | `golpe generate -watch`; the LSP may make it unnecessary | [SPEC-003](/specs/003-cli.md) |

# Rejected

| Idea | Why | From |
|------|-----|------|
| Implicit `context.Context` | Needs special syntax; explicit parameters and writer wrapping suffice | [DEC-008](/decisions/008-no-implicit-context.md) |
| Runtime helper package for generated code | Generated code must use only the standard library | [DEC-005](/decisions/005-standard-library-only-output.md) |
| Conditional attribute groups (`if` inside a tag) | Odd to read and hard to parse; `name={cond}` covers single attributes | Interview round 6 |
| Interpolation inside quoted attribute values | Would break Alpine/Vue values such as `x-data="{ open: false }"`; use `name={expr}` | Interview round 6 |
| Trusted-by-default interpolation | XSS by default | [DEC-009](/decisions/009-escape-by-default.md) |
| Web framework and client-side reactivity | Golpe is server-side templates only | Interview round 1 |
