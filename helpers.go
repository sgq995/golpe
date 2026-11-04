package golpe

import "strings"

// Class joins the non-empty parts with single spaces, to build a class
// attribute value out of optional parts. Empty parts are skipped the way
// [cmp.Or] skips zero values, so Class() and Class("", "") both return "".
//
// Parts are used exactly as given: Class does not trim, split, deduplicate
// or escape them, and a part of only spaces counts as non-empty. A template
// escapes the result like any other dynamic attribute value.
//
// Class is an optional convenience: code generated from .golpe files never
// calls it, and a template can build the value any other way. With [If] and
// [cmp.Or]:
//
//	<button class={golpe.Class("btn", cmp.Or(p.Variant, "primary"), golpe.If(p.Active, "active"))}>
func Class(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(p)
	}
	return b.String()
}

// If returns s when cond is true, and "" otherwise. It turns a conditional
// part of an attribute value into an expression, typically an argument to
// [Class], which skips the part when it is "".
//
// If is an optional convenience: code generated from .golpe files never
// calls it. With [Class] and [cmp.Or]:
//
//	<button class={golpe.Class("btn", cmp.Or(p.Variant, "primary"), golpe.If(p.Active, "active"))}>
func If(cond bool, s string) string {
	if cond {
		return s
	}
	return ""
}
