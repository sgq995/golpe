package parser

// Pass 2 of PLAN-001: the markup parser.

import (
	"bytes"
	"fmt"
	"go/scanner"
	"go/token"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sgq995/golpe/internal/ast"
	golpetoken "github.com/sgq995/golpe/internal/token"
)

// A region is one markup statement as parseMarkup returns it, with what the
// later passes need, as byte offsets into the source of the file:
//
//   - Pass 3, the mask, copies the source and, in src[start:end], replaces
//     every byte by a space except newlines and the bytes of the keep spans,
//     which stay Go. It then writes ';' at start, the placeholder go/parser
//     turns into an explicit EmptyStmt (LAB-001), and at every offset in
//     semis. The '{' of an interpolation is not kept, so {x} reads " x;"
//     (LAB-005). The mask keeps the length and every newline of the source.
//   - Pass 5, the graft, wraps root in an ast.MarkupStmt holding that
//     EmptyStmt, and gives each interpolation of interps the expression of
//     the expression statement go/parser built inside its braces.
type region struct {
	// start and end delimit the markup: src[start] is the '<' of the root;
	// end is after the root's last byte, or where parsing stopped after an
	// error (the end of the source, or the "}" of a line that ends the Go
	// block holding unclosed markup).
	start, end int

	// root is the markup tree: an *ast.Element, *ast.Fragment, *ast.Doctype
	// or *ast.HTMLComment. It is nil after an error that leaves no node, such
	// as an invalid "<!" or a doctype that is not terminated. Its positions
	// are final (see parseMarkup); Interp.X is nil until grafted.
	root ast.Node

	// keep holds the byte ranges that stay Go in the mask, in source order
	// and disjoint: interpolation expressions (the bytes between the braces)
	// and Go comments, which go/parser then records in File.Comments.
	// TASK-001-05 adds the headers of embedded statements.
	keep []span

	// semis holds the offsets, in source order, of the closing '}' of the
	// interpolations: the mask writes ';' there.
	semis []int

	// interps holds the interpolations of the tree, in source order.
	interps []*ast.Interp
}

// A span is the byte range [from, to) of a source.
type span struct{ from, to int }

// lineFile returns a stand-in for the token.File that go/parser.ParseFile
// will create for src, or for its mask, which has the same length and
// lines: same name, size, base and lines.
//
// Markup is parsed before go/parser runs, yet its nodes need their final
// positions. ParseFile adds the file to its file set at the set's Base(), so
// with base read from that set just before calling ParseFile, the positions
// taken from the stand-in are the real ones (TASK-001-06 asserts that
// go/ast.File.FileStart is token.Pos(base)), and the stand-in gives the
// line:col of errors. It knows nothing of //line directives, which go/parser
// learns while scanning.
func lineFile(filename string, base int, src []byte) *token.File {
	f := token.NewFileSet().AddFile(filename, base, len(src))
	f.SetLinesForContent(src)
	return f
}

// parseMarkup parses the markup statement that starts at src[start], a '<'
// at a statement position of a template body: an element, a fragment
// <>…</>, an HTML comment or a doctype (SPEC-001 R7–R9, R13, R15–R17,
// R19–R22). The statement ends at the end of its root.
//
// file gives the positions of the nodes and errors; its size is len(src)
// (see lineFile). Errors are added to errs, and the region holds as much of
// the tree as could be built.
//
// Every tag is parsed as an element: component tags are TASK-005-02's.
// Embedded statements are TASK-001-05's: a line of content whose first word
// is if, for, switch or select is reported as not supported yet.
func parseMarkup(file *token.File, src []byte, start int, errs *errorList) *region {
	p := &markupParser{file: file, src: src, off: start, errs: errs, unwind: noUnwind}
	p.r = &region{start: start}
	p.r.root = p.root()
	p.r.end = p.off
	return p.r
}

