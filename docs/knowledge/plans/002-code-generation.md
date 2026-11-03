---
type: Plan
id: PLAN-002
title: Generating Go from the extended AST
description: Splice the original source, replacing template headers and markup statements with lowered Go, add line directives, then gofmt; the runtime package provides escaping writers.
status: draft
spec: SPEC-002
tags: [generator, runtime]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Approach

Implements [SPEC-002](/specs/002-code-generation.md). The generator is a
**splicer**, not a printer: user Go code is copied byte-for-byte from the
`.golpe` source (keeping its comments and layout) and only three kinds of
span are replaced:

1. a template header `template Name[TP](params) {` →
   `func Name[TP](params) __golpe.Element { return func(__golpeW __golpe.Writer) error {`
   and its closing `}` → `return nil } }`;
2. each `MarkupStmt` span → lowered write statements;
3. each `element` identifier in a type position → `__golpe.Element`.

Every copied span is preceded by a line directive (LAB-006/007 in
[PLAN-001](/plans/001-template-language.md)), then the whole file goes
through `go/format.Source`.

# Architecture

| Package / file                   | Owns                                                         |
|----------------------------------|--------------------------------------------------------------|
| root `golpe` (`write.go`, `escape.go`) | [Runtime](/actuators/runtime.md): `Writer`, `WriteString`, `Write`, `WriteAttr`, escapers |
| `internal/generator`             | [Generator](/actuators/generator.md): `Generate(fset, file, src) ([]byte, error)` |
| `internal/generator/emit.go`     | Buffered emitter: static-byte coalescing, directives          |

# Snippets

## Runtime (root package)

```go
package golpe

type Writer = io.Writer // lets generated code avoid importing "io"

func WriteString(w Writer, s string) error { _, err := io.WriteString(w, s); return err }

// Write renders a content interpolation (DEC-004).
func Write(w Writer, v any) error {
	switch v := v.(type) {
	case nil:
		return nil
	case Element: // == func(io.Writer) error
		if v == nil {
			return nil
		}
		return v(w)
	case string:
		return WriteString(w, html.EscapeString(v))
	case fmt.Stringer:
		return WriteString(w, html.EscapeString(v.String()))
	case error:
		return WriteString(w, html.EscapeString(v.Error()))
	case []byte:
		return WriteString(w, html.EscapeString(string(v)))
	default: // numbers, bools, …
		return WriteString(w, html.EscapeString(fmt.Sprint(v)))
	}
}

// WriteAttr writes a dynamic attribute value, quotes included.
func WriteAttr(w Writer, v any) error // `"` + EscapeAttr(fmt) + `"`
```

## Lowering markup

```go
type emitter struct {
	buf    bytes.Buffer
	static strings.Builder // pending static HTML, flushed as one WriteString
}

func (e *emitter) static(s string) { e.static.WriteString(s) }

func (e *emitter) flush() {
	if e.static.Len() == 0 {
		return
	}
	fmt.Fprintf(&e.buf, "if err := __golpe.WriteString(__golpeW, %s); err != nil { return err }\n",
		strconv.Quote(e.static.String()))
	e.static.Reset()
}

func (g *gen) node(n ast.Node) {
	switch n := n.(type) {
	case *ast.Element:
		g.e.static("<" + n.Name)
		for _, a := range n.Attrs { g.attr(a) }
		g.e.static(">")
		for _, c := range n.Children { g.node(c) }
		if !n.SelfClosing && !token.IsVoid(n.Name) {
			g.e.static("</" + n.Name + ">")
		}
	case *ast.Text:
		g.e.static(html.EscapeString(n.Value)) // escaped at generation time (R7)
	case *ast.Interp:
		g.e.flush()
		g.call("Write", n.X) // __golpe.Write(__golpeW, /*line …*/ <src of X>)
	case *ast.If:
		g.e.flush()
		g.header(n.Stmt)     // "if " + /*line*/ src[Init..Cond] + " {"
		g.nodes(n.Body)
		g.e.flush()
		g.elseBranch(n.Else) // "} else {" / "} else if …"
	// For, Switch, Select, Case: same shape.
	}
}
```

`<!-- comments -->` and the doctype are static bytes. Whitespace (SPEC-001
R18) is already normalised by the parser in `Text.Value`.

## Line directives

```go
// directive returns the text to put immediately before user code that
// starts at p. See LAB-007 for why it targets column C-1.
func directive(p token.Position) string {
	if p.Column == 1 {
		return fmt.Sprintf("\n//line %s:%d:1\n", p.Filename, p.Line)
	}
	return fmt.Sprintf("/*line %s:%d:%d*/ ", p.Filename, p.Line, p.Column-1)
}
```

Filenames in directives are the path relative to the generated file's
directory (i.e. the bare file name), so output is identical on every machine
(R11).

## element rewriting

Collect package-level declarations named `element` across the file; if none,
rewrite every `*ast.Ident` named `element` that appears inside a type
expression (`Field.Type`, `ValueSpec.Type`, `TypeSpec.Type`, composite
literal `Type`, conversion `Fun` resolved syntactically as a type,
`FuncType` params/results, `MapType`, `ArrayType`, `ChanType`,
`StarExpr`, type arguments). Local shadowing in function scopes is handled
with a simple scope stack. Limitation (documented): an `element` declared in
*another* file of the same package is not seen; v1 reports a clear error
if a package declares `element` in a `.go` file (`golpe generate` checks the
package's other files).

# Testing

- **Runtime**: table tests for `Write`/`WriteAttr` over every type case,
  escaping of `& < > " '`; a failing-writer test; `BenchmarkWrite` for the
  DEC-004 overhead.
- **Generator unit tests**: small inputs → expected Go snippets.
- **Golden tests**: `internal/generator/testdata/<case>/x.golpe` →
  `x_golpe.go.golden`; `-update` flag to rewrite.
- **End-to-end**: for each golden case a `render_test.go` in a temporary
  module (`go.mod` with a `replace` to the repo) builds the generated code with
  `go run` and compares rendered HTML with `*.html.golden`. Also asserts
  that `go vet` errors for a deliberately broken case point at `x.golpe:L:C`.
- **Determinism**: generate twice, compare bytes; `gofmt -l` on output.

# Laboratories

- LAB-006 and LAB-007 (done, see PLAN-001): line directives work and survive
  gofmt with the C-1 rule.
- LAB-008 (todo, in TASK-002-01): measure `Write(w, any)` against a direct
  `io.WriteString` + `html.EscapeString`; if overhead > 20% on a 1k-node
  page, raise the priority of type-directed lowering.

# Risks

- **Splicing vs gofmt.** If a directive lands inside a token sequence gofmt
  rewrites (e.g. composite literals split over lines), positions may drift by
  a column. Mitigation: end-to-end position assertions in golden tests.
- **Attribute escaping** needs care for unquoted contexts; we always quote.

# Slices

1. [TASK-002-01 Runtime writers and escaping](/tasks/002-code-generation/01-runtime-writers.md)
2. [TASK-002-02 Lowering templates and markup](/tasks/002-code-generation/02-lowering.md)
3. [TASK-002-03 Line directives and hygiene](/tasks/002-code-generation/03-line-directives.md)
4. [TASK-002-04 End-to-end golden tests](/tasks/002-code-generation/04-end-to-end.md)
