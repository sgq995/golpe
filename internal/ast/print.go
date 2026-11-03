package ast

import (
	"bytes"
	"fmt"
	goast "go/ast"
	"go/printer"
	"go/token"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Fprint writes a debug dump of the tree rooted at node to w, in the format
// described in the package documentation. Positions are resolved with fset;
// if fset is nil they are printed as raw token.Pos values. A nil node prints
// nothing.
func Fprint(w io.Writer, fset *token.FileSet, node goast.Node) error {
	d := &dumper{fset: fset, src: fset}
	if d.src == nil {
		d.src = token.NewFileSet() // go/printer needs one
	}
	if node != nil {
		d.node(node)
	}
	_, err := w.Write(d.buf.Bytes())
	return err
}

// gofmt prints Go source as gofmt does.
var gofmt = printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}

type dumper struct {
	buf   bytes.Buffer
	fset  *token.FileSet // for positions; nil prints raw values
	src   *token.FileSet // for go/printer; never nil
	depth int
}

func (d *dumper) node(n goast.Node) {
	switch n := n.(type) {
	case *goast.File:
		detail := "package"
		if n.Name != nil {
			detail += " " + n.Name.Name
		}
		d.line(n, detail)
		dumpList(d, n.Decls)
		dumpList(d, n.Comments)
	case *TemplateDecl:
		d.line(n, d.signature(n))
		if n.FuncDecl != nil && n.Body != nil {
			d.kids(n.Body)
		}
	case *MarkupStmt:
		d.line(n, "")
		d.kids(children(n)...)
	case *Element:
		detail := n.Name
		if n.SelfClosing {
			detail += " self-closing"
		}
		if n.Close.IsValid() {
			detail += " close=" + d.pos(n.Close)
		}
		d.line(n, detail)
		d.kids(children(n)...)
	case *Attr:
		d.line(n, n.Name)
		d.kids(children(n)...)
	case *StaticValue:
		d.line(n, strconv.Quote(n.Value))
	case *Text:
		d.line(n, strconv.Quote(n.Value))
	case *Interp:
		if n.X == nil {
			d.line(n, "{...}")
		} else if src, ok := d.source(n.X); ok {
			d.line(n, "{"+src+"}")
		} else {
			d.line(n, "")
			d.kids(n.X)
		}
	case *Comment:
		d.line(n, strconv.Quote(n.Text))
	case *Doctype:
		d.line(n, strconv.Quote(n.Text))
	case *If:
		var s goast.Stmt
		if n.Stmt != nil {
			s = n.Stmt
		}
		d.embedded(n, s, "if ...")
	case *For:
		d.embedded(n, n.Stmt, "for ...")
	case *Switch:
		d.embedded(n, n.Stmt, "switch ...")
	case *Select:
		d.line(n, "select") // a select header has no Go parts
		d.kids(children(n)...)
	case *Case:
		missing := "case ...:"
		if n.Default {
			missing = "default:"
		}
		d.embedded(n, n.Clause, missing)
	case *Block:
		d.line(n, "")
		d.kids(children(n)...)
	default:
		d.goNode(n)
	}
}

// embedded dumps an embedded statement or case clause n whose go/ast
// statement is s. With a header that prints as Go, the children are the
// markup body; otherwise the detail is missing and the children include the
// header parts, if any.
func (d *dumper) embedded(n goast.Node, s goast.Stmt, missing string) {
	if s != nil {
		if h, ok := d.header(s); ok {
			d.line(n, h)
			d.kids(afterHeader(n, s)...)
			return
		}
	}
	d.line(n, missing)
	d.kids(children(n)...)
}

// afterHeader returns the children of n, a Golpe or go/ast statement or
// case clause whose go/ast statement is s, that follow the header parts of
// s: its body. Walk visits header parts first, as go/ast.Walk does.
func afterHeader(n goast.Node, s goast.Stmt) []goast.Node {
	var parts counter
	walkHeader(&parts, s)
	return children(n)[parts:]
}

// counter counts the nodes it visits, without descending.
type counter int

func (c *counter) Visit(n goast.Node) goast.Visitor {
	if n != nil {
		*c++
	}
	return nil
}

// goNode dumps a go/ast node: as Go source if no Golpe node is below it,
// structurally otherwise.
func (d *dumper) goNode(n goast.Node) {
	if src, ok := d.source(n); ok {
		d.line(n, src)
		return
	}
	switch n := n.(type) {
	case *goast.Comment:
		d.line(n, strconv.Quote(n.Text))
		return
	case *goast.FuncLit:
		if n.Type != nil && n.Body != nil {
			if typ, ok := d.source(n.Type); ok {
				d.line(n, typ)
				d.kids(n.Body)
				return
			}
		}
	case *goast.LabeledStmt:
		if n.Label != nil && n.Stmt != nil {
			d.line(n, n.Label.Name+":")
			d.kids(n.Stmt)
			return
		}
	case goast.Stmt:
		if h, ok := d.header(n); ok {
			d.line(n, h)
			d.kids(afterHeader(n, n)...)
			return
		}
	}
	d.line(n, operator(n))
	d.kids(children(n)...)
}