// markupParser is a recursive-descent parser over the bytes of one markup
// statement.
type markupParser struct {
	file *token.File
	src  []byte
	off  int // offset of the next byte
	errs *errorList
	r    *region // being built

	// stack holds the open elements and fragments, outermost first.
	stack []*frame

	// unwind is set after a close tag that closes an enclosing frame, the
	// tag being pending: the content loops of the frames deeper than unwind
	// return at once, and the frame at that depth takes pending as its close
	// tag. -1 ends the region early. noUnwind is the normal state.
	unwind  int
	pending *endTag

	// eof records that an error about the end of the source was reported:
	// whatever is still open ends there too, and is not reported again.
	eof bool
}

const noUnwind = math.MaxInt

// A frame is an element or fragment whose content is being parsed.
type frame struct {
	name string // the element's name as written; "" for a fragment

	// textOnly is the name, as written, of the <textarea> or <title> this
	// content is in, which holds only text and interpolations (SPEC-001
	// R20); "" otherwise.
	textOnly string

	pre bool // inside <pre> or <textarea>: whitespace is kept (R20)
}

// closeTag returns the close tag of f as it is expected: </name> or </>.
func (f *frame) closeTag() string { return "</" + f.name + ">" }

// root parses the root of the markup statement.
func (p *markupParser) root() ast.Node {
	if c := p.peek(1); p.peek(0) == '<' && (isLetter(c) || c == '>' || c == '!') {
		return p.node(&frame{}, true)
	}
	// Pass 1 starts markup only at "<" followed by an ASCII letter, ">" or "!".
	p.errorf(p.off, "expected markup")
	if p.off < len(p.src) {
		p.off++
	}
	return nil
}

// node parses the node that starts at p.off with '<' followed by an ASCII
// letter, '>' or '!': an element, a fragment, an HTML comment or, at the
// root of a markup statement, a doctype. parent is the frame whose content
// holds the node. It returns nil after an error that leaves no node.
func (p *markupParser) node(parent *frame, root bool) ast.Node {
	switch c := p.peek(1); {
	case isLetter(c):
		return p.element(parent)
	case c == '>':
		return p.fragment(parent)
	}
	switch rest := p.src[p.off+2:]; { // after "<!"
	case bytes.HasPrefix(rest, []byte("--")):
		if c := p.htmlComment(); c != nil {
			return c
		}
	case hasPrefixFold(rest, "doctype"):
		start := p.off
		if d := p.doctype(); d != nil {
			if root {
				return d
			}
			p.errorf(start, "misplaced <!DOCTYPE>: it must be a markup statement of its own")
		}
	default:
		p.errorf(p.off, `invalid "<!" markup: only <!-- … --> and <!DOCTYPE …> are allowed`)
		p.skipDeclaration()
	}
	return nil
}

// element parses the element at p.off: its start tag and, unless it is
// void or self-closing, its content and close tag (SPEC-001 R13, R19, R20).
func (p *markupParser) element(parent *frame) *ast.Element {
	lt := p.off
	p.off++ // '<'
	el := &ast.Element{Lt: p.pos(lt), Name: p.name()}
	if !p.startTag(el, lt) {
		return el
	}
	switch {
	case el.SelfClosing || golpetoken.IsVoid(el.Name):
		el.EndPos = p.pos(p.off)
	case golpetoken.IsRawText(el.Name):
		p.rawText(el)
	default:
		f := &frame{name: el.Name, pre: parent.pre || isPre(el.Name)}
		if isTextOnly(el.Name) {
			f.textOnly = el.Name
		}
		el.Children, el.Close, el.EndPos = p.content(f)
	}
	return el
}

// fragment parses the fragment <>…</> at p.off (SPEC-001 R8). Its content
// is like its parent's: text-only or preformatted if the parent's is.
func (p *markupParser) fragment(parent *frame) *ast.Fragment {
	fr := &ast.Fragment{Lt: p.pos(p.off)}
	p.off += len("<>")
	f := &frame{textOnly: parent.textOnly, pre: parent.pre}
	fr.Children, fr.Close, fr.EndPos = p.content(f)
	return fr
}

