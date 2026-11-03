---
type: Spec
id: SPEC-001
title: Template language
description: Golpe extends Go with template declarations whose bodies switch from Go to balanced HTML markup with embedded if/for/switch/select, {expr} interpolation, fragments and an explicit writer().
status: draft
plan: PLAN-001
tags: [language, parser]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Summary

Golpe is a superset of Go for writing server-rendered HTML, borrowing part of
the syntax of templ (Go) and Ripple (JavaScript), not their runtimes. It adds
one declaration, `template`, which works like `func`. Inside a template body
everything is normal Go until an HTML tag opens; from there until that tag's
matching close the source is **markup**: tags, raw text, `{expr}`
interpolations and *embedded* `if`, `for`, `switch` and `select`. After the
close tag the body is Go again. A template compiles to a plain Go function
returning `func(io.Writer) error`.

# Motivation

- **Type safety.** Views are checked by the Go compiler: a renamed field is a
  build error, not a blank page in production.
- **Just Go.** No second expression language: the Go around and inside markup
  is real Go, with real imports, generics and tooling.
- **Superset.** Every valid Go file is a valid Golpe file with the same
  meaning, so adoption is incremental and gopls can be reused
  ([SPEC-004](/specs/004-language-server.md)).
- **Simple and explicit, like Go.** No implicit context, no hidden runtime:
  the generated code uses only the standard library
  ([DEC-005](/decisions/005-standard-library-only-output.md)), and the only
  required tool is one CLI ([SPEC-003](/specs/003-cli.md)).

Scope is server-side templates only: no routing, no client-side reactivity
(see [backlog](/backlog.md)).

# Actors

- [Developer](/actors/developer.md) writes `.golpe` files.
- [Parser](/actuators/parser.md) reads them.

# Entities

[Golpe source file](/entities/golpe-file.md) ·
[Template declaration](/entities/template-declaration.md) ·
[Element](/entities/element.md) ·
[writer()](/entities/writer-function.md) ·
[Markup node](/entities/markup-node.md) ·
[Embedded statement](/entities/embedded-statement.md) ·
[Interpolation](/entities/interpolation.md) ·
[Attribute](/entities/attribute.md) ·
[Component tag](/entities/component-tag.md)

# Requirements

## Files and declarations

- **R1 Superset.** Every valid Go source file MUST parse as a Golpe file with
  identical meaning.
- **R2 Template declaration.** A top-level declaration
  `template Name(params) { body }` MUST be accepted. It has the syntax of a
  `func` declaration without a result list; its result is
  `func(io.Writer) error`. Type parameters and method receivers
  (`template (c Card) View()`) SHOULD be accepted.
- **R3 Contextual keyword.** `template` is a keyword only as the first token of
  a top-level declaration followed by an identifier or `(`; everywhere else it
  is an identifier ([DEC-003](/decisions/003-template-is-a-contextual-keyword.md)).
- **R4 element.** `element` MUST be a predeclared type in Golpe files meaning
  `func(io.Writer) error` ([DEC-002](/decisions/002-element-is-an-alias.md)).
  Like Go's predeclared identifiers it MAY be shadowed.
- **R5 writer().** `writer()` MUST be a predeclared function, valid only inside
  template bodies, returning the `io.Writer` the template is rendering to. It
  MAY be shadowed. Calling it outside a template body, or inside an attribute
  value, is an error ([DEC-006](/decisions/006-writer-and-render-calls.md)).

## Template bodies

- **R6 Go mode.** A template body MUST accept any Go statement.
- **R7 Markup statements.** At any statement position in a template body,
  including nested Go blocks but not function literals, `<` immediately
  followed by an ASCII letter, or `<>`, MUST start a markup statement. `<-`
  keeps its Go meaning. The statement ends at the close of its root tag (or
  `</>`); then Go mode resumes. A body MAY contain any number of markup
  statements, interleaved with Go.
- **R8 Fragments.** `<>…</>` MUST render only its content. It is how a
  template renders bare text or a bare interpolation:
  `<>{p.Children(writer())}</>`.
- **R9 Doctype.** `<!DOCTYPE …>` at a statement position MUST be accepted as a
  markup statement of its own.
- **R10 Evaluation.** The whole body runs every time the returned function is
  called (every render). Calling the template function itself only captures
  its arguments.
- **R11 return.** `return` ends rendering successfully; `return err` ends it
  with that error. Output already written stays written (streaming).
- **R12 Reserved names.** Identifiers starting with `__golpe` are reserved for
  generated code and MUST be rejected.

## Markup

- **R13 Balanced tree.** Every non-void tag MUST be closed by a matching
  `</name>` (matched ASCII-case-insensitively, as in HTML) or be self-closing
  (`<x/>`). HTML void elements (`area`, `base`, `br`, `col`, `embed`, `hr`,
  `img`, `input`, `link`, `meta`, `source`, `track`, `wbr`) need no close tag
  and cannot have content. A mismatch MUST be an error at the offending tag
  naming the expected tag.
- **R14 Tag names.** A name of an HTML, SVG or MathML element (matched
  ASCII-case-insensitively) or a name containing `-` (a custom element) is
  markup and is written as authored. Any other name is a component tag
  ([SPEC-005](/specs/005-component-tags.md)).
