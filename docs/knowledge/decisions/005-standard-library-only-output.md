---
type: Decision
id: DEC-005
title: Generated code uses only the standard library
description: Generated _gen.go files import only fmt, html and io; there is no runtime package, the CLI is the only required tool, and golpe's Go API is optional.
status: stable
decision: accepted
tags: [generator, architecture]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
verified: { by: human:sgq995, at: 2026-11-03T03:12:42Z }
---
# Context

[DEC-004](/decisions/004-runtime-dispatch-for-interpolation.md) planned a
runtime package (`golpe.Write`, `golpe.WriteAttr`) that every generated file
would import, and TASK-002-01 built it. In the design interview of 2026-11-03
the maintainer asked for the opposite: "just use std fmt.Fprintf, fmt.Sprintf,
io.WriteString", a single CLI installable as a tool, and golpe APIs that are
"completely optional, just optional helpers, not hard dependencies", following
Go's philosophy of simplicity.

# Options

1. **Runtime package** (DEC-004): one type switch per value at runtime.
2. **Standard library only**: generated code spells everything out with
   `fmt`, `html` and `io`.
3. **Type-check with go/types**: type-specific code per value; a complex,
   slow generator.

# Decision

Option 2.

- Values: `html.EscapeString(fmt.Sprint(x))`, skipping nil interfaces.
- Child elements are rendered explicitly through `writer()`
  ([DEC-006](/decisions/006-writer-and-render-calls.md)), so the generator never
  needs to know whether a value is an element.
- Dynamic attributes use plain `if` statements for true/false/nil.
- Errors are wrapped with `fmt.Errorf` and the `.golpe` position.

Performance is whatever the standard library gives. LAB-008 measured the
runtime-dispatch alternative at +58% over type-aware writes on an
interpolation-heavy page, and an fmt-only variant at about +122%; the
maintainer accepted this cost for simplicity.

# Consequences

- No golpe import in generated code; a project can drop Golpe and keep its
  generated files.
- The root package only holds optional helpers
  ([SPEC-006](/specs/006-optional-helpers.md)); the TASK-002-01 runtime
  writers were removed.
- `{props.Children}` without `(writer())` prints a func value; it cannot be
  detected without types.
- Type-aware lowering remains a possible optimization in the
  [backlog](/backlog.md); it must emit identical bytes.
