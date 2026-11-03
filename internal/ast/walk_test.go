package ast

import (
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"reflect"
	"slices"
	"strings"
	"testing"
)

// TestGoASTWalkPanicsOnPage confirms LAB-003 on a realistic tree: go/ast's
// walker cannot be used on a grafted file, which is why this package exists.
func TestGoASTWalkPanicsOnPage(t *testing.T) {
	pg := newPage(t)
	defer func() {
		if recover() == nil {
			t.Fatal("go/ast.Inspect did not panic on a grafted file")
		}
	}()
	goast.Inspect(pg.file, func(goast.Node) bool { return true })
}

// TestGoASTWalkVisitsBeforeTypeSwitch pins the go/ast behaviour Walk relies
// on: go/ast.Walk calls Visit before switching on the node type and returns
// at once when Visit returns nil, so it never reaches its panicking default
// for a node the visitor handled itself.
func TestGoASTWalkVisitsBeforeTypeSwitch(t *testing.T) {
	var seen []goast.Node
	goast.Walk(visitFunc(func(n goast.Node) goast.Visitor {
		seen = append(seen, n)
		return nil
	}), &Element{Name: "div"})
	if len(seen) != 1 {
		t.Fatalf("Visit called %d times, want 1", len(seen))
	}
}

type visitFunc func(goast.Node) goast.Visitor

func (f visitFunc) Visit(n goast.Node) goast.Visitor { return f(n) }

// TestInspectPage walks a grafted file: no panic, every node exactly once,
// every Go identifier exactly once and in source order, and Golpe nodes in
// source order among them.
func TestInspectPage(t *testing.T) {
	pg := newPage(t)
	visits := map[goast.Node]int{}
	var order []goast.Node // Golpe nodes, identifiers and literals
	Inspect(pg.file, func(n goast.Node) bool {
		if n == nil {
			return true
		}
		visits[n]++
		switch n.(type) {
		case *goast.Ident, *goast.BasicLit:
			order = append(order, n)
		default:
			if isGolpe(n) {
				order = append(order, n)
			}
		}
		return true
	})
	for n, c := range visits {
		if c != 1 {
			t.Errorf("%T at %v visited %d times", n, pg.fset.Position(n.Pos()), c)
		}
	}
	for _, n := range pg.nodes {
		if visits[n] != 1 {
			t.Errorf("%T at %v not visited", n, pg.fset.Position(n.Pos()))
		}
	}
	for i := 1; i < len(order); i++ {
		if order[i].Pos() < order[i-1].Pos() {
			t.Errorf("%T at %v visited after %T at %v", order[i], pg.fset.Position(order[i].Pos()),
				order[i-1], pg.fset.Position(order[i-1].Pos()))
		}
	}

	// The identifiers of the skeleton, scanned in order, are exactly the
	// identifiers Inspect visits: header parts and interpolations once each,
	// skeleton bodies of embedded statements never.
	var want []int
	var s scanner.Scanner
	fs := token.NewFileSet()
	file := fs.AddFile("", fs.Base(), len(pg.mask))
	s.Init(file, pg.mask, nil, 0)
	for {
		p, tok, _ := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.IDENT {
			want = append(want, file.Offset(p))
		}
	}
	var got []int
	for _, n := range order {
		if id, ok := n.(*goast.Ident); ok {
			got = append(got, int(id.Pos())-pg.base)
		}
	}
	if !slices.Equal(got, want) {
		t.Errorf("identifier offsets visited\n%v\nwant (scanned from the skeleton)\n%v", got, want)
	}
}

// TestInspectAncestors checks markup reached through go/ast containers: a Go
// if, a function literal, a Go case clause, and a function literal used as
// an interpolation inside markup.
func TestInspectAncestors(t *testing.T) {
	pg := newPage(t)
	want := map[*MarkupStmt]string{
		pg.ul:      "File TemplateDecl BlockStmt IfStmt BlockStmt",
		pg.p:       "File TemplateDecl BlockStmt AssignStmt FuncLit BlockStmt",
		pg.hr:      "File TemplateDecl BlockStmt SwitchStmt BlockStmt CaseClause",
		pg.doctype: "File TemplateDecl BlockStmt",
		pg.div:     "File TemplateDecl BlockStmt",
		pg.lit:     "File TemplateDecl BlockStmt MarkupStmt Element Interp FuncLit BlockStmt",
	}
	var stack []goast.Node
	got := map[*MarkupStmt]string{}
	Inspect(pg.file, func(n goast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if ms, ok := n.(*MarkupStmt); ok {
			var names []string
			for _, a := range stack {
				names = append(names, reflect.TypeOf(a).Elem().Name())
			}
			got[ms] = strings.Join(names, " ")
		}
		stack = append(stack, n)
		return true
	})
	if len(stack) != 0 {
		t.Errorf("unbalanced Visit(nil) calls, %d left on the stack", len(stack))
	}
	for ms, w := range want {
		if got[ms] != w {
			t.Errorf("markup at %v: ancestors %q, want %q", pg.fset.Position(ms.Pos()), got[ms], w)
		}
	}
}

