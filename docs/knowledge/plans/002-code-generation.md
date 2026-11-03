---
type: Plan
id: PLAN-002
title: Generating Go from the extended AST
description: Splice the original source, replacing template headers, markup statements, element and writer() with standard-library Go; add line directives and position-wrapped errors; gofmt the result into x_gen.go.
status: draft
spec: SPEC-002
tags: [generator]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Approach

Implements [SPEC-002](/specs/002-code-generation.md) under
[DEC-005](/decisions/005-standard-library-only-output.md): the output imports
only `fmt`, `html` and `io`. The generator is a **splicer**, not a printer:
user Go code is copied byte-for-byte from the `.golpe` source (keeping its
comments and layout) and only these spans are replaced:

1. a template header `template Name[TP](params) {` →
   `func Name[TP](params) func(io.Writer) error { return func(__golpeW io.Writer) error {`,
   and its closing `}` → `return nil } }`;
2. each `MarkupStmt` span → lowered write statements;
3. `element` identifiers in type positions → `func(io.Writer) error`;
4. `writer()` calls → `__golpeW`.

Every copied span is preceded by a line directive (LAB-006/007), then the
file goes through `go/format.Source`.

# Architecture

| File in `internal/generator` | Owns |
|------------------------------|------|
| `generate.go` | `Generate(fset, file, src, resolve) ([]byte, error)`; splicing |
| `emit.go` | buffered emitter: static coalescing, checked writes, error wrapping |
| `names.go` | scope tracking; finding unshadowed `element` and `writer()` |
| `imports.go` | adding `fmt`/`html`/`io` hygienically |
| `directive.go` | line directives (C-1 rule) |
| `filename.go` | `GeneratedName("x_test.golpe") == "x_gen_test.go"` |

`resolve` is the component resolver of [PLAN-005](/plans/005-component-tags.md),
injected so the generator stays a pure function of its inputs.

# Snippets

## Emitter

```go
type emitter struct {
	buf      bytes.Buffer
	static   strings.Builder // pending static HTML
	staticAt token.Position  // where the pending static bytes start
	tmpl     string          // template name, for error wrapping
}

func (e *emitter) flush() {
	if e.static.Len() == 0 {
		return
	}
	e.checkedWrite(strconv.Quote(e.static.String()), e.staticAt)
	e.static.Reset()
}

// checkedWrite emits: if _, err := io.WriteString(__golpeW, <expr>); err != nil { return <wrap> }
func (e *emitter) checkedWrite(expr string, at token.Position) {
	fmt.Fprintf(&e.buf, "if _, err := %s.WriteString(__golpeW, %s); err != nil {\nreturn %s\n}\n",
		e.pkg("io"), expr, e.wrap(at))
}

// wrap returns fmt.Errorf("nav.golpe:8:6: NavItem: %w", err) (SPEC-002 R13).
func (e *emitter) wrap(at token.Position) string
```

## Values, render calls, attributes

```go
func (g *gen) interp(n *ast.Interp) {
	g.e.flush()
	src := g.directive(n.X) + g.source(n.X) // verbatim user expression
	if g.names.containsWriter(n.X) {       // DEC-006: writer() anywhere
		fmt.Fprintf(&g.e.buf, "if err := %s; err != nil {\nreturn %s\n}\n", src, g.e.wrap(g.pos(n)))
		return
	}
	fmt.Fprintf(&g.e.buf, "if __golpeV := any(%s); __golpeV != nil {\n", src)
	g.e.checkedWrite(g.e.pkg("html")+".EscapeString("+g.e.pkg("fmt")+".Sprint(__golpeV))", g.pos(n))
	g.e.buf.WriteString("}\n")
}

// <tag name={expr}>: plain if statements (SPEC-002 R11).
func (g *gen) dynamicAttr(a *ast.Attr, x *ast.Interp) {
	g.e.flush()
	escaped := g.e.pkg("html") + ".EscapeString(" + g.e.pkg("fmt") + ".Sprint(__golpeV))"
	fmt.Fprintf(&g.e.buf, "if __golpeV := any(%s); __golpeV == true {\n", g.directive(x.X)+g.source(x.X))
	g.e.checkedWrite(strconv.Quote(" "+a.Name), g.pos(a))
	g.e.buf.WriteString("} else if __golpeV != false && __golpeV != nil {\n")
	g.e.checkedWrite(strconv.Quote(" "+a.Name+`="`)+"+"+escaped+`+"\""`, g.pos(a))
	g.e.buf.WriteString("}\n")
}
```

