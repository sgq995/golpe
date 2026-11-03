---
type: Plan
id: PLAN-001
title: Parsing the template language
description: Pre-scan with go/scanner, parse markup with a hand-written parser, mask it into a same-length Go skeleton, parse once with go/parser, then graft markup nodes that embed go/ast nodes.
status: draft
spec: SPEC-001
tags: [parser, token, ast]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Approach

Implements [SPEC-001](/specs/001-template-language.md) following
[DEC-001](/decisions/001-mask-and-reuse-go-parser.md): we never fork
`go/parser`. A Golpe file is parsed in five passes, all over the same bytes
and the same `token.FileSet`, so every position is a real `.golpe` position.

```text
 x.golpe bytes
   │ 1. scan      go/scanner finds top-level "template" and, inside template
   │              bodies, "<" + letter at statement starts      (internal/parser/scan.go)
   │ 2. markup    hand-written markup parser from each start → tree + extent
   │              + list of Go fragments (interp exprs, stmt headers) (markup.go)
   │ 3. mask      same-length copy: template→"func    ", first "<"→";",
   │              tags/text→spaces (newlines kept), {e}→" e;", headers kept (mask.go)
   │ 4. go/parser ParseFile(masked) → *go/ast.File with exact positions
   │ 5. graft     FuncDecl→TemplateDecl, placeholder EmptyStmt + the Go stmts
   │              inside the region → MarkupStmt; tree nodes get their ast.Expr /
   │              header stmts from the parsed skeleton                (graft.go)
   ▼
 *internal/ast.File  (go/ast.File whose Decls may hold *TemplateDecl)
```

# Architecture

| Package          | Owns                                                                   |
|------------------|------------------------------------------------------------------------|
| `internal/token` | [Golpe tokens](/entities/golpe-token.md): Go's tokens + Golpe's, `Lookup`, `IsVoid`, `IsRawText` |
| `internal/ast`   | `TemplateDecl`, `MarkupStmt`, markup node types, `Walk`, `Inspect`     |
| `internal/parser`| `ParseFile`, scanning, markup parsing, masking, grafting, error list   |

Actuator: [Parser](/actuators/parser.md). Entities produced:
[Template declaration](/entities/template-declaration.md),
[Markup node](/entities/markup-node.md),
[Embedded statement](/entities/embedded-statement.md),
[Interpolation](/entities/interpolation.md),
[Attribute](/entities/attribute.md).

# Snippets

## internal/token: extending go/token

```go
package token

import gotoken "go/token"

// Token is a go/token.Token or a Golpe token. Values below golpeBase are
// exactly go/token's, so conversion both ways is a no-op.
type Token int

const golpeBase Token = 1 << 10 // well above go/token's last value

const (
	TEMPLATE Token = golpeBase + iota // contextual keyword
	TAG_OPEN                          // <name
	TAG_END                           // >
	TAG_SELF_CLOSE                    // />
	TAG_CLOSE                         // </name>
	TEXT
	INTERP_OPEN  // {
	INTERP_CLOSE // }
	HTML_COMMENT // <!-- … -->
	DOCTYPE      // <!DOCTYPE …>
)

func FromGo(t gotoken.Token) Token { return Token(t) }

func (t Token) Go() (gotoken.Token, bool) {
	if t < golpeBase {
		return gotoken.Token(t), true
	}
	return gotoken.ILLEGAL, false
}

func (t Token) String() string {
	if g, ok := t.Go(); ok {
		return g.String()
	}
	return golpeNames[t-golpeBase]
}

func IsVoid(tag string) bool    // br, img, input, …  (SPEC-001 R10)
func IsRawText(tag string) bool // script, style        (R16)
```

## internal/ast: new nodes through embedding

