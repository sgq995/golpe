---
type: Plan
id: PLAN-001
title: Parsing the template language
description: Pre-scan with go/scanner, parse markup with a hand-written parser, mask it into a same-length Go skeleton, parse once with go/parser, then graft markup nodes that embed go/ast nodes.
status: draft
spec: SPEC-001
tags: [parser, token, ast]
generated: { by: claude-code/agent, at: 2026-11-04T20:27:08Z }
---
# Approach

Implements [SPEC-001](/specs/001-template-language.md) following
[DEC-001](/decisions/001-mask-and-reuse-go-parser.md) (accepted): we never
fork `go/parser`. A Golpe file is parsed in five passes over the same bytes
and the same `token.FileSet`, so every position is a real `.golpe` position.

```text
 x.golpe bytes
   │ 1. scan      go/scanner finds top-level "template" (token.IsTemplateKeyword)
   │              and, in template bodies, "<"+letter or "<>" at statement starts
   │ 2. markup    hand-written markup parser from each start → tree + extent
   │              + Go fragments (interp exprs, statement headers)      (markup.go)
   │ 3. mask      same-length copy: template→"func    ", first "<"→";",
   │              tags/text→spaces (newlines kept), {e}→" e;", headers kept (mask.go)
   │ 4. go/parser ParseFile(masked) → *go/ast.File with exact positions
   │ 5. graft     FuncDecl→TemplateDecl, placeholder EmptyStmt + the Go stmts
   │              inside the region → MarkupStmt; nodes get their go/ast parts (graft.go)
   ▼
 *go/ast.File whose Decls may hold *ast.TemplateDecl
```

`element` and `writer()` need nothing special here: they are ordinary
identifiers to the parser; the generator handles them
([PLAN-002](/plans/002-code-generation.md)).

# Architecture

| Package | Owns | State |
|---------|------|-------|
| `internal/token` | [Golpe tokens](/entities/golpe-token.md); `FromGo`, `Go`, `String`, `Lookup`, `IsTemplateKeyword`, `IsVoid`, `IsRawText`, `IsEmbeddedKeyword`; element-name lists (TASK-005-01) | done (TASK-001-02) |
| `internal/ast` | `TemplateDecl`, `MarkupStmt`, markup nodes, `Walk`, `Inspect`, `Fprint` | done (TASK-001-03) |
| `internal/parser` | markup parsing and the error list (TASK-001-04); embedded statements (TASK-001-05); scanning, masking, grafting, `ParseFile` to come | in progress |

Actuator: [Parser](/actuators/parser.md).

# As built: internal/token

See the [entity](/entities/golpe-token.md). Notable refinements over the
first sketch: `Lookup` is exactly `go/token.Lookup` and the contextual check
is `IsTemplateKeyword(lit string, next Token) bool` (DEC-003); names print as
go/token does (delimiters by spelling, classes by name, unknown `token(N)`);
`IsVoid`/`IsRawText` fold ASCII case only (no Unicode folding, so the Kelvin
sign does not make `linK` void); a test fails if go/token ever grows past
the Golpe base (1024).

# As built: internal/ast

