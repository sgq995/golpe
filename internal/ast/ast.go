// Package ast declares the syntax tree of Golpe files: go/ast extended with
// template declarations and markup.
//
// # Nodes
//
// A parsed .golpe file is a *go/ast.File. Two node types bridge Go and
// markup. Each embeds a go/ast node, which promotes go/ast's unexported
// marker methods and so satisfies its sealed interfaces (LAB-002):
//
//   - [TemplateDecl] embeds the *go/ast.FuncDecl that go/parser built from
//     the masked declaration, where "template" is overwritten by "func" and
//     four spaces. It is a go/ast.Decl and sits in File.Decls.
//   - [MarkupStmt] embeds the explicit *go/ast.EmptyStmt that the mask put
//     on the first byte of the markup. It is a go/ast.Stmt and sits in a
//     statement list of a template body: the body itself, a nested block, a
//     case clause or a function literal body.
//
// Below a MarkupStmt the tree is made of markup nodes, which implement
// [Node]: [Element], [Attr], [StaticValue], [Text], [Interp], [Comment],
// [Doctype], and the embedded statements [If], [For], [Switch] and [Select]
// with their parts [Case] and [Block]. Go code inside markup is held as
// go/ast nodes: Interp.X is a go/ast.Expr, and each embedded statement keeps
// the go/ast statement that go/parser built from the mask.
//
// Embedded statements use only the header of that go/ast statement: Init
// and Cond of an IfStmt; Init, Cond and Post of a ForStmt; Key, Value and X
// of a RangeStmt; Init and Tag of a SwitchStmt; Init and Assign of a
// TypeSwitchStmt; List of a CaseClause; Comm of a CommClause (a SelectStmt
// has no header parts). Its Body, and an IfStmt's Else, hold skeleton left
// over from the mask, which the parser may keep or drop: the markup body
// lives in the Golpe node, and [Walk], [Inspect] and [Fprint] never look at
// the go/ast one.
//
// The markup parser builds markup trees before go/parser runs, so Interp.X,
// If.Stmt, For.Stmt, Switch.Stmt, Select.Stmt and Case.Clause are nil until
// the parser grafts them. Every position a node needs (keywords, braces,
// colons) is recorded on the Golpe node itself, so positions, walking and
// dumping all work on trees that are not grafted yet.
//
// Nodes record syntax only. An Element with SelfClosing false and Close ==
// token.NoPos had no end tag; whether that was allowed (a void element) is
// the parser's decision.
//
// Positions are go/token positions in the .golpe file. Pos is the position
// of a node's first byte and End the position just after its last byte, as
// in go/ast. Trees built while recovering from syntax errors may lack
// closing positions; End then falls back to the end of the last part that
// is known, and is token.NoPos only when Pos is.
//
// # Walking
//
// go/ast.Walk and go/ast.Inspect panic on Golpe nodes (LAB-003). [Walk] and
// [Inspect] keep their contract and accept both kinds of node mixed to any
// depth: Golpe nodes inside go/ast nodes (TemplateDecls in File.Decls,
// MarkupStmts in statement lists) and go/ast nodes inside Golpe nodes
// (interpolations and headers), including markup inside a function literal
// inside an interpolation. Children are visited in source order; for an
// embedded statement they are the header parts listed above, then the
// markup body.
//
// # Debug dump
//
// [Fprint] writes the compact, stable text dump used by golden tests. Each
// node is one line,
//
//	Kind pos-end detail
//
// indented two spaces per level, with its children on the following lines,
// one level deeper, in source order. Kind is the node's type name (Element,
// If, AssignStmt, ...). pos and end are line:col of Pos and End, end being
// exclusive; an invalid position prints as "?", and without a file set
// positions print as raw token.Pos values. The detail depends on the node:
//
//	File            "package" and the package name; after the declarations,
//	                File.Comments are listed as CommentGroup lines
//	TemplateDecl    the signature, as in "template Page[T any](title string)"
//	MarkupStmt      none; the child is the root
//	Element         the tag name, then "self-closing" and close=line:col of
//	                the end tag, when present; children are the attributes,
//	                then the content
//	Attr            the name; the child is the value, if any
//	StaticValue     the value, Go-quoted
//	Text            the value (after whitespace normalisation), Go-quoted
//	Interp          {X} with X as Go source, or {...} while X is nil
//	Comment         the comment text, Go-quoted (HTML and go/ast comments)
//	Doctype         the doctype text, Go-quoted
//	If, For,        the header as Go source: "if x := f(); x > 0",
//	Switch, Select, "for i, v := range items", "case 0, 1:", "default:".
//	Case            Go parts not attached yet print as "...": "if ...",
//	                "case ...:"
//	Block           none
//
// A go/ast node with no Golpe node below it prints as Go source, formatted by
// go/printer like gofmt; go/printer includes the doc and line comments
// attached to the node, so those may also appear under File.Comments. Source
// on several lines follows the node's line instead, one level deeper, each
// line prefixed by "| " (an empty line by a lone "|"). A go/ast node with
// Golpe nodes below it prints structurally: a statement with a header (if,
// for, switch, select, case) shows the header and then its body; a function
// literal its type and then its body; a labeled statement its label and
// then the statement; any other node its operator, if it has one, and then
// all its children.
//
// For example, this markup statement, on lines 6 to 10 of a file and
// indented by two tabs,
//
//	<ul class="list">
//	for i, it := range items {
//		<li>{i}: {it}</li>
//	}
//	</ul>
//
// dumps as
//
//	MarkupStmt 6:3-10:8
//	  Element 6:3-10:8 ul close=10:3
//	    Attr 6:7-6:19 class
//	      StaticValue 6:13-6:19 "list"
//	    For 7:3-9:4 for i, it := range items
//	      Block 7:28-9:4
//	        Element 8:4-8:22 li close=8:17
//	          Interp 8:8-8:11 {i}
//	          Text 8:11-8:13 ": "
//	          Interp 8:13-8:17 {it}
//
// See docs/knowledge/plans/001-template-language.md.
package ast