// recorder logs the Walk protocol: which visitor saw which node, and the
// Visit(nil) calls.
type recorder struct {
	name string
	log  *[]string
	skip func(goast.Node) bool // return nil for these nodes
	next *recorder             // visitor for children; itself if nil
}

func (r *recorder) Visit(n goast.Node) goast.Visitor {
	if n == nil {
		*r.log = append(*r.log, r.name+":end")
		return nil
	}
	*r.log = append(*r.log, fmt.Sprintf("%s:%s", r.name, describe(n)))
	if r.skip != nil && r.skip(n) {
		return nil
	}
	if r.next != nil {
		return r.next
	}
	return r
}

func describe(n goast.Node) string {
	switch n := n.(type) {
	case *goast.Ident:
		return n.Name
	case *Element:
		return "<" + n.Name + ">"
	case *Text:
		return fmt.Sprintf("%q", n.Value)
	}
	return reflect.TypeOf(n).Elem().Name()
}

// TestWalkProtocol checks that Walk keeps go/ast.Walk's contract on Golpe
// nodes: the visitor returned by Visit(node) walks the children and gets
// the closing Visit(nil); returning nil prunes the subtree and suppresses
// Visit(nil).
func TestWalkProtocol(t *testing.T) {
	tree := &MarkupStmt{Root: &Element{Name: "ul", Children: []Node{
		&Element{Name: "li", Children: []Node{&Text{Value: "a"}}},
		&Interp{X: &goast.Ident{Name: "x"}},
		&Element{Name: "b", Children: []Node{&Text{Value: "pruned"}}},
	}}}
	var log []string
	child := &recorder{name: "c", log: &log, skip: func(n goast.Node) bool {
		e, ok := n.(*Element)
		return ok && e.Name == "b"
	}}
	Walk(&recorder{name: "r", log: &log, next: child}, tree)
	want := []string{
		"r:MarkupStmt",
		"c:<ul>",
		"c:<li>", "c:\"a\"", "c:end", "c:end",
		"c:Interp", "c:x", "c:end", "c:end",
		"c:<b>",
		"c:end",
		"c:end",
	}
	if !slices.Equal(log, want) {
		t.Errorf("log\n%s\nwant\n%s", strings.Join(log, "\n"), strings.Join(want, "\n"))
	}

	// Pruning a go/ast node that contains markup.
	log = nil
	fn := &goast.FuncLit{Type: &goast.FuncType{}, Body: &goast.BlockStmt{List: []goast.Stmt{tree}}}
	Walk(&recorder{name: "r", log: &log, skip: func(n goast.Node) bool {
		_, ok := n.(*goast.BlockStmt)
		return ok
	}}, fn)
	want = []string{"r:FuncLit", "r:FuncType", "r:end", "r:BlockStmt", "r:end"}
	if !slices.Equal(log, want) {
		t.Errorf("log\n%s\nwant\n%s", strings.Join(log, "\n"), strings.Join(want, "\n"))
	}
}

