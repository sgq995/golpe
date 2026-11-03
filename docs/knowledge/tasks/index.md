# Tasks
Plan slices as todo lists. Pick a `todo` slice whose `depends_on` are all `done`.

* [SPEC-001 Template language](001-template-language/) - 7 slices; 01–03 done, 04 ready
* [SPEC-002 Code generation](002-code-generation/) - 4 slices; 01 done and superseded
* [SPEC-003 Golpe CLI](003-cli/) - 1 slice
* [SPEC-004 Language server](004-language-server/) - 1 slice (after the generator)
* [SPEC-005 Component tags](005-component-tags/) - 4 slices; 01 ready
* [SPEC-006 Optional helpers](006-optional-helpers/) - 1 slice; ready

# Ready now (disjoint touches, can run in parallel)
TASK-001-04 (internal/parser, internal/ast) · TASK-005-01 (internal/token) · TASK-006-01 (root package)

# Dependencies

| Slice | Depends on | State |
|-------|------------|-------|
| 001-01 Project scaffold | — | done |
| 001-02 Token extension | 001-01 | done |
| 001-03 AST nodes | 001-01 | done |
| 001-04 Markup parser | 001-02, 001-03 | ready |
| 001-05 Embedded statements | 001-04 | |
| 001-06 File parser | 001-05 | |
| 001-07 Conformance | 001-06 | |
| 002-01 Runtime writers | 001-01 | done (superseded) |
| 002-02 Lowering | 001-06 | |
| 002-03 Directives, wrapped errors | 002-02 | |
| 002-04 End-to-end | 002-03 | |
| 003-01 generate command | 002-02 | |
| 004-01 LSP proxy skeleton | 002-03, 003-01 | |
| 005-01 Element name lists | 001-02 | ready |
| 005-02 Parse component tags | 001-04, 005-01 | |
| 005-03 Declaration resolver | 001-06 | |
| 005-04 Lower component tags | 002-03, 005-02, 005-03 | |
| 006-01 Class and If | 001-01 | ready |
