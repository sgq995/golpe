---
type: Task Slice
id: TASK-006-01
title: Class and If
description: Add golpe.Class and golpe.If next to the Element alias in the optional root package, with table tests and runnable examples.
spec: SPEC-006
plan: PLAN-006
state: done
depends_on: [TASK-001-01]
touches: [golpe.go, helpers.go, helpers_test.go, example_test.go]
generated: { by: claude-code/agent, at: 2026-11-04T10:02:58Z }
---
# Goal
[SPEC-006](/specs/006-optional-helpers.md) R2–R5 via [PLAN-006](/plans/006-optional-helpers.md).

# Depends on
[TASK-001-01](/tasks/001-template-language/01-project-scaffold.md) ✅

# Todo
- [x] Tests first: `Class()` → `""`; `Class("", "a", "", "b")` → `"a b"`; `If(true, "x")` → `"x"`; `If(false, "x")` → `""`
- [x] `Class`, `If`; runnable examples, including the `cmp.Or` example from the spec
- [x] Package doc: optional; generated code never imports it

# Done when
`go test .` passes with 100% coverage.

# Notes
Implemented in [`helpers.go`](../../../../helpers.go) exactly as the PLAN-006
snippet; table tests in `helpers_test.go`, runnable examples in
`example_test.go` (`ExampleClass`, `ExampleClass_optionalParts` with the
spec's `cmp.Or` button, `ExampleIf`). Coverage 100%. No deviation from the
plan; for the coordinator to fold into SPEC-006 if wanted:

- **Parts are used as given.** `Class` does not trim, split, deduplicate or
  escape its parts, and a part of only spaces is non-empty, so it is kept
  (`Class("a", " ", "b")` is `"a   b"`). The doc comment says so and the
  tests pin it; R3 could state it too.
- **Package doc** (`golpe.go`) now lists `Element`, `Class` and `If` and
  repeats R1: removing the package only affects code the developer wrote.
- SPEC-006 acceptance criteria 1 and 2 are met; criterion 3 (no golden file
  imports the package) belongs to the generator slices.
