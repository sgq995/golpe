package token

import (
	"strings"
	"testing"
)

// The element lists as the standards publish them, in their order and
// spelling, retrieved 2026-11-04 from the sources named on the elements
// table. TestElementSources checks that the table holds exactly these
// names; refresh them together.
var (
	// The "Elements" index of the WHATWG HTML Living Standard, without its
	// "MathML math", "SVG svg" and "autonomous custom elements" rows.
	htmlElements = []string{
		"a", "abbr", "address", "area", "article", "aside", "audio", "b", "base", "bdi", "bdo",
		"blockquote", "body", "br", "button", "canvas", "caption", "cite", "code", "col",
		"colgroup", "data", "datalist", "dd", "del", "details", "dfn", "dialog", "div", "dl", "dt",
		"em", "embed", "fieldset", "figcaption", "figure", "footer", "form", "h1", "h2", "h3", "h4",
		"h5", "h6", "head", "header", "hgroup", "hr", "html", "i", "iframe", "img", "input", "ins",
		"kbd", "label", "legend", "li", "link", "main", "map", "mark", "menu", "meta", "meter",
		"nav", "noscript", "object", "ol", "optgroup", "option", "output", "p", "picture", "pre",
		"progress", "q", "rp", "rt", "ruby", "s", "samp", "script", "search", "section", "select",
		"selectedcontent", "slot", "small", "source", "span", "strong", "style", "sub", "summary",
		"sup", "table", "tbody", "td", "template", "textarea", "tfoot", "th", "thead", "time",
		"title", "tr", "track", "u", "ul", "var", "video", "wbr",
	}
	// The element index of SVG 2.
	svgElements = []string{
		"a", "animate", "animateMotion", "animateTransform", "circle", "clipPath", "defs", "desc",
		"ellipse", "feBlend", "feColorMatrix", "feComponentTransfer", "feComposite",
		"feConvolveMatrix", "feDiffuseLighting", "feDisplacementMap", "feDistantLight",
		"feDropShadow", "feFlood", "feFuncA", "feFuncB", "feFuncG", "feFuncR", "feGaussianBlur",
		"feImage", "feMerge", "feMergeNode", "feMorphology", "feOffset", "fePointLight",
		"feSpecularLighting", "feSpotLight", "feTile", "feTurbulence", "filter", "foreignObject",
		"g", "image", "line", "linearGradient", "marker", "mask", "metadata", "mpath", "path",
		"pattern", "polygon", "polyline", "radialGradient", "rect", "script", "set", "stop",
		"style", "svg", "switch", "symbol", "text", "textPath", "title", "tspan", "use", "view",
	}
	// The MathML Core elements.
	mathmlElements = []string{
		"a", "annotation", "annotation-xml", "maction", "math", "merror", "mfrac", "mi",
		"mmultiscripts", "mn", "mo", "mover", "mpadded", "mphantom", "mprescripts", "mroot", "mrow",
		"ms", "mspace", "msqrt", "mstyle", "msub", "msubsup", "msup", "mtable", "mtd", "mtext",
		"mtr", "munder", "munderover", "semantics",
	}
)

// TestElementsTable checks what IsElement's binary search relies on: the
// table is in strictly increasing compareLowerASCII order, that is sorted by
// ASCII-lower-cased name with no two names equal ignoring ASCII case. Each
// name must also look like one: an ASCII lower-case letter, then ASCII
// letters, digits and '-' (which catches a stray space, or the ‘quotes’ the
// SVG index puts around names).
func TestElementsTable(t *testing.T) {
	const lower = "abcdefghijklmnopqrstuvwxyz"
	const nameBytes = lower + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
	for i, name := range elements {
		if name == "" || !strings.Contains(lower, name[:1]) || strings.Trim(name, nameBytes) != "" {
			t.Errorf("elements[%d] = %+q: want an ASCII lower-case letter followed by ASCII letters, digits and '-'", i, name)
		}
		if i > 0 && compareLowerASCII(elements[i-1], name) >= 0 {
			t.Errorf("elements[%d] = %q is not after elements[%d] = %q ignoring ASCII case: keep the table sorted by lower-cased name, without duplicates", i, name, i-1, elements[i-1])
		}
	}
}

