---
type: Playbook
id: PROC-GIT
title: Git workflow
description: Serial work is committed straight to main; parallel agents each work on a feature branch in their own git worktree and merge back to main.
tags: [process, git, agents]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
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
   [implementation workflow](/process/implementation-workflow.md), commits on
   its branch and pushes it: `git push -u origin feat/TASK-001-02-token-extension`.
3. The coordinator merges finished branches into `main` one at a time:
   rebase the branch on the latest `main`, run `go test ./...`, then
   fast-forward `main` and push. Knowledge-base conflicts (index.md,
   log.md) are resolved by keeping both entries.
4. Clean up: `git worktree remove .worktrees/TASK-001-02`,
   `git branch -d feat/TASK-001-02-token-extension`, and delete the remote
   branch.

# Never

- Force-push `main`.
- Merge a branch whose `go test ./...` fails.
- Have two agents edit the same files at the same time.

# Note for cloud sessions

Claude Code cloud sessions may be assigned a session branch
(`claude/…`). Agents commit there as well and then fast-forward `main`
to it, so `main` stays the source of truth.
