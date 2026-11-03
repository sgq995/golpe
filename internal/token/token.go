// Package token extends go/token with the tokens Golpe adds to Go.
//
// A [Token] holds either a go/token token or a Golpe one. Values below an
// internal base are exactly go/token's: [FromGo] and [Token.Go] convert
// without changing the value, and [Token.String] prints what go/token
// prints, including "token(N)" for values go/token has no name for. Golpe's
// tokens live above the base: the contextual keyword [TEMPLATE] and the
// markup tokens [TAG_OPEN] through [DOCTYPE].
//
// "template" is a keyword only as the first token of a top-level
// declaration, followed by a name or a method receiver (SPEC-001 R3,
// DEC-003). So [Lookup], like go/token.Lookup, returns IDENT for it, and
// the parser asks [IsTemplateKeyword] where a declaration starts.
//
// [IsVoid], [IsRawText] and [IsEmbeddedKeyword] classify tag names and
// words for the markup parser and the code generator.
//
// Positions are not part of this package: Golpe uses go/token.Pos in a
// shared go/token.FileSet throughout.
package token

import (
	gotoken "go/token"
	"strconv"
)

// Token is a lexical token of a .golpe file: one of go/token's tokens, or
// one Golpe adds. The zero value is go/token's ILLEGAL.
type Token int

// golpeBase is the first Golpe token. go/token's tokens end far below it;
// TestGoTokenEndsBelowGolpeBase guards that.
const golpeBase Token = 1 << 10

// The tokens Golpe adds to Go. The comments show what each stands for.
const (
	// TEMPLATE is the contextual keyword that starts a template
	// declaration; see IsTemplateKeyword.
	TEMPLATE Token = golpeBase + iota

	TAG_OPEN       // <name, the start of an open tag
	TAG_END        // >, the end of an open tag
	TAG_SELF_CLOSE // />, the end of a self-closing tag
	TAG_CLOSE      // </name>
	TEXT           // text content
	INTERP_OPEN    // {, the start of an interpolation
	INTERP_CLOSE   // }, the end of an interpolation
	HTML_COMMENT   // <!-- … -->
	DOCTYPE        // <!DOCTYPE …>

	golpeEnd // one past the last Golpe token
)

const templateKeyword = "template"

// golpeNames follows go/token's String convention: keywords and delimiters
// print as their spelling, token classes as their constant name.
var golpeNames = [golpeEnd - golpeBase]string{
	TEMPLATE - golpeBase:       templateKeyword,
	TAG_OPEN - golpeBase:       "TAG_OPEN",
	TAG_END - golpeBase:        ">",
	TAG_SELF_CLOSE - golpeBase: "/>",
	TAG_CLOSE - golpeBase:      "TAG_CLOSE",
	TEXT - golpeBase:           "TEXT",
	INTERP_OPEN - golpeBase:    "{",
	INTERP_CLOSE - golpeBase:   "}",
	HTML_COMMENT - golpeBase:   "HTML_COMMENT",
	DOCTYPE - golpeBase:        "DOCTYPE",
}

// FromGo returns t as a Token. The value is unchanged, so FromGo(t).Go()
// returns t, true for every token go/token defines.
func FromGo(t gotoken.Token) Token { return Token(t) }

// Go returns t as a go/token token and true when t is one of Go's tokens,
// that is, any value below the Golpe range; the value is unchanged. For
// Golpe's tokens, and any other value at or above their range, it returns
// ILLEGAL and false.
func (t Token) Go() (gotoken.Token, bool) {
	if t < golpeBase {
		return gotoken.Token(t), true
	}
	return gotoken.ILLEGAL, false
}

// String returns the string for t. Below the Golpe range it is exactly
// go/token's String. Golpe's keyword and delimiters print as their spelling
// ("template", ">", "/>", "{", "}"), its token classes as their constant
// name ("TAG_OPEN", "TEXT", …), and values with no token as "token(N)".
func (t Token) String() string {
	if g, ok := t.Go(); ok {
		return g.String()
	}
	if t < golpeEnd {
		return golpeNames[t-golpeBase]
	}
	return "token(" + strconv.Itoa(int(t)) + ")"
}

// Lookup maps an identifier to its Go keyword token, or to IDENT if it is
// not a keyword, exactly like go/token.Lookup. Lookup("template") is IDENT:
// template is a keyword only where IsTemplateKeyword says so.
func Lookup(ident string) Token { return FromGo(gotoken.Lookup(ident)) }

// IsTemplateKeyword reports whether the identifier lit, followed by the
// token next, is the contextual keyword [TEMPLATE] (SPEC-001 R3, DEC-003):
// lit is "template" and next is an identifier (the template's name, as in
// "template Card(…)") or "(" (a method receiver, as in
// "template (c Card) View()").
//
// The rule applies only to the first token of a top-level declaration, and
// checking that is the caller's job. Everywhere else "template" is an
// ordinary identifier, as in template.Must(…).
func IsTemplateKeyword(lit string, next Token) bool {
	return lit == templateKeyword && (next == Token(gotoken.IDENT) || next == Token(gotoken.LPAREN))
}