// startTag parses the attributes of el's start tag, whose '<' is at lt, and
// its end, ">" or "/>". It reports whether content may follow; it does not
// when the source ends first.
func (p *markupParser) startTag(el *ast.Element, lt int) bool {
	for {
		p.skipSpace()
		if p.off >= len(p.src) {
			p.errorEOF(lt, "start tag <%s> not terminated", el.Name)
			return false
		}
		switch c := p.src[p.off]; {
		case c == '>':
			p.off++
			return true
		case c == '/' && p.peek(1) == '>':
			p.off += len("/>")
			el.SelfClosing = true
			return true
		case c == '<':
			// Most likely a forgotten '>': the content starts here.
			p.errorf(lt, "start tag <%s> not terminated", el.Name)
			return true
		case c == '{':
			p.errorf(p.off, "attribute spread is not supported; write name={expr}")
			p.skipBraces()
		case isNameByte(c):
			p.attr(el)
		default:
			p.errorf(p.off, "unexpected %q in start tag <%s>", rune(c), el.Name)
			p.off++
		}
	}
}

// attr parses the attribute at p.off, name, name="v", name='v' or
// name={expr} (SPEC-001 R15), and appends it to el.Attrs. Names are kept as
// written; two names that differ only in ASCII case are duplicates, as in
// HTML, where the second would be dropped.
func (p *markupParser) attr(el *ast.Element) {
	start := p.off
	a := &ast.Attr{NamePos: p.pos(start), Name: p.name()}
	for _, b := range el.Attrs {
		if equalFoldASCII(a.Name, b.Name) {
			p.errorf(start, "duplicate attribute %q", a.Name)
			break
		}
	}
	el.Attrs = append(el.Attrs, a)
	afterName := p.off
	p.skipSpace()
	if p.peek(0) != '=' {
		p.off = afterName // a boolean attribute
		return
	}
	eq := p.off
	p.off++
	p.skipSpace()
	switch c := p.peek(0); {
	case p.off >= len(p.src):
		// The start tag reports the end of the source.
	case c == '"' || c == '\'':
		if v := p.quoted(); v != nil {
			a.Value = v
		}
	case c == '{':
		if x := p.interp(false); x != nil {
			a.Value = x
		}
	case c == '>' || c == '<' || c == '/' && p.peek(1) == '>':
		p.errorf(eq, "missing value for attribute %q", a.Name)
	default:
		from := p.off
		for p.off < len(p.src) && !endsUnquoted(p.src[p.off]) && !(p.src[p.off] == '/' && p.peek(1) == '>') {
			p.off++
		}
		v, q := string(p.src[from:p.off]), `"`
		if strings.Contains(v, q) {
			q = "'"
		}
		p.errorf(from, "unquoted value for attribute %q; write %s=%s%s%s", a.Name, a.Name, q, v, q)
	}
}

// quoted parses the quoted attribute value at p.off. Its content is
// literal: braces are text (SPEC-001 R15).
func (p *markupParser) quoted() *ast.StaticValue {
	open := p.off
	n := bytes.IndexByte(p.src[open+1:], p.src[open])
	if n < 0 {
		p.errorEOF(open, "attribute value not terminated")
		p.off = len(p.src)
		return nil
	}
	p.off = open + 1 + n + 1
	raw := string(p.src[open:p.off])
	return &ast.StaticValue{ValuePos: p.pos(open), Raw: raw, Value: raw[1 : len(raw)-1]}
}

// interp parses the interpolation at p.off, a '{', and records it for the
// mask and the graft (SPEC-001 R17). content tells whether it is content,
// rather than an attribute value. It returns nil after an error.
func (p *markupParser) interp(content bool) *ast.Interp {
	lbrace := p.off
	rbrace, empty, ok := p.braces(lbrace)
	if !ok {
		p.errorEOF(lbrace, "interpolation not terminated")
		p.off = len(p.src)
		return nil
	}
	p.off = rbrace + 1
	if empty {
		if content {
			p.errorf(lbrace, `empty interpolation; write {"{}"} for literal braces`)
		} else {
			p.errorf(lbrace, "empty interpolation")
		}
		return nil
	}
	x := &ast.Interp{Lbrace: p.pos(lbrace), Rbrace: p.pos(rbrace)}
	p.keepGo(lbrace+1, rbrace)
	p.r.semis = append(p.r.semis, rbrace)
	p.r.interps = append(p.r.interps, x)
	return x
}

