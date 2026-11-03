---
type: Task Slice
id: TASK-003-01
title: generate command
description: Add golpe generate with pattern expansion, parallel per-file generation, write-if-changed, -check, -clean and sorted diagnostics.
spec: SPEC-003
plan: PLAN-003
state: todo
depends_on: [TASK-002-02]
touches: [cmd/cli, internal/build]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
All of [SPEC-003](/specs/003-cli.md) R1–R5, R7 using [PLAN-003](/plans/003-cli.md).

# Depends on
[TASK-002-02](/tasks/002-code-generation/02-lowering.md) (directives from TASK-002-03 are nice to have, not required).

# Todo
- [ ] Tests first: `Expand` (`./...`, files, dirs, skipped dirs); write-if-changed keeps mtimes; `-check` exit codes; `-clean` removes orphaned generated files only when they have the header; diagnostics sorted and relative
- [ ] `internal/build`: `Expand`, `GenerateFile`, worker pool
- [ ] `cmd/cli/generate.go` and wiring in `run`
- [ ] Decide the binary name question with the maintainer (SPEC-003 open question 1)

# Done when
SPEC-003 acceptance criteria ticked.