import (
	goast "go/ast"
	"go/token"
)

// Node is a node of a markup tree. Fields of type Node document which
// concrete types they hold.
type Node interface {
	goast.Node
	markupNode()
}

// AttrValue is the value of an attribute: a *StaticValue or an *Interp.
type AttrValue interface {
	Node
	attrValueNode()
}

// A TemplateDecl is a template declaration,
//
//	template Name[TypeParams](params) { body }
//
// or a method form with a receiver. The embedded FuncDecl is the one
// go/parser built from the mask: Type.Results is always nil, Type.Func is
// the position of the "template" keyword, and Body holds Go statements and
// MarkupStmts.
type TemplateDecl struct {
	*goast.FuncDecl
	Template token.Pos // position of the "template" keyword
}

// Pos returns the position of the "template" keyword.
func (d *TemplateDecl) Pos() token.Pos {
	if d.Template.IsValid() || d.FuncDecl == nil {
		return d.Template
	}
	return d.FuncDecl.Pos()
}

// End returns the position after the closing brace of the body.
func (d *TemplateDecl) End() token.Pos {
	if d.FuncDecl == nil || d.Body == nil && d.Type == nil {
		return offset(d.Template, len("template"))
	}
	return d.FuncDecl.End()
}

// A MarkupStmt is a markup statement: a markup tree at a statement position
// of a template body. The embedded EmptyStmt is the explicit ";" the mask
// put on the first byte of the markup, so Pos is the position of the root's
// "<". End is the end of the root.
type MarkupStmt struct {
	*goast.EmptyStmt
	Root Node // *Element or *Doctype
}

// Pos returns the position of the first byte of the markup.
func (s *MarkupStmt) Pos() token.Pos {
	switch {
	case s.EmptyStmt != nil:
		return s.Semicolon
	case s.Root != nil:
		return s.Root.Pos()
	}
	return token.NoPos
}

// End returns the end of the root node.
func (s *MarkupStmt) End() token.Pos {
	switch {
	case s.Root != nil:
		return s.Root.End()
	case s.EmptyStmt != nil:
		return s.EmptyStmt.End()
	}
	return token.NoPos
}

// An Element is a tag element: <name attrs>children</name>, a start tag
// without end tag (<br>) or a self-closing tag (<x/>). The name starts right
// after "<", at Lt+1.
type Element struct {
	Lt          token.Pos // position of "<"
	Name        string    // as written: div, my-widget, svg:path
	Attrs       []*Attr
	SelfClosing bool      // the start tag ends with "/>"
	Children    []Node    // *Element, *Text, *Interp, *Comment, *Doctype, *If, *For, *Switch, *Select
	Close       token.Pos // position of "<" in "</name>"; NoPos without end tag
	EndPos      token.Pos // position after the last ">" of the element
}

