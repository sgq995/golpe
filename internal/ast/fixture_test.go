package ast

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"
)

// pageSrc exercises every node type. Markup sits inside a Go if, a function
// literal, a Go case clause and the template body, and inside a function
// literal that is itself an interpolation.
const pageSrc = `package p

template Page[T any](title string, items []T, sel bool, ch chan string) {
	n := len(items)
	if n > 0 {
		<ul class="list" hidden data-n={n}>
		for i, it := range items {
			<li>{i}: {it}</li>
		}
		</ul>
	}
	f := func() element {
		<p>inner</p>
		return nil
	}
	switch {
	case sel:
		<hr/>
	}
	<!DOCTYPE html>
	<div>
		<!-- note -->
		{f()}
		{func() element { <b>lit</b> }}
		if sel {
			<b>on</b>
		} else if n == 0 {
			<i>empty</i>
		} else {
			off
		}
		switch x := n; x {
		case 0, 1:
			<span>few</span>
		default:
			{title}
		}
		select {
		case v := <-ch:
			{v}
		default:
		}
		<br>
	</div>
}
`

// builder builds markup trees the way the markup parser will: positions come
// from a cursor moving forward through the source, so nodes are built in
// source order.
type builder struct {
	t     testing.TB
	src   string
	base  int // base of the token.File holding src
	off   int // cursor
	nodes []goast.Node
}

func (b *builder) add(n goast.Node) { b.nodes = append(b.nodes, n) }

// at returns the position of the next occurrence of s and moves the cursor
// past it.
func (b *builder) at(s string) token.Pos {
	b.t.Helper()
	i := strings.Index(b.src[b.off:], s)
	if i < 0 {
		b.t.Fatalf("fixture: %q not found after offset %d", s, b.off)
	}
	p := token.Pos(b.base + b.off + i)
	b.off += i + len(s)
	return p
}

func (b *builder) startTag(name string, attrs []func() *Attr) *Element {
	b.t.Helper()
	e := &Element{Lt: b.at("<" + name), Name: name}
	if c := b.src[b.off]; !strings.ContainsRune(" >/\n\t", rune(c)) {
		b.t.Fatalf("fixture: <%s matched a longer tag name", name)
	}
	for _, a := range attrs {
		e.Attrs = append(e.Attrs, a())
	}
	b.add(e)
	return e
}

// el builds <name attrs>kids</name>.
func (b *builder) el(name string, attrs []func() *Attr, kids ...func() Node) *Element {
	e := b.startTag(name, attrs)
	b.at(">")
	for _, k := range kids {
		e.Children = append(e.Children, k())
	}
	e.Close = b.at("</" + name)
	e.EndPos = b.at(">") + 1
	return e
}

// void builds a start tag without an end tag, like <br>.
func (b *builder) void(name string) *Element {
	e := b.startTag(name, nil)
	e.EndPos = b.at(">") + 1
	return e
}

func (b *builder) selfClosing(name string) *Element {
	e := b.startTag(name, nil)
	e.SelfClosing = true
	e.EndPos = b.at("/>") + 2
	return e
}

func (b *builder) static(name, raw string) func() *Attr {
	return func() *Attr {
		a := &Attr{NamePos: b.at(name), Name: name}
		v := &StaticValue{ValuePos: b.at(raw), Raw: raw, Value: raw[1 : len(raw)-1]}
		a.Value = v
		b.add(a)
		b.add(v)
		return a
	}
}

func (b *builder) boolean(name string) func() *Attr {
	return func() *Attr {
		a := &Attr{NamePos: b.at(name), Name: name}
		b.add(a)
		return a
	}
}

func (b *builder) dynamic(name string) func() *Attr {
	return func() *Attr {
		a := &Attr{NamePos: b.at(name), Name: name}
		a.Value = b.interp()
		b.add(a)
		return a
	}
}

// interp builds an interpolation without nested braces.
func (b *builder) interp() *Interp {
	in := &Interp{Lbrace: b.at("{"), Rbrace: b.at("}")}
	b.add(in)
	return in
}

func (b *builder) text(raw, value string) func() Node {
	return func() Node {
		t := &Text{ValuePos: b.at(raw), Raw: raw, Value: value}
		b.add(t)
		return t
	}
}

func (b *builder) block(kids ...func() Node) *Block {
	bl := &Block{Lbrace: b.at("{")}
	for _, k := range kids {
		bl.List = append(bl.List, k())
	}
	bl.Rbrace = b.at("}")
	b.add(bl)
	return bl
}

