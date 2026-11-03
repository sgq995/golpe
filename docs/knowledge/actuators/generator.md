---
type: Actuator
id: ACT-generator
title: Generator
description: Lowers the extended AST into plain, gofmt-ed Go that imports only the standard library: templates become functions returning func(io.Writer) error.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Responsibilities
- Lower each [template declaration](/entities/template-declaration.md) into a func returning `func(io.Writer) error`.
- Lower markup into checked writes: static HTML coalesced, values escaped, render calls through [writer()](/entities/writer-function.md), dynamic attributes with plain `if` statements.
- Lower [component tags](/entities/component-tag.md) into calls, resolving props types from declarations.
- Emit [line directives](/entities/source-map.md) and position-wrapped errors; write the [generated file](/entities/generated-file.md).

# Code
`internal/generator` (planned).

# Involved in
[SPEC-002](/specs/002-code-generation.md), [SPEC-005](/specs/005-component-tags.md), [PLAN-002](/plans/002-code-generation.md).
