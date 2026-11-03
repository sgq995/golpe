---
type: Entity
id: ENT-template-declaration
title: Template declaration
description: A top-level 'template Name(params) { … }' declaration: a function whose body mixes Go statements and markup and which yields an Element.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Shape
```golpe
template Name[TypeParams](params) {
    // Go statements …
    <root> … </root>   // markup statement(s)
    // Go statements …
}
```
Like a `func` declaration without results; the result is implicitly an [Element](/entities/element.md). The body contains Go statements and markup statements ([markup nodes](/entities/markup-node.md)). `template` is a contextual keyword ([DEC-003](/decisions/003-template-is-a-contextual-keyword.md)).

In the AST it is `ast.TemplateDecl`, which embeds `*go/ast.FuncDecl`.

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md).
