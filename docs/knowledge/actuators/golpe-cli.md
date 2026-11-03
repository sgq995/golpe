---
type: Actuator
id: ACT-golpe-cli
title: Golpe CLI
description: The golpe command (cmd/golpe), the only tool a project needs: generates _gen.go files from .golpe files and will host the language server.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Responsibilities
- `golpe generate`: find [Golpe source files](/entities/golpe-file.md), run the [parser](/actuators/parser.md) and [generator](/actuators/generator.md), write [generated files](/entities/generated-file.md).
- `golpe lsp` (later): start the [language server](/actuators/language-server.md).
- `golpe version`, `golpe help`.

# Install
`go install github.com/sgq995/golpe/cmd/golpe@latest`, or as a module tool: `go get -tool github.com/sgq995/golpe/cmd/golpe` then `go tool golpe`.

# Code
`cmd/golpe`.

# Involved in
[SPEC-003](/specs/003-cli.md).
