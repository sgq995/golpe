package ast

import (
	"errors"
	goast "go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func dump(t *testing.T, fset *token.FileSet, n goast.Node) string {
	t.Helper()
	var b strings.Builder
	if err := Fprint(&b, fset, n); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestFprintPage(t *testing.T) {
	pg := newPage(t)
	want := `File 1:1-45:2 package p
  TemplateDecl 3:1-45:2 template Page[T any](title string, items []T, sel bool, ch chan string)
    BlockStmt 3:73-45:2
      AssignStmt 4:2-4:17 n := len(items)
      IfStmt 5:2-11:3 if n > 0
        BlockStmt 5:11-11:3
          MarkupStmt 6:3-10:8
            Element 6:3-10:8 ul close=10:3
              Attr 6:7-6:19 class
                StaticValue 6:13-6:19 "list"
              Attr 6:20-6:26 hidden
              Attr 6:27-6:37 data-n
                Interp 6:34-6:37 {n}
              For 7:3-9:4 for i, it := range items
                Block 7:28-9:4
                  Element 8:4-8:22 li close=8:17
                    Interp 8:8-8:11 {i}
                    Text 8:11-8:13 ": "
                    Interp 8:13-8:17 {it}
      AssignStmt 12:2-15:3 :=
        Ident 12:2-12:3 f
        FuncLit 12:7-15:3 func() element
          BlockStmt 12:22-15:3
            MarkupStmt 13:3-13:15
              Element 13:3-13:15 p close=13:11
                Text 13:6-13:11 "inner"
            ReturnStmt 14:3-14:13 return nil
      SwitchStmt 16:2-19:3 switch
        BlockStmt 16:9-19:3
          CaseClause 17:2-18:8 case sel:
            MarkupStmt 18:3-18:8
              Element 18:3-18:8 hr self-closing
      MarkupStmt 20:2-20:17
        Doctype 20:2-20:17 "<!DOCTYPE html>"
      MarkupStmt 21:2-44:8
        Element 21:2-44:8 div close=44:2
          Comment 22:3-22:16 "<!-- note -->"
          Interp 23:3-23:8 {f()}
          Interp 24:3-24:34
            FuncLit 24:4-24:33 func() element
              BlockStmt 24:19-24:33
                MarkupStmt 24:21-24:31
                  Element 24:21-24:31 b close=24:27
                    Text 24:24-24:27 "lit"
          If 25:3-31:4 if sel
            Block 25:10-27:4
              Element 26:4-26:13 b close=26:9
                Text 26:7-26:9 "on"
            If 27:10-31:4 if n == 0
              Block 27:20-29:4
                Element 28:4-28:16 i close=28:12
                  Text 28:7-28:12 "empty"
              Block 29:10-31:4
                Text 29:11-31:3 "off"
          Switch 32:3-37:4 switch x := n; x
            Case 33:3-34:20 case 0, 1:
              Element 34:4-34:20 span close=34:13
                Text 34:10-34:13 "few"
            Case 35:3-36:11 default:
              Interp 36:4-36:11 {title}
          Select 38:3-42:4 select
            Case 39:3-40:7 case v := <-ch:
              Interp 40:4-40:7 {v}
            Case 41:3-41:11 default:
          Element 43:3-43:7 br
`
	if got := dump(t, pg.fset, pg.file); got != want {
		t.Errorf("dump:\n%s\nwant:\n%s", got, want)
	}
}

// TestFprintPreGraft dumps a markup tree as the markup parser builds it,
// before go/parser has run: Go parts are shown as "...".
func TestFprintPreGraft(t *testing.T) {
	src := "<ul>\n" +
		"\tfor i := range n {\n" +
		"\t\t<li>{i}</li>\n" +
		"\t}\n" +
		"\tswitch x {\n" +
		"\tcase 1:\n" +
		"\tdefault:\n" +
		"\t}\n" +
		"\tselect {\n" +
		"\tcase v := <-c:\n" +
		"\t}\n" +
		"\tif a {\n" +
		"\t}\n" +
		"</ul>"
	fset := token.NewFileSet()
	file := fset.AddFile("m.golpe", fset.Base(), len(src))
	file.SetLinesForContent([]byte(src))
	b := &builder{t: t, src: src, base: file.Base()}
	root := b.markup(b.el("ul", nil,
		func() Node {
			f := &For{Keyword: b.at("for")}
			f.Body = b.block(b.elem("li", nil, func() Node { return b.interp() }))
			return f
		},
		func() Node {
			s := &Switch{Keyword: b.at("switch x"), Lbrace: b.at("{")}
			s.Cases = []*Case{b.clause("case 1", false), b.clause("default", true)}
			s.Rbrace = b.at("}")
			return s
		},
		func() Node {
			s := &Select{Keyword: b.at("select"), Lbrace: b.at("{")}
			s.Cases = []*Case{b.clause("case v := <-c", false)}
			s.Rbrace = b.at("}")
			return s
		},
		func() Node {
			i := &If{Keyword: b.at("if a")}
			i.Body = b.block()
			return i
		},
	))
	want := `MarkupStmt 1:1-14:6
  Element 1:1-14:6 ul close=14:1
    For 2:2-4:3 for ...
      Block 2:19-4:3
        Element 3:3-3:15 li close=3:10
          Interp 3:7-3:10 {...}
    Switch 5:2-8:3 switch ...
      Case 6:2-6:9 case ...:
      Case 7:2-7:10 default:
    Select 9:2-11:3 select
      Case 10:2-10:16 case ...:
    If 12:2-13:3 if ...
      Block 12:7-13:3
`
	if got := dump(t, fset, root); got != want {
		t.Errorf("dump:\n%s\nwant:\n%s", got, want)
	}
}

// TestFprintHeaders checks the header text of every statement form against
// its gofmt'ed source line.
func TestFprintHeaders(t *testing.T) {
	src := `package p

func f() {
	if x {
	}
	if y := g(); y {
	}
	for {
	}
	for x < 3 {
	}
	for i := 0; i < 3; i++ {
	}
	for ; ; i++ {
	}
	for i := 0; i < 3; {
	}
	for range ch {
	}
	for k := range m {
	}
	for k, v = range m {
	}
	switch {
	case a, b:
	default:
	}
	switch x {
	}
	switch x := g(); x {
	}
	switch x := g(); {
	}
	switch v := x.(type) {
	}
	switch y := 1; x.(type) {
	}
	select {
	case v := <-ch:
	case ch <- 1:
	case <-ch:
	default:
	}
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "h.go", src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(src, "\n")
	p := &dumper{src: fset}
	n := 0
	goast.Inspect(f, func(node goast.Node) bool {
		s, ok := node.(goast.Stmt)
		if !ok {
			return true
		}
		h, ok := p.header(s)
		if !ok {
			return true
		}
		n++
		line := strings.TrimSpace(lines[fset.Position(s.Pos()).Line-1])
		if want := strings.TrimSuffix(line, " {"); h != want {
			t.Errorf("header %q, want %q", h, want)
		}
		return true
	})
	if n != 23 {
		t.Errorf("checked %d headers, want 23", n)
	}
	if _, ok := p.header(&goast.BlockStmt{}); ok {
		t.Error("a BlockStmt has no header")
	}
}

// TestFprintGo covers go/ast nodes: pure ones as Go source (several lines
// as a "|" block), comments, a header that holds markup, a bad header,
// nil file sets and write errors.
func TestFprintGo(t *testing.T) {
	src := `package p

// F is pure Go.
func F() {
	x := 1

	_ = x
}

var v = 1 // trailing
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "g.go", src, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	want := `File 1:1-10:10 package p
  FuncDecl 4:1-8:2
    | // F is pure Go.
    | func F() {
    | 	x := 1
    |
    | 	_ = x
    | }
  GenDecl 10:1-10:10 var v = 1 // trailing
  CommentGroup 3:1-3:17
    Comment 3:1-3:17 "// F is pure Go."
  CommentGroup 10:11-10:22
    Comment 10:11-10:22 "// trailing"
`
	if got := dump(t, fset, f); got != want {
		t.Errorf("dump:\n%s\nwant:\n%s", got, want)
	}

	// A header part holding markup is printed structurally, after "...".
	markup := &MarkupStmt{EmptyStmt: &goast.EmptyStmt{Semicolon: 30}, Root: &Text{ValuePos: 30, Raw: "hi", Value: "hi"}}
	cond := &goast.CallExpr{
		Fun:    &goast.FuncLit{Type: &goast.FuncType{Func: 20, Params: &goast.FieldList{Opening: 24, Closing: 25}}, Body: &goast.BlockStmt{Lbrace: 27, List: []goast.Stmt{markup}, Rbrace: 34}},
		Lparen: 35, Rparen: 36,
	}
	impure := &If{Keyword: 17, Stmt: &goast.IfStmt{If: 17, Cond: cond, Body: &goast.BlockStmt{}}, Body: &Block{Lbrace: 38, Rbrace: 39}}
	bad := &For{Keyword: 50, Stmt: &goast.BadStmt{From: 50, To: 60}, Body: &Block{Lbrace: 61, Rbrace: 62}}
	labeled := &goast.LabeledStmt{Label: &goast.Ident{NamePos: 70, Name: "L"}, Colon: 71, Stmt: &MarkupStmt{EmptyStmt: &goast.EmptyStmt{Semicolon: 73}, Root: &Text{ValuePos: 73, Raw: "x", Value: "x"}}}
	unary := &goast.UnaryExpr{OpPos: 80, Op: token.AND, X: &goast.CompositeLit{Type: &goast.Ident{NamePos: 81, Name: "T"}, Lbrace: 82, Elts: []goast.Expr{&goast.FuncLit{Type: &goast.FuncType{Func: 83, Params: &goast.FieldList{Opening: 87, Closing: 88}}, Body: &goast.BlockStmt{Lbrace: 90, List: []goast.Stmt{&MarkupStmt{Root: &Text{ValuePos: 91, Raw: "y", Value: "y"}}}, Rbrace: 93}}}, Rbrace: 94}}
	want = `If 17-40 if ...
  CallExpr 20-37
    FuncLit 20-35 func()
      BlockStmt 27-35
        MarkupStmt 30-32
          Text 30-32 "hi"
  Block 38-40
For 50-63 for ...
  BadStmt 50-60 BadStmt
  Block 61-63
LabeledStmt 70-74 L:
  MarkupStmt 73-74
    Text 73-74 "x"
UnaryExpr 80-95 &
  CompositeLit 81-95
    Ident 81-82 T
    FuncLit 83-94 func()
      BlockStmt 90-94
        MarkupStmt 91-92
          Text 91-92 "y"
`
	var got strings.Builder
	for _, n := range []goast.Node{impure, bad, labeled, unary} {
		got.WriteString(dump(t, nil, n))
	}
	if got.String() != want {
		t.Errorf("dump:\n%s\nwant:\n%s", got.String(), want)
	}

	if got := dump(t, fset, nil); got != "" {
		t.Errorf("dump of nil = %q", got)
	}
	if got := dump(t, nil, &Element{Name: "p"}); got != "Element ?-? p\n" {
		t.Errorf("dump without positions = %q", got)
	}
	errWrite := errors.New("disk full")
	if err := Fprint(failWriter{errWrite}, fset, f); !errors.Is(err, errWrite) {
		t.Errorf("Fprint error = %v, want %v", err, errWrite)
	}
}

// TestFprintGoMode dumps Go statements of a template body that hold markup:
// each shows its header, then its body.
func TestFprintGoMode(t *testing.T) {
	src := `package p

template List(items []string, ch chan string, x any) {
	for _, it := range items {
		<li>{it}</li>
	}
	for i := 0; i < 2; i++ {
		<br>
	}
	switch v := x.(type) {
	case int:
		<b>{v}</b>
	}
	select {
	case s := <-ch:
		<i>{s}</i>
	}
	var g = func() element {
		<hr/>
		return nil
	}
	_ = g
	if x == nil {
		<p/>
	} else {
		<a/>
	}
}
`
	fset := token.NewFileSet()
	b := &builder{t: t, src: src, base: fset.Base()}
	interp := func() Node { return b.interp() }
	stmts := []*MarkupStmt{
		b.markup(b.el("li", nil, interp)),
		b.markup(b.void("br")),
		b.markup(b.el("b", nil, interp)),
		b.markup(b.el("i", nil, interp)),
		b.markup(b.selfClosing("hr")),
		b.markup(b.selfClosing("p")),
		b.markup(b.selfClosing("a")),
	}
	f, _, _ := parseSkeleton(t, fset, src, stmts)
	want := `File 1:1-28:2 package p
  TemplateDecl 3:1-28:2 template List(items []string, ch chan string, x any)
    BlockStmt 3:54-28:2
      RangeStmt 4:2-6:3 for _, it := range items
        BlockStmt 4:27-6:3
          MarkupStmt 5:3-5:16
            Element 5:3-5:16 li close=5:11
              Interp 5:7-5:11 {it}
      ForStmt 7:2-9:3 for i := 0; i < 2; i++
        BlockStmt 7:25-9:3
          MarkupStmt 8:3-8:7
            Element 8:3-8:7 br
      TypeSwitchStmt 10:2-13:3 switch v := x.(type)
        BlockStmt 10:23-13:3
          CaseClause 11:2-12:13 case int:
            MarkupStmt 12:3-12:13
              Element 12:3-12:13 b close=12:9
                Interp 12:6-12:9 {v}
      SelectStmt 14:2-17:3 select
        BlockStmt 14:9-17:3
          CommClause 15:2-16:13 case s := <-ch:
            MarkupStmt 16:3-16:13
              Element 16:3-16:13 i close=16:9
                Interp 16:6-16:9 {s}
      DeclStmt 18:2-21:3
        GenDecl 18:2-21:3 var
          ValueSpec 18:6-21:3
            Ident 18:6-18:7 g
            FuncLit 18:10-21:3 func() element
              BlockStmt 18:25-21:3
                MarkupStmt 19:3-19:8
                  Element 19:3-19:8 hr self-closing
                ReturnStmt 20:3-20:13 return nil
      AssignStmt 22:2-22:7 _ = g
      IfStmt 23:2-27:3 if x == nil
        BlockStmt 23:14-25:3
          MarkupStmt 24:3-24:7
            Element 24:3-24:7 p self-closing
        BlockStmt 25:9-27:3
          MarkupStmt 26:3-26:7
            Element 26:3-26:7 a self-closing
`
	if got := dump(t, fset, f); got != want {
		t.Errorf("dump:\n%s\nwant:\n%s", got, want)
	}
}

// TestFprintDocExample keeps the example in the package documentation true.
func TestFprintDocExample(t *testing.T) {
	src := `package p

template T(items []string) {
	if len(items) > 0 {
		// the example in the package documentation
		<ul class="list">
		for i, it := range items {
			<li>{i}: {it}</li>
		}
		</ul>
	}
}
`
	fset := token.NewFileSet()
	b := &builder{t: t, src: src, base: fset.Base()}
	interp := func() Node { return b.interp() }
	ul := b.markup(b.el("ul", []func() *Attr{b.static("class", `"list"`)}, func() Node {
		f := &For{Keyword: b.at("for")}
		f.Body = b.block(b.elem("li", nil, interp, b.text(": ", ": "), interp))
		return f
	}))
	parseSkeleton(t, fset, src, []*MarkupStmt{ul})
	want := `MarkupStmt 6:3-10:8
  Element 6:3-10:8 ul close=10:3
    Attr 6:7-6:19 class
      StaticValue 6:13-6:19 "list"
    For 7:3-9:4 for i, it := range items
      Block 7:28-9:4
        Element 8:4-8:22 li close=8:17
          Interp 8:8-8:11 {i}
          Text 8:11-8:13 ": "
          Interp 8:13-8:17 {it}
`
	if got := dump(t, fset, ul); got != want {
		t.Errorf("dump:\n%s\nwant:\n%s", got, want)
	}
}

func TestOperator(t *testing.T) {
	cases := []struct {
		n    goast.Node
		want string
	}{
		{&goast.AssignStmt{Tok: token.DEFINE}, ":="},
		{&goast.BinaryExpr{Op: token.ADD}, "+"},
		{&goast.UnaryExpr{Op: token.ARROW}, "<-"},
		{&goast.IncDecStmt{Tok: token.INC}, "++"},
		{&goast.GenDecl{Tok: token.CONST}, "const"},
		{&goast.CallExpr{}, ""},
	}
	for _, c := range cases {
		if got := operator(c.n); got != c.want {
			t.Errorf("operator(%T) = %q, want %q", c.n, got, c.want)
		}
	}
}

// TestFprintPositionOutsideFileSet prints positions the file set does not
// know as "?".
func TestFprintPositionOutsideFileSet(t *testing.T) {
	if got := dump(t, token.NewFileSet(), &Element{Lt: 5, Name: "p"}); got != "Element ?-? p\n" {
		t.Errorf("dump = %q", got)
	}
}

type failWriter struct{ err error }

func (w failWriter) Write([]byte) (int, error) { return 0, w.err }

// TestFprintTemplateSignatures covers receivers, a template without a body
// and declarations missing parts.
func TestFprintTemplateSignatures(t *testing.T) {
	src := "package p\n\nfunc     (c Card) View() {\n}\n"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "t.golpe", src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	fd := f.Decls[0].(*goast.FuncDecl)
	cases := []struct {
		n    *TemplateDecl
		want string
	}{
		{&TemplateDecl{FuncDecl: fd, Template: fd.Pos()}, "TemplateDecl 3:1-4:2 template (c Card) View()\n  BlockStmt 3:26-4:2\n    | {\n    | }\n"},
		{&TemplateDecl{FuncDecl: &goast.FuncDecl{Name: fd.Name, Type: fd.Type}, Template: fd.Pos()}, "TemplateDecl 3:1-3:25 template View()\n"},
		{&TemplateDecl{FuncDecl: &goast.FuncDecl{Name: fd.Name}, Template: fd.Pos()}, "TemplateDecl 3:1-3:9 template View\n"},
		{&TemplateDecl{Template: fd.Pos()}, "TemplateDecl 3:1-3:9 template\n"},
	}
	for _, c := range cases {
		if got := dump(t, fset, c.n); got != c.want {
			t.Errorf("dump:\n%s\nwant:\n%s", got, c.want)
		}
	}
}
