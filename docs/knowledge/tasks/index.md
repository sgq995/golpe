# Tasks
Plan slices as todo lists. Pick a `todo` slice whose `depends_on` are all `done`.

* [SPEC-001 Template language](001-template-language/) - 7 slices; 01 done, 02 and 03 ready (parallel)
* [SPEC-002 Code generation](002-code-generation/) - 4 slices; 01 ready (parallel with SPEC-001 work)
* [SPEC-003 Golpe CLI](003-cli/) - 1 slice
* [SPEC-004 Language server](004-language-server/) - 1 slice (spec still draft)

# Dependency graph
```text
001-01 ─┬─ 001-02 ─┐
        ├─ 001-03 ─┴─ 001-04 ── 001-05 ── 001-06 ─┬─ 001-07
        └─ 002-01 ────────────────────────────────┴─ 002-02 ─┬─ 002-03 ── 002-04
                                                             └─ 003-01 ─┐
                                                       002-03 ──────────┴─ 004-01
```
