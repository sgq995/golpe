---
type: Task Slice
id: TASK-005-01
title: Element name lists
description: Add IsElement (HTML, SVG and MathML element names, ASCII-case-insensitive) and IsCustomElement (names containing '-') to internal/token, from a hand-maintained sorted table of the current standards' element names.
spec: SPEC-005
plan: PLAN-005
state: done
depends_on: [TASK-001-02]
touches: [internal/token/elements.go, internal/token/elements_test.go, internal/token/example_test.go, internal/token/token.go]
generated: { by: claude-code/agent, at: 2026-11-04T10:11:15Z }
---
# Goal
The classification half of [SPEC-005](/specs/005-component-tags.md) R1, as sketched in [PLAN-005](/plans/005-component-tags.md) ("Classification").

# Depends on
[TASK-001-02](/tasks/001-template-language/02-token-extension.md) ✅. Can run in parallel with TASK-001-04.

# Todo
- [x] Tests first: every listed name in lower, upper and mixed case; SVG camelCase (`linearGradient`, `foreignObject`); MathML (`math`, `mrow`); non-names (`Card`, `myComponent`, `ui.Button`, empty); `IsCustomElement("my-widget")`; no Unicode folding (Kelvin sign)
- [x] Element lists from the WHATWG HTML element index (including obsolete-but-parsed ones the HTML parser still knows, e.g. `center`? decide and note), SVG 2 element index and MathML Core; record the source URLs and retrieval date in a comment
- [x] A small generator under `internal/token/gen` (run with `go generate`) or a hand-maintained sorted table; a test checks the table is sorted and duplicate-free
- [x] `IsElement(name string) bool` (binary search, ASCII folding, no allocation), `IsCustomElement(name string) bool`
- [x] Update the package doc

# Done when
`go test ./internal/token` passes with 100% coverage.

# Notes
Where the code ([internal/token](../../../../internal/token/)) differs from
the plan sketch, for the coordinator to fold into PLAN-005 (and to raise with
the maintainer where marked):

- **Hand-maintained table, no generator.** The run brief chose this so builds
  and CI stay offline and deterministic. PLAN-005's Classification snippet
  ("generated … (go:generate, checked in)") and its Risks line ("regenerate
  them per release") should say: one hand-maintained table in
  `elements.go`, sorted by ASCII-lower-cased name, refreshed per release by
  re-reading the three sources. The lists as published are kept in
  `elements_test.go`. `TestElementSources` checks that the table holds
  exactly their names, spelled as published (SVG camelCase kept), and
  `TestElementsTable` checks the order, duplicates and name syntax. No
  `internal/token/gen` exists; `touches` now names the files changed.
- **Sources, retrieved 2026-11-04.**
  - HTML: 113 names from the WHATWG "Elements" index. Its "MathML math",
    "SVG svg" and "autonomous custom elements" rows were dropped: the first
    two come from the lists below, and the third is `IsCustomElement`'s job.
  - SVG: 63 names from the SVG 2 element index (Candidate Recommendation
    Snapshot of 6 October 2026).
  - MathML: 31 MathML Core elements (Editor's Draft of 1 October 2026),
    which includes `a`.
  - `a`, `script`, `style` and `title` are shared, so the table has 202 names.
- **Excluded: maintainer not yet asked** (current standards only):
  - The 29 obsolete, non-conforming HTML elements: acronym, applet, basefont,
    bgsound, big, blink, center, dir, font, frame, frameset, isindex, keygen,
    listing, marquee, menuitem, multicol, nextid, nobr, noembed, noframes,
    param, plaintext, rb, rtc, spacer, strike, tt, xmp. Browsers still parse
    these as elements, but here `<Center/>`, `<Font/>`, `<Big/>` and
    `<Frame/>` are component tags.
  - The 18 SVG 1.1 elements that SVG 2 dropped: altGlyph, altGlyphDef,
    altGlyphItem, animateColor, color-profile, cursor, font, font-face,
    font-face-format, font-face-name, font-face-src, font-face-uri, glyph,
    glyphRef, hkern, missing-glyph, tref, vkern. The 7 names with a dash are
    still markup because they count as custom elements.
  - MathML elements outside MathML Core, for example none, mfenced,
    menclose, mglyph, mlabeledtr, mstack and Content MathML (apply, ci, cn).
- **Custom elements.** `IsCustomElement` is `strings.Contains(name, "-")`.
  The HTML standard's stricter rules for a valid custom element name are not
  applied, and the doc comment says so: the name may start with any
  character, ASCII upper case is allowed, and the eight reserved names such
  as `annotation-xml` and `font-face` are accepted.
- **No allocation.** `IsElement` binary-searches with `compareLowerASCII`,
  which folds A-Z byte by byte as it compares. `TestIsElementDoesNotAllocate`
  (`testing.AllocsPerRun`) and `BenchmarkIsElement` (`b.ReportAllocs`) both
  report 0 allocations. As with `IsVoid`, only ASCII letters fold.
- **SPEC-005.** The test confirms that all 21 names in "Known consequence of
  R1" are elements. The classification half of the first acceptance
  criterion holds: `div`, `DIV`, `linearGradient` and `math` are elements,
  `my-widget` is a custom element, and `Card`, `card`, `ui.Button` and
  `svg:path` are neither. The parser (TASK-005-02) still has to turn that
  into components and errors.
