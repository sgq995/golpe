// Package golpe holds optional helpers for projects that use Golpe templates.
//
// Nothing here is required: code generated from .golpe files imports only
// the standard library and never this package (see
// docs/knowledge/decisions/005-standard-library-only-output.md), so removing
// it from a project only affects code the developer wrote. Import it only
// for the conveniences it offers (see
// docs/knowledge/specs/006-optional-helpers.md):
//
//   - [Element] names the type every template returns;
//   - [Class] builds a class attribute value from optional parts;
//   - [If] keeps a string only when a condition holds.
package golpe

import "io"

// Element is a renderable piece of HTML: a function that writes itself to w.
// It is the type every template returns, spelled element inside .golpe files.
//
// It is a type alias (not a defined type), so any plain
// func(io.Writer) error is an Element and vice versa, without conversions.
// See docs/knowledge/decisions/002-element-is-an-alias.md.
type Element = func(w io.Writer) error
