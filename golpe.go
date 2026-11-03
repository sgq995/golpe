// Package golpe is the public runtime API used by code generated from .golpe
// files, and by users who compose generated templates by hand.
//
// See docs/knowledge/specs/001-template-language.md for the language and
// docs/knowledge/specs/002-code-generation.md for what the generator emits.
package golpe

import "io"

// Element is a renderable piece of HTML: a function that writes itself to w.
//
// It is a type alias (not a defined type) so that any plain
// func(io.Writer) error can be used where an Element is expected, and vice
// versa, without conversions. See docs/knowledge/decisions/002-element-is-an-alias.md.
type Element = func(w io.Writer) error

// Writer is io.Writer under a name in this package, so generated code can
// spell its closures func(__golpeW __golpe.Writer) error without importing
// "io", which could clash with the user's own imports. Being an alias, that
// spelling is exactly the type Element.
type Writer = io.Writer
