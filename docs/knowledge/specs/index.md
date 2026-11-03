# Specs
What we build and why. `status: draft` until the maintainer approves.

* [SPEC-001 Template language](001-template-language.md) - template declarations whose bodies switch from Go to balanced HTML with embedded if/for/switch/select, {expr}, fragments and writer()
* [SPEC-002 Code generation](002-code-generation.md) - each .golpe compiles to a committed x_gen.go using only the standard library
* [SPEC-003 Golpe CLI](003-cli.md) - one golpe binary (go install or go tool): generate, -check, -clean
* [SPEC-004 Language server](004-language-server.md) - editor intelligence by proxying gopls over generated Go (after the generator)
* [SPEC-005 Component tags](005-component-tags.md) - non-HTML tags call templates with props built from attributes
* [SPEC-006 Optional helpers](006-optional-helpers.md) - Element, Class and If in the root package; never needed by generated code
