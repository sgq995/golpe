---
type: Actor
id: ACTOR-go-toolchain
title: Go toolchain
description: The standard go command and compiler that build, vet and test the Go files Golpe generates.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Role
External system. Golpe never replaces it: [generated files](/entities/generated-file.md) are ordinary Go that `go build`, `go vet` and `go test` consume. Its diagnostics point back at `.golpe` files through [line directives](/entities/source-map.md).

Golpe also reuses its standard library packages `go/scanner`, `go/parser`, `go/ast`, `go/token` and `go/format` (see [DEC-001](/decisions/001-mask-and-reuse-go-parser.md)).

# Involved in
[SPEC-002](/specs/002-code-generation.md), [SPEC-003](/specs/003-cli.md).