```go
package ast

import goast "go/ast"

// TemplateDecl embeds the FuncDecl go/parser produced from the masked
// "func    Name(...)": it satisfies goast.Decl via the promoted declNode().
type TemplateDecl struct {
	*goast.FuncDecl           // Type.Results is always nil
	Template token.Pos        // position of the "template" keyword
}

// MarkupStmt replaces the placeholder EmptyStmt (";") the mask left where
// the root tag started; it satisfies goast.Stmt via promoted stmtNode().
type MarkupStmt struct {
	*goast.EmptyStmt
	Root Node   // *Element or *Doctype
	EndPos token.Pos
}

func (s *MarkupStmt) End() token.Pos { return s.EndPos }

// Node is any markup node.
type Node interface {
	goast.Node
	markupNode()
}

type Element struct {
	Lt          token.Pos // <
	Name        string
	NamePos     token.Pos
	Attrs       []*Attr
	SelfClosing bool
	Children    []Node
	Close       token.Pos // < of </name>, or NoPos for void/self-closing
	EndPos      token.Pos
}

type Attr struct {
	Name    string
	NamePos token.Pos
	Value   AttrValue // nil (boolean) | *StaticValue | *Interp
}

type Text struct{ ValuePos token.Pos; Raw, Value string } // Value after R18
type Interp struct{ Lbrace token.Pos; X goast.Expr; Rbrace token.Pos }
type Comment struct{ Start token.Pos; Text string }       // <!-- -->
type Doctype struct{ Start token.Pos; Text string }

// Embedded statements reuse the go/ast statement go/parser built from the
// skeleton, for its header; bodies are markup.
type If struct {
	Stmt *goast.IfStmt // Init, Cond (Body/Else ignored)
	Body []Node
	Else Node          // nil | *If | *Block
}
type For struct {
	Stmt goast.Stmt // *goast.ForStmt or *goast.RangeStmt (header only)
	Body []Node
}
type Switch struct {
	Stmt  goast.Stmt // *goast.SwitchStmt or *goast.TypeSwitchStmt
	Cases []*Case
}
type Select struct {
	Stmt  *goast.SelectStmt
	Cases []*Case
}
type Case struct {
	Clause goast.Stmt // *goast.CaseClause or *goast.CommClause (header only)
	Body   []Node
}
type Block struct{ Lbrace token.Pos; Body []Node; Rbrace token.Pos }

// Walk/Inspect understand goast nodes AND these; goast.Walk panics on
// them (LAB-003). Implemented as a type switch on our nodes that falls back
// to goast.Inspect for subtrees that are pure go/ast.
func Inspect(n goast.Node, f func(goast.Node) bool)
```

## internal/parser: entry point

```go
package parser

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

## Pass 1: finding templates and markup starts

```go
// At top level (depth 0), IDENT "template" followed by IDENT or "(" is a
// template declaration (DEC-003). Inside its body, track the previous token:
// a statement starts after ";" (incl. auto-inserted), "{", "}", ":" (labels,
// case clauses) or at body start. There, LSS immediately followed by an
// ASCII letter or "!" starts markup; ARROW ("<-") is Go (LAB-004).
```

go/scanner cannot continue past markup (an apostrophe in text would be
scanned as a rune literal), so after each markup region the scanner is
re-initialised on `src[end:]` with a fresh `token.File` of the right base.

## Pass 2: markup parsing

A byte-level recursive-descent parser:

- `<name` attrs (`>` | `/>`), children until `</name>`; void elements end at `>`.
- Children: `<` → element/close/comment; `{` → interpolation; a line whose
  first token is `if`, `for`, `switch`, `select` → embedded statement;
  inside an embedded body `}` closes it (then optional `else`), and in
  switch/select bodies `case`/`default` at line start open a clause;
  `//` or `/*` after whitespace at a node boundary → Go comment; otherwise text.
- Interpolation extent: run go/scanner from the `{` counting braces, so
  strings, runes, comments and composite/func literals are handled exactly.
- Embedded header extent: go/scanner from the keyword to the first `{` at
  paren/bracket depth 0, which is Go's own rule (composite literals in
  headers need parentheses in Go too).
- Raw text elements: find `</script` case-insensitively; no scanning inside.

Each region records the tree plus a list of **fragments** (byte ranges that
stay as Go in the mask) and **separators** (where ` ` / `;` go).

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

Invariant (fuzzed): `len(mask(src)) == len(src)` and every `'\n'` stays at
the same offset, so line:col never moves.

## Pass 5: grafting

