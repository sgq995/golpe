# Golpe — agent guide (AGENTS.md mirrors CLAUDE.md; keep them identical)

Golpe is a meta-language over Go: `.golpe` files are Go plus `template`
declarations whose bodies switch into balanced HTML markup (with embedded
`if`/`for`/`switch`/`select` and `{expr}`), compiled to plain Go functions
returning `func(io.Writer) error`.

## The knowledge base is the source of truth

`docs/knowledge/` is an [OKF v0.2](https://okf.md) bundle with every spec,
plan, task, decision, actor, actuator and entity. **Read
`docs/knowledge/index.md` first and descend; never bulk-read the tree.**

| When the maintainer…                          | Do this                                                        |
|-----------------------------------------------|----------------------------------------------------------------|
| describes a feature in chat                   | `/feature-intake` → `docs/knowledge/process/feature-workflow.md` |
| asks to implement / continue work             | `/implement-task` → `docs/knowledge/process/implementation-workflow.md` |
| asks for parallel work                        | `docs/knowledge/process/git-workflow.md` (worktrees)           |

Conventions (types, frontmatter, IDs, links): `docs/knowledge/process/conventions.md`.

## Git

- Serial work: commit straight to `main` and push.
- Parallel agents: one feature branch + git worktree per task slice under
  `.worktrees/`, merged back to `main` by the coordinator.
- Conventional commit subjects (`feat(parser): TASK-001-04 …`,
  `docs(knowledge): …`). Details in the git workflow playbook.

## Layout

| Path              | What                                                  |
|-------------------|-------------------------------------------------------|
| `/` (package `golpe`) | public API and runtime used by generated code     |
| `cmd/cli`         | the `golpe` CLI (`go build -o golpe ./cmd/cli`)       |
| `internal/token`  | go/token extended with Golpe tokens                   |
| `internal/ast`    | go/ast extended with template and markup nodes        |
| `internal/parser` | `.golpe` parser (masks markup, reuses go/parser)      |
| `internal/generator` | lowering to Go (planned)                           |
| `internal/okflint`| knowledge-base linter                                 |
| `internal/labs`   | executable laboratories backing plan assumptions      |
| `lsp`             | language server proxying gopls (planned)              |

## Checks before every commit

```sh
gofmt -l .          # must print nothing
go vet ./...
go test ./...       # includes okflint over docs/knowledge
```

Standard library only unless a plan says otherwise.
