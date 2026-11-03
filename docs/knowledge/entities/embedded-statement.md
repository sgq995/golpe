---
type: Entity
id: ENT-embedded-statement
title: Embedded statement
description: An if, for, switch or select inside markup whose header is Go and whose bodies contain only markup.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Shape
```golpe
<ul>
  for _, item := range items {
    <li>{item.Name}</li>
  }
</ul>
```
Headers follow Go's grammar exactly (including init statements, `range`, type switches and `case`/`default` clauses). Bodies contain only [markup nodes](/entities/markup-node.md) and further embedded statements; no Go statements.

# Involved in
[SPEC-001](/specs/001-template-language.md).
