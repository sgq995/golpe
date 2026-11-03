package ast

import (
	"fmt"
	goast "go/ast"
	"go/token"
	"testing"
)

func span(fset *token.FileSet, n goast.Node) string {
	pos := func(p token.Pos) string {
		if !p.IsValid() {
			return "?"
		}
		q := fset.Position(p)
		return fmt.Sprintf("%d:%d", q.Line, q.Column)
	}
	return pos(n.Pos()) + "-" + pos(n.End())
}

// TestPosEnd checks the extent of every node type in a realistic tree. End
// is exclusive, as in go/ast.
func TestPosEnd(t *testing.T) {
	pg := newPage(t)
	elem := func(s *MarkupStmt) *Element { return s.Root.(*Element) }
	cases := []struct {
		name string
		n    goast.Node
		want string
	}{
		{"TemplateDecl", pg.tmpl, "3:1-45:2"},
		{"MarkupStmt", pg.ul, "6:3-10:8"},
		{"Element", elem(pg.ul), "6:3-10:8"},
		{"Element nested", pg.li, "8:4-8:22"},
		{"Element void", pg.br, "43:3-43:7"},
		{"Element self-closing", pg.hrEl, "18:3-18:8"},
		{"MarkupStmt self-closing", pg.hr, "18:3-18:8"},
		{"Attr static", pg.attrs[0], "6:7-6:19"},
		{"StaticValue", pg.attrs[0].Value, "6:13-6:19"},
		{"Attr boolean", pg.attrs[1], "6:20-6:26"},
		{"Attr dynamic", pg.attrs[2], "6:27-6:37"},
		{"Interp", pg.attrs[2].Value, "6:34-6:37"},
		{"Interp func literal", pg.funcInterp, "24:3-24:34"},
		{"MarkupStmt in func literal interp", pg.lit, "24:21-24:31"},
		{"Text", pg.li.Children[1], "8:11-8:13"},
		{"Text raw spans lines", pg.off, "29:11-31:3"},
		{"Comment", pg.comment, "22:3-22:16"},
		{"MarkupStmt doctype", pg.doctype, "20:2-20:17"},
		{"Doctype", pg.doctype.Root, "20:2-20:17"},
		{"For", pg.forNode, "7:3-9:4"},
		{"Block", pg.forNode.Body, "7:28-9:4"},
		{"If with else chain", pg.ifSel, "25:3-31:4"},
		{"If body", pg.ifSel.Body, "25:10-27:4"},
		{"If else if", pg.ifElse, "27:10-31:4"},
		{"Block else", pg.elseBlock, "29:10-31:4"},
		{"Switch", pg.sw, "32:3-37:4"},
		{"Case", pg.cases[0], "33:3-34:20"},
		{"Case default", pg.cases[1], "35:3-36:11"},
		{"Select", pg.sel, "38:3-42:4"},
		{"Case comm", pg.cases[2], "39:3-40:7"},
		{"Case empty default", pg.cases[3], "41:3-41:11"},
		{"MarkupStmt div", pg.div, "21:2-44:8"},
		{"MarkupStmt in func literal", pg.p, "13:3-13:15"},
		{"File", pg.file, "1:1-45:2"},
	}
	for _, c := range cases {
		if got := span(pg.fset, c.n); got != c.want {
			t.Errorf("%s: extent %s, want %s", c.name, got, c.want)
		}
	}
	if pg.ul.Pos() != pg.ul.Semicolon {
		t.Errorf("MarkupStmt.Pos %d != placeholder %d", pg.ul.Pos(), pg.ul.Semicolon)
	}
	if pg.tmpl.Pos() != pg.tmpl.Type.Func {
		t.Errorf("TemplateDecl.Pos %d != masked func keyword %d", pg.tmpl.Pos(), pg.tmpl.Type.Func)
	}
}