Walk every `BlockStmt` of each template body. Statements whose position
lies inside a region's `[start, end)` are removed; the explicit EmptyStmt at
`start` is replaced by `&ast.MarkupStmt{EmptyStmt: e, Root: …}`. Interpolation
nodes take the `ExprStmt.X` found at their offset; embedded statement nodes
take the `IfStmt`/`ForStmt`/`RangeStmt`/`SwitchStmt`/`TypeSwitchStmt`/
`SelectStmt` found at their offset (their bodies, which only contain more
skeleton, are discarded). Any *other* Go statement found inside an embedded
body is an error (R15: "Go statements are not allowed inside markup").
Then each template's `*goast.FuncDecl` is wrapped in `*ast.TemplateDecl`.

# Testing

- **Unit tests** per pass, table-driven: scanner (template detection,
  `<-`, labels), markup parser (every R10–R18 rule), mask (exact bytes),
  graft.
- **Golden tests**: `internal/parser/testdata/*.golpe` → a stable textual dump
  of the AST (`*.ast.golden`), updated with `go test ./internal/parser -update`.
  Includes the three examples of SPEC-001 and every error case
  (`*.err.golden` holds `file:line:col: msg` lines).
- **Superset test (R1)**: parse every `.go` file under `$(go env GOROOT)/src`
  with both `go/parser` and `internal/parser`; the ASTs must print
  identically. Skipped with `-short`.
- **Fuzz**: `FuzzParseFile` (never panics; mask invariant holds; positions in
  range), seeded from testdata.

# Laboratories

Executable in `internal/labs/parserlab` (run with `go test ./internal/labs/...`):

| Lab     | Question                                                            | Result |
|---------|---------------------------------------------------------------------|--------|
| LAB-001 | Does a `;` placeholder survive go/parser as an explicit EmptyStmt at the exact offset? | ✅ yes, `Implicit=false`, line:col exact |
| LAB-002 | Do embedded go/ast nodes satisfy the sealed `ast.Decl`/`ast.Stmt`?   | ✅ yes, promoted unexported methods |
| LAB-003 | Does `ast.Walk` tolerate foreign nodes?                            | ❌ panics → own `Walk`/`Inspect` needed |
| LAB-004 | Can go/scanner tell `<-ch` from `<li`?                              | ✅ `ARROW` vs `LSS IDENT` |
| LAB-005 | Does one go/parser call over the skeleton mask yield every interpolation and header with original offsets? | ✅ yes, including `if/else` and comments |
| LAB-006 | Do `/*line*/` directives move reported positions to `.golpe`?       | ✅ yes (used by [PLAN-002](/plans/002-code-generation.md)) |
| LAB-007 | Do they survive gofmt?                                               | ⚠️ gofmt adds a space after the comment (column +1); emit `/*line f:L:C-1*/ ` instead, or `//line f:L:1` for column-1 code |

# Risks

- **Ambiguous text.** A text line starting with `if `/`for ` is parsed as an
  embedded statement. Mitigation: only when the keyword is the first token
  of the line *and* the header scans as Go up to a `{`; otherwise text. The
  spec documents `{"if"}` as the escape.
- **Error quality.** go/parser messages about the skeleton may mention
  `;` or `func`. Mitigation: a small rewrite table in the error list,
  covered by `*.err.golden` tests.
- **Go grammar changes** are inherited for free, but a new statement form
  starting with `<` would break R6 (none exists or is proposed).

# Slices

1. [TASK-001-01 Project scaffold](/tasks/001-template-language/01-project-scaffold.md) ✅
2. [TASK-001-02 Token extension](/tasks/001-template-language/02-token-extension.md)
3. [TASK-001-03 AST nodes and walker](/tasks/001-template-language/03-ast-nodes.md)
4. [TASK-001-04 Markup parser](/tasks/001-template-language/04-markup-parser.md)
5. [TASK-001-05 Embedded statements](/tasks/001-template-language/05-embedded-statements.md)
6. [TASK-001-06 File parser: scan, mask, graft](/tasks/001-template-language/06-file-parser.md)
7. [TASK-001-07 Conformance: superset, golden, fuzz](/tasks/001-template-language/07-conformance.md)

02 and 03 can run in parallel; so can TASK-002-01 (runtime) with all of them.
