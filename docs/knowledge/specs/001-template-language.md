---
type: Spec
id: SPEC-001
title: Template language
description: Golpe extends Go with template declarations whose bodies switch from Go to balanced HTML markup with embedded if/for/switch/select and {expr} interpolation.
status: draft
plan: PLAN-001
tags: [language, parser]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Summary

Golpe is a superset of Go for writing HTML views, in the spirit of templ (Go)
and Ripple (JavaScript). It adds one declaration, `template`, which works like
`func`. Inside a template body everything is normal Go until an HTML tag
opens; from there until that tag's matching close the source is **markup**,
where only tags, text, `{expr}` interpolations and *embedded* `if`, `for`,
`switch` and `select` are allowed. After the close tag the body is Go again.

# Motivation

- **Type safety.** Views are checked by the Go compiler: a renamed field or
  wrong argument is a build error, not a blank page in production.
- **Just Go.** No second expression language (unlike `html/template`): the
  Go you write around and inside markup is real Go, with real imports,
  generics and tooling.
- **Superset, not a new language.** Any valid Go file is a valid Golpe file
  with the same meaning, so adoption is incremental and gopls-based tooling
  can be reused ([SPEC-004](/specs/004-language-server.md)).
- **Plain output.** Templates compile to ordinary functions returning
  `func(io.Writer) error` that users compose freely
  ([SPEC-002](/specs/002-code-generation.md)).

# Actors

- [Developer](/actors/developer.md) writes `.golpe` files.
- [Parser](/actuators/parser.md) reads them.

# Entities

[Golpe source file](/entities/golpe-file.md) ·
[Template declaration](/entities/template-declaration.md) ·
[Element](/entities/element.md) ·
[Markup node](/entities/markup-node.md) ·
[Embedded statement](/entities/embedded-statement.md) ·
[Interpolation](/entities/interpolation.md) ·
[Attribute](/entities/attribute.md)

# Requirements

## Files and declarations

- **R1 Superset.** Every valid Go source file MUST parse as a Golpe file
  with identical meaning.
- **R2 Template declaration.** A top-level declaration
  `template Name(params) { body }` MUST be accepted. It has the syntax of a
  `func` declaration without a result list; its result is implicitly
  `element`. Type parameters (`template List[T any](…)`) SHOULD and method
  receivers (`template (c Card) View()`) SHOULD be accepted.
- **R3 Contextual keyword.** `template` MUST be a keyword only as the first
  token of a top-level declaration followed by an identifier or `(`;
  everywhere else it is an identifier
  ([DEC-003](/decisions/003-template-is-a-contextual-keyword.md)).
- **R4 element type.** `element` MUST be a predeclared type in Golpe files,
  identical to `golpe.Element` (`func(io.Writer) error`,
  [DEC-002](/decisions/002-element-is-an-alias.md)). Like Go's predeclared
  identifiers it MAY be shadowed.

## Template bodies

- **R5 Go mode.** A template body MUST accept any Go statement.
- **R6 Markup statements.** At any statement position in a template body
  (including inside nested Go blocks), `<` immediately followed by an ASCII
  letter MUST start a markup statement. `<-` keeps its Go meaning. The markup
  statement ends at the close of its root tag; then Go mode resumes. A body
  MAY contain any number of markup statements, interleaved with Go.
- **R7 Doctype.** `<!DOCTYPE …>` at a statement position MUST be accepted
  as a markup statement of its own.
- **R8 return.** `return` ends rendering successfully; `return err` (a value
  of type `error`) ends it with that error. No other result forms are allowed.
- **R9 Reserved names.** Identifiers starting with `__golpe` are reserved for
  generated code and MUST be rejected in Golpe files.

## Markup

- **R10 Balanced tree.** Every non-void tag MUST be closed by a matching
  `</name>`, or be self-closing (`<x/>`). HTML void elements (`area`, `base`,
  `br`, `col`, `embed`, `hr`, `img`, `input`, `link`, `meta`, `source`,
  `track`, `wbr`) need no close tag. A mismatch MUST be an error at the
  offending tag's position naming the expected tag.
- **R11 Tag names.** Lowercase HTML names, custom elements (`my-widget`) and
  namespaced names (`svg:path`) MUST be accepted.
