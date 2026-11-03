package golpe

import "html"

// All escaping of dynamic values happens in this file: one audited place
// (DEC-004).
//
// Golpe writes dynamic values into two contexts only: element content, and
// attribute values, which it always double-quotes. One escaper is correct in
// both. It replaces the five characters that are special in either context,
// exactly as html.EscapeString does:
//
//	&  →  &amp;
//	<  →  &lt;
//	>  →  &gt;
//	"  →  &#34;
//	'  →  &#39;
//
// The generator is to escape static text and static attribute values with
// html.EscapeString too, at generation time, so that static and dynamic
// output agree byte for byte.

// writeText writes s to w HTML-escaped. An empty s writes nothing.
func writeText(w Writer, s string) error {
	if s == "" {
		return nil
	}
	return WriteString(w, html.EscapeString(s))
}

// writeAttrValue writes s to w HTML-escaped and between double quotes.
func writeAttrValue(w Writer, s string) error {
	if err := WriteString(w, `"`); err != nil {
		return err
	}
	if err := writeText(w, s); err != nil {
		return err
	}
	return WriteString(w, `"`)
}