// TestPosEndPartial checks the fallbacks for trees built while recovering
// from errors or before grafting: End never panics, never returns a position
// before Pos, and is NoPos only when nothing is known.
func TestPosEndPartial(t *testing.T) {
	const b = 100 // positions are offsets from b
	at := func(off int) token.Pos { return token.Pos(b + off) }
	text := &Text{ValuePos: at(5), Raw: "abc", Value: "abc"}
	x := &goast.Ident{NamePos: at(11), Name: "xy"}
	cases := []struct {
		name     string
		n        goast.Node
		pos, end int // offsets; -1 for NoPos
	}{
		{"Element unterminated with children", &Element{Lt: at(0), Name: "div", Children: []Node{text}}, 0, 8},
		{"Element end tag without >", &Element{Lt: at(0), Name: "div", Close: at(20)}, 0, 25},
		{"Element attrs only", &Element{Lt: at(0), Name: "a", Attrs: []*Attr{{NamePos: at(3), Name: "href"}}}, 0, 7},
		{"Element name only", &Element{Lt: at(0), Name: "span"}, 0, 5},
		{"Attr boolean", &Attr{NamePos: at(3), Name: "hidden"}, 3, 9},
		{"Interp unterminated", &Interp{Lbrace: at(10), X: x}, 10, 13},
		{"Interp pre-graft", &Interp{Lbrace: at(10), Rbrace: at(13)}, 10, 14},
		{"Interp bare", &Interp{Lbrace: at(10)}, 10, 11},
		{"StaticValue", &StaticValue{ValuePos: at(2), Raw: `'a"b'`, Value: `a"b`}, 2, 7},
		{"Comment", &Comment{Start: at(0), Text: "<!-- x -->"}, 0, 10},
		{"Doctype", &Doctype{Start: at(0), Text: "<!doctype html>"}, 0, 15},
		{"If without body", &If{Keyword: at(4)}, 4, 6},
		{"If without else", &If{Keyword: at(4), Body: &Block{Lbrace: at(10), Rbrace: at(20)}}, 4, 21},
		{"For without body", &For{Keyword: at(4)}, 4, 7},
		{"Switch unterminated", &Switch{Keyword: at(0), Lbrace: at(9), Cases: []*Case{{Keyword: at(11), Colon: at(17)}}}, 0, 18},
		{"Switch without cases", &Switch{Keyword: at(0), Lbrace: at(9)}, 0, 10},
		{"Switch header only", &Switch{Keyword: at(0)}, 0, 6},
		{"Select unterminated", &Select{Keyword: at(0), Lbrace: at(7)}, 0, 8},
		{"Select header only", &Select{Keyword: at(0)}, 0, 6},
		{"Select", &Select{Keyword: at(0), Lbrace: at(7), Rbrace: at(9)}, 0, 10},
		{"Case without colon", &Case{Keyword: at(0)}, 0, 4},
		{"Default without colon", &Case{Keyword: at(0), Default: true}, 0, 7},
		{"Case with body", &Case{Keyword: at(0), Colon: at(6), Body: []Node{text}}, 0, 8},
		{"Block unterminated", &Block{Lbrace: at(4), List: []Node{text}}, 4, 8},
		{"Block empty unterminated", &Block{Lbrace: at(4)}, 4, 5},
		{"MarkupStmt pre-graft", &MarkupStmt{Root: text}, 5, 8},
		{"MarkupStmt without root", &MarkupStmt{EmptyStmt: &goast.EmptyStmt{Semicolon: at(5)}}, 5, 6},
		{"MarkupStmt empty", &MarkupStmt{}, -1, -1},
		{"TemplateDecl without FuncDecl", &TemplateDecl{Template: at(0)}, 0, 8},
		{"TemplateDecl without template pos", &TemplateDecl{FuncDecl: &goast.FuncDecl{
			Name: &goast.Ident{NamePos: at(9), Name: "F"},
			Type: &goast.FuncType{Func: at(0), Params: &goast.FieldList{Opening: at(10), Closing: at(11)}},
		}}, 0, 12},
		{"TemplateDecl without type and body", &TemplateDecl{Template: at(0), FuncDecl: &goast.FuncDecl{}}, 0, 8},
		{"Text NoPos", &Text{Raw: "x"}, -1, -1},
		{"If NoPos", &If{}, -1, -1},
		{"Element NoPos", &Element{Name: "p"}, -1, -1},
	}
	off := func(p token.Pos) int {
		if !p.IsValid() {
			return -1
		}
		return int(p) - b
	}
	for _, c := range cases {
		if pos, end := off(c.n.Pos()), off(c.n.End()); pos != c.pos || end != c.end {
			t.Errorf("%s: extent [%d, %d), want [%d, %d)", c.name, pos, end, c.pos, c.end)
		}
	}
}

// TestInterfaces documents which node types implement which interface; the
// compile-time assertions live in ast.go.
func TestInterfaces(t *testing.T) {
	var decl goast.Decl = &TemplateDecl{}
	var stmt goast.Stmt = &MarkupStmt{}
	if _, ok := decl.(Node); ok {
		t.Error("TemplateDecl must not be a markup Node")
	}
	if _, ok := stmt.(Node); ok {
		t.Error("MarkupStmt must not be a markup Node")
	}
	markup := []goast.Node{&Element{}, &Attr{}, &StaticValue{}, &Text{}, &Interp{}, &Comment{},
		&Doctype{}, &If{}, &For{}, &Switch{}, &Select{}, &Case{}, &Block{}}
	for _, n := range markup {
		if _, ok := n.(Node); !ok {
			t.Errorf("%T is not a Node", n)
		}
		_, isValue := n.(AttrValue)
		switch n.(type) {
		case *StaticValue, *Interp:
			if !isValue {
				t.Errorf("%T is not an AttrValue", n)
			}
		default:
			if isValue {
				t.Errorf("%T must not be an AttrValue", n)
			}
		}
		if _, ok := n.(goast.Expr); ok {
			t.Errorf("%T must not be a go/ast.Expr", n)
		}
	}
}
