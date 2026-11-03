---
type: Entity
id: ENT-golpe-token
title: Golpe token
description: internal/token.Token: go/token's token kinds extended with TEMPLATE and the markup tokens, plus the name classifiers the parser uses.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Shape
A `Token` (int) whose values below 1024 are exactly go/token's, and above it Golpe's: `TEMPLATE`, `TAG_OPEN`, `TAG_END`, `TAG_SELF_CLOSE`, `TAG_CLOSE`, `TEXT`, `INTERP_OPEN`, `INTERP_CLOSE`, `HTML_COMMENT`, `DOCTYPE`. `String()` follows go/token: delimiters print their spelling (`{`), classes their name (`TEXT`), unknown values `token(N)`. Positions are plain `go/token.Pos`.

# API (TASK-001-02)
`FromGo`, `Go`, `String`, `Lookup` (= go/token.Lookup, so "template" is IDENT), `IsTemplateKeyword(lit, next)` (the contextual check), `IsVoid`, `IsRawText` (ASCII-case-insensitive, no Unicode folding), `IsEmbeddedKeyword` (if/for/switch/select). Planned: element-name classification for component tags (TASK-005-01).

# Involved in
[PLAN-001](/plans/001-template-language.md), [PLAN-005](/plans/005-component-tags.md).
