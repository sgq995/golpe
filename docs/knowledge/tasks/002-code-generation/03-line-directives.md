---
type: Task Slice
id: TASK-002-03
title: Line directives and hygiene
description: Emit /*line*/ and //line directives before every piece of user code using the LAB-007 C-1 rule, and enforce the __golpe prefix.
spec: SPEC-002
plan: PLAN-002
state: todo
depends_on: [TASK-002-02]
touches: [internal/generator/directive.go, internal/generator/directive_test.go]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
[SPEC-002](/specs/002-code-generation.md) R9, R10 using "Line directives" in [PLAN-002](/plans/002-code-generation.md). Produces the [source map](/entities/source-map.md).

# Depends on
[TASK-002-02](/tasks/002-code-generation/02-lowering.md)

# Todo
- [ ] Tests first: for each golden case, parse the generated file and check that every user identifier maps back to its `.golpe` line:col (column-1 code: line only)
- [ ] `directive()` as in the plan; relative filenames
- [ ] Check user code never declares `__golpe*` (parser error) and generated code never emits user-visible names

# Done when
Position tests pass for all goldens.
