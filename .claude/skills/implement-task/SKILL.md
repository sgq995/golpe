---
name: implement-task
description: Implement one or more Golpe task slices from docs/knowledge/tasks test-first, keep the knowledge base in sync, and commit per the git workflow (main for serial work, feature branches in git worktrees for parallel agents). Use when the maintainer asks to build, continue, or parallelise work on Golpe.
---

# Implement a task slice

1. Read `docs/knowledge/process/implementation-workflow.md` and
   `docs/knowledge/process/git-workflow.md`.
2. Pick the slice (or the one the maintainer named) from
   `docs/knowledge/tasks/index.md`: `state: todo`, all `depends_on` done.
   Read its spec and plan sections only.
3. **Several ready slices with disjoint `touches`, and the maintainer wants
   speed?** Coordinate: one worker agent per slice, each on
   `feat/<TASK-ID>-<slug>` in its own git worktree (Agent tool with
   `isolation: "worktree"`, or `git worktree add .worktrees/<TASK-ID> …`).
   Merge finished branches into `main` one at a time after `go test ./...`.
   Otherwise work serially on `main`.
4. Tests first, then code. Stay within `touches`.
5. Before committing: `gofmt -l .` empty, `go vet ./...`, `go test ./...`.
6. Update the slice (`state`, ticked checklist, `generated.at`), fix the plan
   or spec if reality differed, add a `docs/knowledge/log.md` entry.
7. Commit (`feat(<pkg>): <TASK-ID> <title>`) and push.
