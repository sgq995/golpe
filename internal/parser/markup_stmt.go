package parser

// Embedded statements in markup: the line rule (SPEC-001 R18).
//
// Inside markup, a line whose first word is if, for, switch or select starts
// an embedded statement. Its header is Go up to the '{' that opens its body;
// its bodies hold only markup; a '}' closes a body anywhere on a line; else
// follows that '}' on its line; in the body of a switch or select, a line
// starting with case or default opens a clause, Go up to its ':'.
//
// The Go parts of a statement (header, else, clause headers, closing
// braces) are kept verbatim for the mask, so that go/parser parses them in
// place (LAB-005). A statement whose header is not valid Go keeps none of
// them: its error is reported here, and go/parser never sees it, so it
// cannot cascade into more errors.

import (
	"bytes"
	goast "go/ast"
	goparser "go/parser"
	"go/scanner"
	"go/token"

	"github.com/sgq995/golpe/internal/ast"
)

// embedded parses the embedded statement whose keyword, if, for, switch or
// select, is the first word of the line at p.off, in the content of f, and
// returns it, or nil after an error: the line is then text.
//
// If the header is not valid Go, the error carries the quoting hint. When
// the keyword's line ends with '{', the statement is parsed all the same, so
// that its body and closing '}' do not cause more errors, but none of its
// Go parts is kept for the mask; otherwise the line is text.
func (p *markupParser) embedded(f *frame, keyword string) ast.Node {
	kw := p.off
	lbrace, ok := p.header(kw, keyword)
	if !ok {
		p.errorf(kw, `invalid %s header: a line starting with '%[1]s' is an embedded statement; write {"%[1]s"} to start text with it`, keyword)
		if lbrace < 0 {
			return nil
		}
	}
	switch keyword {
	case "if":
		return p.ifStmt(f, kw, lbrace, !ok)
	case "for":
		return p.forStmt(f, kw, lbrace, !ok)
	case "switch":
		s := &ast.Switch{Keyword: p.pos(kw), Lbrace: p.pos(lbrace)}
		s.Cases, s.Rbrace = p.clauses(f, s, kw, lbrace, false, !ok)
		return s
	default:
		s := &ast.Select{Keyword: p.pos(kw), Lbrace: p.pos(lbrace)}
		s.Cases, s.Rbrace = p.clauses(f, s, kw, lbrace, true, !ok)
		return s
	}
}

// ifStmt parses the if statement whose keyword is at kw and whose body
// opens at lbrace, with its else branch, in the content of f. With bad, none
// of its Go parts is kept.
func (p *markupParser) ifStmt(f *frame, kw, lbrace int, bad bool) *ast.If {
	s := &ast.If{Keyword: p.pos(kw)}
	p.keepHeader(s, kw, lbrace+1, bad)
	s.Body = p.block(f, lbrace, bad)
	if !s.Body.Rbrace.IsValid() {
		return s
	}
	rbrace := p.off - 1
	if bad {
		rbrace = -1
	}
	if e := p.skipBlanks(p.off); p.wordAt(e) == "else" {
		p.elseBranch(f, s, e, rbrace)
	} else if !bad {
		p.endStmt(rbrace)
	}
	return s
}

// elseBranch parses the else branch of s, whose else keyword is at e:
// "else {…}" or "else if …", on the same line. rbrace is the offset of the
// '}' that the else follows when it is kept, and -1 when it is not (s is
// not kept, or the else does not follow a '}' on its line): then none of
// the Go parts of the branch is kept either, nor are they when the header of
// an else if is invalid. After an error that leaves no branch, s.Else is nil
// and p.off has not moved.
func (p *markupParser) elseBranch(f *frame, s *ast.If, e, rbrace int) {
	i := p.skipBlanks(e + len("else"))
	block := i < len(p.src) && p.src[i] == '{'
	lbrace, ok := i, true
	switch {
	case block:
	case p.wordAt(i) == "if":
		if lbrace, ok = p.header(i, "if"); !ok {
			p.errorf(i, "invalid if header")
		}
	default:
		p.errorf(i, "expected if statement or block after else")
		lbrace, ok = -1, false
	}
	kept := ok && rbrace >= 0
	switch {
	case kept && block:
		p.keepGo(e, i+1) // "else {"
	case kept:
		p.keepGo(e, e+len("else"))
	case rbrace >= 0:
		p.endStmt(rbrace) // in the mask, s ends at that '}'
	}
	switch {
	case lbrace < 0:
	case block:
		b := p.block(f, i, !kept)
		s.Else = b
		if b.Rbrace.IsValid() {
			p.closeStmt(f, p.off-1, !kept)
		}
	default:
		s.Else = p.ifStmt(f, i, lbrace, !kept)
	}
}

