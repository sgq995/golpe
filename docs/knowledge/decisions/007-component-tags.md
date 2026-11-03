---
type: Decision
id: DEC-007
title: Component tags
description: Tags that are not HTML/SVG/MathML elements (matched ignoring case) or dashed custom elements are components; their props type is read from the declaration; attributes map verbatim to fields; no children in v1.
status: stable
decision: accepted
tags: [language, composition]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
verified: { by: human:sgq995, at: 2026-11-03T03:12:42Z }
---
# Context

The maintainer first deferred composition, then asked for it: "anything else
called as an element `<myComponent>` or `<MyComponent>` becomes
`myComponent(w)` and `MyComponent(w)`; props are passed in a structured way".

# Options

- Which tags: React's rule (capitalized only), or "not a known element".
- Props type: a naming convention (`NameProps`), or read the declaration.
- Children: a `Children` field, slots (`map[string]func(io.Writer) error`,
  default slot `""`), or none for now.

# Decision

- A tag is markup when its name matches an HTML, SVG or MathML element
  ignoring ASCII case, or contains `-`. Any other Go identifier, or `pkg.Name`,
  is a component.
- The generator reads the component's declaration (template or func, in this
  package or the imported one) to get its single parameter type; no
  convention, no type-checking.
- Attributes become fields verbatim: `"a"` → string, `{expr}` → the value,
  bare → `true`.
- No children in v1: content in a component tag is an error; developers pass
  markup through props. Slots are in the [backlog](/backlog.md).

# Consequences

- Components named like elements (Button, Header, Form, Input, Table, Image…)
  cannot be unqualified tags; see [SPEC-005](/specs/005-component-tags.md).
- The CLI must read other files and packages (PLAN-005), which the LSP must
  also do on every change.