func (b *builder) clause(header string, def bool, kids ...func() Node) *Case {
	c := &Case{Keyword: b.at(header), Default: def}
	c.Colon = b.at(":")
	for _, k := range kids {
		c.Body = append(c.Body, k())
	}
	b.add(c)
	return c
}

// markup wraps root in a MarkupStmt with a provisional placeholder; graft
// swaps in the EmptyStmt go/parser built.
func (b *builder) markup(root Node) *MarkupStmt {
	s := &MarkupStmt{EmptyStmt: &goast.EmptyStmt{Semicolon: root.Pos()}, Root: root}
	b.add(s)
	return s
}

// elem defers el, so that it runs when its parent reaches it.
func (b *builder) elem(name string, attrs []func() *Attr, kids ...func() Node) func() Node {
	return func() Node { return b.el(name, attrs, kids...) }
}

// page is pageSrc parsed the way PLAN-001 describes: markup trees built
// first, then masked, parsed once by go/parser and grafted.
type page struct {
	fset  *token.FileSet
	base  int
	mask  []byte
	file  *goast.File
	tmpl  *TemplateDecl
	nodes []goast.Node // every Golpe node, TemplateDecl included

	ul, p, hr, doctype, div, lit *MarkupStmt
	li, br, hrEl                 *Element
	attrs                        []*Attr // class, hidden, data-n
	ifSel, ifElse                *If
	elseBlock                    *Block
	off                          *Text
	forNode                      *For
	sw                           *Switch
	sel                          *Select
	cases                        []*Case // switch case, switch default, select case, select default
	funcInterp                   *Interp
	comment                      *Comment
}

func newPage(t testing.TB) *page {
	t.Helper()
	pg := &page{fset: token.NewFileSet()}
	// go/parser will add the file at the current base; markup positions are
	// computed against it before parsing, as the real parser must do.
	pg.base = pg.fset.Base()
	b := &builder{t: t, src: pageSrc, base: pg.base}

	b.at("if n > 0 {")
	pg.ul = b.markup(b.el("ul",
		[]func() *Attr{b.static("class", `"list"`), b.boolean("hidden"), b.dynamic("data-n")},
		func() Node {
			pg.forNode = &For{Keyword: b.at("for i, it := range items")}
			pg.forNode.Body = b.block(func() Node {
				pg.li = b.el("li", nil, func() Node { return b.interp() }, b.text(": ", ": "), func() Node { return b.interp() })
				return pg.li
			})
			b.add(pg.forNode)
			return pg.forNode
		}))
	pg.attrs = pg.ul.Root.(*Element).Attrs
	pg.p = b.markup(b.el("p", nil, b.text("inner", "inner")))
	pg.hrEl = b.selfClosing("hr")
	pg.hr = b.markup(pg.hrEl)
	pg.doctype = b.markup(func() Node {
		d := &Doctype{Start: b.at("<!DOCTYPE html>"), Text: "<!DOCTYPE html>"}
		b.add(d)
		return d
	}())
	pg.div = b.markup(b.el("div", nil,
		func() Node {
			pg.comment = &Comment{Start: b.at("<!-- note -->"), Text: "<!-- note -->"}
			b.add(pg.comment)
			return pg.comment
		},
		func() Node { return b.interp() },
		func() Node {
			pg.funcInterp = &Interp{Lbrace: b.at("{func")}
			pg.lit = b.markup(b.el("b", nil, b.text("lit", "lit")))
			pg.funcInterp.Rbrace = b.at("}}") + 1
			b.add(pg.funcInterp)
			return pg.funcInterp
		},
		func() Node {
			pg.ifSel = &If{Keyword: b.at("if sel")}
			pg.ifSel.Body = b.block(b.elem("b", nil, b.text("on", "on")))
			pg.ifElse = &If{Keyword: b.at("if n == 0")}
			pg.ifElse.Body = b.block(b.elem("i", nil, b.text("empty", "empty")))
			pg.ifElse.Else = b.block(func() Node {
				pg.off = b.text("\n\t\t\toff\n\t\t", "off")().(*Text)
				return pg.off
			})
			pg.elseBlock = pg.ifElse.Else.(*Block)
			pg.ifSel.Else = pg.ifElse
			b.add(pg.ifSel)
			b.add(pg.ifElse)
			return pg.ifSel
		},
		func() Node {
			pg.sw = &Switch{Keyword: b.at("switch x := n; x"), Lbrace: b.at("{")}
			pg.sw.Cases = []*Case{
				b.clause("case 0, 1", false, b.elem("span", nil, b.text("few", "few"))),
				b.clause("default", true, func() Node { return b.interp() }),
			}
			pg.sw.Rbrace = b.at("}")
			b.add(pg.sw)
			return pg.sw
		},
		func() Node {
			pg.sel = &Select{Keyword: b.at("select"), Lbrace: b.at("{")}
			pg.sel.Cases = []*Case{
				b.clause("case v := <-ch", false, func() Node { return b.interp() }),
				b.clause("default", true),
			}
			pg.sel.Rbrace = b.at("}")
			b.add(pg.sel)
			return pg.sel
		},
		func() Node { pg.br = b.void("br"); return pg.br },
	))
	pg.cases = append(slices.Clone(pg.sw.Cases), pg.sel.Cases...)

	stmts := []*MarkupStmt{pg.ul, pg.p, pg.hr, pg.doctype, pg.div, pg.lit}
	pg.file, pg.tmpl, pg.mask = parseSkeleton(t, pg.fset, pageSrc, stmts)
	pg.nodes = append(b.nodes, pg.tmpl)
	return pg
}

