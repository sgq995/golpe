---
type: Actuator
id: ACT-parser
title: Parser
description: internal/parser, internal/token and internal/ast: turn a .golpe file into a Go AST extended with template and markup nodes.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Responsibilities
- Find [template declarations](/entities/template-declaration.md) and [markup](/entities/markup-node.md) regions using [Golpe tokens](/entities/golpe-token.md).
- Mask markup out of the source and parse the remaining Go with the standard go/parser ([DEC-001](/decisions/001-mask-and-reuse-go-parser.md)).
- Parse markup: elements, [component tags](/entities/component-tag.md), fragments, text, [attributes](/entities/attribute.md), [interpolations](/entities/interpolation.md) and [embedded statements](/entities/embedded-statement.md) (line rule).
- Report syntax errors at `.golpe` positions.

# Code
`internal/parser`, `internal/token`, `internal/ast`.

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-005](/specs/005-component-tags.md), [PLAN-001](/plans/001-template-language.md).
