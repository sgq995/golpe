---
type: Actuator
id: ACT-language-server
title: Language server
description: The golpe LSP (lsp/): proxies editor requests to gopls over generated Go and maps positions back to .golpe files.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Responsibilities
- Speak LSP to the [editor](/actors/editor.md).
- Regenerate Go in memory on every change and forward requests to [gopls](/actors/gopls.md).
- Translate positions with the [source map](/entities/source-map.md).

# Code
`lsp/`.

# Involved in
[SPEC-004](/specs/004-language-server.md).
