---
type: Task Slice
id: TASK-002-02
title: Lowering templates and markup
description: Implement internal/generator: splice template headers, lower markup to coalesced standard-library writes, values vs writer() render calls, dynamic attributes with if statements, element/writer rewriting, imports, file naming.
spec: SPEC-002
plan: PLAN-002
state: todo
depends_on: [TASK-001-06]
touches: [internal/generator]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
The [generator](/actuators/generator.md) core of [PLAN-002](/plans/002-code-generation.md). Covers [SPEC-002](/specs/002-code-generation.md) R1, R2, R4–R12, R14, R16. Output imports only `fmt`, `html`, `io` ([DEC-005](/decisions/005-standard-library-only-output.md)).

# Depends on
[TASK-001-06](/tasks/001-template-language/06-file-parser.md)

# Todo
- [ ] Tests first: header/footer splice; static coalescing (one write for `<li><span>`, also across consecutive markup statements); static text verbatim (`&nbsp;` kept); value `{x}` → nil skip + `html.EscapeString(fmt.Sprint(x))` evaluated once; render calls when `writer()` appears anywhere (`{p.C(writer())}`, `{json.NewEncoder(writer()).Encode(v)}`); shadowed `writer` is a value; dynamic attributes true/false/nil/other with if statements; every embedded statement form; `return err` preserved
- [ ] Tests: `element` rewritten to `func(io.Writer) error` in params, fields, var types, generics; not when shadowed; `writer()` outside templates or in attributes is an error
- [ ] Tests: imports reuse existing names, add missing ones, alias with `__golpe` on collision; `GeneratedName` (`x.golpe`→`x_gen.go`, `x_test.golpe`→`x_gen_test.go`, `x_linux_amd64.golpe`→`x_gen_linux_amd64.go`)
- [ ] `Generate(fset, file, src, resolve) ([]byte, error)` with `go/format.Source`; header line and build-constraint copy
- [ ] Error wrapping with positions can be a stub here (`return err`); TASK-002-03 completes it

# Done when
The SPEC-002 `NavItem` example generates the documented output (modulo directives and wrapping).