// forStmt parses the for statement whose keyword is at kw and whose body
// opens at lbrace, in the content of f. With bad, none of its Go parts is
// kept.
func (p *markupParser) forStmt(f *frame, kw, lbrace int, bad bool) *ast.For {
	s := &ast.For{Keyword: p.pos(kw)}
	p.keepHeader(s, kw, lbrace+1, bad)
	s.Body = p.block(f, lbrace, bad)
	if s.Body.Rbrace.IsValid() {
		p.closeStmt(f, p.off-1, bad)
	}
	return s
}

// block parses the body at lbrace of an if, else or for in the content of
// f, up to its closing '}', which is kept unless bad.
func (p *markupParser) block(f *frame, lbrace int, bad bool) *ast.Block {
	b := &ast.Block{Lbrace: p.pos(lbrace)}
	p.off = lbrace + 1
	if f.pre {
		p.skipLineEnd()
	}
	bf := &frame{body: true, pre: f.pre}
	b.List = p.content(bf)
	if bf.rbrace >= 0 {
		b.Rbrace = p.pos(bf.rbrace)
		if !bad {
			p.keepGo(bf.rbrace, bf.rbrace+1)
		}
	}
	return b
}

// clauses parses the body at lbrace of the switch or select s (sel) whose
// keyword is at kw, in the content of f: its case and default clauses up to
// its closing '}'. Content before the first clause is an error, and is
// dropped with its Go parts. With bad, none of the Go parts of s or of its
// clauses is kept.
func (p *markupParser) clauses(f *frame, s ast.Node, kw, lbrace int, sel, bad bool) (cases []*ast.Case, rbrace token.Pos) {
	p.keepHeader(s, kw, lbrace+1, bad)
	p.off = lbrace + 1
	if f.pre {
		p.skipLineEnd()
	}
	cf := &frame{body: true, clauses: true, sel: sel, bad: bad, pre: f.pre}
	errs, keep, semis, interps, stmts := len(p.errs.list), len(p.r.keep), len(p.r.semis), len(p.r.interps), len(p.r.stmts)
	if off := p.firstContent(p.content(cf)); off >= 0 {
		if len(p.errs.list) == errs { // one error is enough
			p.errorf(off, "expected case or default")
		}
		p.r.keep, p.r.semis, p.r.interps, p.r.stmts = p.r.keep[:keep], p.r.semis[:semis], p.r.interps[:interps], p.r.stmts[:stmts]
	}
	for cf.next != nil {
		c := cf.next
		p.keepHeader(c, p.file.Offset(c.Keyword), p.file.Offset(c.Colon)+1, bad)
		if f.pre {
			p.skipLineEnd()
		}
		c.Body = p.content(cf)
		cases = append(cases, c)
	}
	if cf.rbrace >= 0 {
		rbrace = p.pos(cf.rbrace)
		if !bad {
			p.keepGo(cf.rbrace, cf.rbrace+1)
		}
		p.closeStmt(f, cf.rbrace, bad)
	}
	return cases, rbrace
}

// clause parses the header of the case or default clause (word) that is
// the first word of the line at p.off, in the body f of a switch or select,
// up to its ':', and returns the clause, without body; clauses keeps its
// header. It returns nil after an error, leaving p.off: the line is then
// text.
func (p *markupParser) clause(f *frame, word string) *ast.Case {
	kw := p.off
	colon := p.colon(kw)
	if colon < 0 || !validClause(f.sel, p.src[kw:colon+1]) {
		p.errorf(kw, `invalid %s clause: a line starting with '%[1]s' opens a clause; write {"%[1]s"} to start text with it`, word)
		return nil
	}
	p.off = colon + 1
	return &ast.Case{Keyword: p.pos(kw), Default: word == "default", Colon: p.pos(colon)}
}

// keepHeader keeps src[from:to], the header of the embedded statement or
// clause n, for the mask, and records n for the graft, unless bad.
func (p *markupParser) keepHeader(n ast.Node, from, to int, bad bool) {
	if !bad {
		p.keepGo(from, to)
		p.r.stmts = append(p.r.stmts, n)
	}
}

// closeStmt ends the embedded statement in the content of f whose last
// '}', kept unless bad, is at rbrace: a for, switch or select, or an if
// whose last branch is an else block. An else after it on the line is an
// error, as in Go; the branch is parsed for recovery and dropped.
func (p *markupParser) closeStmt(f *frame, rbrace int, bad bool) {
	if !bad {
		p.endStmt(rbrace)
	}
	if e := p.skipBlanks(rbrace + 1); p.wordAt(e) == "else" {
		p.errorf(e, "else without if")
		p.elseBranch(f, &ast.If{}, e, -1)
	}
}

