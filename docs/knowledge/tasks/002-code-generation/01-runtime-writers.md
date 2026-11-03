---
type: Task Slice
id: TASK-002-01
title: Runtime writers and escaping
description: Add Writer, WriteString, Write and WriteAttr to the root golpe package with HTML/attribute escaping, tests and the LAB-008 benchmark.
spec: SPEC-002
plan: PLAN-002
state: todo
depends_on: [TASK-001-01]
touches: [golpe.go, write.go, write_test.go, escape.go]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
The [runtime](/actuators/runtime.md) half of [PLAN-002](/plans/002-code-generation.md) ("Runtime"), per [DEC-004](/decisions/004-runtime-dispatch-for-interpolation.md). Covers [SPEC-002](/specs/002-code-generation.md) R7, R12.

# Depends on
[TASK-001-01](/tasks/001-template-language/01-project-scaffold.md) only, so it can run in parallel with the parser slices.

# Todo
- [ ] Tests first: `Write` for nil, nil Element, Element, string, Stringer, error, []byte, int, float, bool; `<script>` → `&lt;script&gt;`
- [ ] Tests: `WriteAttr` quotes and escapes `" ' & < >`; writer errors are returned
- [ ] `Writer` alias, `WriteString`, `Write`, `WriteAttr`
- [ ] LAB-008: `BenchmarkWrite` vs direct escaping; record numbers in PLAN-002
- [ ] Resolve SPEC-002 open question 1 (root package vs `golpe/runtime`) with the maintainer

# Done when
Tests pass; benchmark numbers recorded.
