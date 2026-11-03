---
type: Entity
id: ENT-interpolation
title: Interpolation
description: A {expr} Go expression in markup: a render call if it contains writer(), otherwise a value written as escaped text.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Kinds
| Kind | Example | Generated |
|------|---------|-----------|
| Render call | `{props.Children(writer())}`, `{Button(p)(writer())}` | `if err := …; err != nil { return wrapped }` |
| Value | `{props.Title}`, `{len(items)}` | nil interface → nothing; else `html.EscapeString(fmt.Sprint(x))` |
| Quoted text | `{"if "}`, `{"{"}` | a value; used to escape the line rule and braces |

As an [attribute](/entities/attribute.md) value, `{expr}` is always a value; `writer()` is not allowed there ([DEC-006](/decisions/006-writer-and-render-calls.md)).

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md).
