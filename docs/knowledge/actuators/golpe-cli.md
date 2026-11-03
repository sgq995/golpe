---
type: Actuator
id: ACT-golpe-cli
title: Golpe CLI
description: The golpe command (cmd/cli) that generates Go from .golpe files and hosts the language server.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Responsibilities
- `golpe generate`: find [Golpe source files](/entities/golpe-file.md), run the [parser](/actuators/parser.md) and [generator](/actuators/generator.md), write [generated files](/entities/generated-file.md).
- `golpe lsp`: start the [language server](/actuators/language-server.md).
- `golpe version`, `golpe help`.

# Code
`cmd/cli` (binary name `golpe`).

# Involved in
[SPEC-003](/specs/003-cli.md).