```go
type TemplateDecl struct {
	*goast.FuncDecl           // from the masked "func    Name(...)"; Results nil
	Template token.Pos        // the "template" keyword; Pos() returns it
}

type MarkupStmt struct {
	*goast.EmptyStmt          // the ";" placeholder; Pos() is the markup start
	Root Node                 // *Element, *Fragment, *Doctype or *HTMLComment; End() is Root.End()
}

type Node interface{ goast.Node; markupNode() }   // every markup node
type AttrValue interface{ Node; attrValueNode() } // *StaticValue | *Interp

type Element struct {
	Lt          token.Pos; Name string; Attrs []*Attr
	SelfClosing bool; Children []Node; Close token.Pos; EndPos token.Pos
}
type Attr        struct{ Name string; NamePos token.Pos; Value AttrValue } // nil = boolean
type StaticValue struct{ ValuePos token.Pos; Raw, Value string }           // Raw has quotes
type Text        struct{ ValuePos token.Pos; Raw, Value string }
type Interp      struct{ Lbrace token.Pos; X goast.Expr; Rbrace token.Pos }
type Fragment    struct{ Lt token.Pos; Children []Node; Close, EndPos token.Pos } // <>…</>
type HTMLComment struct{ Start token.Pos; Text string } // <!-- --> incl. delimiters
type Doctype     struct{ Start token.Pos; Text string }

// Embedded statements hold the go/ast statement go/parser built from the
// skeleton, but only its HEADER is meaningful; positions live on our nodes.
type If     struct{ Keyword token.Pos; Stmt *goast.IfStmt; Body *Block; Else Node } // nil | *If | *Block
type For    struct{ Keyword token.Pos; Stmt goast.Stmt; Body *Block }              // ForStmt | RangeStmt
type Switch struct{ Keyword token.Pos; Stmt goast.Stmt; Lbrace token.Pos; Cases []*Case; Rbrace token.Pos }
type Select struct{ Keyword token.Pos; Stmt *goast.SelectStmt; Lbrace token.Pos; Cases []*Case; Rbrace token.Pos }
type Case   struct{ Keyword token.Pos; Clause goast.Stmt; Default bool; Colon token.Pos; Body []Node }
type Block  struct{ Lbrace token.Pos; List []Node; Rbrace token.Pos }
```

`Walk(v, n)` is `go/ast.Walk(walker{v}, n)`: go/ast.Walk calls `Visit` before
its type switch and stops when `Visit` returns nil, so the adapter handles our
nodes itself (visit, children, `Visit(nil)`) and returns nil, and wraps go/ast
nodes again (LAB-011). Embedded statements walk header fields only. `Fprint`
prints one line per node, `Kind start-end detail`, positions as line:col.

Still to come: `Component` (TASK-005-02).

# As built: markup parsing (TASK-001-04)

- `lineFile(filename, base, src) *token.File` is a stand-in for the file
  go/parser will create; with `base` read from the shared set just before
  `ParseFile`, node positions are final (LAB-014) and errors get line:col.
