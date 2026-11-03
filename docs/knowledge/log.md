# Knowledge base update log

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
