---
type: Task Slice
id: TASK-002-04
title: End-to-end golden tests
description: Golden .golpe → _golpe.go cases that are compiled and rendered in a temporary module, plus go vet position checks.
spec: SPEC-002
plan: PLAN-002
state: todo
depends_on: [TASK-002-03]
touches: [internal/generator/testdata, internal/generator/e2e_test.go]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
Close [SPEC-002](/specs/002-code-generation.md) acceptance criteria.

# Depends on
[TASK-002-03](/tasks/002-code-generation/03-line-directives.md)

# Todo
- [ ] Cases: Button, NavItem, List (from SPEC-001), escaping, failing writer, nested templates via `{Button(...)}`
- [ ] `x_golpe.go.golden` and `x.html.golden` with `-update`
- [ ] Temporary module with `replace github.com/sgq995/golpe => <repo>`; build and render
- [ ] Broken case: `go vet` reports `x.golpe:L:C`
- [ ] Determinism and gofmt checks
- [ ] Tick SPEC-002 acceptance criteria

# Done when
`go test ./internal/generator` passes.
