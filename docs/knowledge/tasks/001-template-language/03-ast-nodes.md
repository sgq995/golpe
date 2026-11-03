---
type: Task Slice
id: TASK-001-03
title: AST nodes and walker
description: Implement internal/ast: TemplateDecl and MarkupStmt embedding go/ast nodes, the markup node types, and Walk/Inspect over both.
spec: SPEC-001
plan: PLAN-001
state: todo
depends_on: [TASK-001-01]
touches: [internal/ast]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
The node types from [PLAN-001](/plans/001-template-language.md) ("internal/ast: new nodes through embedding"), modelling [template declarations](/entities/template-declaration.md), [markup nodes](/entities/markup-node.md), [embedded statements](/entities/embedded-statement.md), [interpolations](/entities/interpolation.md) and [attributes](/entities/attribute.md).

# Depends on
[TASK-001-01](/tasks/001-template-language/01-project-scaffold.md). Can run in parallel with TASK-001-02.

# Todo
- [ ] Tests first: compile-time assertions that `*TemplateDecl` is a `go/ast.Decl` and `*MarkupStmt` a `go/ast.Stmt` (LAB-002)
- [ ] Tests: `Pos`/`End` of every node; `Inspect` visits nodes in source order across Go and markup subtrees and never panics (LAB-003)
- [ ] Node types: `TemplateDecl`, `MarkupStmt`, `Element`, `Attr`, `StaticValue`, `Text`, `Interp`, `Comment`, `Doctype`, `If`, `For`, `Switch`, `Select`, `Case`, `Block`
- [ ] `Walk(Visitor, Node)` and `Inspect`, delegating pure go/ast subtrees to go/ast
- [ ] `Fprint`-style debug dump used by golden tests (stable, positions as line:col)

# Done when
`go test ./internal/ast` passes; dump format documented in package doc.
