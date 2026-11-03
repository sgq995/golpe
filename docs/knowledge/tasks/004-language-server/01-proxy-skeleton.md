---
type: Task Slice
id: TASK-004-01
title: Proxy skeleton
description: Stdlib JSON-RPC framing, gopls subprocess, overlay sync of generated Go, source map, and mapped diagnostics.
spec: SPEC-004
plan: PLAN-004
state: todo
depends_on: [TASK-002-03, TASK-003-01]
touches: [lsp, cmd/cli]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
First usable `golpe lsp`: [SPEC-004](/specs/004-language-server.md) R1–R3, R5 via [PLAN-004](/plans/004-language-server.md). R4/R6 follow in later slices.

# Depends on
[TASK-002-03](/tasks/002-code-generation/03-line-directives.md), [TASK-003-01](/tasks/003-cli/01-generate-command.md). SPEC-004 is still draft: confirm scope with the maintainer before starting.

# Todo
- [ ] LAB-009 and LAB-010 from PLAN-004; record results
- [ ] Tests first: JSON-RPC framing; source map round-trip; proxy with a fake gopls
- [ ] `lsp/jsonrpc.go`, `lsp/protocol.go`, `lsp/sourcemap.go`, `lsp/proxy.go`, `lsp/server.go`
- [ ] `golpe lsp` subcommand

# Done when
A type error in an interpolation shows up at the right range in a scripted client test.
