---
type: Playbook
id: PROC-IMPLEMENT
title: Implementation workflow
description: How an agent picks a task slice, implements it test-first and closes it out in the knowledge base.
tags: [process, tasks, agents]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Picking work

Open [/tasks/](/tasks/) and pick a slice whose `state` is `todo` and whose
`depends_on` slices are all `done`. Read its spec and plan sections (linked
from the slice), not the whole bundle. The `implement-task` skill
(`.claude/skills/implement-task/SKILL.md`) points here.

# Doing it

1. Set `state: doing` in the slice and commit that with the first code
   change (not as a separate commit).
2. Write the tests listed in the slice first; watch them fail.
3. Implement until they pass. Keep to the slice's `touches`; if you must
   touch more, say so in the slice's Notes.
4. Run the checks: `gofmt -l .` prints nothing, `go vet ./...` and
   `go test ./...` pass (the latter also lints this bundle).
5. Tick the checklist items you finished. If the plan turned out wrong,
   fix the plan (and spec, if behaviour changed) in the same change.
6. Set `state: done`, update `generated.at`, add a log entry (in a parallel
   run, leave the log, indexes and other shared docs to the coordinator; see
   the [git workflow](/process/git-workflow.md)).
7. Commit and push per the [git workflow](/process/git-workflow.md):
   `feat(parser): TASK-001-04 markup parser`.

# Blocked

Set `state: blocked`, write why under Notes, and tell the maintainer.
