---
type: Decision
id: DEC-001
title: Mask markup and reuse go/parser instead of forking it
description: Parse .golpe files by masking markup into a same-length Go skeleton, parsing it with the standard go/parser, and grafting markup nodes that embed go/ast nodes back in.
status: stable
decision: accepted
tags: [parser, architecture]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
verified: { by: human:sgq995, at: 2026-11-03T03:12:42Z }
---
# Context

The idea is to build Golpe on `go/parser` and `go/token`, extending them with
our own packages and using struct embedding to satisfy their interfaces. In
practice:

- `go/parser` cannot be extended: its parser type and productions are
  unexported, and it has no hooks.
- `go/ast`'s `Decl`, `Stmt` and `Expr` interfaces are sealed by unexported
  marker methods, **but** embedding a `go/ast` node promotes those methods,
  so our own nodes do satisfy them (LAB-002).
- `ast.Walk`/`ast.Inspect` panic on unknown node types (LAB-003), and
  `go/printer` cannot print them.
- `go/scanner` is reusable as-is and tells `<-ch` (`ARROW`) apart from
  `<li` (`LSS IDENT`) (LAB-004).

# Options

1. **Fork go/parser** into `internal/parser` and add productions.
   Full control, but ~3k lines to maintain and re-sync with every Go release
   (new syntax, generics changes).
2. **Write our own Go parser.** Far too much work; always behind Go.
3. **Mask and reuse.** Pre-scan with `go/scanner`, find template
   declarations and markup regions, and build a *same-length* masked copy of
   the source: `template` becomes `func    `; the first byte of a markup
   statement becomes `;`; tags and text become spaces (newlines kept);
   embedded statement headers stay verbatim; `{expr}` becomes ` expr;`.
   One `go/parser.ParseFile` call parses every Go fragment with exact
   positions (LAB-001, LAB-005). A markup parser then builds the HTML tree
   and grafts it into the AST as `internal/ast` nodes that embed
   `go/ast` nodes.

# Decision

Option 3. It keeps the user's idea (extend `go/token` / `go/ast` through
embedding, reuse the standard parser), keeps us on the official Go grammar
for free, and gives byte-exact positions for diagnostics and the LSP.

# Consequences

Accepted by the maintainer in the design interview of 2026-11-03, after
TASK-001-02 and TASK-001-03 confirmed the approach in code.


- `internal/ast` needs its own `Walk`/`Inspect`; anything handed to
  `go/printer` must be lowered to pure `go/ast` first. Implemented as an
  adapter around `go/ast.Walk`, which calls `Visit` before its type switch
  (LAB-011 in [PLAN-001](/plans/001-template-language.md)).
- Masking must preserve length and newlines exactly; this is the invariant
  the fuzz test guards (see [PLAN-001](/plans/001-template-language.md)).
- Go syntax errors inside markup (for example in an `if` header) come from
  go/parser with correct positions, but the message talks about Go, not
  markup; the parser may rewrite a few common ones.
- If a future need cannot be expressed as a mask, we revisit option 1.
