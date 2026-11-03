---
type: Task Slice
id: TASK-002-02
title: Lowering templates and markup
description: Implement internal/generator: splice template headers, lower MarkupStmt trees to coalesced escaped writes, lower embedded statements, rewrite element.
spec: SPEC-002
plan: PLAN-002
state: todo
depends_on: [TASK-001-06, TASK-002-01]
touches: [internal/generator]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
The [generator](/actuators/generator.md) core of [PLAN-002](/plans/002-code-generation.md) ("Lowering markup", "element rewriting"). Covers [SPEC-002](/specs/002-code-generation.md) R1–R6, R8, R9, R11.

# Depends on
[TASK-001-06](/tasks/001-template-language/06-file-parser.md), [TASK-002-01](/tasks/002-code-generation/01-runtime-writers.md)

# Todo
- [ ] Tests first: header/footer splice; static coalescing (one write for `<li><span>`); text escaped at generation time; interpolation → `__golpe.Write`; attribute → `__golpe.WriteAttr`; every embedded statement form; `return err` preserved
- [ ] Tests: `element` rewritten in params, fields, var types, generics; not rewritten when shadowed or declared in the package
- [ ] `Generate(fset, file, src) ([]byte, error)` with `go/format.Source`
- [ ] Generated header and build-constraint copy (R1)

# Done when
The SPEC-002 `NavItem` example generates the documented output (modulo directives).
