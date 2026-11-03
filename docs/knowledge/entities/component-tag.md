---
type: Entity
id: ENT-component-tag
title: Component tag
description: A tag whose name is not an HTML/SVG/MathML element or custom element: a call to a template or func with props built from its attributes.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Shape
`<Card title="Hi" featured/>`, `<ui.Button label={l}/>`, `<Footer/>`.

- Classified by name ([DEC-007](/decisions/007-component-tags.md)): element names match ignoring case; dashed names are custom elements; anything else is a component.
- Its declaration is read syntactically to find the props type: zero params → `Name()(w)`; `T` → `Name(T{…})(w)`; `*T` → `Name(&T{…})(w)`.
- [Attributes](/entities/attribute.md) map verbatim to fields; no children in v1.

# Involved in
[SPEC-005](/specs/005-component-tags.md).
