---
type: Spec
id: SPEC-006
title: Optional helpers
description: The root package github.com/sgq995/golpe offers small, optional helpers (Element, Class, If) that projects may import; generated code never depends on it.
status: draft
plan: PLAN-006
tags: [api]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Summary

Golpe works with the CLI alone ([DEC-005](/decisions/005-standard-library-only-output.md)).
The root package is a small toolbox for developers who want it.

# Motivation

Some patterns repeat in every project, such as building a `class` value from
optional parts. A tiny optional helper is clearer than ad-hoc string
concatenation, as long as it never becomes a dependency of generated code.

# Actors

[Developer](/actors/developer.md) · [Runtime helpers](/actuators/runtime.md)

# Entities

[Element](/entities/element.md) · [Attribute](/entities/attribute.md)

# Requirements

- **R1 Optional.** Generated code MUST NOT import this package. Removing it
  from a project must only affect code the developer wrote.
- **R2 Element.** `type Element = func(io.Writer) error`
  ([DEC-002](/decisions/002-element-is-an-alias.md)).
- **R3 Class.** `Class(parts ...string) string` joins the non-empty parts
  with single spaces, skipping empty ones the way `cmp.Or` skips zero values.
- **R4 If.** `If(cond bool, s string) string` returns `s` when `cond` is true,
  otherwise `""`.
- **R5** Standard library only; no state; no generated-code entry points.

# Examples

```golpe
<button class={golpe.Class("btn", cmp.Or(p.Variant, "primary"), golpe.If(p.Active, "active"))}>
```

`Class("btn", "primary", "")` returns `"btn primary"`.

# Acceptance criteria

- [ ] `Class()` is `""`; `Class("", "a", "", "b")` is `"a b"`.
- [ ] `If(true, "x")` is `"x"`; `If(false, "x")` is `""`.
- [ ] No generated golden file imports `github.com/sgq995/golpe`.

# Out of scope

Anything generated code would need. Attribute spread types
([backlog](/backlog.md)).
