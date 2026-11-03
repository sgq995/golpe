---
type: Task Slice
id: TASK-002-01
title: Runtime writers and escaping
description: Add Writer, WriteString, Write and WriteAttr to the root golpe package with HTML/attribute escaping, tests and the LAB-008 benchmark.
spec: SPEC-002
plan: PLAN-002
state: done
depends_on: [TASK-001-01]
touches: [golpe.go, write.go, write_test.go, escape.go]
generated: { by: claude-code/agent, at: 2026-11-03T01:58:38Z }
---
# Goal
The [runtime](/actuators/runtime.md) half of [PLAN-002](/plans/002-code-generation.md) ("Runtime"), per [DEC-004](/decisions/004-runtime-dispatch-for-interpolation.md). Covers [SPEC-002](/specs/002-code-generation.md) R7, R12.

# Depends on
[TASK-001-01](/tasks/001-template-language/01-project-scaffold.md) only, so it can run in parallel with the parser slices.

# Todo
- [x] Tests first: `Write` for nil, nil Element, Element, string, Stringer, error, []byte, int, float, bool; `<script>` → `&lt;script&gt;`
- [x] Tests: `WriteAttr` quotes and escapes `" ' & < >`; writer errors are returned
- [x] `Writer` alias, `WriteString`, `Write`, `WriteAttr`
- [x] LAB-008: `BenchmarkWrite` vs direct escaping; record numbers in PLAN-002
- [ ] Resolve SPEC-002 open question 1 (root package vs `golpe/runtime`) with the maintainer

# Done when
Tests pass; benchmark numbers recorded.

# Notes
- **API.** `type Writer = io.Writer` (`golpe.go`); `WriteString(w Writer, s string) error`,
  `Write(w Writer, v any) error` and `WriteAttr(w Writer, v any) error` (`write.go`). The escaping
  primitives in `escape.go` are unexported and produce exactly `html.EscapeString`'s bytes
  (`&amp; &lt; &gt; &#34; &#39;`); no escaper is exported.
- **Deviations from the plan's first snippet** (the plan now shows the built shape):
  - Every value that is not nil, an Element, a string or a `[]byte` renders exactly as
    `fmt.Sprint` formats it, then escaped. So fmt's rules hold: Formatter before error before
    Stringer, a nil pointer prints `<nil>` (escaped), and a panicking String/Error method is
    reported instead of crashing the render. The snippet called `String()` itself, before
    `Error()`, which panics on nil pointer receivers.
  - Values of named func types whose underlying type is `func(io.Writer) error` (such as
    `type Icon func(io.Writer) error`) are Elements: a reflect check in the slow path calls
    them, and a typed nil renders nothing. `case Element:` alone misses them, and fmt would
    print a code address. Being an Element wins over also having a String method.
  - Fast paths chosen from LAB-008 numbers, all byte-identical to fmt
    (`TestWriteMatchesFmt`): strconv for predeclared numbers and bools; Error or String called
    directly, unless the value is a `fmt.Formatter` or a `reflect.Value` (left to fmt). If
    the method panics, fmt calls it a second time and prints the panic its usual way.
  - `WriteAttr` writes the quotes itself (three writes), so the generator emits only `name=`
    before it; nil writes `""`. An Element (nil or not, named or not) is an error that names
    its type; nothing is written and the Element is not called.
  - An empty escaped value (`""`, a nil `[]byte`) makes no write at all.
- **For TASK-002-02.** Escape static text and static attribute values with
  `html.EscapeString` so that static and dynamic output agree byte for byte.
- **Spec gaps for the coordinator.** SPEC-002, DEC-004 and ENT-interpolation do not say that an
  Element as an attribute value is an error, or that fmt's precedence (error before Stringer)
  applies. SPEC-002's acceptance examples (`&lt;script&gt;`, `href="&#34;onmouseover=x"`,
  stopping at the first failing write) hold at the runtime level. They stay unticked until
  the end-to-end tests (TASK-002-04) cover them.
- **LAB-008.** +58% on the 1,005-node page (threshold 20%). Interface boxing at the call site
  is most of it, and the runtime cannot remove it. Numbers and conclusion are in PLAN-002.
  Raising type-directed lowering's priority is the maintainer's call.
- **Open question 1** (root package vs `golpe/runtime`) is left for the coordinator to settle
  with the maintainer; the helpers live in the root package for now.
- Extra file outside `touches`: `example_test.go`, with runnable `Example` (a hand-lowered
  template), `ExampleWrite` and `ExampleWriteAttr`.