// braces finds the '}' that matches the '{' at src[lbrace] by counting the
// braces of the Go tokens go/scanner finds from there, so braces in
// strings, runes and comments, and those of composite and function
// literals, count exactly as in Go. It reports whether there is no token
// between the two braces, and false if the source ends first.
func (p *markupParser) braces(lbrace int) (rbrace int, empty, ok bool) {
	src := p.src[lbrace:]
	file := token.NewFileSet().AddFile("", -1, len(src))
	var s scanner.Scanner
	s.Init(file, src, nil, 0) // Go errors are go/parser's to report, on the mask
	depth := 0
	for n := 1; ; n++ {
		pos, tok, _ := s.Scan()
		switch tok {
		case token.EOF:
			return 0, false, false
		case token.LBRACE:
			depth++
		case token.RBRACE:
			if depth--; depth == 0 {
				return lbrace + file.Offset(pos), n == 2, true
			}
		}
	}
}

// skipBraces skips the braces at p.off and what they hold, as an
// interpolation.
func (p *markupParser) skipBraces() {
	if rbrace, _, ok := p.braces(p.off); ok {
		p.off = rbrace + 1
		return
	}
	p.eof = true // the error already reported explains the rest
	p.off = len(p.src)
}

// content parses the content of the open element or fragment f up to and
// including its close tag (SPEC-001 R13, R16–R18, R20–R22). It returns the
// content, with its whitespace normalised, and the positions of the close
// tag's '<' and of its end, each NoPos when missing.
func (p *markupParser) content(f *frame) (list []ast.Node, closePos, endPos token.Pos) {
	p.stack = append(p.stack, f)
	depth := len(p.stack) - 1
	defer func() { p.stack = p.stack[:depth] }()

	var (
		text      = -1    // offset of the pending text, or -1
		boundary  = true  // at a node boundary: only whitespace since the start tag, a node or a Go comment (R21)
		space     = false // the previous byte is whitespace
		lineStart = false // only whitespace and Go comments since the last newline
	)
	flushAt := func(end int) { // the pending text ends at end
		if text >= 0 {
			list = append(list, &ast.Text{ValuePos: p.pos(text), Raw: string(p.src[text:end])})
			text = -1
		}
	}
	addText := func() { // the byte at p.off is text
		if text < 0 {
			text = p.off
		}
		p.off++
	}
loop:
	for {
		if p.unwind <= depth {
			flushAt(p.off)
			if p.unwind == depth { // an inner frame met f's close tag
				t := p.pending
				p.unwind, p.pending = noUnwind, nil
				p.off = t.end
				closePos, endPos = t.closePos, t.endPos
			}
			break loop
		}
		if p.off >= len(p.src) {
			p.errorEOF(p.off, "expected %s, found 'EOF'", f.closeTag())
			break loop
		}
		c := p.src[p.off]
		if isSpace(c) {
			addText()
			space = true
			lineStart = lineStart || c == '\n'
			continue
		}

		// The line rule (R18) looks at the first word of every line.
		atLineStart := lineStart
		lineStart = false
		if atLineStart {
			if c == '}' {
				// A "}" line closes an embedded body (TASK-001-05). With none
				// open it most likely ends the Go block that holds this
				// markup: f was not closed, and the region ends here.
				p.errorf(p.off, "expected %s, found '}'", f.closeTag())
				p.unwind = -1
				break loop
			}
			if f.textOnly == "" && golpetoken.IsEmbeddedKeyword(p.word()) {
				kw := p.off
				if n := p.embedded(); n != nil {
					flushAt(kw)
					list = append(list, n)
					boundary, space = true, false
					continue
				}
			}
		}

		switch {
		case c == '<':
			switch next := p.peek(1); {
			case next == '/' && (isLetter(p.peek(2)) || p.peek(2) == '>'):
				flushAt(p.off)
				t := p.scanEndTag()
				switch {
				case equalFoldASCII(t.name, f.name):
					p.off = t.end
					closePos, endPos = t.closePos, t.endPos
					break loop
				case t.name != "" && golpetoken.IsVoid(t.name):
					p.errorf(t.lt, "void element <%s> cannot have content or a close tag", t.name)
				default:
					p.errorf(t.lt, "expected %s, found </%s>", f.closeTag(), t.name)
					if i := p.enclosing(t.name); i >= 0 {
						p.unwind, p.pending = i, t
						break loop
					}
				}
				p.off = t.end // skip the stray close tag
				boundary = true
			case isLetter(next) || next == '>' || next == '!':
				flushAt(p.off)
				if f.textOnly != "" {
					p.errorf(p.off, "<%s> can contain only text and interpolations", f.textOnly)
				}
				if n := p.node(f, false); n != nil {
					list = append(list, n)
				}
				boundary = true
			default:
				p.errorf(p.off, `unexpected '<' in text; write {"<"} for a literal '<'`)
				addText()
				boundary = false
			}
		case c == '{':
			flushAt(p.off)
			if x := p.interp(true); x != nil {
				list = append(list, x)
			}
			boundary = true
		case c == '}':
			p.errorf(p.off, `unexpected '}' in text; write {"}"} for a literal '}'`)
			addText()
			boundary = false
		case c == '/' && f.textOnly == "" && (atLineStart || boundary && space) && (p.peek(1) == '/' || p.peek(1) == '*'):
			flushAt(p.off)
			p.goComment()
			boundary, lineStart = true, atLineStart // a comment is invisible to the line rule
		default:
			addText()
			boundary = false
		}
		space = false
	}
	flushAt(p.off)
	return normalize(list, f.pre), closePos, endPos
}

