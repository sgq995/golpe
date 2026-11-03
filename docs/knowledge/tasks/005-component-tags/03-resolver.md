---
type: Task Slice
id: TASK-005-03
title: Declaration resolver
description: Implement internal/resolve: find a component's template or func declaration in the package or an imported package, syntactically, and return its parameter count and props type.
spec: SPEC-005
plan: PLAN-005
state: todo
depends_on: [TASK-001-06]
touches: [internal/resolve]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
[SPEC-005](/specs/005-component-tags.md) R2, R3, R7 as in [PLAN-005](/plans/005-component-tags.md) ("Resolution").

# Depends on
[TASK-001-06](/tasks/001-template-language/06-file-parser.md) (to read `.golpe` declarations).

# Todo
- [ ] Tests first, on a testdata module: same-package template; `func` in a `.go` file; `*Props` parameter; zero-parameter template; template in another package imported under its name and under an alias; errors for not found, generic, two parameters, a func with the wrong result; `_gen.go` files ignored
- [ ] `Resolver.Lookup` with a per-directory cache invalidated by mtime
- [ ] LAB-013: `go/build.Import(…, FindOnly)` latency; record it in PLAN-005

# Done when
All resolver tests pass.
