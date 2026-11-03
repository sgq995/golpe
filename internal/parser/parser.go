// Package parser parses .golpe files.
//
// Go itself is parsed by the standard go/parser: markup regions are first
// masked out of the source (keeping byte offsets and line numbers intact),
// the masked source is parsed as plain Go, and the markup is parsed by this
// package and grafted back into the AST as internal/ast nodes.
//
// Status: scaffold. See docs/knowledge/plans/001-template-language.md.
package parser