// enclosing returns the depth of the innermost frame enclosing the current
// one that the close tag </name> closes, or -1.
func (p *markupParser) enclosing(name string) int {
	for i := len(p.stack) - 2; i >= 0; i-- {
		if equalFoldASCII(p.stack[i].name, name) {
			return i
		}
	}
	return -1
}

// embedded parses the embedded statement whose keyword, if, for, switch or
// select, is the first word of the line at p.off (the line rule, SPEC-001
// R18), and returns it.
//
// It is TASK-001-05's. Until then it reports an error and returns nil, and
// the line is parsed as text.
func (p *markupParser) embedded() ast.Node {
	p.errorf(p.off, "embedded statements are not supported yet")
	return nil
}

// An endTag is a close tag, </name> or </>, as scanEndTag found it.
type endTag struct {
	name     string // as written; "" for </>
	lt, end  int    // offsets of its '<' and after it
	closePos token.Pos
	endPos   token.Pos // after its '>'; NoPos without one
}

// scanEndTag scans the close tag at p.off, "</" followed by a name or '>',
// without consuming it. It reports an error if the tag does not end with
// '>' after the name and optional whitespace; the tag then stops at the
// next '>' on its line, if one comes before a '<'.
func (p *markupParser) scanEndTag() *endTag {
	t := &endTag{lt: p.off, closePos: p.pos(p.off)}
	i := p.off + len("</")
	j := i
	for j < len(p.src) && isNameByte(p.src[j]) {
		j++
	}
	t.name = string(p.src[i:j])
	for j < len(p.src) && isSpace(p.src[j]) {
		j++
	}
	t.end = j
	switch {
	case j < len(p.src) && p.src[j] == '>':
		t.end++
	case j == len(p.src):
		p.errorEOF(t.lt, "close tag </%s> not terminated", t.name)
		return t
	default:
		p.errorf(t.lt, "close tag </%s> not terminated", t.name)
		for k := j; k < len(p.src) && p.src[k] != '<' && p.src[k] != '\n'; k++ {
			if p.src[k] == '>' {
				t.end = k + 1
				break
			}
		}
		if t.end == j {
			return t
		}
	}
	t.endPos = p.pos(t.end)
	return t
}

