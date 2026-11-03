---
type: Task Slice
id: TASK-001-07
title: Conformance: superset, golden, fuzz
description: Prove the superset property against GOROOT, add FuzzParseFile, and close SPEC-001's acceptance criteria.
spec: SPEC-001
plan: PLAN-001
state: todo
depends_on: [TASK-001-06]
touches: [internal/parser/conformance_test.go, internal/parser/fuzz_test.go]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
Close the acceptance criteria of [SPEC-001](/specs/001-template-language.md).

# Depends on
[TASK-001-06](/tasks/001-template-language/06-file-parser.md)

# Todo
- [ ] `TestSupersetGOROOT` (skipped with `-short`): every `.go` under GOROOT/src parses identically with go/parser and internal/parser
- [ ] `FuzzParseFile`: no panics, mask invariant, all positions within the file; seed corpus from testdata
- [ ] Tick SPEC-001 acceptance criteria; move SPEC-001 and PLAN-001 to `stable` after maintainer review

# Done when
`go test ./internal/parser` (full) passes and 10 minutes of fuzzing find nothing.
