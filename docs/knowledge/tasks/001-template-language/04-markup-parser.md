---
type: Task Slice
id: TASK-001-04
title: Markup parser
description: Parse a markup region from a byte offset: tags, fragments, literal and dynamic attributes, void/raw-text/text-only elements, text, interpolation extents, HTML and Go comments, and the whitespace rule.
spec: SPEC-001
plan: PLAN-001
state: done
depends_on: [TASK-001-02, TASK-001-03]
touches: [internal/parser/markup.go, internal/parser/markup_test.go, internal/parser/errors.go, internal/ast]
generated: { by: claude-code/agent, at: 2026-11-04T10:32:30Z }
---
# Goal
"Pass 2: markup parsing" of [PLAN-001](/plans/001-template-language.md), without embedded statements and without component classification (TASK-005-02). Covers [SPEC-001](/specs/001-template-language.md) R8, R9, R13, R15–R17, R19–R22.

# Depends on
[TASK-001-02](/tasks/001-template-language/02-token-extension.md) ✅, [TASK-001-03](/tasks/001-template-language/03-ast-nodes.md) ✅

# Todo
- [x] AST: add `Fragment{Lt, Children, Close, EndPos}`; rename `Comment` to `HTMLComment` (and its dump kind) so it is not confused with go/ast.Comment
- [x] Tests first (table-driven, expected `ast.Fprint` dump or error `line:col: msg`):
  - [x] nested elements; void (`<br>`, error on `<br>x</br>`); self-closing (`<x/>`); custom (`my-el`); close tags matched ignoring ASCII case (`<DIV></div>`)
  - [x] fragments `<>…</>`, nested and with text
  - [x] attributes: double/single quoted are literal (`x-data="{ open: false }"` keeps braces); boolean; `{expr}`; names `hx-get`, `@click`, `:class`, `viewBox`, `xlink:href`; errors for duplicates and `class=btn`
  - [x] interpolation with nested braces, strings containing `}`, func literals
  - [x] `<script>`/`<style>` raw content with braces and `<`; `<textarea>`/`<title>` reject child tags; `<pre>`/`<textarea>` keep whitespace
  - [x] `<!-- -->` kept as a node; `<!DOCTYPE html>`; Go comments at node boundaries (kept for go/parser, not rendered); `https://x` stays text
  - [x] whitespace (R22): `<p>⏎ Hello,⏎ {name}⏎</p>` → text `Hello, `; `<b>Hi</b>⏎ there` → ` there`; start/end runs dropped
  - [x] errors: `expected </span>`, unterminated tag/attribute/interpolation/comment, stray `}`
- [x] Error list type (`scanner.ErrorList` wrapper) with sorting and de-dup
- [x] `parseMarkup(src []byte, start int) (region, []error)` returning the tree, extent, Go fragments and separator offsets
- [x] Interpolation extent via go/scanner brace counting

# Done when
All tests pass; the SPEC-001 `Button` and `Layout` examples' markup parses.

# Notes
Implemented in [`internal/parser`](../../../../internal/parser/) (`markup.go`, `errors.go`) and [`internal/ast`](../../../../internal/ast/). Coverage 99.0% of the parser; the only uncovered block is the success branch of the `embedded` hook, which TASK-001-05 fills.

**API for TASK-001-06** (all unexported):

- `lineFile(filename string, base int, src []byte) *token.File`: a stand-in for the file go/parser will create (same name, size, base, lines). With `base := fset.Base()` read just before `ParseFile`, node positions are final; assert `f.FileStart == token.Pos(base)`. The tests check this with other files already in the set.
- `parseMarkup(file *token.File, src []byte, start int, errs *errorList) *region` (the checklist had `(src, start) (region, []error)`: positions need the file, and errors go to the file's one list).
- `region{start, end int; root ast.Node; keep []span; semis []int; interps []*ast.Interp}`: mask `src[start:end]` to spaces except newlines and `keep` (interpolation expressions between the braces, Go comments), then write `;` at `start` and at each `semis` offset (interpolation `}`), so `{x}` reads ` x;`. `root` is nil only after an invalid `<!` or an unterminated doctype. `markupMask` in `markup_test.go` is a working reference, checked with go/parser on the SPEC-001 `Button` and `Layout` examples.
- `errorList{add, addGo, err}`: `err` sorts and keeps the first error per line, as go/parser does (it does not use `RemoveMultiples`, which drops errors without position).

**Decisions and deviations**, for the coordinator:

- Duplicate attributes ignore ASCII case, as HTML does (it keeps the first). Whitespace is allowed around `=` and not required between attributes. Names are any bytes but whitespace, controls and `/ > < = " ' { }`; tags start with an ASCII letter. Quoted values may span lines.
- `<textarea>` and `<title>` hold only text and interpolations: `//`, `/*` and line-rule keywords are text there; tags, fragments and HTML comments are errors (still parsed, for recovery). `<pre>` is normal content with whitespace kept, so Go comments and the line rule apply inside it.
- R22: a whitespace-only run that is dropped makes no `Text` node; otherwise `Raw` is the whole source run. Text splits only across a Go comment or a skipped bad tag; a run spanning both sides is rendered once, in the first node.
- R21: a Go comment is also recognised after an HTML comment or another Go comment, and a `/* */` at the start of a line leaves the line start for the line rule (`/* c */ for` is a statement line).
- Recovery: a close tag of an enclosing element reports `expected </x>, found </y>` once and leaves the inner elements unclosed; a stray close tag is reported and skipped; the end of the source gives one error. A `}` as the first non-blank of a content line (no embedded body open) reports `expected </x>, found '}'` and ends the region, so the Go block's `}` stays Go.
- `<script/>` is accepted (`<x/>` is fine for any element); the generator must write `<script …></script>`. Raw text ends at `</script` followed by `>`, `/`, whitespace or the end of the source.
- TASK-001-05 must edit `markup.go` (the `embedded` hook, the `}` line branch of `content`, frames for bodies) and the line-rule test in `markup_test.go`; its `touches` lists only `markup_stmt.go`.
- PLAN-001 "As built: internal/ast" still shows `Comment` and `Root` without `Fragment`/`HTMLComment`.
