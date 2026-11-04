---
type: Task Slice
id: TASK-001-05
title: Embedded statements
description: Implement the line rule in the markup parser: if/for/switch/select lines start embedded statements with Go headers, case/default clauses, } closes, else on the same line, one-line bodies, quoting hints.
spec: SPEC-001
plan: PLAN-001
state: done
depends_on: [TASK-001-04]
touches: [internal/parser/markup.go, internal/parser/markup_test.go, internal/parser/markup_stmt.go, internal/parser/markup_stmt_test.go]
generated: { by: claude-code/agent, at: 2026-11-04T20:27:08Z }
---
# Goal
[Embedded statements](/entities/embedded-statement.md) by the line rule, [SPEC-001](/specs/001-template-language.md) R18. Header text stays in the mask as Go so go/parser parses it (LAB-005).

# Depends on
[TASK-001-04](/tasks/001-template-language/04-markup-parser.md) ✅. Replaces the `embedded` hook and the `}`-line branch in `markup.go` (see PLAN-001 "As built: markup parsing"). Shares `markup.go` with TASK-005-02, so the two run one after the other.

# Todo
- [x] Tests first: `if`/`else if`/`else`; `for` 3-clause, condition-only, infinite, `range` (incl. `range 10`); expression switch, type switch, switch with init; `select` with send/receive cases; nested statements; one-line body `if ok { <b>yes</b> }`
- [x] Tests for the rule itself: a line whose first word is `if`/`for`/`switch`/`select` is a statement even in prose (error with the quoting hint `write {"if"} …`); `format`, `iffy` are text; `if` mid-line is text; `{"if "}you can` is text; `else` on its own line is an error
- [x] Header extent: the first `{` outside brackets that go/parser accepts as the body's (the plan's "first `{` at depth 0" was wrong, see Notes)
- [x] Body parsing until the matching `}`; `else` chaining on the same line; `case`/`default` clause splitting in switch/select bodies
- [x] Record header fragments for the mask

# Done when
The SPEC-001 `NavItem` and `List` examples parse into the expected trees.

# Notes
Implemented in [`internal/parser`](../../../../internal/parser/): `markup_stmt.go` (statements, header and clause extents) and `markup.go` (frames for bodies, the content loop). Coverage 100% of the parser. Besides the table tests, a throwaway fuzz harness (over 25 million runs) and a grammar-based generator (200,000 nested templates) checked that a clean region always masks to a skeleton go/parser accepts and that every statement grafts; neither is committed (fuzzing is TASK-001-07's).

**Header extent.** PLAN-001's "first `{` at paren/bracket depth 0" is not Go's rule: `for _, x := range []int{1, 2} {` and `if f := func() bool { return true }; f() {` are valid. Every `{` token outside brackets is a candidate, in order; the first for which `src[keyword:candidate] + "{}"`, wrapped in `package p;func _(){…}`, parses with go/parser as that one statement with its body at the candidate ends the header (`for x := range Map[K, V]{} {` then splits as Go splits it). The search stops at EOF, an unmatched closing bracket, more `;` outside brackets than the header can hold (if/switch 1, for 2, select 0), or `</` (never Go). Clause headers end at the first `:` token outside brackets (`case a[1:2] == b:`, `case "x:y":`, `case T{A: 1}.B:`), before any `;` outside them, and are checked with go/parser in a `switch{…}`/`select{…}` wrapper.

**Errors and recovery.** An invalid header reports, at the keyword, `invalid if header: a line starting with 'if' is an embedded statement; write {"if"} to start text with it` (`invalid case clause: a line starting with 'case' opens a clause; write {"case"} …` for clauses). If the keyword's line ends with a `{` token (`if x = 1 {`, `for x in xs {`, `if f(x {`), the statement is still parsed so its body and `}` cause no further errors, but none of its Go parts is kept, so go/parser never sees the bad header (keeping it cascaded into many errors). Otherwise the line is text. `else` starting a line reports `else must follow '}' on the same line; write {"else"} …` and, if only whitespace follows an if, becomes that if's else (not kept). `else` after the `}` of a for, switch, select or else block reports `else without if`. Content before the first clause reports `expected case or default` (once) and is dropped with its Go parts. Bodies report `expected '}', found </x>` / `'EOF'`.

**Decisions where SPEC-001 was silent** (R18 and R20 updated):
- After the last `}` of a statement, the rest of the line is content of the enclosing element (`if ok { yes } and more`); the mask writes `;` right after that `}` when the line goes on, so `if ok { };  x;` stays Go.
- `<pre>`: a statement line adds no whitespace. The indentation before a keyword, clause or line-start `}`, and the line end after a header's `{`, a clause's `:` or a closing `}` (when only blanks or a `//` comment follow), are not content; everything else is verbatim. One-line bodies keep their inner spaces. Outside `<pre>`, statements behave like elements for R22 (`Hello⏎if x {…}⏎!` renders `Hello ` … ` !`).
- `else` and its `{` or `if` must be on the `}` line (Go also allows `else⏎{`). A `case`/`default` line is a clause only directly in a switch/select body (inside `<p>` in a clause it is text); a statement keyword after `{` or `:` on the same line is text (`if a { if b {x} }` renders `if b ` and `{x}`).

**For TASK-001-06**: `region.keep` now also holds each valid header through its `{`, each `else` / `else {`, each clause through its `:`, and each closing `}`; `region.semis` also holds the byte after a statement's last `}` when the line goes on; new `region.stmts` lists the kept `*ast.If/For/Switch/Select/Case` in keyword order: graft each with the go/ast statement at its `Pos()` (IfStmt; ForStmt or RangeStmt; SwitchStmt or TypeSwitchStmt; SelectStmt; CaseClause or CommClause). Statements with invalid headers are in the tree, not in `stmts` (Stmt stays nil). `markupGraft` in `markup_stmt_test.go` is a working reference.

**For TASK-001-07**: the fuzzer showed that a Go comment kept in the mask may hold a NUL byte or invalid UTF-8, which go/parser rejects, as it would in a `.go` file; a skeleton-parses property must allow for it.
