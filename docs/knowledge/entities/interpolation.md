---
type: Entity
id: ENT-interpolation
title: Interpolation
description: A {expr} Go expression inside markup or as an attribute value, rendered as an Element or as escaped text.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Shape
`{props.Title}`, `{len(items)}`, `{Button(ButtonProps{Children: icon})}`.

- In content: an [Element](/entities/element.md) is rendered; any other value is formatted and HTML-escaped.
- As an [attribute](/entities/attribute.md) value: formatted, attribute-escaped and double-quoted.

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md).
