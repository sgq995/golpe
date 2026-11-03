---
type: Playbook
id: PROC-CONVENTIONS
title: Knowledge base conventions
description: Concept types, required frontmatter, IDs, file naming and link style for the Golpe OKF bundle.
tags: [process, okf]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Format

The bundle follows OKF v0.2: one concept per markdown file, YAML frontmatter
with a non-empty `type`, `index.md` (no frontmatter, except `okf_version` at
the root) and `log.md` reserved at every level. On top of OKF this project
adds the rules below. `go run ./internal/okflint` checks them, and
`go test ./...` runs the same check against this bundle.

# Concept types

| `type`       | Folder        | Holds                                                        | Extra required keys            |
|--------------|---------------|--------------------------------------------------------------|--------------------------------|
| `Actor`      | `actors/`     | A person or external system that interacts with Golpe        | `id`                           |
| `Actuator`   | `actuators/`  | A Golpe component that performs work (CLI, parser, …)        | `id`                           |
| `Entity`     | `entities/`   | A thing Golpe reads, produces or reasons about               | `id`                           |
| `Spec`       | `specs/`      | **What** and **why**: problem, requirements, acceptance     | `id`, `status`                 |
| `Plan`       | `plans/`      | **How**: architecture, snippets, tests, laboratories, risks  | `id`, `status`, `spec`         |
| `Task Slice` | `tasks/<n>/`  | One shippable slice of a plan, as a todo list                | `id`, `spec`, `plan`, `state` |
| `Decision`   | `decisions/`  | An architecture decision record                              | `id`, `status`, `decision`     |
| `Playbook`   | `process/`    | How we work                                                  | `id`                           |
| `Backlog`    | `backlog.md`  | Deferred and rejected ideas, with the decision that parked them | `id`                        |

Every concept also needs `title` and a one-sentence `description` (it is what
index entries and search show).

# Field values

- `status` is OKF lifecycle: `draft` (written, not yet approved),
  `stable` (approved / current), `deprecated` (kept for history).
- `state` (task slices only): `todo`, `doing`, `done`, `blocked`.
- `decision` (decisions only): `proposed`, `accepted`, `superseded`, `rejected`.
- `generated: { by, at }` on every concept an agent writes or meaningfully
  changes. Agents use `by: claude-code/agent`; people use `human:<github-login>`.
- `verified: [{ by: human:sgq995, at: … }]` is added when the maintainer
  approves a spec, plan or decision. A spec moves to `status: stable` only with
  a human `verified` entry (OKF trust tier *human-reviewed*).
- Task slices also carry `depends_on: [TASK-…]` and `touches: [paths]`.
  Slices with no unmet dependency and disjoint `touches` may run in parallel
  (see [git workflow](/process/git-workflow.md)).

# IDs and file names

| Kind     | ID pattern                  | File                                       |
|----------|-----------------------------|--------------------------------------------|
| Spec     | `SPEC-001`                  | `specs/001-short-slug.md`                  |
| Plan     | `PLAN-001` (same number)    | `plans/001-short-slug.md`                  |
| Task     | `TASK-001-02`               | `tasks/001-short-slug/02-short-slug.md`    |
| Decision | `DEC-001`                   | `decisions/001-short-slug.md`              |
| Actor    | `ACTOR-<slug>`              | `actors/<slug>.md`                         |
| Actuator | `ACT-<slug>`                | `actuators/<slug>.md`                      |
| Entity   | `ENT-<slug>`                | `entities/<slug>.md`                       |

Numbers are never reused. A spec and its plan share a number; one spec has
exactly one plan.

# Links

- Concepts relate through **markdown links in the body**, using
  bundle-absolute paths (`/entities/element.md`), as OKF §6.1 recommends.
  The kind of relationship is stated by the surrounding prose or the section
  heading ("Actors", "Entities", "Depends on").
- Frontmatter `spec`, `plan` and `depends_on` repeat the most important
  links as IDs so tools can filter without reading bodies.
- Links to source code use repo-relative paths that leave the bundle
  (`../../internal/parser/`); the linter ignores those.
- Every concept must be listed in its folder's `index.md` with its
  description. Every change is recorded in [the log](/log.md), newest first.

# Body sections

- **Spec**: Summary · Motivation (why) · Actors · Entities · Requirements
  (numbered `R1`…, MUST/SHOULD/MAY) · Examples · Acceptance criteria ·
  Out of scope · Open questions.
- **Plan**: Approach · Architecture · Snippets · Testing · Laboratories ·
  Risks · Slices (links to task files).
- **Task Slice**: Goal · Depends on · Todo (`- [ ]` checklist, tests first) ·
  Done when · Notes.
- **Decision**: Context · Options · Decision · Consequences.
- **Actor / Actuator / Entity**: what it is, its responsibilities or shape,
  and which specs involve it.
