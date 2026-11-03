---
type: Task Slice
id: TASK-006-01
title: Class and If
description: Add golpe.Class and golpe.If next to the Element alias in the optional root package, with table tests and runnable examples.
spec: SPEC-006
plan: PLAN-006
state: todo
depends_on: [TASK-001-01]
touches: [golpe.go, helpers.go, helpers_test.go, example_test.go]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
[SPEC-006](/specs/006-optional-helpers.md) R2–R5 via [PLAN-006](/plans/006-optional-helpers.md).

# Depends on
[TASK-001-01](/tasks/001-template-language/01-project-scaffold.md) ✅

# Todo
- [ ] Tests first: `Class()` → `""`; `Class("", "a", "", "b")` → `"a b"`; `If(true, "x")` → `"x"`; `If(false, "x")` → `""`
- [ ] `Class`, `If`; runnable examples, including the `cmp.Or` example from the spec
- [ ] Package doc: optional; generated code never imports it

# Done when
`go test .` passes with 100% coverage.