- `parseMarkup(file, src, start, errs) *region` parses one markup statement.
  `region{start, end int; root ast.Node; keep []span; semis []int; interps []*ast.Interp}`:
  the mask blanks `src[start:end]` except newlines and `keep` (interpolation
  expressions, Go comments), then writes `;` at `start` and at each `semis`
  offset (each interpolation's `}`).
- `errorList` sorts and keeps the first error per line, as go/parser does.
- Recovery: a close tag for an enclosing element reports once; a stray close
  tag is reported and skipped; a content line starting with `}` reports
  `expected </x>` and ends the region, so the Go block's `}` stays Go.
- Choices recorded in SPEC-001: duplicate attributes ignore ASCII case,
  whitespace around `=`, text-only `<textarea>`/`<title>`, `<pre>` as ordinary
  content. Whitespace dropped by R22 produces no Text node. `<script/>` is
  accepted, so the generator writes `<script></script>` (SPEC-002 R8).
- TASK-001-05 replaced the `embedded` hook and the `}`-line branch (see
  below): a content line starting with `}` now closes the innermost open
  body, and ends the region only when no body is open.

# As built: embedded statements (TASK-001-05)

In `markup_stmt.go`; details and decisions in the
[slice](/tasks/001-template-language/05-embedded-statements.md).

- **Frames.** A body is a frame too (`body`; `clauses` for switch/select,
  where a `case`/`default` line ends the content). A `}` closes a body
  anywhere on a line; a close tag of an enclosing element ends a body with
  `expected '}', found </x>`; a `}` line inside an element nested in a body
  reports `expected </x>, found '}'` and closes that body (unwinding).
- **Header extent.** Every `{` token outside brackets is a candidate, in
  order; the first for which `src[keyword:candidate] + "{}"` parses with
  go/parser (in `package p;func _(){…}`) as the statement with its body there
  ends the header. Bounded by EOF, an unmatched closing bracket, too many `;`
  outside brackets (if/switch 1, for 2, select 0) and `</`. Clause headers end
  at the first `:` token outside brackets and are checked in a
  `switch{…}`/`select{…}` wrapper.
- **Extra keep spans.** Each valid header through its `{`; `else {` or
  `else` (before an `if` header); each clause through its `:`; each closing
  `}`. **Extra semis:** the byte after a statement's last `}` when the line
  goes on (`if ok { } x;` is not Go). **New `region.stmts`:** the kept
  `If`/`For`/`Switch`/`Select`/`Case` nodes, in keyword order, for the graft.
- **Invalid headers** keep nothing (go/parser cascades badly on them): the
  error carries the quoting hint; if the line ends with `{` the statement is
  parsed for recovery, else the line is text. Content before the first
  clause is reported and dropped with its Go parts.
- **Whitespace.** Bodies are normalised like element content (R22); in
  `<pre>` a statement line adds no whitespace (SPEC-001 R20).

# internal/parser

## Entry point

```go
// ParseFile parses a .golpe file. It returns as much of the AST as it could
// build and a scanner.ErrorList sorted by position.
func ParseFile(fset *token.FileSet, filename string, src []byte) (*goast.File, error) {
	starts, errs := scanTemplates(src)          // pass 1
	regions := parseMarkup(src, starts, &errs)  // pass 2
	masked := mask(src, regions)                // pass 3 (len(masked)==len(src))
	f, err := goparser.ParseFile(fset, filename, masked, goparser.ParseComments|goparser.SkipObjectResolution)
	errs.addGo(err)                             // pass 4
	graft(f, fset, regions, &errs)              // pass 5
	return f, errs.Err()
}
```

## Pass 1: templates and markup starts

At top level, IDENT `template` with `IsTemplateKeyword(lit, next)` starts a
template declaration. Inside its body, a statement starts after `;` (incl.
auto-inserted), `{`, `}`, `:` or at body start. There, `LSS` immediately
followed by an ASCII letter, `>` (fragment) or `!` (doctype/comment) starts
markup; `ARROW` (`<-`) is Go (LAB-004). Markup inside a function literal is an
error. go/scanner cannot continue past markup (an apostrophe in text would be
a rune literal), so after each region it is re-initialised on `src[end:]`.

## Pass 2: markup parsing

A byte-level recursive-descent parser:

- **Tags:** `<name` attrs (`>` | `/>`), children until `</name>` (matched
  ignoring ASCII case); void elements end at `>` and cannot have content;
  `<>`…`</>` is a fragment. Raw text elements (`script`, `style`) end at
  `</script` (case-insensitive) with no scanning inside; `textarea`/`title`
  allow text and interpolations only; `pre`/`textarea` keep whitespace.
  Component tags (non-element names) are classified with TASK-005-01's lists
  and parsed by TASK-005-02.
- **Attributes:** names as authored; `"…"`/`'…'` literal (no interpolation);
  `{expr}`; bare. Duplicates and unquoted values are errors.
- **Children, by line:** if the first Go token of a line is `if`, `for`,
  `switch` or `select`, it starts an embedded statement (the line rule); in a
  switch/select body `case`/`default` lines open clauses; a `}` line closes
  the body and may continue with `else`. Otherwise: `<` → tag, comment or
  close; `{` → interpolation; `//` or `/*` at a node boundary → Go comment
  (kept in the mask so go/parser records it, dropped from output); anything
  else → text.
- **Interpolation extent:** go/scanner from the `{` counting braces, so
  strings, runes, comments and composite/func literals are exact.
- **Header extent:** not simply the first `{` at paren/bracket depth 0:
  `for _, x := range []int{1, 2} {` and
  `if f := func() bool { return true }; f() {` are valid Go. Each `{` outside
  brackets is a candidate, and the first one go/parser accepts as the body's
  ends the header (see "As built: embedded statements").
- **Whitespace (R22):** text runs collapse to one space; runs at the start or
  end of an element's, fragment's or body's content are dropped.
- **Errors:** line-rule failures carry the quoting hint (`{"for"}`).

Each region records the tree plus the **fragments** (byte ranges that stay Go
in the mask) and **separators** (where ` `/`;` go).

## Pass 3: the mask

```go
func mask(src []byte, regions []region) []byte {
	out := bytes.Clone(src)
	for _, r := range regions {
		for i := r.start; i < r.end; i++ {
			if out[i] != '\n' && !r.isFragment(i) {
				out[i] = ' '
			}
		}
		out[r.start] = ';'               // placeholder EmptyStmt (LAB-001)
		for _, in := range r.interps {   // {expr} → " expr;" (LAB-005)
			out[in.lbrace], out[in.rbrace] = ' ', ';'
		}
	}
	for _, t := range templateKeywords {
		copy(out[t:], "func    ")        // same length as "template"
	}
	return out
}
```

Invariant (fuzzed): `len(mask(src)) == len(src)` and every `'\n'` keeps its
offset, so line:col never moves. The test-only mask/graft in
`internal/ast/fixture_test.go` is a working reference.

## Pass 5: grafting

Walk every `BlockStmt` of each template body. Statements inside a region's
`[start, end)` are removed; the explicit EmptyStmt at `start` becomes
`&ast.MarkupStmt{EmptyStmt: e, Root: …}`. Interpolations take the
`ExprStmt.X` at their offset; embedded statements (`region.stmts`) take the
statement at their keyword (header only). Any other Go statement inside an embedded body is
an error ("Go statements are not allowed inside markup"). Each template's
`*goast.FuncDecl` is wrapped in `*ast.TemplateDecl`.

# Testing

- **Unit tests** per pass, table-driven.
- **Golden tests:** `internal/parser/testdata/*.golpe` → `*.ast.golden`
  (`ast.Fprint`) and `*.err.golden` (`file:line:col: msg`), `-update` flag.
- **Superset test (R1):** every `.go` file under GOROOT/src parses identically
  with go/parser and internal/parser (skipped with `-short`).
- **Fuzz:** `FuzzParseFile` (no panics; mask invariant; positions in range).

# Laboratories

Executable in `internal/labs/parserlab` and `internal/ast`:

| Lab | Question | Result |
|-----|----------|--------|
| LAB-001 | Does a `;` placeholder survive go/parser as an explicit EmptyStmt at the exact offset? | ✅ yes |
| LAB-002 | Do embedded go/ast nodes satisfy the sealed `ast.Decl`/`ast.Stmt`? | ✅ yes |
| LAB-003 | Does `ast.Walk` tolerate foreign nodes? | ❌ panics |
| LAB-004 | Can go/scanner tell `<-ch` from `<li`? | ✅ `ARROW` vs `LSS IDENT` |
| LAB-005 | Does one go/parser call over the skeleton mask yield every interpolation and header at original offsets? | ✅ yes |
| LAB-006 | Do `/*line*/` directives move reported positions to `.golpe`? | ✅ yes |
| LAB-007 | Do they survive gofmt? | ⚠️ gofmt adds a space; emit `/*line f:L:C-1*/ ` or `//line f:L:1` |
| LAB-011 | Does go/ast.Walk call `Visit` before its type switch, so an adapter can intercept foreign nodes? | ✅ yes (`TestGoASTWalkVisitsBeforeTypeSwitch` in internal/ast) |
| LAB-014 | Can markup positions be computed before go/parser creates the file? | ✅ yes: ParseFile adds the file at `fset.Base()`; pinned by internal/parser tests |

# Risks

- **Line rule surprises.** A text line starting with `for`/`if` is a statement
  by design; mitigated by the quoting hint in errors and golden tests.
- **Error quality.** go/parser messages about the skeleton may mention `;` or
  `func`; a small rewrite table, covered by `*.err.golden`.
- **Go grammar changes** come for free, unless a statement form starting with
  `<` appears (none exists or is proposed).

# Slices

1. [TASK-001-01 Project scaffold](/tasks/001-template-language/01-project-scaffold.md) ✅
2. [TASK-001-02 Token extension](/tasks/001-template-language/02-token-extension.md) ✅
3. [TASK-001-03 AST nodes and walker](/tasks/001-template-language/03-ast-nodes.md) ✅
4. [TASK-001-04 Markup parser](/tasks/001-template-language/04-markup-parser.md) ✅
5. [TASK-001-05 Embedded statements](/tasks/001-template-language/05-embedded-statements.md) ✅
6. [TASK-001-06 File parser: scan, mask, graft](/tasks/001-template-language/06-file-parser.md) — ready
7. [TASK-001-07 Conformance: superset, golden, fuzz](/tasks/001-template-language/07-conformance.md)