Static parts (`<li>`, ` class="btn"`, text after R22) only go to
`e.static`, so `<li><span>` becomes one write, also across consecutive
markup statements. Embedded statements print their header verbatim, then the
lowered body: `if <init; cond> {` … `} else {` … `}`.

## Names and scopes

`names` walks the file with a scope stack (package declarations, params,
`:=`/`var` in blocks, range keys, func literal params). `element` and
`writer` are rewritten only when they resolve to the universe scope, like
Go's predeclared names. `writer()` outside a template body, or inside an
attribute value, is an error.

## Imports

If the file already imports `"io"` (or `fmt`, `html`) under some name, use
that name; otherwise add the import under its default name, or as
`__golpe_io` when `io` is already taken by another package. `html` is added
only when a value is written; `fmt` and `io` whenever a template exists.

# Testing

- **Unit tests:** header/footer splice; coalescing; value vs render-call
  detection (`{p.C(writer())}`, `{f(g(writer()))}`, `{json.NewEncoder(writer()).Encode(v)}`,
  shadowed `writer`); attributes true/false/nil/other; nil interface skip;
  embedded statements; `element` rewriting incl. shadowing; imports with
  name collisions; `GeneratedName` suffix cases.
- **Golden tests:** `internal/generator/testdata/<case>/x.golpe` →
  `x_gen.go.golden` with `-update`.
- **End-to-end:** each golden case is compiled in a temporary module (no
  dependency needed: std only) and rendered; HTML compared with
  `x.html.golden`. A broken case asserts that `go vet` reports `x.golpe:L:C`.
- **Determinism:** generate twice, compare bytes; `gofmt -l` on output.

# Laboratories

- LAB-006, LAB-007 (done, PLAN-001): line directives work and survive gofmt
  with the C-1 rule.
- LAB-008 (done, history): measured the runtime-dispatch design of
  [DEC-004](/decisions/004-runtime-dispatch-for-interpolation.md) against
  type-aware writes on an interpolation-heavy 1,005-node page: 50.9 µs vs
  32.2 µs (+58%), 762 vs 350 allocations; formatting everything with fmt was
  about +122%. Most of the cost is boxing values into `any`. DEC-005 accepts
  standard-library cost (`fmt.Sprint` per value) for simplicity; type-aware
  lowering is in the [backlog](/backlog.md).

# Risks

- **Verbose output.** Every write is checked and wrapped; acceptable for
  generated code, and gofmt keeps it readable.
- **Splicing vs gofmt.** A directive inside a construct gofmt rewrites could
  drift a column; end-to-end position assertions catch it.
- **Func values printed by mistake** (`{p.Children}`); documented, and a
  future vet-like check could flag fields of type `element` used as values.

# Slices

1. [TASK-002-01 Runtime writers and escaping](/tasks/002-code-generation/01-runtime-writers.md) ✅ (superseded by DEC-005; code removed)
2. [TASK-002-02 Lowering templates and markup](/tasks/002-code-generation/02-lowering.md)
3. [TASK-002-03 Line directives and wrapped errors](/tasks/002-code-generation/03-line-directives.md)
4. [TASK-002-04 End-to-end golden tests](/tasks/002-code-generation/04-end-to-end.md)
