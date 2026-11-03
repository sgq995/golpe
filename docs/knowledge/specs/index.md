# Specs
What we build and why. `status: draft` until the maintainer approves.

* [SPEC-001 Template language](001-template-language.md) - template declarations whose bodies switch from Go to balanced HTML with embedded if/for/switch/select and {expr}
* [SPEC-002 Code generation](002-code-generation.md) - each template compiles to a plain func returning golpe.Element that writes escaped HTML
* [SPEC-003 Golpe CLI](003-cli.md) - golpe generate across packages, errors at .golpe positions, -check for CI
* [SPEC-004 Language server](004-language-server.md) - editor intelligence by proxying gopls over generated Go