// header returns the header of a statement as Go source, without the body:
// "if x := f(); x > 0", "for k, v := range m", "case a, b:", "default:".
// It reports false for statements without a header and when a header part
// does not print as Go.
func (d *dumper) header(s goast.Stmt) (string, bool) {
	ok := true
	src := func(n goast.Node) string {
		text, printed := d.source(n)
		ok = ok && printed
		return text
	}
	opt := func(n goast.Node) string {
		if n == nil {
			return ""
		}
		return src(n)
	}
	var h string
	switch s := s.(type) {
	case *goast.IfStmt:
		h = "if "
		if s.Init != nil {
			h += src(s.Init) + "; "
		}
		h += src(s.Cond)
	case *goast.ForStmt:
		h = "for"
		if s.Init != nil || s.Post != nil {
			h = strings.TrimRight(h+" "+opt(s.Init)+"; "+opt(s.Cond)+"; "+opt(s.Post), " ")
		} else if s.Cond != nil {
			h += " " + src(s.Cond)
		}
	case *goast.RangeStmt:
		h = "for "
		if s.Key != nil {
			h += src(s.Key)
			if s.Value != nil {
				h += ", " + src(s.Value)
			}
			h += " " + s.Tok.String() + " "
		}
		h += "range " + src(s.X)
	case *goast.SwitchStmt:
		h = "switch"
		if s.Init != nil {
			h += " " + src(s.Init) + ";"
		}
		if s.Tag != nil {
			h += " " + src(s.Tag)
		}
	case *goast.TypeSwitchStmt:
		h = "switch "
		if s.Init != nil {
			h += src(s.Init) + "; "
		}
		h += src(s.Assign)
	case *goast.SelectStmt:
		h = "select"
	case *goast.CaseClause:
		if s.List == nil {
			return "default:", true
		}
		list := make([]string, len(s.List))
		for i, x := range s.List {
			list[i] = src(x)
		}
		h = "case " + strings.Join(list, ", ") + ":"
	case *goast.CommClause:
		if s.Comm == nil {
			return "default:", true
		}
		h = "case " + src(s.Comm) + ":"
	default:
		return "", false
	}
	return h, ok
}

// signature returns the signature of a template declaration.
func (d *dumper) signature(t *TemplateDecl) string {
	if t.FuncDecl == nil || t.Name == nil {
		return "template"
	}
	if t.Type != nil {
		fd := &goast.FuncDecl{Recv: t.Recv, Name: t.Name, Type: t.Type}
		if src, ok := d.source(fd); ok {
			if rest, ok := strings.CutPrefix(src, "func"); ok {
				return "template" + rest
			}
		}
	}
	return "template " + t.Name.Name
}

// source returns n as Go source. It reports false if n is not a kind of
// node go/printer prints, or if a Golpe node is below it.
func (d *dumper) source(n goast.Node) (string, bool) {
	switch n.(type) {
	case goast.Expr, goast.Stmt, goast.Decl, goast.Spec:
	default:
		return "", false
	}
	if !pure(n) {
		return "", false
	}
	var b bytes.Buffer
	if err := gofmt.Fprint(&b, d.src, n); err != nil {
		return "", false
	}
	// go/printer ends a trailing line comment with a newline.
	return strings.TrimRight(b.String(), "\n"), true
}

// operator returns the operator or keyword token of a go/ast node, if it
// has one that its children do not show.
func operator(n goast.Node) string {
	switch n := n.(type) {
	case *goast.AssignStmt:
		return n.Tok.String()
	case *goast.BinaryExpr:
		return n.Op.String()
	case *goast.UnaryExpr:
		return n.Op.String()
	case *goast.IncDecStmt:
		return n.Tok.String()
	case *goast.GenDecl:
		return n.Tok.String()
	}
	return ""
}

// line writes the line of node n.
func (d *dumper) line(n goast.Node, detail string) {
	indent := strings.Repeat("  ", d.depth)
	fmt.Fprintf(&d.buf, "%s%s %s-%s", indent, kind(n), d.pos(n.Pos()), d.pos(n.End()))
	if !strings.Contains(detail, "\n") {
		if detail != "" {
			d.buf.WriteString(" " + detail)
		}
		d.buf.WriteByte('\n')
		return
	}
	d.buf.WriteByte('\n')
	for _, l := range strings.Split(detail, "\n") {
		if l == "" {
			fmt.Fprintf(&d.buf, "%s  |\n", indent)
		} else {
			fmt.Fprintf(&d.buf, "%s  | %s\n", indent, l)
		}
	}
}

func (d *dumper) kids(nodes ...goast.Node) {
	d.depth++
	for _, n := range nodes {
		d.node(n)
	}
	d.depth--
}

func dumpList[N goast.Node](d *dumper, list []N) {
	d.depth++
	for _, n := range list {
		d.node(n)
	}
	d.depth--
}

func (d *dumper) pos(p token.Pos) string {
	if !p.IsValid() {
		return "?"
	}
	if d.fset == nil {
		return strconv.Itoa(int(p))
	}
	q := d.fset.Position(p)
	if !q.IsValid() {
		return "?"
	}
	return fmt.Sprintf("%d:%d", q.Line, q.Column)
}

// kind returns the type name of a node: Element, IfStmt, ...
func kind(n goast.Node) string {
	t := reflect.TypeOf(n)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
}

// children returns the children of n, as Walk visits them.
func children(n goast.Node) []goast.Node {
	var kids []goast.Node
	root := true
	Inspect(n, func(c goast.Node) bool {
		if root {
			root = false
			return true
		}
		if c != nil {
			kids = append(kids, c)
		}
		return false
	})
	return kids
}

// pure reports whether no Golpe node is in the tree rooted at n.
func pure(n goast.Node) bool {
	ok := true
	Inspect(n, func(c goast.Node) bool {
		if isGolpe(c) {
			ok = false
		}
		return ok
	})
	return ok
}

// isGolpe reports whether n is a node of this package.
func isGolpe(n goast.Node) bool {
	switch n.(type) {
	case Node, *TemplateDecl, *MarkupStmt:
		return true
	}
	return false
}
