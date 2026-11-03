---
type: Entity
id: ENT-element
title: Element
description: The renderable unit: a func(io.Writer) error, spelled element inside .golpe files; golpe.Element is an optional alias.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Shape
```go
func(w io.Writer) error
```
Every [template declaration](/entities/template-declaration.md) returns one. In `.golpe` files the predeclared type `element` means exactly this type, and generated code spells it out, so it needs no import ([DEC-005](/decisions/005-standard-library-only-output.md)). The optional root package offers `golpe.Element` as an alias of the same type ([DEC-002](/decisions/002-element-is-an-alias.md)).

An element is rendered by calling it with the current writer: `{props.Children(writer())}` ([writer()](/entities/writer-function.md)). Without the call, `{props.Children}` is a plain value and prints fmt's representation of a func.

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md), [SPEC-006](/specs/006-optional-helpers.md).