// parseSkeleton masks src around the markup statements stmts (built against
// the file set's current base, in source order), parses the skeleton with
// go/parser and grafts the markup in.
func parseSkeleton(t testing.TB, fset *token.FileSet, src string, stmts []*MarkupStmt) (*goast.File, *TemplateDecl, []byte) {
	t.Helper()
	base := fset.Base()
	m := mask(t, src, base, stmts)
	f, err := parser.ParseFile(fset, "x.golpe", m, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("skeleton does not parse: %v\n%s", err, m)
	}
	if got := fset.File(f.Pos()).Base(); got != base {
		t.Fatalf("go/parser used base %d, markup positions assumed %d", got, base)
	}
	return f, graft(t, f, src, base, stmts), m
}

// mask is a test-only version of PLAN-001's pass 3, driven by already built
// markup trees: markup bytes become spaces (newlines kept) except the Go
// fragments, the first byte of each markup statement becomes ";" and an
// interpolation {x} becomes " x;". stmts must be in source order, so that a
// markup statement nested in an interpolation is masked after its parent.
func mask(t testing.TB, src string, base int, stmts []*MarkupStmt) []byte {
	out := []byte(src)
	off := func(p token.Pos) int { return int(p) - base }
	for _, ms := range stmts {
		keep := make([]bool, len(src))
		span := func(from, to token.Pos) {
			for i := off(from); i < off(to); i++ {
				keep[i] = true
			}
		}
		var interps []*Interp
		Inspect(ms.Root, func(n goast.Node) bool {
			switch n := n.(type) {
			case *Interp:
				span(n.Pos(), n.End())
				interps = append(interps, n)
				return false
			case *If:
				span(n.Keyword, n.Body.Pos())
				if n.Else != nil {
					span(n.Body.End(), n.Else.Pos()) // "else"
				}
			case *For:
				span(n.Keyword, n.Body.Pos())
			case *Switch:
				span(n.Keyword, n.Lbrace+1)
				span(n.Rbrace, n.Rbrace+1)
			case *Select:
				span(n.Keyword, n.Lbrace+1)
				span(n.Rbrace, n.Rbrace+1)
			case *Case:
				span(n.Keyword, n.Colon+1)
			case *Block:
				span(n.Lbrace, n.Lbrace+1)
				span(n.Rbrace, n.Rbrace+1)
			}
			return true
		})
		for i := off(ms.Pos()); i < off(ms.End()); i++ {
			if !keep[i] && out[i] != '\n' {
				out[i] = ' '
			}
		}
		out[off(ms.Pos())] = ';'
		for _, in := range interps {
			out[off(in.Lbrace)], out[off(in.Rbrace)] = ' ', ';'
		}
	}
	i := strings.Index(src, "template ")
	copy(out[i:], "func    ")
	if len(out) != len(src) {
		t.Fatal("mask changed the length")
	}
	return out
}

