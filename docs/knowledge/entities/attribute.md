---
type: Entity
id: ENT-attribute
title: Attribute
description: A name or name=value pair on a tag; the value is a static string or an interpolation.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Forms
| Form             | Example              |
|------------------|----------------------|
| Static, quoted   | `class="btn"`, `class='btn'` |
| Boolean          | `disabled`           |
| Dynamic          | `href={props.Route}` |

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md).