- **R12 Attributes.** `name="v"`, `name='v'`, boolean `name` and dynamic
  `name={expr}` MUST be accepted. Names follow HTML (`data-x`, `aria-label`,
  `hx-get`, `@click`, `:class`).
- **R13 Text.** Characters other than `<`, `{` and `}` are text. A literal
  brace is written as an interpolation of a string: `{"{"}`.
- **R14 Interpolation.** `{expr}` MUST accept any Go expression, including
  composite and function literals containing braces. In content, an
  [Element](/entities/element.md) is rendered (nil renders nothing); any
  other value is written as escaped text. As an attribute value it is written
  attribute-escaped and double-quoted.
- **R15 Embedded statements.** Inside markup, `if`/`else if`/`else`, all
  forms of `for` (including `range`), expression and type `switch`, and
  `select` MUST be accepted when the keyword is the first non-blank token of
  a line. Their headers, `case` and `default` clauses follow Go's grammar;
  their bodies contain only markup content (tags, text, interpolations,
  embedded statements). Go statements inside embedded bodies are an error.
- **R16 Raw text elements.** The content of `<script>` and `<style>` MUST be
  taken literally up to the matching close tag: no interpolation, no embedded
  statements.
- **R17 Comments.** HTML comments `<!-- … -->` are markup nodes and are kept
  in the output. Go comments (`//`, `/* */`) MAY appear in markup where a node
  could start, after whitespace; inside a run of text `//` is literal text (so
  `https://example.com` is text).
- **R18 Whitespace.** Whitespace-only text between tags that contains a
  newline is dropped. Other text has leading and trailing whitespace that
  contains a newline trimmed and inner whitespace runs collapsed to one space.

## Diagnostics

- **R19** All syntax errors MUST be reported as `file.golpe:line:col: message`
  with the position of the offending byte, and the parser SHOULD report all
  errors in a file rather than stopping at the first.

# Examples

```golpe
type ButtonProps struct {
    Children element
}

template Button(props ButtonProps) {
    <button>{props.Children}</button>
}
```

```golpe
template NavItem(props NavItemProps) {
    <li>
    if props.Selected {
        <span>{props.Children}</span>
    } else {
        <a href={props.Route}>{props.Children}</a> // a Go comment
    }
    </li>
}
```

```golpe
template List[T fmt.Stringer](title string, items []T) {
    heading := strings.ToUpper(title)    // Go mode
    <section>
        <h2>{heading}</h2>
        switch len(items) {
        case 0:
            <p>Nothing here yet.</p>
        default:
            <ul>
            for _, it := range items {
                <li>{it}</li>
            }
            </ul>
        }
    </section>
}
```

# Acceptance criteria

- [ ] Every `.go` file in GOROOT/src parses as Golpe with an AST equal to
      go/parser's (R1).
- [ ] The three examples above parse into the expected tree (golden test).
- [ ] `template.Must(x)` inside a function body and `var template = 1` at top
      level parse as Go (R3).
- [ ] `<-ch` in a template body is a receive statement (R6).
- [ ] `<div><span></div>` reports `expected </span>` at the `</div>` (R10).
- [ ] `<p>if you can</p>` is text; `<script>if (a) { b() }</script>` is raw (R15, R16).
- [ ] `<a href="https://x.dev">see https://x.dev</a>` keeps both URLs as text (R17).
- [ ] Errors carry `.golpe` line and column (R19).

# Out of scope

- Calling components with tag syntax (`<Button>…</Button>`) and writing
  markup as an expression (`Children: <b>hi</b>`). Today a template is
  called through interpolation: `{Button(ButtonProps{Children: icon})}`.
- Attribute spreading and conditional attributes.
- Go statements, `break` and `continue` inside embedded bodies.
- Context-aware escaping inside `<script>`/`<style>` and URL sanitising of
  `href`/`src` (planned as a follow-up security spec).

# Open questions

1. **Keyword.** `template` (chosen) or `component`, or both? See
   [DEC-003](/decisions/003-template-is-a-contextual-keyword.md).
2. **context.Context.** Should Element be `func(io.Writer) error` (as
   specified) or take a `context.Context` too (templ does) so templates can
   read request-scoped values and honour cancellation?
3. **Whitespace (R18).** Is the collapsing rule right, or should text be
   emitted verbatim?
4. **Component syntax.** When should `<Button>` call syntax and markup
   expressions arrive (a separate spec)?
