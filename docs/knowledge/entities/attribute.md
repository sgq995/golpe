---
type: Entity
id: ENT-attribute
title: Attribute
description: A name, name="literal" or name={expr} on a tag; on markup tags dynamic values follow the true/false/nil rule, on component tags they become props fields.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# On markup tags
| Form | Example | Output |
|------|---------|--------|
| Literal, quoted | `class="btn"`, `x-data='{ open: false }'` | verbatim, no interpolation inside quotes |
| Boolean | `disabled` | ` disabled` |
| Dynamic | `href={p.Route}`, `disabled={p.Busy}` | `true` → bare name; `false`/`nil` → omitted; else `name="escaped"` |

Duplicate names and unquoted values are errors. Only the predeclared `bool` gets boolean treatment; ARIA-style `"true"`/`"false"` attributes need a string.

# On component tags
`name="a"` → field `name: "a"`; `name={expr}` → `name: expr` (value as is); bare `name` → `name: true`. See [component tag](/entities/component-tag.md).

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md), [SPEC-005](/specs/005-component-tags.md).