// headerFixture parses Go statements whose bodies hold a "skeleton"
// identifier that a walker must never reach.
func headerFixture(t *testing.T, body string) []goast.Stmt {
	t.Helper()
	src := "package p\nfunc f() {\n" + body + "\n}\n"
	f, err := parser.ParseFile(token.NewFileSet(), "h.go", src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	return f.Decls[0].(*goast.FuncDecl).Body.List
}

// TestWalkEmbeddedHeaders checks that embedded statements expose only the
// header of their go/ast statement, in source order, then the markup body.
func TestWalkEmbeddedHeaders(t *testing.T) {
	stmts := headerFixture(t, `
	if a := b; c { skeleton } else { skeleton }
	for d := e; g; h++ { skeleton }
	for k, v := range m { skeleton }
	switch s := u; w { case x, y: skeleton }
	switch z := q.(type) { default: skeleton }
	select { case r := <-ch: skeleton; default: skeleton }
	for range n { skeleton }
	`)
	body := func(s string) *Block { return &Block{List: []Node{&Text{Value: s}}} }
	sw := stmts[3].(*goast.SwitchStmt)
	ts := stmts[4].(*goast.TypeSwitchStmt)
	sel := stmts[5].(*goast.SelectStmt)
	cases := []struct {
		name string
		n    Node
		want string
	}{
		{"If", &If{Stmt: stmts[0].(*goast.IfStmt), Body: body("then"), Else: body("else")},
			`If AssignStmt a b c Block "then" Block "else"`},
		{"If pre-graft", &If{Body: body("then")}, `If Block "then"`},
		{"For", &For{Stmt: stmts[1], Body: body("loop")}, `For AssignStmt d e g IncDecStmt h Block "loop"`},
		{"For range", &For{Stmt: stmts[2], Body: body("loop")}, `For k v m Block "loop"`},
		{"For range without key", &For{Stmt: stmts[6]}, `For n`},
		{"For bad header", &For{Stmt: &goast.BadStmt{}}, `For BadStmt`},
		{"For pre-graft", &For{}, `For`},
		{"Switch", &Switch{Stmt: sw, Cases: []*Case{{Clause: sw.Body.List[0], Body: []Node{&Text{Value: "one"}}}}},
			`Switch AssignStmt s u w Case x y "one"`},
		{"Switch type", &Switch{Stmt: ts, Cases: []*Case{{Clause: ts.Body.List[0], Default: true}}},
			`Switch AssignStmt z TypeAssertExpr q Case`},
		{"Select", &Select{Stmt: sel, Cases: []*Case{
			{Clause: sel.Body.List[0], Body: []Node{&Text{Value: "recv"}}},
			{Clause: sel.Body.List[1], Default: true},
		}}, `Select Case AssignStmt r UnaryExpr ch "recv" Case`},
		{"Select pre-graft", &Select{Cases: []*Case{{}}}, `Select Case`},
		{"Block", &Block{List: []Node{&Text{Value: "a"}, nil, &Text{Value: "b"}}}, `Block "a" "b"`},
	}
	for _, c := range cases {
		var got []string
		Inspect(c.n, func(n goast.Node) bool {
			if n != nil {
				got = append(got, describe(n))
			}
			return true
		})
		if s := strings.Join(got, " "); s != c.want {
			t.Errorf("%s: visited %s, want %s", c.name, s, c.want)
		}
	}
}

// TestWalkAllNodes visits a hand-built tree holding every node type with all
// optional parts present, and a TemplateDecl with its declaration parts.
func TestWalkAllNodes(t *testing.T) {
	stmts := headerFixture(t, "if c {}")
	decl := &TemplateDecl{FuncDecl: &goast.FuncDecl{
		Doc:  &goast.CommentGroup{List: []*goast.Comment{{Text: "// doc"}}},
		Recv: &goast.FieldList{List: []*goast.Field{{Names: []*goast.Ident{{Name: "c"}}, Type: &goast.Ident{Name: "Card"}}}},
		Name: &goast.Ident{Name: "View"},
		Type: &goast.FuncType{Params: &goast.FieldList{}},
		Body: &goast.BlockStmt{List: []goast.Stmt{&MarkupStmt{Root: &Element{
			Name: "a",
			Attrs: []*Attr{
				{Name: "href", Value: &Interp{X: &goast.Ident{Name: "url"}}},
				{Name: "class", Value: &StaticValue{Value: "x"}},
				{Name: "hidden"},
			},
			Children: []Node{
				&Comment{Text: "<!-- c -->"},
				&Doctype{Text: "<!DOCTYPE html>"},
				&If{Stmt: stmts[0].(*goast.IfStmt), Body: &Block{List: []Node{&Text{Value: "t"}}}},
			},
		}}}},
	}}
	var got []string
	Inspect(&goast.File{Name: &goast.Ident{Name: "p"}, Decls: []goast.Decl{decl}}, func(n goast.Node) bool {
		if n != nil {
			got = append(got, describe(n))
		}
		return true
	})
	want := `File p TemplateDecl CommentGroup Comment FieldList Field c Card View FuncType FieldList BlockStmt ` +
		`MarkupStmt <a> Attr Interp url Attr StaticValue Attr Comment Doctype If c Block "t"`
	if s := strings.Join(got, " "); s != want {
		t.Errorf("visited\n%s\nwant\n%s", s, want)
	}

	// Empty shells never panic.
	for _, n := range []goast.Node{&TemplateDecl{}, &TemplateDecl{FuncDecl: &goast.FuncDecl{}}, &MarkupStmt{},
		&Attr{}, &Interp{}, &If{}, &For{}, &Switch{}, &Select{}, &Case{}, &Block{}, &Element{}} {
		Inspect(n, func(goast.Node) bool { return true })
	}
}
