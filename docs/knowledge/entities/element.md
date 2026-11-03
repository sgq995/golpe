---
type: Entity
id: ENT-element
title: Element
description: The renderable unit: golpe.Element = func(io.Writer) error, spelled element inside .golpe files.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Shape
```go
type Element = func(w io.Writer) error
```
A type alias, so any `func(io.Writer) error` is an Element and vice versa ([DEC-002](/decisions/002-element-is-an-alias.md)). Every [template declaration](/entities/template-declaration.md) returns one. Interpolating an Element renders it; a nil Element renders nothing.

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md).