- **R15 Attributes.** `name="v"` and `name='v'` are literal and written
  verbatim: there is no interpolation inside quotes, so values such as
  Alpine's `x-data="{ open: false }"` pass through. A bare `name` is a boolean
  attribute. `name={expr}` is dynamic: `true` writes the bare name, `false` or
  `nil` omits the attribute, anything else writes `name="…"` escaped.
  Only the predeclared `bool` type gets boolean treatment. Attributes whose
  values are the strings `"true"`/`"false"` (ARIA, `draggable`,
  `contenteditable`) must be given a string:
  `aria-busy={strconv.FormatBool(busy)}`. Attribute names are written as
  authored (`viewBox`, `hx-get`, `@click`, `:class`). Duplicate names and
  unquoted values (`class=btn`) are errors.
- **R16 Text.** Characters other than `<`, `{` and `}` are text, written
  verbatim, so HTML entities such as `&nbsp;` work. A literal brace is written
  as an interpolated string: `{"{"}`.
- **R17 Interpolation.** `{expr}` MUST accept any single Go expression.
  - If `expr` contains a call to `writer()` anywhere, it is a **render call**:
    it must evaluate to exactly one `error`, which is checked and returned.
    Example: `{props.Children(writer())}`.
  - Otherwise `expr` is a **value**: written as
    `html.EscapeString(fmt.Sprint(expr))`, except that a nil interface value
    writes nothing.
  - `{props.Children}` without `(writer())` is a value and prints fmt's
    representation of a func; the generator cannot detect this without type
    information (known limitation).
- **R18 Embedded statements: the line rule.** Inside markup, a line whose
  first word (first Go token) is `if`, `for`, `switch` or `select` starts an
  embedded statement; any other line is markup content. Whole words only:
  `format` is text. To start a text line with one of those words, write it as
  an interpolated string: `{"if "}you can…`.
  - Headers follow Go's grammar up to the `{` that opens the body.
  - Bodies contain only markup: tags, text, interpolations, fragments,
    component tags and nested embedded statements. Go statements are errors.
  - Inside an embedded `switch`/`select` body, lines starting with `case` or
    `default` open clauses (Go grammar up to `:`).
  - A line starting with `}` closes the body; `else` / `else if` must follow
    that `}` on the same line, as in Go.
  - A short body may stay on one line: `if ok { <b>yes</b> }`.
- **R19 Raw text elements.** The content of `<script>` and `<style>` MUST be
  taken literally up to the matching close tag: no interpolation, no embedded
  statements, no whitespace changes.
- **R20 Text-only and preformatted elements.** `<textarea>` and `<title>`
  contain only text and interpolations. `<pre>` and `<textarea>` keep their
  whitespace exactly as written.
- **R21 Comments.** Go comments (`//`, `/* */`) MAY appear in markup at a node
  boundary (at the start of a line, or after whitespace following a tag or an
  interpolation); they are template-only and never written. Inside a run of
  text `//` is text (`https://example.com`). HTML comments `<!-- … -->` are
  copied to the output.
- **R22 Whitespace, as browsers treat it.** In text, every whitespace run,
  newlines included, becomes one space. Runs at the very start or end of an
  element's, fragment's or embedded body's content are dropped. Exceptions:
  R19 and R20.
  `<p>⏎  Hello,⏎  {name}⏎</p>` renders `<p>Hello, Ana</p>`;
  `<b>Hi</b>⏎  there` renders `<b>Hi</b> there`.

## Diagnostics

- **R23** All errors MUST be reported as `file.golpe:line:col: message` at the
  offending byte, and the parser SHOULD report every error in a file. Errors
  caused by the line rule SHOULD hint at quoting (`write {"for"} to start a
  text line with 'for'`).

# Examples

```golpe
type ButtonProps struct {
    Children element
}

template Button(props ButtonProps) {
    <button>{props.Children(writer())}</button>
}
```

```golpe
template NavItem(props NavItemProps) {
    <li>
    if props.Selected {
        <span>{props.Children(writer())}</span>
    } else {
        <a href={props.Route}>{props.Children(writer())}</a> // a Go comment
    }
    </li>
}
```

```golpe
template List[T fmt.Stringer](title string, items []T, hidden bool) {
    heading := strings.ToUpper(title)    // Go mode
    <section hidden={hidden}>
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

```golpe
template Layout(p LayoutProps) {
    if p.Header != nil {
        if err := p.Header(writer()); err != nil {   // writer() in Go mode
            return err
        }
    }
    <>{p.Children(writer())}</>                      // fragment
}
```

# Acceptance criteria

- [ ] Every `.go` file in GOROOT/src parses as Golpe with an AST equal to
      go/parser's (R1).
- [ ] The four examples above parse into the expected trees (golden tests).
- [ ] `template.Must(x)` in a function body and `var template = 1` at top level
      parse as Go (R3).
- [ ] `<-ch` in a template body is a receive statement (R7).
- [ ] `<div><span></div>` reports `expected </span>` at the `</div>` (R13);
      `<DIV></div>` is valid.
- [ ] `<p>⏎  if you can⏎</p>` reports a line-rule error with the quoting hint;
      `<p>{"if"} you can</p>` is text (R18).
- [ ] `<script>if (a) { b() }</script>` is raw (R19); `x-data="{ open: false }"`
      is a literal attribute (R15).
- [ ] `<a href="https://x.dev">see https://x.dev</a>` keeps both URLs as text
      (R21).
- [ ] Whitespace examples of R22 render as stated.
- [ ] Errors carry `.golpe` line and column (R23).

# Out of scope

Recorded with reasons in the [backlog](/backlog.md): component children and
slots, attribute spread, conditional attribute groups, interpolation inside
quoted attribute values, an implicit context, `<script>`/`<style>` helpers,
URL and JavaScript sanitizing (a later security spec), client-side
reactivity, `golpe fmt`.

# Open questions

None. All questions were settled in the design interview of 2026-11-03 (see
[log](/log.md)).
