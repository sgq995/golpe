---
type: Entity
id: ENT-template-declaration
title: Template declaration
description: A top-level 'template Name(params) { … }' declaration: a function whose body mixes Go statements and markup and which returns func(io.Writer) error.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Shape
```golpe
template Name[TypeParams](params) {
    // Go statements …
    <root> … </root>   // markup statement(s)
    // Go statements …
}
```
Like a `func` declaration without results; the result is `func(io.Writer) error` ([Element](/entities/element.md)). The body runs on every render. `template` is a contextual keyword ([DEC-003](/decisions/003-template-is-a-contextual-keyword.md)). Templates with zero or one parameter can be used as [component tags](/entities/component-tag.md).

In the AST it is `ast.TemplateDecl`, which embeds `*go/ast.FuncDecl`.

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md), [SPEC-005](/specs/005-component-tags.md).
