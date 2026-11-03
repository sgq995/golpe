# Plans
How we build each spec. One plan per spec, same number.

* [PLAN-001 Parsing the template language](001-template-language.md) - scan, parse markup, mask to a Go skeleton, go/parser once, graft markup nodes
* [PLAN-002 Generating Go](002-code-generation.md) - splice source, lower markup to escaped writes, line directives, gofmt; runtime writers
* [PLAN-003 golpe generate](003-cli.md) - stdlib flag dispatcher, internal/build shared with the LSP, write-if-changed
* [PLAN-004 Language server](004-language-server.md) - proxy to gopls with overlays and a line-directive source map
