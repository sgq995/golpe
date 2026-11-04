---
type: Task Slice
id: TASK-005-02
title: Parse component tags
description: Classify tag names in the markup parser and parse non-element tags into ast.Component: identifier or pkg.Name, attributes, self-closing or empty only.
spec: SPEC-005
plan: PLAN-005
state: todo
depends_on: [TASK-001-04, TASK-005-01]
touches: [internal/ast, internal/parser/markup.go, internal/parser/markup_component_test.go]
generated: { by: claude-code/agent, at: 2026-11-04T10:36:16Z }
---
# Goal
[SPEC-005](/specs/005-component-tags.md) R1, R4 (syntax), R5 in the parser.

# Depends on
[TASK-001-04](/tasks/001-template-language/04-markup-parser.md) ✅, [TASK-005-01](/tasks/005-component-tags/01-element-names.md) ✅. Shares `internal/parser/markup.go` with TASK-001-05: run after it, not in parallel.

# Todo
- [ ] AST: `Component{Lt, Name goast.Expr, Attrs, SelfClosing, Close, EndPos}` with Walk/Fprint support
- [ ] Tests first: `<Card/>`, `<card></card>`, `<ui.Button label="x"/>`, `<Card title={t} featured/>`; `<div>`/`<DIV>`/`<my-widget>` stay elements; `<svg:path>` and `<a.b.c>` are errors; content in a component tag is an error suggesting a props field; non-identifier attribute names (`aria-label`) and duplicates are errors
- [ ] Mask: `{expr}` attribute values become skeleton expressions like other interpolations

# Done when
The SPEC-005 example parses into the expected tree.
