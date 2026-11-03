---
type: Task Slice
id: TASK-001-02
title: Token extension
description: Implement internal/token: Go's tokens plus TEMPLATE and markup tokens, Lookup, IsVoid and IsRawText.
spec: SPEC-001
plan: PLAN-001
state: todo
depends_on: [TASK-001-01]
touches: [internal/token]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
Implement [Golpe tokens](/entities/golpe-token.md) as sketched in [PLAN-001](/plans/001-template-language.md) ("internal/token: extending go/token"). Serves [SPEC-001](/specs/001-template-language.md) R3, R10, R16.

# Depends on
[TASK-001-01](/tasks/001-template-language/01-project-scaffold.md)

# Todo
- [ ] Tests first: `String()` for every Go token equals go/token's; Golpe tokens have names; `FromGo`/`Go` round-trip
- [ ] Tests: `IsVoid` for the 13 void elements (case-insensitive), `IsRawText` for script/style
- [ ] `Token`, `golpeBase`, constants, `String`, `FromGo`, `Go`
- [ ] `IsVoid`, `IsRawText`, `IsEmbeddedKeyword` (if/for/switch/select)
- [ ] Package doc and examples

# Done when
`go test ./internal/token` passes with 100% statement coverage.
