---
type: Task Slice
id: TASK-002-03
title: Line directives and wrapped errors
description: Emit line directives before every piece of user code using the LAB-007 C-1 rule, wrap every write and render-call error with the .golpe position and template name, and enforce the __golpe prefix.
spec: SPEC-002
plan: PLAN-002
state: todo
depends_on: [TASK-002-02]
touches: [internal/generator/directive.go, internal/generator/emit.go, internal/generator/directive_test.go]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
[SPEC-002](/specs/002-code-generation.md) R13–R15: positions for the compiler and for runtime errors. Produces the [source map](/entities/source-map.md).

# Depends on
[TASK-002-02](/tasks/002-code-generation/02-lowering.md)

# Todo
- [ ] Tests first: for each golden case, parse the generated file and check that every user identifier maps back to its `.golpe` line:col (column-1 code: line only)
- [ ] Tests: a failing writer makes the error read `x.golpe:L:C: Template: <cause>`; nested templates chain positions; `errors.Is` still finds the cause; a user `return err` is not wrapped
- [ ] `directive()` as in PLAN-001 LAB-007; relative file names
- [ ] `wrap()` with `fmt.Errorf("…: %w", err)`

# Done when
Position and wrapping tests pass for all goldens.
