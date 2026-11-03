---
type: Task Slice
id: TASK-002-04
title: End-to-end golden tests
description: Golden .golpe → _gen.go cases that are compiled and rendered in a temporary standard-library-only module, plus go vet position checks.
spec: SPEC-002
plan: PLAN-002
state: todo
depends_on: [TASK-002-03]
touches: [internal/generator/testdata, internal/generator/e2e_test.go]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
Close [SPEC-002](/specs/002-code-generation.md) acceptance criteria.

# Depends on
[TASK-002-03](/tasks/002-code-generation/03-line-directives.md)

# Todo
- [ ] Cases: Button, NavItem, List, Layout (SPEC-001), escaping, boolean attributes, nil values, failing writer, nested templates via `{Button(p)(writer())}`
- [ ] `x_gen.go.golden` and `x.html.golden` with `-update`
- [ ] Temporary module with no dependencies; build and render each case
- [ ] Assert generated files import only the standard library
- [ ] Broken case: `go vet` reports `x.golpe:L:C`
- [ ] Determinism and gofmt checks; tick SPEC-002 acceptance criteria

# Done when
`go test ./internal/generator` passes.
