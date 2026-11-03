---
name: feature-intake
description: Turn a feature the maintainer describes in chat into OKF knowledge — actors, actuators, entities, a spec (what/why), a plan (how, with snippets, tests and labs), task slices (todo lists) and decisions — under docs/knowledge. Use whenever the maintainer describes, changes or extends a Golpe feature, even casually, and before writing any code for a new feature.
---

# Feature intake

1. Read `docs/knowledge/process/feature-workflow.md` and
   `docs/knowledge/process/conventions.md` in full and follow them exactly.
2. Orient through `docs/knowledge/index.md` and the folder `index.md` files;
   open only the concepts the feature touches. Update existing concepts
   rather than creating near-duplicates.
3. Use the maintainer's own wording and examples in the spec. Keep the spec
   to *what* and *why*; put *how* in the plan.
4. When a plan rests on an assumption about Go's standard library or tools,
   write a laboratory: a small test in `internal/labs/<area>lab/`, run it,
   and record the result in the plan's Laboratories table.
5. Slice the plan into task files with `depends_on` and disjoint `touches`
   so independent slices can run in parallel.
6. Run `go run ./internal/okflint` and `go test ./...`; fix every problem.
7. Commit per `docs/knowledge/process/git-workflow.md`
   (`docs(knowledge): spec NNN <slug>`) and push.
8. Reply with: a short summary of the requirements, the open questions that
   need the maintainer, and the list of slices with which can start now.
