---
type: Task Slice
id: TASK-001-04
title: Markup parser
description: Parse a markup region from a byte offset: tags, attributes, void/self-closing/raw-text elements, text, interpolation extents, HTML and Go comments, whitespace rules.
spec: SPEC-001
plan: PLAN-001
state: todo
depends_on: [TASK-001-02, TASK-001-03]
touches: [internal/parser/markup.go, internal/parser/markup_test.go, internal/parser/errors.go]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Goal
"Pass 2: markup parsing" of [PLAN-001](/plans/001-template-language.md), without embedded statements. Covers [SPEC-001](/specs/001-template-language.md) R7, R10–R14, R16–R18.

# Depends on
[TASK-001-02](/tasks/001-template-language/02-token-extension.md), [TASK-001-03](/tasks/001-template-language/03-ast-nodes.md)

# Todo
- [ ] Tests first (table-driven, each with expected tree dump or error `line:col: msg`):
  - [ ] nested elements, void (`<br>`), self-closing (`<x/>`), custom (`my-el`), namespaced (`svg:path`)
  - [ ] attributes: double/single quoted, boolean, `{expr}`, names like `hx-get`, `@click`, `:class`
  - [ ] interpolation with nested braces, strings containing `}`, func literals
  - [ ] `<script>`/`<style>` raw content with braces and `<`
  - [ ] `<!-- -->`, `<!DOCTYPE html>`, Go comments at node boundaries, `https://x` stays text
  - [ ] whitespace normalisation (R18)
  - [ ] errors: `expected </span>`, unterminated tag/attribute/interpolation/comment, stray `}`
- [ ] Error list type (`scanner.ErrorList` wrapper) with sorting and de-dup
- [ ] `parseMarkup(src []byte, start int) (region, []error)` returning the tree, extent, Go fragments and separator offsets
- [ ] Interpolation extent via go/scanner brace counting

# Done when
All tests pass; the SPEC-001 `Button` example parses.
