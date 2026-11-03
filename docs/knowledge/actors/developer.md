---
type: Actor
id: ACTOR-developer
title: Developer
description: A Go programmer who writes .golpe files to build HTML views and calls the generated functions from ordinary Go.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Who
Anyone building a web application in Go who wants HTML views that are type-checked by the Go compiler and composable as plain functions.

# Interacts with
- Writes [Golpe source files](/entities/golpe-file.md) containing [template declarations](/entities/template-declaration.md).
- Runs the [Golpe CLI](/actuators/golpe-cli.md) (`golpe generate`) and the [Go toolchain](/actors/go-toolchain.md).
- Edits in an [editor](/actors/editor.md) backed by the [language server](/actuators/language-server.md).
- Composes [Elements](/entities/element.md) freely in Go code.

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md), [SPEC-003](/specs/003-cli.md), [SPEC-004](/specs/004-language-server.md).