// rawText parses the content of the raw text element el (<script>,
// <style>), taken literally up to its close tag (SPEC-001 R19): "</", the
// element's name in any ASCII case, then '>', '/', whitespace or the end of
// the source. Then it parses the close tag.
func (p *markupParser) rawText(el *ast.Element) {
	start, end := p.off, len(p.src)
	for i := start; ; {
		j := bytes.Index(p.src[i:], []byte("</"))
		if j < 0 {
			break
		}
		j += i
		k := j + len("</") + len(el.Name)
		if hasPrefixFold(p.src[j+len("</"):], el.Name) &&
			(k == len(p.src) || p.src[k] == '>' || p.src[k] == '/' || isSpace(p.src[k])) {
			end = j
			break
		}
		i = j + len("</")
	}
	if end > start {
		raw := string(p.src[start:end])
		el.Children = []ast.Node{&ast.Text{ValuePos: p.pos(start), Raw: raw, Value: raw}}
	}
	p.off = end
	if end == len(p.src) {
		p.errorEOF(end, "expected </%s>, found 'EOF'", el.Name)
		return
	}
	t := p.scanEndTag()
	p.off = t.end
	el.Close, el.EndPos = t.closePos, t.endPos
}

// htmlComment parses the HTML comment <!-- … --> at p.off (SPEC-001 R21).
func (p *markupParser) htmlComment() *ast.HTMLComment {
	start := p.off
	n := bytes.Index(p.src[start+len("<!--"):], []byte("-->"))
	if n < 0 {
		p.errorEOF(start, "HTML comment not terminated")
		p.off = len(p.src)
		return nil
	}
	p.off = start + len("<!--") + n + len("-->")
	return &ast.HTMLComment{Start: p.pos(start), Text: string(p.src[start:p.off])}
}

// doctype parses the doctype <!DOCTYPE …> at p.off (SPEC-001 R9), which
// ends at the first '>'.
func (p *markupParser) doctype() *ast.Doctype {
	start := p.off
	n := bytes.IndexByte(p.src[start:], '>')
	if n < 0 {
		p.errorEOF(start, "doctype not terminated")
		p.off = len(p.src)
		return nil
	}
	p.off = start + n + 1
	return &ast.Doctype{Start: p.pos(start), Text: string(p.src[start:p.off])}
}

// skipDeclaration skips an invalid "<!" at p.off: up to the next '>' if one
// comes on the same line before a '<', else just "<!".
func (p *markupParser) skipDeclaration() {
	p.off += len("<!")
	for i := p.off; i < len(p.src) && p.src[i] != '<' && p.src[i] != '\n'; i++ {
		if p.src[i] == '>' {
			p.off = i + 1
			return
		}
	}
}

// goComment skips the Go comment at p.off, // or /* */, and keeps it as Go
// in the mask, so that go/parser records it (SPEC-001 R21).
func (p *markupParser) goComment() {
	start := p.off
	if p.src[start+1] == '/' {
		end := len(p.src)
		if n := bytes.IndexByte(p.src[start:], '\n'); n >= 0 {
			end = start + n
		}
		p.keepGo(start, end)
		p.off = end
		return
	}
	n := bytes.Index(p.src[start+len("/*"):], []byte("*/"))
	if n < 0 {
		p.errorEOF(start, "comment not terminated")
		p.off = len(p.src)
		return
	}
	p.off = start + len("/*") + n + len("*/")
	p.keepGo(start, p.off)
}

// keepGo records src[from:to] as Go for the mask.
func (p *markupParser) keepGo(from, to int) {
	p.r.keep = append(p.r.keep, span{from, to})
}

// normalize sets the Value of the Text nodes of a content list, which
// follow each other only across a Go comment, and drops those left empty
// (SPEC-001 R22): every whitespace run becomes one space, and the runs at
// the start and at the end of the content are dropped. A Go comment is
// invisible: the whitespace on both sides of it is one run, rendered once.
// With pre, inside <pre> or <textarea>, text is kept as written (R20).
func normalize(list []ast.Node, pre bool) []ast.Node {
	if pre {
		for _, n := range list {
			if t, ok := n.(*ast.Text); ok {
				t.Value = t.Raw
			}
		}
		return list
	}
	space := true // the rendered content is empty so far, or ends with a space
	for _, n := range list {
		t, ok := n.(*ast.Text)
		if !ok {
			space = false
			continue
		}
		b := make([]byte, 0, len(t.Raw))
		for i := 0; i < len(t.Raw); i++ {
			switch c := t.Raw[i]; {
			case !isSpace(c):
				b = append(b, c)
				space = false
			case !space:
				b = append(b, ' ')
				space = true
			}
		}
		t.Value = string(b)
	}
	// The run at the end, if any, ends the last Text that is not empty,
	// which has no other node after it.
	for i := len(list) - 1; i >= 0; i-- {
		t, ok := list[i].(*ast.Text)
		if !ok {
			break
		}
		if t.Value != "" {
			t.Value = strings.TrimSuffix(t.Value, " ")
			break
		}
	}
	out := list[:0]
	for _, n := range list {
		if t, ok := n.(*ast.Text); !ok || t.Value != "" {
			out = append(out, n)
		}
	}
	return out
}

