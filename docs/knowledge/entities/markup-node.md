---
type: Entity
id: ENT-markup-node
title: Markup node
description: A node of the HTML tree inside a template: element tag, text, interpolation, embedded statement, HTML comment or doctype.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Kinds
| Node                | Source                       | Notes                                   |
|---------------------|------------------------------|-----------------------------------------|
| Tag element         | `<a href="/">…</a>`, `<br>` | Balanced; void and self-closing allowed |
| Text                | `Hello, world`               | Literal, HTML-escaped at compile time   |
| [Interpolation](/entities/interpolation.md) | `{expr}`  | Go expression                           |
| [Embedded statement](/entities/embedded-statement.md) | `if c { … }` | Body is markup only        |
| HTML comment        | `<!-- … -->`                 | Copied to the output                    |
| Doctype             | `<!DOCTYPE html>`            | Its own markup statement                |

A **markup statement** is the outermost tag element (plus its subtree) found at a statement position in a template body. Tags carry [attributes](/entities/attribute.md).

# Involved in
[SPEC-001](/specs/001-template-language.md), [PLAN-001](/plans/001-template-language.md).
