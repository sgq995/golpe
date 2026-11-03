---
type: Playbook
id: PROC-FEATURE
title: Feature workflow
description: How an agent turns a feature described in chat into actors, actuators, entities, a spec, a plan and task slices.
tags: [process, specs, agents]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Trigger

The maintainer ([ACTOR-maintainer](/actors/maintainer.md)) describes a
feature in chat, in their own words. They do not write specs; an
[AI agent](/actors/ai-agent.md) does. The `feature-intake` skill
(`.claude/skills/feature-intake/SKILL.md`) points here.

# Steps

1. **Orient, don't bulk-read.** Read [/index.md](/index.md), then the
   `index.md` of actors, actuators, entities and specs. Open only the
   concepts the feature touches.
2. **Clarify only what blocks you.** If the description leaves a choice that
   changes the design and has no sensible default, ask. Otherwise pick the
   default and record it under *Open questions* in the spec.
3. **Model the domain (update over create).** For every person or external
   system involved, link an existing [actor](/actors/) or add one. Same for
   Golpe components ([actuators](/actuators/)) and data
   ([entities](/entities/)). Extend an existing concept rather than creating a
   near-duplicate.
4. **Write the spec** (`specs/NNN-slug.md`, `status: draft`): *what* and
   *why* only, no implementation. Numbered requirements, concrete examples
   (use the maintainer's own examples), acceptance criteria that a test can
   check, out of scope, open questions. Link every actor, actuator and entity
   it involves.
5. **Write the plan** (`plans/NNN-slug.md`, same number): *how*.
   Architecture, the packages it touches, short code snippets of the key
   types and functions, the test strategy (unit, golden, fuzz), and
   **laboratories**: small experiments that validate risky assumptions
   before slices depend on them. Run labs when you can and record the result;
   executable labs live in `internal/labs/`.
6. **Slice the plan** into `tasks/NNN-slug/MM-slug.md`. One slice is one
   reviewable change (roughly one PR, ideally under a day of agent work),
   with `depends_on` and `touches`, and a checklist that starts with tests.
   Slices that can run in parallel must have disjoint `touches`.
7. **Record decisions.** Any non-obvious choice with real alternatives gets a
   `decisions/NNN-slug.md` (`decision: proposed` until the maintainer agrees).
8. **Index and log.** Add every new concept to its folder's `index.md`, add a
   dated entry to [/log.md](/log.md).
9. **Validate.** `go run ./internal/okflint` must pass (`go test ./...` too).
10. **Commit** following the [git workflow](/process/git-workflow.md):
    `docs(knowledge): spec NNN <slug>`.
11. **Report** to the maintainer: the spec's requirements in one paragraph,
    the open questions, and the slices. The maintainer approves by saying so;
    the agent then sets `status: stable` and adds a `verified` entry with
    `human:sgq995`.

# Changing an existing feature

Edit the existing spec and plan instead of writing new ones; note what
changed in the log. If a requirement is dropped, keep it struck through with
a reason rather than deleting it. Task slices already `done` are never
rewritten: add new slices.
