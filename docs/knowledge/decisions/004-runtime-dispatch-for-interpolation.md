---
type: Decision
id: DEC-004
title: Interpolations dispatch on type at runtime in v1
description: Lower {expr} to a call into a runtime helper that renders Elements and escapes everything else, instead of type-checking .golpe files at generation time.
status: deprecated
decision: superseded
tags: [generator, runtime, security]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
> **Superseded** on 2026-11-03 by
> [DEC-005](/decisions/005-standard-library-only-output.md): the maintainer
> chose standard-library-only output with explicit render calls through
> `writer()` ([DEC-006](/decisions/006-writer-and-render-calls.md)). The
> runtime writers built for this decision (TASK-002-01) were removed; their
> LAB-008 measurements stay in [PLAN-002](/plans/002-code-generation.md).

# Context

`{props.Children}` must call the Element; `{props.Name}` must write escaped
text. The maintainer's example lowered `href={props.Route}` to
`fmt.Fprintf(w, `<a href="%v">`, props.Route)`, which is an XSS hole: values
must be escaped. The generator needs to know, or not need to know, the type of
each expression.

# Options

1. **Type-check at generation time** with `go/types` (needs the whole
   package, including other `.golpe` files, loaded and partially generated).
   Fastest output, much more complex generator.
2. **Runtime dispatch**: emit `__golpe.Write(w, expr)`; the helper type-switches:
   Element / `func(io.Writer) error` → call (nil → nothing), `string` →
   escape, `fmt.Stringer`, `error`, numbers, `bool` → format then escape.
   Attributes use `__golpe.WriteAttr`.

# Decision

Option 2 for v1. It is simple, always correct, and keeps generation a pure
per-file syntax transformation (fast, and usable by the LSP on every
keystroke). Option 1 stays a later optimisation that must produce identical
output bytes.

# Consequences

- One interface conversion per interpolation; acceptable for v1, measured by
  a benchmark in [PLAN-002](/plans/002-code-generation.md).
- Escaping lives in one audited place, the [runtime](/actuators/runtime.md).