// endStmt follows the last '}' of an embedded statement, at rbrace and
// kept. If the line goes on after it, Go needs a ';' before what follows
// ("if ok { } x;" is not Go), which the mask writes on the next byte: not a
// kept byte, since a Go comment or an interpolation expression cannot start
// right after a '}'. A '}' that follows needs none.
func (p *markupParser) endStmt(rbrace int) {
	if i := rbrace + 1; !p.endsLine(i) && p.src[i] != '}' {
		p.r.semis = append(p.r.semis, i)
	}
}

// maxSemis is the number of ';' that a header may hold outside parentheses,
// brackets and braces: if and switch have an optional init statement, for
// has three clauses and select has no header.
var maxSemis = map[string]int{"if": 1, "switch": 1, "for": 2, "select": 0}

// header finds the '{' that opens the body of the embedded statement whose
// keyword, if, for, switch or select, is at kw, and reports whether the
// header is valid Go.
//
// The first '{' after the keyword is not always the body's: it may open a
// composite literal (for _, x := range []int{1, 2} {) or a function literal
// (if f := func() bool { return true }; f() {). So every '{' token outside
// parentheses, brackets and braces is a candidate, in order, and the first
// one such that the source from the keyword up to it, followed by "{}",
// parses as the statement with its body there ends the header: Go's own
// rule, decided by go/parser. The search stops at the end of the source, at
// an unmatched closing bracket, at more ';' (explicit or at the end of a
// line) than the header may hold, or at "</", which Go never has (so a
// header is never looked for past the close tag of its element).
//
// When no candidate is valid, lbrace is the '{' that ends the keyword's line,
// for recovery (see lineBrace), or -1.
func (p *markupParser) header(kw int, keyword string) (lbrace int, ok bool) {
	s, file := p.scan(kw, len(p.src))
	s.Scan() // the keyword
	depth, semis := 0, 0
	for {
		pos, tok, _ := s.Scan()
		switch tok {
		case token.EOF:
			return p.lineBrace(kw), false
		case token.LSS:
			if p.closeTagAt(kw + file.Offset(pos)) {
				return p.lineBrace(kw), false
			}
		case token.LPAREN, token.LBRACK:
			depth++
		case token.RPAREN, token.RBRACK, token.RBRACE:
			if depth--; depth < 0 {
				return p.lineBrace(kw), false
			}
		case token.SEMICOLON:
			if depth > 0 {
				break
			}
			if semis++; semis > maxSemis[keyword] {
				return p.lineBrace(kw), false
			}
		case token.LBRACE:
			if depth == 0 {
				if c := kw + file.Offset(pos); validHeader(p.src[kw:c]) {
					return c, true
				}
			}
			depth++
		}
	}
}

// lineBrace returns the offset of the '{' that ends the line at kw, the
// last Go token on it (a // comment may follow), or -1. When the header of
// a statement is invalid and its line ends with '{', the statement is most
// likely meant (if x = 1 {, for x in xs {, if f(x {), so its body is parsed
// for recovery.
func (p *markupParser) lineBrace(kw int) int {
	eol := len(p.src)
	if i := bytes.IndexByte(p.src[kw:], '\n'); i >= 0 {
		eol = kw + i
	}
	s, file := p.scan(kw, eol)
	last, at := token.ILLEGAL, -1
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok != token.SEMICOLON || lit != "\n" { // not inserted at the end
			last, at = tok, kw+file.Offset(pos)
		}
	}
	if last != token.LBRACE {
		return -1
	}
	return at
}

// colon returns the offset of the ':' that ends the header of the case or
// default clause at kw: the first ':' token outside parentheses, brackets
// and braces (case a[1:2] == b:, case T{K: v}.F:). It returns -1 if the end
// of the source, an unmatched closing bracket, a ';' outside them (explicit
// or at the end of a line) or "</" comes first.
func (p *markupParser) colon(kw int) int {
	s, file := p.scan(kw, len(p.src))
	depth := 0
	for {
		pos, tok, _ := s.Scan()
		switch tok {
		case token.EOF:
			return -1
		case token.LSS:
			if p.closeTagAt(kw + file.Offset(pos)) {
				return -1
			}
		case token.LPAREN, token.LBRACK, token.LBRACE:
			depth++
		case token.RPAREN, token.RBRACK, token.RBRACE:
			if depth--; depth < 0 {
				return -1
			}
		case token.SEMICOLON:
			if depth == 0 {
				return -1
			}
		case token.COLON:
			if depth == 0 {
				return kw + file.Offset(pos)
			}
		}
	}
}

// closeTagAt reports whether the '<' at src[i] starts "</" (a close tag),
// which Go never has: '/' cannot start an operand, and "<//" or "</*" is a
// '<' and a comment.
func (p *markupParser) closeTagAt(i int) bool {
	rest := p.src[i:]
	return bytes.HasPrefix(rest, []byte("</")) &&
		!bytes.HasPrefix(rest, []byte("<//")) && !bytes.HasPrefix(rest, []byte("</*"))
}

