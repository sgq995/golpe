---
type: Task Slice
id: TASK-003-01
title: generate command
description: Add golpe generate to cmd/golpe: pattern expansion, parallel per-file generation, _gen.go writes only when changed, -check, -clean for golpe-headed orphans, sorted diagnostics.
spec: SPEC-003
plan: PLAN-003
state: todo
depends_on: [TASK-002-02]
touches: [cmd/golpe, internal/build]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
[SPEC-003](/specs/003-cli.md) R2–R6, R8 using [PLAN-003](/plans/003-cli.md).

# Depends on
[TASK-002-02](/tasks/002-code-generation/02-lowering.md) (directives from TASK-002-03 are nice to have, not required).

# Todo
- [ ] Tests first: `Expand` (`./...`, files, dirs, skipped dirs); write-if-changed keeps mtimes; `-check` exit codes; `-clean` removes only orphans whose first line is golpe's header; diagnostics sorted and relative
- [ ] `internal/build`: `Expand`, `GenerateFile`, `IsGolpeGenerated`, worker pool
- [ ] `cmd/golpe/generate.go` and wiring in `run`
- [ ] LAB-012: `go get -tool` + `go tool golpe version` in a fresh module

# Done when
SPEC-003 acceptance criteria ticked.
