---
type: Decision
id: DEC-006
title: writer() and render calls
description: A predeclared writer() exposes the current io.Writer explicitly; any {expr} containing writer() is a render call whose single error result is checked; every other {expr} is an escaped value.
status: stable
decision: accepted
tags: [language, generator]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
verified: { by: human:sgq995, at: 2026-11-03T03:12:42Z }
---
# Context

Templates must render other elements (`props.Children`, a child template)
without a runtime that inspects types ([DEC-005](/decisions/005-standard-library-only-output.md)).
An implicit, magic name for the writer was rejected for the same reason as an
implicit context ([DEC-008](/decisions/008-no-implicit-context.md)), but the
maintainer asked for a way to reach the writer, alongside fragments.

# Options

- Name: expose the generated `__golpeW`; a variable `w`; a predeclared
  function `writer()`.
- Render-call detection: `writer()` as a direct argument of the outermost
  call, or `writer()` anywhere in the expression.

# Decision

- `writer()` is a predeclared function, valid only inside template bodies,
  returning the `io.Writer` being rendered to. It can be shadowed, like Go's
  predeclared names, and the generator's `__golpe` names stay private.
- An interpolation that contains `writer()` **anywhere** is a render call:
  `if err := <expr>; err != nil { return … }`. So
  `{json.NewEncoder(writer()).Encode(v)}` is a render call too. The Go
  compiler rejects a render expression that is not a single `error`.
- Every other interpolation is a value, written escaped.
- `writer()` is not allowed in attribute values.
- Go-mode code may call `writer()` freely:
  `if err := p.Header(writer()); err != nil { return err }`.

# Consequences

- The generator finds `writer()` syntactically, honouring shadowing.
- `{p.Children}` without the call prints a func value (no types, no
  detection).
- Unescaped output is possible on purpose through `writer()`
  (`io.WriteString(writer(), trustedHTML)`), which is easy to grep.