// graft is a test-only version of PLAN-001's pass 5. It attaches the go/ast
// nodes go/parser built from the mask to the markup trees, replaces each
// placeholder EmptyStmt by its MarkupStmt (dropping the skeleton statements
// of the region from that list) and wraps the template's FuncDecl. Skeleton
// bodies of embedded statements are left untouched on purpose: they share
// expressions with the interpolations, so a walker that descended into them
// would visit those expressions twice.
func graft(t testing.TB, f *goast.File, src string, base int, stmts []*MarkupStmt) *TemplateDecl {
	t.Helper()
	at := map[token.Pos]goast.Stmt{}
	var lists []*[]goast.Stmt
	goast.Inspect(f, func(n goast.Node) bool { // pure go/ast until mutated below
		switch n := n.(type) {
		case *goast.BlockStmt:
			lists = append(lists, &n.List)
		case *goast.CaseClause:
			lists = append(lists, &n.Body)
		case *goast.CommClause:
			lists = append(lists, &n.Body)
		}
		if s, ok := n.(goast.Stmt); ok {
			if _, dup := at[s.Pos()]; !dup {
				at[s.Pos()] = s
			}
		}
		return true
	})
	stmtAt := func(p token.Pos) goast.Stmt {
		t.Helper()
		s := at[p]
		if s == nil {
			t.Fatalf("graft: no skeleton statement at offset %d", int(p)-base)
		}
		return s
	}
	for _, ms := range stmts {
		Inspect(ms.Root, func(n goast.Node) bool {
			switch n := n.(type) {
			case *Interp:
				for p := n.Lbrace + 1; p < n.Rbrace && n.X == nil; p++ {
					if es, ok := at[p].(*goast.ExprStmt); ok {
						n.X = es.X
					}
				}
				if n.X == nil {
					t.Fatalf("graft: no expression for interpolation at offset %d", int(n.Lbrace)-base)
				}
				return false
			case *If:
				n.Stmt = stmtAt(n.Keyword).(*goast.IfStmt)
			case *For:
				n.Stmt = stmtAt(n.Keyword)
			case *Switch:
				n.Stmt = stmtAt(n.Keyword)
			case *Select:
				n.Stmt = stmtAt(n.Keyword).(*goast.SelectStmt)
			case *Case:
				n.Clause = stmtAt(n.Keyword)
			}
			return true
		})
	}
	starts := map[token.Pos]*MarkupStmt{}
	for _, ms := range stmts {
		starts[ms.Pos()] = ms
	}
	inRegion := func(p token.Pos) bool {
		for _, ms := range stmts {
			if ms.Pos() <= p && p < ms.End() {
				return true
			}
		}
		return false
	}
	placeholder := func(s goast.Stmt) *MarkupStmt {
		if e, ok := s.(*goast.EmptyStmt); ok && !e.Implicit && starts[e.Pos()] != nil {
			starts[e.Pos()].EmptyStmt = e
			return starts[e.Pos()]
		}
		return nil
	}
	for _, list := range lists {
		if !slices.ContainsFunc(*list, func(s goast.Stmt) bool { return placeholder(s) != nil }) {
			continue
		}
		var out []goast.Stmt
		for _, s := range *list {
			if ms := placeholder(s); ms != nil {
				out = append(out, ms)
			} else if !inRegion(s.Pos()) {
				out = append(out, s)
			}
		}
		*list = out
	}
	for _, ms := range stmts {
		if ms.EmptyStmt.Implicit || ms.Semicolon != ms.Root.Pos() {
			t.Fatalf("graft: bad placeholder for markup at offset %d", int(ms.Root.Pos())-base)
		}
	}
	for i, d := range f.Decls {
		if fd, ok := d.(*goast.FuncDecl); ok && strings.HasPrefix(src[int(fd.Pos())-base:], "template") {
			td := &TemplateDecl{FuncDecl: fd, Template: fd.Pos()}
			f.Decls[i] = td
			return td
		}
	}
	t.Fatal("graft: no template declaration")
	return nil
}

// TestPageFixture checks the fixture itself: the skeleton keeps length and
// newlines, has the shape of PLAN-001's mask (shown with whitespace runs
// collapsed), and the graft left six statements in the template body.
func TestPageFixture(t *testing.T) {
	pg := newPage(t)
	for i := range pageSrc {
		if (pageSrc[i] == '\n') != (pg.mask[i] == '\n') {
			t.Fatalf("mask moved a newline at offset %d", i)
		}
	}
	var lines []string
	for _, l := range strings.Split(string(pg.mask), "\n") {
		lines = append(lines, strings.Join(strings.Fields(l), " "))
	}
	want := `package p

func Page[T any](title string, items []T, sel bool, ch chan string) {
n := len(items)
if n > 0 {
; n;
for i, it := range items {
i; it;
}

}
f := func() element {
;
return nil
}
switch {
case sel:
;
}
;
;

f();
func() element { ; };
if sel {

} else if n == 0 {

} else {

}
switch x := n; x {
case 0, 1:

default:
title;
}
select {
case v := <-ch:
v;
default:
}


}
`
	if got := strings.Join(lines, "\n"); got != want {
		t.Errorf("mask (whitespace collapsed) =\n%s\nwant\n%s", got, want)
	}
	if got := len(pg.tmpl.Body.List); got != 6 {
		t.Errorf("template body has %d statements, want 6 (n :=, if, f :=, switch, doctype, div)", got)
	}
}
