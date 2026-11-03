---
type: Decision
id: DEC-002
title: golpe.Element is a type alias for func(io.Writer) error
description: element in .golpe files means func(io.Writer) error; the optional root package offers golpe.Element as an alias of that same type, so writer funcs interoperate without conversions.
status: stable
decision: accepted
tags: [runtime, api]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
verified: { by: human:sgq995, at: 2026-11-03T03:12:42Z }
---
# Context

The `element` type in `.golpe` files maps to `func(io.Writer) error`, so
users can "use the generated funcs freely".

# Options

1. `type Element = func(io.Writer) error` (alias).
2. `type Element func(io.Writer) error` (defined type, can have methods
   such as `Render` or `String`).

# Decision

Option 1. A named `func(io.Writer) error` from any library (or a method
value like `buf.WriteTo`-style adapters) is usable as an Element with no
conversion, and generated signatures read exactly like the maintainer's
example. Helpers that a defined type would offer as methods are package
functions instead (`golpe.Render`, `golpe.String` later).

# Consequences

- No methods on Element, ever, without a breaking change.
- Revised on 2026-11-03 by [DEC-005](/decisions/005-standard-library-only-output.md):
  generated code never imports golpe, so `element` is rewritten to the
  literal type `func(io.Writer) error`. `golpe.Element` stays as an optional
  alias for developers ([SPEC-006](/specs/006-optional-helpers.md)).
- The maintainer kept `func(io.Writer) error` over a context-taking or
  interface-based Element ([DEC-008](/decisions/008-no-implicit-context.md)).
