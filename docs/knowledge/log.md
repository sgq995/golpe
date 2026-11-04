# Knowledge base update log

## 2026-11-04 (later)
* **Update**: [TASK-001-05 Embedded statements](/tasks/001-template-language/05-embedded-statements.md)
  is done: the line rule in the markup parser, with header extents decided by go/parser.
* **Update**: [PLAN-001](/plans/001-template-language.md): corrected the header-extent rule
  (the first `{` at depth 0 is not Go's rule) and added "As built: embedded statements"
  (frames for bodies, extra keep spans and semis, `region.stmts` for the graft).
* **Update**: [SPEC-001](/specs/001-template-language.md) R18 (content before the first clause
  and a line starting with `else` are errors; content may follow a statement's last `}`) and
  R20 (in `<pre>`, the lines of embedded statements add no whitespace).
* **Update**: TASK-001-06 and TASK-005-02 are ready; they touch different files.

## 2026-11-04
* **Update**: Merged the second parallel run:
  [TASK-001-04 Markup parser](/tasks/001-template-language/04-markup-parser.md),
  [TASK-005-01 Element name lists](/tasks/005-component-tags/01-element-names.md) and
  [TASK-006-01 Class and If](/tasks/006-optional-helpers/01-class-and-if.md).
* **Update**: Folded the agents' notes into [SPEC-001](/specs/001-template-language.md) (R15, R20, R23),
  [SPEC-002](/specs/002-code-generation.md) (R8: self-closing output),
  [SPEC-005](/specs/005-component-tags.md) (R1: the element lists),
  [SPEC-006](/specs/006-optional-helpers.md) (R3; criteria 1–2 met),
  [PLAN-001](/plans/001-template-language.md) (as built: markup parsing, LAB-014) and
  [PLAN-005](/plans/005-component-tags.md) (hand-maintained table).
* **Update**: TASK-001-05 is ready; TASK-005-02 follows it because both edit `markup.go`.

## 2026-11-03 (later)
* **Update**: Merged the first parallel run: [TASK-001-02](/tasks/001-template-language/02-token-extension.md),
  [TASK-001-03](/tasks/001-template-language/03-ast-nodes.md) and
  [TASK-002-01](/tasks/002-code-generation/01-runtime-writers.md); folded their notes into
  [PLAN-001](/plans/001-template-language.md) (walker adapter, LAB-011, AST shape, token API).
* **Update**: Design interview with the maintainer (11 rounds). Scope is server-side templates
  only. Outcomes: [DEC-001](/decisions/001-mask-and-reuse-go-parser.md),
  [DEC-002](/decisions/002-element-is-an-alias.md) and
  [DEC-003](/decisions/003-template-is-a-contextual-keyword.md) accepted;
  [DEC-004](/decisions/004-runtime-dispatch-for-interpolation.md) superseded by
  [DEC-005](/decisions/005-standard-library-only-output.md) (standard-library-only output);
  new [DEC-006](/decisions/006-writer-and-render-calls.md) (writer() and render calls),
  [DEC-007](/decisions/007-component-tags.md) (component tags),
  [DEC-008](/decisions/008-no-implicit-context.md) (no implicit context) and
  [DEC-009](/decisions/009-escape-by-default.md) (escape by default).
* **Update**: Rewrote [SPEC-001](/specs/001-template-language.md) (line rule, fragments, writer(),
  literal quoted attributes, boolean attributes, comments, whitespace),
  [SPEC-002](/specs/002-code-generation.md) (`_gen.go`, standard library only, wrapped errors) and
  [SPEC-003](/specs/003-cli.md) (`cmd/golpe`, go tool); updated [SPEC-004](/specs/004-language-server.md).
* **Creation**: [SPEC-005 Component tags](/specs/005-component-tags.md) and
  [SPEC-006 Optional helpers](/specs/006-optional-helpers.md) with plans and slices; the
  [backlog](/backlog.md) of deferred and rejected ideas; entities
  [writer()](/entities/writer-function.md) and [component tag](/entities/component-tag.md).
* **Update**: Code: `cmd/cli` renamed to `cmd/golpe`; the root runtime writers removed (DEC-005);
  MIT license added.

## 2026-11-03
* **Initialization**: Created the bundle, [conventions](/process/conventions.md) and the
  [feature](/process/feature-workflow.md), [implementation](/process/implementation-workflow.md)
  and [git](/process/git-workflow.md) workflows.
* **Creation**: Captured the first feature from chat as
  [SPEC-001 Template language](/specs/001-template-language.md),
  [SPEC-002 Code generation](/specs/002-code-generation.md),
  [SPEC-003 CLI](/specs/003-cli.md) and the draft
  [SPEC-004 Language server](/specs/004-language-server.md), with plans, task slices and
  decisions.
* **Creation**: Ran parser laboratories LAB-001..006 (`internal/labs/parserlab`), which
  shaped [DEC-001](/decisions/001-mask-and-reuse-go-parser.md).
