package ast

import goast "go/ast"

// A Visitor's Visit method is invoked for each node encountered by Walk. It
// is go/ast's Visitor, so existing visitors work with Walk unchanged.
type Visitor = goast.Visitor

// Walk traverses a tree of go/ast and Golpe nodes in depth-first order, with
// go/ast.Walk's contract: it starts by calling v.Visit(node); node must not
// be nil. If the visitor w returned by v.Visit(node) is not nil, Walk is
// invoked recursively with visitor w for each of the non-nil children of
// node, followed by a call of w.Visit(nil).
//
// The children of a Golpe node are walked in source order. Those of an
// embedded statement are the header parts of its go/ast statement (see the
// package documentation) and then its markup body; the go/ast statement's
// own body is never walked.
func Walk(v Visitor, node goast.Node) {
	goast.Walk(walker{v}, node)
}

// walker lets go/ast.Walk drive a tree that mixes go/ast and Golpe nodes.
//
// go/ast.Walk calls Visit before it switches on the node type, and returns
// at once if Visit returns nil. For a Golpe node, the walker does all the
// work itself (Visit, children, Visit(nil)) and returns nil, so go/ast.Walk
// never reaches the default case of its switch, which panics. For a go/ast
// node it returns the user's visitor wrapped again, so go/ast.Walk walks the
// children, and any Golpe node among or below them comes back here. Golpe
// nodes can therefore sit inside go/ast containers (File.Decls, statement
// lists, function literals) and go/ast nodes inside Golpe nodes, to any
// depth, without re-implementing go/ast.Walk.
type walker struct{ v Visitor }

func (w walker) Visit(node goast.Node) goast.Visitor {
	if node == nil {
		w.v.Visit(nil) // end of a go/ast node's children
		return nil
	}
	v := w.v.Visit(node)
	if v == nil {
		return nil
	}
	if walkGolpe(v, node) {
		v.Visit(nil)
		return nil
	}
	return walker{v}
}

// walkGolpe walks the children of node with v if node is a Golpe node, and
// reports whether it was one.
func walkGolpe(v Visitor, node goast.Node) bool {
	switch n := node.(type) {
	case *TemplateDecl:
		if d := n.FuncDecl; d != nil {
			// The children of the FuncDecl, not the FuncDecl itself: the
			// TemplateDecl is the declaration.
			walkOpt(v, d.Doc)
			walkOpt(v, d.Recv)
			walkOpt(v, d.Name)
			walkOpt(v, d.Type)
			walkOpt(v, d.Body)
		}
	case *MarkupStmt:
		walkOpt(v, n.Root)
	case *Element:
		walkList(v, n.Attrs)
		walkList(v, n.Children)
	case *Fragment:
		walkList(v, n.Children)
	case *Attr:
		walkOpt(v, n.Value)
	case *StaticValue, *Text, *HTMLComment, *Doctype:
		// no children
	case *Interp:
		walkOpt(v, n.X)
	case *If:
		if n.Stmt != nil {
			walkHeader(v, n.Stmt)
		}
		walkOpt(v, n.Body)
		walkOpt(v, n.Else)
	case *For:
		walkHeader(v, n.Stmt)
		walkOpt(v, n.Body)
	case *Switch:
		walkHeader(v, n.Stmt)
		walkList(v, n.Cases)
	case *Select:
		walkList(v, n.Cases) // a SelectStmt has no header parts
	case *Case:
		walkHeader(v, n.Clause)
		walkList(v, n.Body)
	case *Block:
		walkList(v, n.List)
	default:
		return false
	}
	return true
}

// walkHeader walks the header parts of the go/ast statement of an embedded
// statement or case clause, never its body, which is skeleton.
func walkHeader(v Visitor, s goast.Stmt) {
	switch s := s.(type) {
	case nil:
	case *goast.IfStmt:
		walkOpt(v, s.Init)
		walkOpt(v, s.Cond)
	case *goast.ForStmt:
		walkOpt(v, s.Init)
		walkOpt(v, s.Cond)
		walkOpt(v, s.Post)
	case *goast.RangeStmt:
		walkOpt(v, s.Key)
		walkOpt(v, s.Value)
		walkOpt(v, s.X)
	case *goast.SwitchStmt:
		walkOpt(v, s.Init)
		walkOpt(v, s.Tag)
	case *goast.TypeSwitchStmt:
		walkOpt(v, s.Init)
		walkOpt(v, s.Assign)
	case *goast.SelectStmt:
		// no header parts
	case *goast.CaseClause:
		walkList(v, s.List)
	case *goast.CommClause:
		walkOpt(v, s.Comm)
	default:
		// Not a statement with a header, such as a BadStmt left by a
		// syntax error: visit it whole.
		Walk(v, s)
	}
}

// optional is a node field: an interface, or a pointer to a node type.
type optional interface {
	goast.Node
	comparable
}

// walkOpt walks n unless it is nil. Comparing in the field's own type
// catches nil pointers as well as nil interfaces.
func walkOpt[N optional](v Visitor, n N) {
	var zero N
	if n != zero {
		Walk(v, n)
	}
}

// walkList walks the non-nil nodes of list.
func walkList[N optional](v Visitor, list []N) {
	for _, n := range list {
		walkOpt(v, n)
	}
}

type inspector func(goast.Node) bool

func (f inspector) Visit(node goast.Node) goast.Visitor {
	if f(node) {
		return f
	}
	return nil
}

// Inspect traverses a tree of go/ast and Golpe nodes in depth-first order,
// with go/ast.Inspect's contract: it starts by calling f(node); node must
// not be nil. If f returns true, Inspect invokes f recursively for each of
// the non-nil children of node, followed by a call of f(nil). The children
// are the ones Walk visits.
func Inspect(node goast.Node, f func(goast.Node) bool) {
	Walk(inspector(f), node)
}
