---
type: Entity
id: ENT-markup-node
title: Markup node
description: A node of the HTML tree inside a template: element tag, component tag, fragment, text, interpolation, embedded statement, HTML comment or doctype.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Kinds
| Node | Source | Notes |
|------|--------|-------|
| Tag element | `<a href="/">…</a>`, `<br>` | Balanced; void and self-closing allowed; close tags match ignoring case |
| [Component tag](/entities/component-tag.md) | `<Card title="x"/>` | A call; no content in v1 |
| Fragment | `<>…</>` | Renders only its content |
| Text | `Hello, world` | Verbatim (entities work); whitespace collapsed as browsers do |
| [Interpolation](/entities/interpolation.md) | `{expr}` | Render call or escaped value |
| [Embedded statement](/entities/embedded-statement.md) | `if c { … }` | Starts a line; body is markup only |
| HTML comment | `<!-- … -->` | Copied to the output |
| Doctype | `<!DOCTYPE html>` | Its own markup statement |

Go comments in markup are template-only and produce no node in the output. A **markup statement** is the outermost tag, fragment or doctype found at a statement position in a template body. Tags carry [attributes](/entities/attribute.md).

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-005](/specs/005-component-tags.md), [PLAN-001](/plans/001-template-language.md).