// TestElementSources checks the table against the published lists: each
// table name is a listed name, spelled as its standard spells it (SVG keeps
// its camelCase), and each listed name is in the table.
func TestElementSources(t *testing.T) {
	lists := [][]string{htmlElements, svgElements, mathmlElements}
	listed := map[string]map[string]bool{} // lowerASCII(name) → spellings
	for _, list := range lists {
		for _, name := range list {
			key := lowerASCII(name)
			if listed[key] == nil {
				listed[key] = map[string]bool{}
			}
			listed[key][name] = true
		}
	}
	inTable := map[string]bool{}
	for _, name := range elements {
		key := lowerASCII(name)
		inTable[key] = true
		if !listed[key][name] {
			t.Errorf("table name %q is not in the published HTML, SVG or MathML lists, spelled that way", name)
		}
	}
	for _, list := range lists {
		for _, name := range list {
			if !inTable[lowerASCII(name)] {
				t.Errorf("published name %q is missing from the table", name)
			}
		}
	}
}

func TestIsElement(t *testing.T) {
	// SPEC-005 R1: every name of the three standards, matched ignoring ASCII
	// case: as spelled, lower-cased, upper-cased, capitalised and in
	// alternating case.
	for _, list := range [][]string{htmlElements, svgElements, mathmlElements} {
		for _, name := range list {
			for _, s := range append(spellings(name), strings.ToLower(name)) {
				if !IsElement(s) {
					t.Errorf("IsElement(%+q) = false, want true", s)
				}
			}
		}
	}
	for _, name := range []string{
		// SPEC-005 acceptance criteria
		"div", "DIV", "linearGradient", "math",
		// SVG camelCase, in any ASCII case
		"linearGradient", "lineargradient", "LINEARGRADIENT", "LinearGradient",
		"foreignObject", "foreignobject", "FOREIGNOBJECT", "clipPath", "textPath", "feGaussianBlur",
		// MathML Core
		"math", "Math", "mrow", "MROW", "mi", "semantics", "annotation-xml",
		// SPEC-005 "Known consequence of R1": these component names are elements
		"Button", "Header", "Footer", "Nav", "Main", "Section", "Form", "Input", "Label",
		"Select", "Option", "Table", "Dialog", "Menu", "Search", "Title", "Link", "Data",
		"Time", "Image", "Text",
		// every IsVoid and IsRawText name
		"area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source",
		"track", "wbr", "script", "style",
	} {
		if !IsElement(name) {
			t.Errorf("IsElement(%+q) = false, want true", name)
		}
	}
	for _, name := range []string{
		"",
		// component tags (SPEC-005 R1)
		"Card", "card", "CARD", "myComponent", "MyComponent", "PrimaryButton", "SiteHeader",
		"ui.Button", "ui.button", "html.Div",
		// custom elements are IsCustomElement's
		"my-widget", "x-div", "div-", "-div",
		// namespaced, padded or partial names
		"svg:path", "svg:svg", "math:mrow", " div", "div ", "di v", "<div>", "div/", "div\x00",
		// near misses, and names past either end of the table
		"abb", "abbrr", "aa", "h", "h0", "h7", "h10", "linear", "Gradient", "mrows", "0", "@", "~", "zz",
		// left out: obsolete, non-conforming HTML elements (TASK-005-01 Notes)
		"acronym", "applet", "basefont", "bgsound", "big", "blink", "center", "dir", "font",
		"frame", "frameset", "isindex", "keygen", "listing", "marquee", "menuitem", "multicol",
		"nextid", "nobr", "noembed", "noframes", "param", "plaintext", "rb", "rtc", "spacer",
		"strike", "tt", "xmp",
		// left out: SVG 1.1 elements that SVG 2 dropped (the dashed ones are
		// still markup, as custom elements)
		"altGlyph", "altGlyphDef", "altGlyphItem", "animateColor", "color-profile", "cursor",
		"font-face", "font-face-format", "font-face-name", "font-face-src", "font-face-uri",
		"glyph", "glyphRef", "hkern", "missing-glyph", "tref", "vkern",
		// left out: MathML elements outside MathML Core
		"none", "mfenced", "menclose", "mglyph", "mlabeledtr", "mstack", "mlongdiv",
		"maligngroup", "malignmark", "apply", "ci", "cn",
		// only ASCII letters fold: U+212A KELVIN SIGN is not "k" and U+017F
		// LATIN SMALL LETTER LONG S is not "s" (strings.EqualFold would say
		// they are)
		"lin\u212a", "LIN\u212a", "\u212abd", "mar\u212a", "trac\u212a",
		"\u017fvg", "\u017fpan", "m\u017f", "\u017fcript",
	} {
		if IsElement(name) {
			t.Errorf("IsElement(%+q) = true, want false", name)
		}
	}
}

