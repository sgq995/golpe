---
type: Task Slice
id: TASK-001-02
title: Token extension
description: Implement internal/token: Go's tokens plus TEMPLATE and markup tokens, Lookup, IsVoid and IsRawText.
spec: SPEC-001
plan: PLAN-001
state: done
depends_on: [TASK-001-01]
touches: [internal/token]
generated: { by: claude-code/agent, at: 2026-11-03T01:32:57Z }
---
# Goal
Implement [Golpe tokens](/entities/golpe-token.md) as sketched in [PLAN-001](/plans/001-template-language.md) ("internal/token: extending go/token"). Serves [SPEC-001](/specs/001-template-language.md) R3, R10, R16.

# Depends on
[TASK-001-01](/tasks/001-template-language/01-project-scaffold.md)

# Todo
- [x] Tests first: `String()` for every Go token equals go/token's; Golpe tokens have names; `FromGo`/`Go` round-trip
- [x] Tests: `IsVoid` for the 13 void elements (case-insensitive), `IsRawText` for script/style
- [x] `Token`, `golpeBase`, constants, `String`, `FromGo`, `Go`
- [x] `IsVoid`, `IsRawText`, `IsEmbeddedKeyword` (if/for/switch/select)
- [x] Package doc and examples

# Done when
`go test ./internal/token` passes with 100% statement coverage.

# Notes
Where the code ([internal/token](../../../../internal/token/)) refines the
plan sketch, for the coordinator to fold into PLAN-001 and the entity:

- **Contextual keyword (DEC-003).** `Lookup(ident) Token` is exactly
  `go/token.Lookup`, so `Lookup("template")` is IDENT. The contextual check
  is a separate exported helper the parser calls at the first token of a
  top-level declaration: `IsTemplateKeyword(lit string, next Token) bool`,
  true when `lit == "template"` and `next` is IDENT or LPAREN. Checking the
  "first token of a top-level declaration" part stays with the caller. The
  plan's Architecture table lists only `Lookup`, `IsVoid`, `IsRawText`; it
  should also name `IsTemplateKeyword` and `IsEmbeddedKeyword`.
- **Golpe token names.** `String()` follows go/token's documented convention:
  keyword and delimiters print as their spelling (`template`, `>`, `/>`, `{`,
  `}`), token classes as their constant name (`TAG_OPEN`, `TAG_CLOSE`,
  `TEXT`, `HTML_COMMENT`, `DOCTYPE`). So `INTERP_OPEN` and `LBRACE` both
  print `{`; the values differ.
- **Values outside both ranges.** The sketch's `golpeNames[t-golpeBase]`
  panics past the last Golpe token. Such values now print `token(N)` like
  go/token's unnamed values, and `Go()` returns `ILLEGAL, false`. Negative
  values are below the base, so they count as Go values (`Go()` is a no-op,
  `String()` is go/token's `token(-1)`). An unexported `golpeEnd` sentinel
  bounds the range.
- **Layout guard.** go/token (Go 1.24) names values up to 88 (`~`). A test
  fails if go/token ever names a value at or above `golpeBase` (1024).
- **Case folding.** `IsVoid` and `IsRawText` ignore ASCII case only, as HTML
  does for tag names. `strings.ToLower`/`EqualFold` would also fold U+212A
  KELVIN SIGN to `k` (making `linK` void) and U+017F LONG S to `s`.
- `IsEmbeddedKeyword(word string) bool` is case-sensitive and covers only
  the statement starters `if`, `for`, `switch`, `select` (not `else`,
  `case`, `default`).
