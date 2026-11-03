---
type: Entity
id: ENT-writer-function
title: writer()
description: The predeclared function that returns the io.Writer a template is rendering to; any interpolation containing it is a render call.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Shape
`writer() io.Writer`, valid only inside template bodies (Go mode and interpolations, not attribute values). It can be shadowed by a user declaration ([DEC-006](/decisions/006-writer-and-render-calls.md)).

# Uses
- Render a child element: `{props.Children(writer())}`.
- In Go mode: `if err := p.Header(writer()); err != nil { return err }`.
- Explicit trusted output: `io.WriteString(writer(), trustedHTML)`.

Generated code replaces it with the closure's writer variable.

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md).
