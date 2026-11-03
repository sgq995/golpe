---
type: Task Slice
id: TASK-001-05
title: Embedded statements
description: Implement the line rule in the markup parser: if/for/switch/select lines start embedded statements with Go headers, case/default clauses, } closes, else on the same line, one-line bodies, quoting hints.
spec: SPEC-001
plan: PLAN-001
state: todo
depends_on: [TASK-001-04]
touches: [internal/parser/markup_stmt.go, internal/parser/markup_stmt_test.go]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
[Embedded statements](/entities/embedded-statement.md) by the line rule, [SPEC-001](/specs/001-template-language.md) R18. Header text stays in the mask as Go so go/parser parses it (LAB-005).

# Depends on
[TASK-001-04](/tasks/001-template-language/04-markup-parser.md)

# Todo
- [ ] Tests first: `if`/`else if`/`else`; `for` 3-clause, condition-only, infinite, `range` (incl. `range 10`); expression switch, type switch, switch with init; `select` with send/receive cases; nested statements; one-line body `if ok { <b>yes</b> }`
- [ ] Tests for the rule itself: a line whose first word is `if`/`for`/`switch`/`select` is a statement even in prose (error with the quoting hint `write {"if"} …`); `format`, `iffy` are text; `if` mid-line is text; `{"if "}you can` is text; `else` on its own line is an error
- [ ] Header extent: go/scanner from the keyword to the first `{` at depth 0
- [ ] Body parsing until the matching `}`; `else` chaining on the same line; `case`/`default` clause splitting in switch/select bodies
- [ ] Record header fragments for the mask

# Done when
The SPEC-001 `NavItem` and `List` examples parse into the expected trees.
