---
type: Plan
id: PLAN-005
title: Implementing component tags
description: Classify tag names with element lists in internal/token, parse component tags into an ast.Component node, resolve declarations syntactically in internal/resolve, and lower each tag to one checked call.
status: draft
spec: SPEC-005
tags: [parser, generator, composition]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Approach

Implements [SPEC-005](/specs/005-component-tags.md) and
[DEC-007](/decisions/007-component-tags.md) in four steps: name lists, parsing,
resolution, lowering. Everything is syntactic: no go/types.

# Architecture

| Package | Adds |
|---------|------|
| `internal/token` | `IsElement(name string) bool` (HTML + SVG + MathML, ASCII-case-insensitive), `IsCustomElement(name string) bool` (contains `-`) |
| `internal/ast` | `Component{Lt, Name goast.Expr /* *Ident or *SelectorExpr */, Attrs []*Attr, SelfClosing bool, Close, EndPos token.Pos}` |
| `internal/parser` | in pass 2, non-element tag names become `Component`; content is an error |
| `internal/resolve` | `Resolver.Lookup(file *goast.File, dir string, name goast.Expr) (Signature, error)` |
| `internal/generator` | lowering of `Component` using the resolver |

# Snippets

## Classification

```go
// IsElement reports whether name is an HTML, SVG or MathML element name,
// ignoring ASCII case (DEC-007). Lists are generated from the WHATWG HTML,
// SVG 2 and MathML Core indexes into a sorted table (go:generate, checked in).
func IsElement(name string) bool
```

## Resolution

```go
type Signature struct {
	NumParams int        // 0 or 1
	PropsType goast.Expr // the parameter's type expression, e.g. CardProps or *CardProps
	PropsPkg  string     // import path when the props type is qualified elsewhere
}

// Lookup finds a top-level template or func named Name in dir's .golpe and
// .go files (not _gen.go), or, for pkg.Name, in the directory of the import
// with that name (go/build.Import(path, dir, build.FindOnly)). Parsed
// declarations are cached per directory and invalidated by mtime.
func (r *Resolver) Lookup(file *goast.File, dir string, name goast.Expr) (Signature, error)
```

`.golpe` files are parsed with `internal/parser` (declarations only);
`.go` files with `go/parser` (`SkipObjectResolution`). Errors: not found,
generic, more than one parameter, a `func` not returning
`func(io.Writer) error`/`element`.

## Lowering

```go
// <Card title="Hi" featured={x}/>  →
//   if err := Card(CardProps{title: "Hi", featured: x})(__golpeW); err != nil { return wrap }
// *T parameter → &T{…}; zero parameters → Footer()(__golpeW).
// A props type written as CardProps in package ui is emitted as ui.CardProps,
// reusing the file's import name for ui.
```

# Testing

- `IsElement`: every name in the three lists (mixed case), plus non-names.
- Parser: classification table; `<Card>text</Card>` error; attribute forms.
- Resolver: testdata module with a same-package template, a `func` in a `.go`
  file, a pointer-props template, a template in another package, generic and
  two-parameter errors.
- Generator golden: the SPEC-005 example; e2e render.

# Laboratories

- LAB-013 (todo): measure `go/build.Import(…, FindOnly)` latency per import in
  module mode (it runs `go list`); if slow for the LSP, cache per session.

# Risks

- **Name collisions** with elements (`Button`, `Header`) surprise users;
  the parser error for content in such tags should mention the rule.
- **Element lists drift** as HTML evolves; regenerate them per release.

# Slices

1. [TASK-005-01 Element name lists](/tasks/005-component-tags/01-element-names.md)
2. [TASK-005-02 Parse component tags](/tasks/005-component-tags/02-parse-components.md)
3. [TASK-005-03 Declaration resolver](/tasks/005-component-tags/03-resolver.md)
4. [TASK-005-04 Lower component tags](/tasks/005-component-tags/04-lowering.md)
