---
type: Task Slice
id: TASK-001-06
title: File parser: scan, mask, graft
description: Implement parser.ParseFile: template and markup-start detection with go/scanner, the same-length mask, a single go/parser call, and grafting markup into the AST.
spec: SPEC-001
plan: PLAN-001
state: todo
depends_on: [TASK-001-05]
touches: [internal/parser/parser.go, internal/parser/scan.go, internal/parser/mask.go, internal/parser/graft.go, internal/parser/testdata]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
Passes 1, 3, 4 and 5 of [PLAN-001](/plans/001-template-language.md) and the public `ParseFile`. Covers [SPEC-001](/specs/001-template-language.md) R1–R9, R19.

# Depends on
[TASK-001-05](/tasks/001-template-language/05-embedded-statements.md)

# Todo
- [ ] Tests first:
  - [ ] `template` detection: decl, generic, method receiver; `template.Must` in bodies and `var template` stay Go (R3)
  - [ ] markup start after `;` `{` `}` `:`; `<-ch` is Go (R6); markup inside nested Go blocks; several markup statements per body
  - [ ] `mask`: exact expected bytes for the SPEC-001 examples; length and newline invariant
  - [ ] graft: `TemplateDecl` wraps FuncDecl; `MarkupStmt` replaces placeholder; interpolations and headers carry go/ast nodes; Go statements in embedded bodies are errors
  - [ ] `return` / `return err` accepted, `return 1, 2` rejected (R8); `__golpe` identifiers rejected (R9)
  - [ ] golden `testdata/*.golpe` → `*.ast.golden` / `*.err.golden` with `-update`
- [ ] `scan.go`, `mask.go`, `graft.go`, `ParseFile`
- [ ] Rewrite the most common skeleton-specific go/parser messages

# Done when
All three SPEC-001 examples have AST goldens; error goldens cover every rule.
