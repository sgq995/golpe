---
type: Task Slice
id: TASK-005-04
title: Lower component tags
description: Lower ast.Component to one checked call, Name(T{fields})(w), Name(&T{…})(w) or Name()(w), with position-wrapped errors, and add golden and end-to-end cases.
spec: SPEC-005
plan: PLAN-005
state: todo
depends_on: [TASK-002-03, TASK-005-02, TASK-005-03]
touches: [internal/generator/component.go, internal/generator/testdata/components]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
[SPEC-005](/specs/005-component-tags.md) R3, R4, R6 in the generator.

# Depends on
[TASK-002-03](/tasks/002-code-generation/03-line-directives.md), [TASK-005-02](/tasks/005-component-tags/02-parse-components.md), [TASK-005-03](/tasks/005-component-tags/03-resolver.md)

# Todo
- [ ] Tests first: field mapping (`"a"` → string, `{x}` → value, bare → true); pointer props; zero params; qualified props types reuse the file's import name; a misspelled field fails `go build` at the tag's `.golpe` position
- [ ] Golden + e2e: the SPEC-005 example renders
- [ ] Tick SPEC-005 acceptance criteria

# Done when
`go test ./internal/generator` passes with the component cases.
