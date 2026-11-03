---
type: Actuator
id: ACT-generator
title: Generator
description: Lowers the extended AST into plain, gofmt-ed Go: template declarations become functions returning golpe.Element.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Responsibilities
- Lower each [template declaration](/entities/template-declaration.md) into a func returning an [Element](/entities/element.md).
- Lower markup into escaped writes, coalescing static HTML.
- Emit [line directives](/entities/source-map.md) and write the [generated file](/entities/generated-file.md).

# Code
`internal/generator` (planned).

# Involved in
[SPEC-002](/specs/002-code-generation.md), [PLAN-002](/plans/002-code-generation.md).
