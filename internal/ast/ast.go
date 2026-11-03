// Package ast declares the syntax tree nodes Golpe adds to go/ast.
//
// New nodes embed a go/ast node so they satisfy go/ast's sealed interfaces
// (ast.Decl, ast.Stmt) through promoted unexported methods. Because
// ast.Walk and ast.Inspect panic on unknown node types, this package also
// provides its own Walk/Inspect that understand both.
//
// Status: scaffold. See docs/knowledge/plans/001-template-language.md.
package ast
