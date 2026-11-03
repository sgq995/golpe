---
type: Entity
id: ENT-embedded-statement
title: Embedded statement
description: An if, for, switch or select inside markup, recognized by the line rule, whose header is Go and whose bodies contain only markup.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Shape
```golpe
<ul>
  for _, item := range items {
    <li>{item.Name}</li>
  }
</ul>
```
**Line rule:** a line whose first word is `if`, `for`, `switch` or `select` starts one; every other line is markup content (write `{"if "}` to start text with such a word). Headers follow Go's grammar; `case`/`default` lines open clauses; a `}` line closes the body; `else` follows `}` on the same line; one-line bodies are allowed. Bodies contain only [markup nodes](/entities/markup-node.md).

# Involved in
[SPEC-001](/specs/001-template-language.md).
