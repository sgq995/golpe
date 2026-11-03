# Plans
How we build each spec. One plan per spec, same number.

* [PLAN-001 Parsing the template language](001-template-language.md) - scan, parse markup, mask to a Go skeleton, go/parser once, graft markup nodes
* [PLAN-002 Generating Go](002-code-generation.md) - splice source, lower markup to standard-library writes, line directives, wrapped errors
* [PLAN-003 golpe generate](003-cli.md) - stdlib flag dispatcher in cmd/golpe, internal/build shared with the LSP
* [PLAN-004 Language server](004-language-server.md) - proxy to gopls with overlays and a line-directive source map
* [PLAN-005 Component tags](005-component-tags.md) - element lists, ast.Component, syntactic declaration resolver, lowering
* [PLAN-006 Optional helpers](006-optional-helpers.md) - Class and If next to the Element alias
