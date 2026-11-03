---
type: Task Slice
id: TASK-001-05
title: Embedded statements
description: Extend the markup parser with embedded if/else, for, switch and select: header extents via go/scanner, markup-only bodies, case/default clauses.
spec: SPEC-001
plan: PLAN-001
state: todo
depends_on: [TASK-001-04]
touches: [internal/parser/markup_stmt.go, internal/parser/markup_stmt_test.go]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
[Embedded statements](/entities/embedded-statement.md) in markup, [SPEC-001](/specs/001-template-language.md) R15. Header text stays in the mask as Go so go/parser parses it (LAB-005).

# Depends on
[TASK-001-04](/tasks/001-template-language/04-markup-parser.md)

# Todo
- [ ] Tests first: `if`/`else if`/`else`; `for` 3-clause, condition-only, infinite, `range` (incl. `range 10`, `range ch`); expression switch, type switch (`switch v := x.(type)`), switch with init; `select` with send/receive cases
- [ ] Tests: `<p>if you can</p>` is text; `if` mid-line is text; nested embedded statements; error for Go statements inside embedded bodies (reported in graft, TASK-001-06 — add a pending test)
- [ ] Header extent: go/scanner from keyword to first `{` at depth 0
- [ ] Body parsing until matching `}`; `else` chaining; `case`/`default` clause splitting
- [ ] Record header fragments for the mask

# Done when
The SPEC-001 `NavItem` and `List` examples parse into the expected trees.
