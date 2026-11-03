---
type: Task Slice
id: TASK-001-04
title: Markup parser
description: Parse a markup region from a byte offset: tags, fragments, literal and dynamic attributes, void/raw-text/text-only elements, text, interpolation extents, HTML and Go comments, and the whitespace rule.
spec: SPEC-001
plan: PLAN-001
state: todo
depends_on: [TASK-001-02, TASK-001-03]
touches: [internal/parser/markup.go, internal/parser/markup_test.go, internal/parser/errors.go, internal/ast]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Goal
"Pass 2: markup parsing" of [PLAN-001](/plans/001-template-language.md), without embedded statements and without component classification (TASK-005-02). Covers [SPEC-001](/specs/001-template-language.md) R8, R9, R13, R15–R17, R19–R22.

# Depends on
[TASK-001-02](/tasks/001-template-language/02-token-extension.md) ✅, [TASK-001-03](/tasks/001-template-language/03-ast-nodes.md) ✅

# Todo
- [ ] AST: add `Fragment{Lt, Children, Close, EndPos}`; rename `Comment` to `HTMLComment` (and its dump kind) so it is not confused with go/ast.Comment
- [ ] Tests first (table-driven, expected `ast.Fprint` dump or error `line:col: msg`):
  - [ ] nested elements; void (`<br>`, error on `<br>x</br>`); self-closing (`<x/>`); custom (`my-el`); close tags matched ignoring ASCII case (`<DIV></div>`)
  - [ ] fragments `<>…</>`, nested and with text
  - [ ] attributes: double/single quoted are literal (`x-data="{ open: false }"` keeps braces); boolean; `{expr}`; names `hx-get`, `@click`, `:class`, `viewBox`, `xlink:href`; errors for duplicates and `class=btn`
  - [ ] interpolation with nested braces, strings containing `}`, func literals
  - [ ] `<script>`/`<style>` raw content with braces and `<`; `<textarea>`/`<title>` reject child tags; `<pre>`/`<textarea>` keep whitespace
  - [ ] `<!-- -->` kept as a node; `<!DOCTYPE html>`; Go comments at node boundaries (kept for go/parser, not rendered); `https://x` stays text
  - [ ] whitespace (R22): `<p>⏎ Hello,⏎ {name}⏎</p>` → text `Hello, `; `<b>Hi</b>⏎ there` → ` there`; start/end runs dropped
  - [ ] errors: `expected </span>`, unterminated tag/attribute/interpolation/comment, stray `}`
- [ ] Error list type (`scanner.ErrorList` wrapper) with sorting and de-dup
- [ ] `parseMarkup(src []byte, start int) (region, []error)` returning the tree, extent, Go fragments and separator offsets
- [ ] Interpolation extent via go/scanner brace counting

# Done when
All tests pass; the SPEC-001 `Button` and `Layout` examples' markup parses.