// scan returns a go/scanner over src[from:to], with the file it reports
// offsets in. Go errors are go/parser's to report, on the mask.
func (p *markupParser) scan(from, to int) (*scanner.Scanner, *token.File) {
	src := p.src[from:to]
	file := token.NewFileSet().AddFile("", -1, len(src))
	s := new(scanner.Scanner)
	s.Init(file, src, nil, 0)
	return s, file
}

// validHeader reports whether header, the source of an embedded statement
// from its keyword up to a '{', followed by "{}" parses as that statement
// with its body at the '{'.
func validHeader(header []byte) bool {
	const prefix = "package p;func _(){"
	s, ok := parseOneStmt(prefix + string(header) + "{}}")
	if !ok {
		return false
	}
	var body *goast.BlockStmt
	switch s := s.(type) {
	case *goast.IfStmt:
		body = s.Body
	case *goast.ForStmt:
		body = s.Body
	case *goast.RangeStmt:
		body = s.Body
	case *goast.SwitchStmt:
		body = s.Body
	case *goast.TypeSwitchStmt:
		body = s.Body
	case *goast.SelectStmt:
		body = s.Body
	default:
		return false
	}
	return int(body.Lbrace)-1 == len(prefix)+len(header)
}

// validClause reports whether header, the source of a case or default
// clause up to the first ':' outside brackets (see colon), parses as a
// clause of a switch, or of a select with sel. That ':' is then the
// clause's own: go/parser can take no other ':' before it as the end of
// the clause.
func validClause(sel bool, header []byte) bool {
	prefix := "package p;func _(){switch{"
	if sel {
		prefix = "package p;func _(){select{"
	}
	_, ok := parseOneStmt(prefix + string(header) + "}}")
	return ok
}

// parseOneStmt parses src, a Go file of one function, and returns the only
// statement of its body; it reports false if there is not exactly one, or
// after any syntax error. Positions are offsets plus one (the file's base).
func parseOneStmt(src string) (goast.Stmt, bool) {
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, goparser.SkipObjectResolution)
	if err != nil || len(f.Decls) != 1 {
		return nil, false
	}
	fd, ok := f.Decls[0].(*goast.FuncDecl)
	if !ok || fd.Body == nil || len(fd.Body.List) != 1 {
		return nil, false
	}
	return fd.Body.List[0], true
}

// danglingIf returns the if statement that an else line after list
// continues, for recovery: the last if of the chain that ends list, if it
// has no else yet; nil otherwise. Its body is closed: a body that is not
// ends the content that holds the statement.
func danglingIf(list []ast.Node) *ast.If {
	var s *ast.If
	if n := len(list); n > 0 {
		s, _ = list[n-1].(*ast.If)
	}
	for s != nil && s.Else != nil {
		s, _ = s.Else.(*ast.If)
	}
	return s
}

// firstContent returns the offset of the first byte of list, content
// before the first clause of a switch or select, that is not whitespace, or
// -1.
func (p *markupParser) firstContent(list []ast.Node) int {
	for _, n := range list {
		off := p.file.Offset(n.Pos())
		if t, ok := n.(*ast.Text); ok {
			i := 0
			for i < len(t.Raw) && isSpace(t.Raw[i]) {
				i++
			}
			if i == len(t.Raw) {
				continue
			}
			off += i
		}
		return off
	}
	return -1
}

// skipLineEnd skips the rest of the line at p.off, with its newline, if
// only spaces, tabs and a // comment, kept as Go, are left on it. It reports
// whether it did. In preformatted content, the end of a line of an embedded
// statement (after a header's '{', a clause's ':' or a closing '}') belongs
// to the statement, not to the content.
func (p *markupParser) skipLineEnd() bool {
	if !p.endsLine(p.off) {
		return false
	}
	for p.off < len(p.src) && p.src[p.off] != '\n' && p.src[p.off] != '/' {
		p.off++
	}
	if p.off < len(p.src) && p.src[p.off] == '/' {
		p.goComment()
	}
	if p.off < len(p.src) {
		p.off++ // the newline
	}
	return true
}

// endsLine reports whether only spaces, tabs and a // comment are left on
// the line at src[i:].
func (p *markupParser) endsLine(i int) bool {
	for i < len(p.src) && (p.src[i] == ' ' || p.src[i] == '\t' || p.src[i] == '\r') {
		i++
	}
	return i == len(p.src) || p.src[i] == '\n' || bytes.HasPrefix(p.src[i:], []byte("//"))
}

// skipBlanks returns the offset of the first byte from src[i:] that is not
// a space or a tab.
func (p *markupParser) skipBlanks(i int) int {
	for i < len(p.src) && (p.src[i] == ' ' || p.src[i] == '\t') {
		i++
	}
	return i
}
