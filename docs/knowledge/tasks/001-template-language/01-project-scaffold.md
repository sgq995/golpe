---
type: Task Slice
id: TASK-001-01
title: Project scaffold
description: Create the Go module, package layout, CLI stub, okflint and the parser laboratories.
spec: SPEC-001
plan: PLAN-001
state: done
depends_on: []
touches: [go.mod, golpe.go, cmd/cli, internal/token, internal/parser, internal/ast, internal/okflint, internal/labs, lsp]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
Lay out the repository as the official Go module layout recommends so every later slice has a home. Part of [SPEC-001](/specs/001-template-language.md) / [PLAN-001](/plans/001-template-language.md).

# Depends on
Nothing.

# Todo
- [x] `go.mod` for `github.com/sgq995/golpe` (Go 1.24)
- [x] Root package `golpe` with `Element` ([DEC-002](/decisions/002-element-is-an-alias.md))
- [x] `cmd/cli` with `version`/`help` and tests
- [x] Package docs for `internal/token`, `internal/parser`, `internal/ast`, `lsp`
- [x] `internal/okflint` + `TestKnowledgeBundle`
- [x] Laboratories LAB-001..007 in `internal/labs/parserlab`
- [x] CI workflow running gofmt, vet, test

# Done when
`go vet ./...` and `go test ./...` pass on a clean checkout. ✅
