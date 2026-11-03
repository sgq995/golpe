---
type: Playbook
id: PROC-GIT
title: Git workflow
description: Serial work is committed straight to main; parallel agents each work on a feature branch in their own git worktree and merge back to main.
tags: [process, git, agents]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Serial work (one agent at a time)

Commit directly on `main` and push. Small, focused commits with
conventional-commit subjects:

| Subject prefix         | Used for                                  |
|------------------------|-------------------------------------------|
| `docs(knowledge): …`   | specs, plans, tasks, decisions, actors…   |
| `feat(<pkg>): …`       | new behaviour, usually one task slice     |
| `fix(<pkg>): …`        | bug fixes                                 |
| `test(<pkg>): …`       | tests or labs only                        |
| `chore: …`             | tooling, CI, repo housekeeping            |

Reference the task ID in the subject or body (`TASK-001-04`).

# Parallel work (several agents at once)

Use it only for slices that are ready (dependencies done) and whose
`touches` do not overlap.

1. The coordinating agent picks the slices and, for each, creates a branch
   and a worktree from an up-to-date `main`:

   ```sh
   git fetch origin main
   git worktree add -b feat/TASK-001-02-token-extension \
       .worktrees/TASK-001-02 origin/main
   ```

   `.worktrees/` is git-ignored. (Claude Code's Agent tool with
   `isolation: "worktree"` does the same thing automatically.)
2. Each worker agent works only inside its worktree, follows the
   [implementation workflow](/process/implementation-workflow.md) and commits
   on its branch. Workers that share the clone with the coordinator do not
   push; workers in separate sessions push their branch.
3. **The coordinator owns shared docs during a parallel run:** workers edit
   only their own slice file (and any plan or doc the brief hands them
   explicitly), never `log.md`, `index.md` files, specs, decisions or other
   plans. They report deviations in the slice's Notes and in their final
   report; the coordinator folds them into the shared docs after merging.
4. The coordinator reviews each finished branch, then merges them into `main`
   one at a time: rebase the branch on the latest `main` (inside its
   worktree), run `gofmt`, `go vet ./...` and `go test ./...`, fast-forward
   `main`, and push `main` (plus the session branch in cloud sessions).
5. Clean up: `git worktree remove .worktrees/TASK-001-02` and
   `git branch -d feat/TASK-001-02-token-extension` (and the remote branch if
   one was pushed).

# Never

- Force-push `main`.
- Merge a branch whose `go test ./...` fails.
- Have two agents edit the same files at the same time.

# Note for cloud sessions

Claude Code cloud sessions may be assigned a session branch
(`claude/…`). Agents commit there as well and then fast-forward `main`
to it, so `main` stays the source of truth.