func (e *Element) Pos() token.Pos { return e.Lt }

func (e *Element) End() token.Pos {
	switch {
	case e.EndPos.IsValid():
		return e.EndPos
	case e.Close.IsValid(): // "</name" not closed by ">"
		return offset(e.Close, len("</")+len(e.Name))
	case len(e.Children) > 0:
		return e.Children[len(e.Children)-1].End()
	case len(e.Attrs) > 0:
		return e.Attrs[len(e.Attrs)-1].End()
	}
	return offset(e.Lt, len("<")+len(e.Name))
}

// An Attr is an attribute of a start tag: name, name="v", name='v' or
// name={expr}.
type Attr struct {
	NamePos token.Pos
	Name    string    // as written: class, data-x, hx-get, @click, :class
	Value   AttrValue // nil for a boolean attribute
}

func (a *Attr) Pos() token.Pos { return a.NamePos }

func (a *Attr) End() token.Pos {
	if a.Value != nil {
		return a.Value.End()
	}
	return offset(a.NamePos, len(a.Name))
}

// A StaticValue is a quoted attribute value.
type StaticValue struct {
	ValuePos token.Pos // position of the opening quote
	Raw      string    // source text, quotes included: "btn" or 'btn'
	Value    string    // the value between the quotes
}

func (v *StaticValue) Pos() token.Pos { return v.ValuePos }
func (v *StaticValue) End() token.Pos { return offset(v.ValuePos, len(v.Raw)) }

// A Text is a run of text content.
type Text struct {
	ValuePos token.Pos
	Raw      string // source text
	Value    string // the text to render, as normalised by the parser (SPEC-001 R18)
}

func (t *Text) Pos() token.Pos { return t.ValuePos }
func (t *Text) End() token.Pos { return offset(t.ValuePos, len(t.Raw)) }

// An Interp is an interpolation {X}, as content or as an attribute value.
type Interp struct {
	Lbrace token.Pos  // position of "{"
	X      goast.Expr // nil until grafted
	Rbrace token.Pos  // position of "}"
}

func (x *Interp) Pos() token.Pos { return x.Lbrace }

func (x *Interp) End() token.Pos {
	switch {
	case x.Rbrace.IsValid():
		return x.Rbrace + 1
	case x.X != nil:
		return x.X.End()
	}
	return offset(x.Lbrace, len("{"))
}

// A Comment is an HTML comment, <!-- … -->. Go comments in markup are not
// markup nodes.
type Comment struct {
	Start token.Pos // position of "<"
	Text  string    // source text, delimiters included
}

func (c *Comment) Pos() token.Pos { return c.Start }
func (c *Comment) End() token.Pos { return offset(c.Start, len(c.Text)) }

// A Doctype is a doctype declaration, <!DOCTYPE …>.
type Doctype struct {
	Start token.Pos // position of "<"
	Text  string    // source text, delimiters included
}

func (d *Doctype) Pos() token.Pos { return d.Start }
func (d *Doctype) End() token.Pos { return offset(d.Start, len(d.Text)) }

// An If is an embedded if statement with an optional else branch.
type If struct {
	Keyword token.Pos     // position of "if"
	Stmt    *goast.IfStmt // header: Init and Cond; nil until grafted
	Body    *Block
	Else    Node // nil, *If or *Block
}

func (s *If) Pos() token.Pos { return s.Keyword }

func (s *If) End() token.Pos {
	switch {
	case s.Else != nil:
		return s.Else.End()
	case s.Body != nil:
		return s.Body.End()
	}
	return offset(s.Keyword, len("if"))
}

// A For is an embedded for statement, in any of Go's forms.
type For struct {
	Keyword token.Pos  // position of "for"
	Stmt    goast.Stmt // header: *go/ast.ForStmt or *go/ast.RangeStmt; nil until grafted
	Body    *Block
}

func (s *For) Pos() token.Pos { return s.Keyword }

func (s *For) End() token.Pos {
	if s.Body != nil {
		return s.Body.End()
	}
	return offset(s.Keyword, len("for"))
}

