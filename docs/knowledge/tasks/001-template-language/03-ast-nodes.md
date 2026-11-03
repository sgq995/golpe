---
type: Task Slice
id: TASK-001-03
title: AST nodes and walker
description: Implement internal/ast: TemplateDecl and MarkupStmt embedding go/ast nodes, the markup node types, and Walk/Inspect over both.
spec: SPEC-001
plan: PLAN-001
state: done
depends_on: [TASK-001-01]
touches: [internal/ast]
generated: { by: claude-code/agent, at: 2026-11-03T02:00:38Z }
---
# Goal
The node types from [PLAN-001](/plans/001-template-language.md) ("internal/ast: new nodes through embedding"), modelling [template declarations](/entities/template-declaration.md), [markup nodes](/entities/markup-node.md), [embedded statements](/entities/embedded-statement.md), [interpolations](/entities/interpolation.md) and [attributes](/entities/attribute.md).

# Depends on
[TASK-001-01](/tasks/001-template-language/01-project-scaffold.md). Can run in parallel with TASK-001-02.

# Todo
- [x] Tests first: compile-time assertions that `*TemplateDecl` is a `go/ast.Decl` and `*MarkupStmt` a `go/ast.Stmt` (LAB-002)
- [x] Tests: `Pos`/`End` of every node; `Inspect` visits nodes in source order across Go and markup subtrees and never panics (LAB-003)
- [x] Node types: `TemplateDecl`, `MarkupStmt`, `Element`, `Attr`, `StaticValue`, `Text`, `Interp`, `Comment`, `Doctype`, `If`, `For`, `Switch`, `Select`, `Case`, `Block`
- [x] `Walk(Visitor, Node)` and `Inspect`, delegating pure go/ast subtrees to go/ast
- [x] `Fprint`-style debug dump used by golden tests (stable, positions as line:col)

# Done when
`go test ./internal/ast` passes; dump format documented in package doc.

# Notes
Implemented in [`internal/ast`](../../../../internal/ast/) (`ast.go`, `walk.go`, `print.go`). The tests build a realistic fixture the way PLAN-001 describes: markup trees first, then a test-only mask, one `go/parser` call and a test-only graft (`fixture_test.go`), which TASK-001-06 can use as a reference. Coverage 99.7%.

Deviations from the PLAN-001 snippet, for the coordinator to fold into the plan:

- **Walk.** Not "a type switch on our nodes that falls back to `goast.Inspect` for pure go/ast subtrees": that is unsafe, because go/ast containers hold our nodes (`File.Decls` holds `*TemplateDecl`; any statement list of a template body, including function literal bodies inside interpolations, holds `*MarkupStmt`). Instead `Walk(v, n)` is `goast.Walk(walker{v}, n)`: `go/ast.Walk` calls `Visit` before its type switch and stops when `Visit` returns nil, so the adapter handles Golpe nodes itself (visit, children, `Visit(nil)`) and returns nil, and passes go/ast nodes through wrapped again. No go/ast walking code is copied; go/ast's contract (pruning, `Visit(nil)`) is kept. `TestGoASTWalkVisitsBeforeTypeSwitch` pins the go/ast behaviour this relies on; it deserves a LAB entry in the plan. `Visitor` is an alias of `go/ast.Visitor`.
- **Embedded statements walk headers only**: Init/Cond, Init/Cond/Post, Key/Value/X, Init/Tag, Init/Assign, case List, Comm; never the go/ast statement's skeleton Body/Else. An unexpected statement type (a `BadStmt`) is visited whole.
- **Positions recorded on Golpe nodes**, so Pos/End, Walk and the dump work on trees built before grafting (TASK-001-04/05 tests) and on error-recovery trees: `If`, `For`, `Switch`, `Select`, `Case` have `Keyword`; `Switch`/`Select` have `Lbrace`/`Rbrace`; `Case` has `Colon` and `Default bool`. End falls back to the last known part when closing positions are missing.
- **`If.Body` and `For.Body` are `*Block`**, and `Block` has `List` (not `Body`), mirroring `go/ast.IfStmt.Body`/`BlockStmt.List`; the brace positions live in the Block. The PLAN-002 generator snippet becomes `g.nodes(n.Body.List)`.
- **`MarkupStmt` has no `EndPos`**: `End()` is `Root.End()`. `Element` has no `NamePos` (always `Lt+1`). `TemplateDecl` overrides `Pos` (the `Template` field) and `End` (nil-safe).
- **`StaticValue{ValuePos, Raw, Value}`** (Raw includes the quotes); `Comment.Text` and `Doctype.Text` include their delimiters, so End is `Start+len(Text)`. `Node` covers every markup node (Attr, StaticValue, Case and Block included); `AttrValue` restricts `Attr.Value` to `*StaticValue`/`*Interp`.
- **Dump**: `Fprint(w, fset, node) error`, documented in the package doc. Pure go/ast nodes print as gofmt source; `File.Comments` are listed after the declarations, so a doc or line comment that go/printer prints with a pure node also appears there.

Questions for the maintainer: rename `Comment` to `HTMLComment` (it shares its name, and its dump kind, with `go/ast.Comment`)? Should the mask keep Go comments inside markup so they reach `File.Comments` (R17 goldens could then show them)?
