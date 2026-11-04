---
type: Spec
id: SPEC-005
title: Component tags
description: A tag that is not an HTML, SVG or MathML element or a custom element calls a template (or Go func) found by reading its declaration, with props built from the tag's attributes.
status: draft
plan: PLAN-005
tags: [language, composition]
generated: { by: claude-code/agent, at: 2026-11-04T10:36:16Z }
---
# Summary

`<Card title="Hi" featured/>` calls `Card(CardProps{title: "Hi", featured:
true})(w)`. The generator decides that `Card` is a component (not HTML) from
the element lists, and finds `CardProps` by reading `Card`'s declaration;
there is no naming convention and no type-checking
([DEC-007](/decisions/007-component-tags.md)).

# Motivation

- **Composition is the core of a template language.** Tag syntax reads like
  the markup around it (as in Ripple and JSX), instead of
  `{Card(CardProps{…})(writer())}`.
- **Props stay structured and typed.** Attributes become fields of the
  component's own props struct, so the Go compiler checks every one.
- **Simple.** Components are plain functions; the generator emits one call.

# Actors

[Developer](/actors/developer.md) · [Parser](/actuators/parser.md) ·
[Generator](/actuators/generator.md)

# Entities

[Component tag](/entities/component-tag.md) ·
[Template declaration](/entities/template-declaration.md) ·
[Attribute](/entities/attribute.md) · [Markup node](/entities/markup-node.md)

# Requirements

- **R1 Classification.** A tag name is markup when it matches, ignoring ASCII
  case, an element of HTML, SVG or MathML, or contains `-` (a custom element).
  Any other name must be a Go identifier or a qualified identifier `pkg.Name`
  and is a component tag. A qualified name is always a component. Anything
  else is an error. The lists are the current standards (WHATWG HTML, SVG 2,
  MathML Core; 202 names, retrieved 2026-11-04); obsolete elements such as
  `center`, `font` or `marquee` are not markup, so `<center>` is a component
  tag (pending the maintainer's confirmation).
- **R2 Resolution.** The generator MUST find the declaration syntactically: a
  top-level `template` or `func` named `Name` in the package's `.golpe` and
  `.go` files (not `_gen.go`), or, for `pkg.Name`, in the source of the
  package the file imports as `pkg`.
- **R3 Signature.** The declaration MUST have no type parameters and zero or
  one parameter; a `func` must return `func(io.Writer) error` (or `element`).
  Zero parameters → `Name()(w)`; one parameter of type `T` →
  `Name(T{…})(w)`; of type `*T` → `Name(&T{…})(w)`. Attributes on a
  zero-parameter component are an error.
- **R4 Attributes become fields.** `name="v"` → `name: "v"` (a string);
  `name={expr}` → `name: expr` (the Go value as is: not stringified, not
  escaped); a bare `name` → `name: true`. Names are used verbatim as field
  names, case kept, so they must be Go identifiers (`aria-label` is an
  error). Duplicates are errors. Unknown or mistyped fields are Go compile
  errors reported at the `.golpe` position.
- **R5 No children in v1.** A component tag MUST be self-closing
  (`<Card title="x"/>`) or have empty content (`<Card></Card>`). Content is an
  error whose message suggests a props field (`Children element`). Slots are
  in the [backlog](/backlog.md).
- **R6 Rendering.** The call is a render call: its error is checked, wrapped
  with the tag's position and returned (SPEC-002 R13).
- **R7 Generic components** cannot be used as tags (type arguments cannot be
  inferred from a composite literal); the error suggests interpolation.

# Known consequence of R1

Because element names match ignoring case, a component named like an element
cannot be used as an unqualified tag: `<Button>`, `<Header>`, `<Footer>`,
`<Nav>`, `<Main>`, `<Section>`, `<Form>`, `<Input>`, `<Label>`, `<Select>`,
`<Option>`, `<Table>`, `<Dialog>`, `<Menu>`, `<Search>`, `<Title>`, `<Link>`,
`<Data>`, `<Time>`, `<Image>` and `<Text>` (SVG), and others, are HTML/SVG.
Use a distinct name (`PrimaryButton`, `SiteHeader`), a qualified tag
(`<ui.Button/>`), or interpolation (`{Button(p)(writer())}`).

# Examples

```golpe
type CardProps struct {
    title    string
    featured bool
    Body     element
}

template Card(p CardProps) {
    <article class={golpe.If(p.featured, "featured")}>
        <h2>{p.title}</h2>
        {p.Body(writer())}
    </article>
}

template Footer() {
    <footer>© 2026</footer>
}

template Page(items []Item) {
    <main>
    for _, it := range items {
        <Card title={it.Name} featured={it.Hot} Body={it.Summary}/>
    }
    </main>
    <Footer/>
}
```

`<Card …/>` lowers to
`Card(CardProps{title: it.Name, featured: it.Hot, Body: it.Summary})(__golpeW)`
and `<Footer/>` to `Footer()(__golpeW)`, each with a checked, wrapped error.

# Acceptance criteria

- [ ] `<div>`, `<DIV>`, `<linearGradient>`, `<math>` and `<my-widget>` are
      markup; `<Card>`, `<card>`, `<ui.Button>` are components; `<svg:path>`
      is an error (R1).
- [ ] The example above generates the documented calls, for a template in
      the same package, a `func` in a `.go` file, and a template in another
      package (R2, R3).
- [ ] `<Card>text</Card>` is an error suggesting a props field (R5).
- [ ] `<Card aria-label="x"/>` and `<Card title="a" title="b"/>` are errors (R4).
- [ ] A misspelled attribute is reported by `go build` at the tag's `.golpe`
      position (R4).

# Out of scope

Children and slots, attribute spread, generic components as tags, component
tags as attribute values ([backlog](/backlog.md)).

# Open questions

None.
