---
type: Task Slice
id: TASK-005-01
title: Element name lists
description: Add IsElement (HTML, SVG and MathML element names, ASCII-case-insensitive) and IsCustomElement (names containing '-') to internal/token, with the lists generated from the standards and checked in.
spec: SPEC-005
plan: PLAN-005
state: todo
depends_on: [TASK-001-02]
touches: [internal/token/elements.go, internal/token/elements_test.go, internal/token/gen]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
The classification half of [SPEC-005](/specs/005-component-tags.md) R1, as sketched in [PLAN-005](/plans/005-component-tags.md) ("Classification").

# Depends on
[TASK-001-02](/tasks/001-template-language/02-token-extension.md) ✅. Can run in parallel with TASK-001-04.

# Todo
- [ ] Tests first: every listed name in lower, upper and mixed case; SVG camelCase (`linearGradient`, `foreignObject`); MathML (`math`, `mrow`); non-names (`Card`, `myComponent`, `ui.Button`, empty); `IsCustomElement("my-widget")`; no Unicode folding (Kelvin sign)
- [ ] Element lists from the WHATWG HTML element index (including obsolete-but-parsed ones the HTML parser still knows, e.g. `center`? decide and note), SVG 2 element index and MathML Core; record the source URLs and retrieval date in a comment
- [ ] A small generator under `internal/token/gen` (run with `go generate`) or a hand-maintained sorted table; a test checks the table is sorted and duplicate-free
- [ ] `IsElement(name string) bool` (binary search, ASCII folding, no allocation), `IsCustomElement(name string) bool`
- [ ] Update the package doc

# Done when
`go test ./internal/token` passes with 100% coverage.
