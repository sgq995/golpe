package token

import (
	"strings"
	"testing"
)

// spellings returns name as written, upper-cased, capitalised and in
// alternating case: the variants an ASCII case-insensitive match accepts.
func spellings(name string) []string {
	alt := []byte(name)
	for i := 0; i < len(alt); i += 2 {
		if 'a' <= alt[i] && alt[i] <= 'z' {
			alt[i] -= 'a' - 'A'
		}
	}
	return []string{name, strings.ToUpper(name), strings.ToUpper(name[:1]) + name[1:], string(alt)}
}

func TestIsVoid(t *testing.T) {
	// SPEC-001 R13: exactly these 13, matched ASCII case-insensitively.
	void := []string{"area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr"}
	for _, tag := range void {
		for _, s := range spellings(tag) {
			if !IsVoid(s) {
				t.Errorf("IsVoid(%+q) = false, want true", s)
			}
		}
	}
	notVoid := []string{
		"", "div", "span", "p", "a", "b", "html", "head", "body", "script", "style", "template", "textarea",
		// void in older HTML or legacy parsing only; not in R10
		"param", "keygen", "command", "menuitem", "basefont", "bgsound", "frame", "image", "isindex", "nextid",
		// near misses, custom elements and namespaced names
		"brr", "inputs", "b r", " br", "br ", "br/", "my-br", "br-x", "svg:br", "img:x",
		// only ASCII letters fold: U+212A KELVIN SIGN and U+017F LATIN SMALL
		// LETTER LONG S are not "k" and "s" (strings.ToLower and EqualFold
		// would say they are)
		"lin\u212a", "LIN\u212a", "\u017fource",
	}
	for _, tag := range notVoid {
		if IsVoid(tag) {
			t.Errorf("IsVoid(%+q) = true, want false", tag)
		}
	}
}

func TestIsRawText(t *testing.T) {
	// SPEC-001 R19: script and style, matched ASCII case-insensitively.
	for _, tag := range []string{"script", "style"} {
		for _, s := range spellings(tag) {
			if !IsRawText(s) {
				t.Errorf("IsRawText(%+q) = false, want true", s)
			}
		}
	}
	for _, tag := range []string{
		"", "s", "scrip", "scripts", "styles", "div", "br",
		// raw or escapable text in the HTML parser, but not in R16
		"textarea", "title", "xmp", "iframe", "noembed", "noframes", "noscript", "plaintext",
		"my-script", "svg:style", "\u017fcript", "\u017ftyle",
	} {
		if IsRawText(tag) {
			t.Errorf("IsRawText(%+q) = true, want false", tag)
		}
	}
}

func TestIsEmbeddedKeyword(t *testing.T) {
	// SPEC-001 R18: the Go keywords that start an embedded statement.
	for _, kw := range []string{"if", "for", "switch", "select"} {
		if !IsEmbeddedKeyword(kw) {
			t.Errorf("IsEmbeddedKeyword(%+q) = false, want true", kw)
		}
	}
	for _, s := range []string{
		// Go keywords are case-sensitive
		"If", "IF", "For", "FOR", "Switch", "SELECT",
		// clauses and continuations, not statement starts
		"else", "case", "default", "range",
		// other Go keywords and words
		"go", "func", "return", "defer", "break", "continue", "goto", "fallthrough", "type", "var", "chan",
		"template", "element", "", "if ", "iff", "fo",
	} {
		if IsEmbeddedKeyword(s) {
			t.Errorf("IsEmbeddedKeyword(%+q) = true, want false", s)
		}
	}
}

func TestLowerASCII(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"br", "br"},
		{"BR", "br"},
		{"My-Widget", "my-widget"},
		{"SVG:Path", "svg:path"},
		{"aZ09@[`{", "az09@[`{"}, // only A-Z change
		{"\u212a\u017f", "\u212a\u017f"},
		{"\u00c0B", "\u00c0b"}, // non-ASCII letters are left alone
	}
	for _, tt := range tests {
		if got := lowerASCII(tt.in); got != tt.want {
			t.Errorf("lowerASCII(%+q) = %+q, want %+q", tt.in, got, tt.want)
		}
	}
}