// A Switch is an embedded expression or type switch statement.
type Switch struct {
	Keyword token.Pos  // position of "switch"
	Stmt    goast.Stmt // header: *go/ast.SwitchStmt or *go/ast.TypeSwitchStmt; nil until grafted
	Lbrace  token.Pos
	Cases   []*Case
	Rbrace  token.Pos
}

func (s *Switch) Pos() token.Pos { return s.Keyword }

func (s *Switch) End() token.Pos {
	return clausesEnd(s.Keyword, len("switch"), s.Lbrace, s.Cases, s.Rbrace)
}

// A Select is an embedded select statement.
type Select struct {
	Keyword token.Pos         // position of "select"
	Stmt    *goast.SelectStmt // has no header parts; nil until grafted
	Lbrace  token.Pos
	Cases   []*Case
	Rbrace  token.Pos
}

func (s *Select) Pos() token.Pos { return s.Keyword }

func (s *Select) End() token.Pos {
	return clausesEnd(s.Keyword, len("select"), s.Lbrace, s.Cases, s.Rbrace)
}

// A Case is a case or default clause of an embedded switch or select.
type Case struct {
	Keyword token.Pos  // position of "case" or "default"
	Default bool       // the clause is "default:"
	Clause  goast.Stmt // header: *go/ast.CaseClause (List) or *go/ast.CommClause (Comm); nil until grafted
	Colon   token.Pos
	Body    []Node // content, as in Element.Children
}

func (c *Case) Pos() token.Pos { return c.Keyword }

func (c *Case) End() token.Pos {
	switch {
	case len(c.Body) > 0:
		return c.Body[len(c.Body)-1].End()
	case c.Colon.IsValid():
		return c.Colon + 1
	case c.Default:
		return offset(c.Keyword, len("default"))
	}
	return offset(c.Keyword, len("case"))
}

// A Block is a braced markup body: of an If or a For, or an else branch.
type Block struct {
	Lbrace token.Pos
	List   []Node // content, as in Element.Children
	Rbrace token.Pos
}

func (b *Block) Pos() token.Pos { return b.Lbrace }

func (b *Block) End() token.Pos {
	switch {
	case b.Rbrace.IsValid():
		return b.Rbrace + 1
	case len(b.List) > 0:
		return b.List[len(b.List)-1].End()
	}
	return offset(b.Lbrace, len("{"))
}

func clausesEnd(keyword token.Pos, n int, lbrace token.Pos, cases []*Case, rbrace token.Pos) token.Pos {
	switch {
	case rbrace.IsValid():
		return rbrace + 1
	case len(cases) > 0:
		return cases[len(cases)-1].End()
	case lbrace.IsValid():
		return lbrace + 1
	}
	return offset(keyword, n)
}

// offset returns p+n, or NoPos if p is invalid.
func offset(p token.Pos, n int) token.Pos {
	if !p.IsValid() {
		return token.NoPos
	}
	return p + token.Pos(n)
}

func (*Element) markupNode()     {}
func (*Attr) markupNode()        {}
func (*StaticValue) markupNode() {}
func (*Text) markupNode()        {}
func (*Interp) markupNode()      {}
func (*Comment) markupNode()     {}
func (*Doctype) markupNode()     {}
func (*If) markupNode()          {}
func (*For) markupNode()         {}
func (*Switch) markupNode()      {}
func (*Select) markupNode()      {}
func (*Case) markupNode()        {}
func (*Block) markupNode()       {}

func (*StaticValue) attrValueNode() {}
func (*Interp) attrValueNode()      {}

// The bridge nodes satisfy go/ast's sealed interfaces through embedding.
var (
	_ goast.Decl = (*TemplateDecl)(nil)
	_ goast.Stmt = (*MarkupStmt)(nil)

	_ AttrValue = (*StaticValue)(nil)
	_ AttrValue = (*Interp)(nil)
	_ Node      = (*Element)(nil)
	_ Node      = (*Attr)(nil)
	_ Node      = (*Text)(nil)
	_ Node      = (*Comment)(nil)
	_ Node      = (*Doctype)(nil)
	_ Node      = (*If)(nil)
	_ Node      = (*For)(nil)
	_ Node      = (*Switch)(nil)
	_ Node      = (*Select)(nil)
	_ Node      = (*Case)(nil)
	_ Node      = (*Block)(nil)
)