// name scans a tag or attribute name at p.off.
func (p *markupParser) name() string {
	start := p.off
	for p.off < len(p.src) && isNameByte(p.src[p.off]) {
		p.off++
	}
	return string(p.src[start:p.off])
}

// word returns the word at p.off: letters, digits and '_', as in a Go
// identifier, so that "format" and "iffy" are not "for" and "if".
func (p *markupParser) word() string {
	i := p.off
	for i < len(p.src) {
		r, n := utf8.DecodeRune(p.src[i:])
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			break
		}
		i += n
	}
	return string(p.src[p.off:i])
}

func (p *markupParser) skipSpace() {
	for p.off < len(p.src) && isSpace(p.src[p.off]) {
		p.off++
	}
}

// peek returns the byte i bytes after p.off, or 0 past the end.
func (p *markupParser) peek(i int) byte {
	if p.off+i < len(p.src) {
		return p.src[p.off+i]
	}
	return 0
}

func (p *markupParser) pos(off int) token.Pos { return p.file.Pos(off) }

func (p *markupParser) errorf(off int, format string, args ...any) {
	p.errs.add(p.file.Position(p.pos(off)), fmt.Sprintf(format, args...))
}

// errorEOF reports an error at off about a construct that the end of the
// source cuts. Only the first such error is reported: everything still open
// ends there too.
func (p *markupParser) errorEOF(off int, format string, args ...any) {
	if !p.eof {
		p.eof = true
		p.errorf(off, format, args...)
	}
}

// isSpace reports whether c is ASCII whitespace as HTML defines it.
func isSpace(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\f', '\r':
		return true
	}
	return false
}

// isLetter reports whether c is an ASCII letter, which starts a tag name
// (SPEC-001 R7).
func isLetter(c byte) bool {
	c |= 'a' - 'A'
	return 'a' <= c && c <= 'z'
}

// isNameByte reports whether c may be part of a tag or attribute name:
// anything but whitespace, control characters and / > < = " ' { }. Names
// are checked further where they are classified (TASK-005-02).
func isNameByte(c byte) bool {
	switch c {
	case '/', '>', '<', '=', '"', '\'', '{', '}', 0x7f:
		return false
	}
	return c > ' '
}

// endsUnquoted reports whether c ends an unquoted attribute value, which
// is an error that the parser skips.
func endsUnquoted(c byte) bool {
	return isSpace(c) || c == '>' || c == '<' || c == '{' || c == '}'
}

// isPre reports whether the element name keeps its whitespace (SPEC-001
// R20).
func isPre(name string) bool {
	return equalFoldASCII(name, "pre") || equalFoldASCII(name, "textarea")
}

// isTextOnly reports whether the element name holds only text and
// interpolations (SPEC-001 R20).
func isTextOnly(name string) bool {
	return equalFoldASCII(name, "textarea") || equalFoldASCII(name, "title")
}

// equalFoldASCII reports whether a and b are equal ignoring ASCII case, as
// HTML compares tag and attribute names. Other bytes must match exactly.
func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if lowerASCII(a[i]) != lowerASCII(b[i]) {
			return false
		}
	}
	return true
}

// hasPrefixFold reports whether b starts with prefix, ignoring ASCII case.
func hasPrefixFold(b []byte, prefix string) bool {
	if len(b) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if lowerASCII(b[i]) != lowerASCII(prefix[i]) {
			return false
		}
	}
	return true
}

func lowerASCII(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return c + 'a' - 'A'
	}
	return c
}
