// Package lsp implements the Golpe language server: a proxy that compiles
// .golpe files to Go in memory, forwards requests to gopls, and maps
// positions back through the generator's source map.
//
// Status: draft. See docs/knowledge/specs/004-language-server.md.
package lsp
