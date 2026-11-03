---
type: Entity
id: ENT-golpe-token
title: Golpe token
description: internal/token.Token: go/token's token kinds extended with TEMPLATE and the markup tokens.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Shape
A `Token` whose values below a base are exactly go/token's, and above it Golpe's: `TEMPLATE`, `TAG_OPEN`, `TAG_CLOSE`, `TAG_END`, `TAG_SELF_CLOSE`, `TEXT`, `INTERP_OPEN`, `INTERP_CLOSE`, `HTML_COMMENT`, `DOCTYPE`. Positions are plain `go/token.Pos` in a shared `go/token.FileSet`.

# Involved in
[PLAN-001](/plans/001-template-language.md).
