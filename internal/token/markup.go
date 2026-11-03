package token

// IsVoid reports whether tag names an HTML void element, which has no close
// tag (SPEC-001 R10): area, base, br, col, embed, hr, img, input, link,
// meta, source, track or wbr. As in HTML, ASCII case is ignored.
func IsVoid(tag string) bool {
	switch lowerASCII(tag) {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr":
		return true
	}
	return false
}

// IsRawText reports whether tag names an element whose content is taken
// literally up to its close tag, with no interpolation and no embedded
// statements (SPEC-001 R16): script or style. As in HTML, ASCII case is
// ignored.
func IsRawText(tag string) bool {
	switch lowerASCII(tag) {
	case "script", "style":
		return true
	}
	return false
}

// IsEmbeddedKeyword reports whether word is one of the Go keywords that
// start an embedded statement inside markup (SPEC-001 R15): if, for, switch
// or select. Like Go keywords it is case-sensitive, so "If" is not one.
func IsEmbeddedKeyword(word string) bool {
	switch word {
	case "if", "for", "switch", "select":
		return true
	}
	return false
}

// lowerASCII maps the ASCII letters A-Z in s to lower case and leaves every
// other byte alone, which is how HTML compares tag names. strings.ToLower
// would also fold non-ASCII letters: U+212A KELVIN SIGN becomes "k", so
// "lin\u212a" would pass for "link". s itself is returned, without
// allocating, when it has no upper-case ASCII letter.
func lowerASCII(s string) string {
	i := 0
	for i < len(s) && !isUpperASCII(s[i]) {
		i++
	}
	if i == len(s) {
		return s
	}
	b := []byte(s)
	for ; i < len(b); i++ {
		if isUpperASCII(b[i]) {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

func isUpperASCII(c byte) bool { return 'A' <= c && c <= 'Z' }
