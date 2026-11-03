---
type: Actor
id: ACTOR-editor
title: Editor
description: An LSP client (VS Code, Neovim, Helix, …) the developer uses to edit .golpe files.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Role
External system. Starts `golpe lsp` and talks LSP over stdio to the [language server](/actuators/language-server.md).

# Involved in
[SPEC-004](/specs/004-language-server.md).