// TestCompareLowerASCII checks the comparison IsElement searches with
// against its definition: strings.Compare of both sides after lowerASCII.
func TestCompareLowerASCII(t *testing.T) {
	words := []string{
		"", "a", "A", "b", "B", "z", "Z", "ab", "aB", "AB", "abc", "0", "-", "@", "[", "`", "{",
		"a-", "a-b", "h1", "H1", "k", "K", "\u212a", "s", "\u017f", "\u00c0",
		"line", "link", "lin\u212a", "linearGradient", "LINEARGRADIENT", "lineargradient",
	}
	for _, a := range words {
		for _, b := range words {
			if got, want := compareLowerASCII(a, b), strings.Compare(lowerASCII(a), lowerASCII(b)); got != want {
				t.Errorf("compareLowerASCII(%+q, %+q) = %d, want %d", a, b, got, want)
			}
		}
	}
}

func TestIsElementDoesNotAllocate(t *testing.T) {
	names := []string{"div", "DIV", "linearGradient", "LINEARGRADIENT", "mrow", "Card", "my-widget", "lin\u212a", ""}
	allocs := testing.AllocsPerRun(100, func() {
		for _, name := range names {
			IsElement(name)
		}
	})
	if allocs != 0 {
		t.Errorf("IsElement allocated %v times per %d lookups, want 0", allocs, len(names))
	}
}

func BenchmarkIsElement(b *testing.B) {
	for _, name := range []string{"a", "div", "DIV", "linearGradient", "LINEARGRADIENT", "mrow", "Card", "my-widget"} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				IsElement(name)
			}
		})
	}
}

func TestIsCustomElement(t *testing.T) {
	for _, name := range []string{
		"my-widget", "MY-WIDGET", "My-Widget", "x-div", "a-b-c", "x-", "-x", "-",
		// an element name with '-' is a custom element too; the tag is markup either way
		"annotation-xml",
		// SVG 1.1 names left out of the element table stay markup this way
		"font-face", "color-profile", "missing-glyph",
		// the HTML standard's stricter rules for custom element names are not applied
		"1-x", "my-Widget", "my-widg\u00e9t", "my.widget-x",
	} {
		if !IsCustomElement(name) {
			t.Errorf("IsCustomElement(%+q) = false, want true", name)
		}
	}
	for _, name := range []string{
		"", "div", "Card", "myComponent", "myWidget", "mywidget", "my_widget", "ui.Button", "svg:path",
		// only U+002D HYPHEN-MINUS counts, not other dashes
		"my\u2010widget", // HYPHEN
		"my\u2011widget", // NON-BREAKING HYPHEN
		"my\u2013widget", // EN DASH
		"my\u2212widget", // MINUS SIGN
		"my\ufe63widget", // SMALL HYPHEN-MINUS
		"my\uff0dwidget", // FULLWIDTH HYPHEN-MINUS
	} {
		if IsCustomElement(name) {
			t.Errorf("IsCustomElement(%+q) = true, want false", name)
		}
	}
}
