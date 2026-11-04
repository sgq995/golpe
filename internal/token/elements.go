package token

import (
	"cmp"
	"slices"
	"strings"
)

// IsElement reports whether name is the name of an HTML, SVG or MathML
// element, ignoring ASCII case, which makes a tag with that name markup
// (SPEC-005 R1, DEC-007): "div", "DIV", "linearGradient", "lineargradient"
// and "mrow" are element names; "Card", "myComponent" and "ui.Button" are
// not, and nor is "my-widget", which [IsCustomElement] reports.
//
// The names are those of the current standards: the WHATWG HTML Living
// Standard, SVG 2 and MathML Core. Obsolete HTML elements (center, font,
// marquee, …), SVG 1.1 elements that SVG 2 dropped (altGlyph, tref, …) and
// MathML elements outside MathML Core (none, mfenced, …) are not element
// names here. Because the match ignores case, a component named like an
// element cannot be used as an unqualified tag: <Button> is the HTML button
// element.
//
// As in [IsVoid], only the ASCII letters A-Z fold: "lin\u212a", with U+212A
// KELVIN SIGN, is not "link". IsElement does not allocate.
func IsElement(name string) bool {
	_, found := slices.BinarySearchFunc(elements[:], name, compareLowerASCII)
	return found
}

// IsCustomElement reports whether name contains '-' (U+002D HYPHEN-MINUS),
// which makes a tag with that name a custom element, and so markup
// (SPEC-005 R1). The HTML standard keeps such names for custom elements: no
// element will be added to HTML, SVG or MathML with a '-' in its name.
//
// The check is deliberately just that. The HTML standard's stricter rules
// for a valid custom element name (an ASCII lower-case letter first, no
// ASCII upper-case letters, a valid element local name, and none of eight
// reserved names such as annotation-xml and font-face) are not applied:
// "My-Widget", "x-" and "annotation-xml" are custom elements here. Whether
// a name is a well-formed tag name at all is for the parser to check.
func IsCustomElement(name string) bool {
	return strings.Contains(name, "-")
}

// compareLowerASCII compares a and b as strings.Compare(lowerASCII(a),
// lowerASCII(b)) does, without building either string: byte by byte, with
// the ASCII letters A-Z read as a-z.
func compareLowerASCII(a, b string) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if ca, cb := lowerASCIIByte(a[i]), lowerASCIIByte(b[i]); ca != cb {
			return cmp.Compare(ca, cb)
		}
	}
	return cmp.Compare(len(a), len(b))
}

// lowerASCIIByte maps an ASCII upper-case letter to lower case and returns
// any other byte unchanged.
func lowerASCIIByte(c byte) byte {
	if isUpperASCII(c) {
		return c + 'a' - 'A'
	}
	return c
}

// elements holds the element names of the current HTML, SVG and MathML
// standards, for IsElement. Sources, retrieved 2026-11-04:
//
//   - HTML: the "Elements" index of the WHATWG HTML Living Standard,
//     https://html.spec.whatwg.org/multipage/indices.html#elements-3. Its
//     "MathML math" and "SVG svg" rows come from the lists below, and its
//     "autonomous custom elements" row is IsCustomElement's.
//   - SVG: the element index of SVG 2 (Candidate Recommendation Snapshot,
//     6 October 2026), https://www.w3.org/TR/SVG2/eltindex.html.
//   - MathML: the MathML Core elements (Editor's Draft, 1 October 2026),
//     https://w3c.github.io/mathml-core/#mathml-elements-and-attributes.
//
// Left out as not current: the obsolete HTML elements authors must not use
// (center, font, marquee, frame, …; listed at
// https://html.spec.whatwg.org/multipage/obsolete.html#non-conforming-features),
// the SVG 1.1 elements SVG 2 dropped (altGlyph, font-face, tref, …; see
// https://www.w3.org/TR/SVG11/eltindex.html) and MathML elements outside
// MathML Core (none, mfenced, menclose, …).
//
// Names in several standards (a, script, style, title) appear once, and
// each is spelled as its standard spells it, so SVG keeps its camelCase
// (linearGradient). The table is sorted by compareLowerASCII, that is by
// ASCII-lower-cased name, with no two names equal ignoring ASCII case:
// IsElement's binary search needs both, and TestElementsTable checks them.
// TestElementSources checks the table against the lists as published, kept
// in elements_test.go.
//
// To refresh the table, re-read the three sources, then update their lists
// in elements_test.go, this table and the date above, and run the tests.
var elements = [...]string{
	"a", "abbr", "address", "animate", "animateMotion", "animateTransform", "annotation",
	"annotation-xml", "area", "article", "aside", "audio",
	"b", "base", "bdi", "bdo", "blockquote", "body", "br", "button",
	"canvas", "caption", "circle", "cite", "clipPath", "code", "col", "colgroup",
	"data", "datalist", "dd", "defs", "del", "desc", "details", "dfn", "dialog", "div", "dl",
	"dt",
	"ellipse", "em", "embed",
	"feBlend", "feColorMatrix", "feComponentTransfer", "feComposite", "feConvolveMatrix",
	"feDiffuseLighting", "feDisplacementMap", "feDistantLight", "feDropShadow", "feFlood",
	"feFuncA", "feFuncB", "feFuncG", "feFuncR", "feGaussianBlur", "feImage", "feMerge",
	"feMergeNode", "feMorphology", "feOffset", "fePointLight", "feSpecularLighting",
	"feSpotLight", "feTile", "feTurbulence", "fieldset", "figcaption", "figure", "filter",
	"footer", "foreignObject", "form",
	"g",
	"h1", "h2", "h3", "h4", "h5", "h6", "head", "header", "hgroup", "hr", "html",
	"i", "iframe", "image", "img", "input", "ins",
	"kbd",
	"label", "legend", "li", "line", "linearGradient", "link",
	"maction", "main", "map", "mark", "marker", "mask", "math", "menu", "merror", "meta",
	"metadata", "meter", "mfrac", "mi", "mmultiscripts", "mn", "mo", "mover", "mpadded",
	"mpath", "mphantom", "mprescripts", "mroot", "mrow", "ms", "mspace", "msqrt", "mstyle",
	"msub", "msubsup", "msup", "mtable", "mtd", "mtext", "mtr", "munder", "munderover",
	"nav", "noscript",
	"object", "ol", "optgroup", "option", "output",
	"p", "path", "pattern", "picture", "polygon", "polyline", "pre", "progress",
	"q",
	"radialGradient", "rect", "rp", "rt", "ruby",
	"s", "samp", "script", "search", "section", "select", "selectedcontent", "semantics", "set",
	"slot", "small", "source", "span", "stop", "strong", "style", "sub", "summary", "sup",
	"svg", "switch", "symbol",
	"table", "tbody", "td", "template", "text", "textarea", "textPath", "tfoot", "th", "thead",
	"time", "title", "tr", "track", "tspan",
	"u", "ul", "use",
	"var", "video", "view",
	"wbr",
}
