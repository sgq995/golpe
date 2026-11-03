---
type: Actuator
id: ACT-okflint
title: okflint
description: internal/okflint: validates this knowledge base as an OKF v0.2 bundle plus the project's conventions.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Responsibilities
Checks the rules in [conventions](/process/conventions.md): frontmatter and `type` on every concept, reserved `index.md`/`log.md`, required keys per type, allowed `status`/`state`/`decision` values, unique IDs, every concept listed in its folder index, and no broken bundle links.

# Usage
`go run ./internal/okflint` (also run by `go test ./...`).
