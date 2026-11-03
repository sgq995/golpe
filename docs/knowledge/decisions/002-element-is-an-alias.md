---
type: Decision
id: DEC-002
title: golpe.Element is a type alias for func(io.Writer) error
description: Define Element as an alias rather than a defined type so generated templates and hand-written writer funcs interoperate without conversions.
status: draft
decision: proposed
tags: [runtime, api]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
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
- `element` in a `.golpe` file is rewritten to `golpe.Element` (an alias,
  so it is the same type as `func(io.Writer) error`).
