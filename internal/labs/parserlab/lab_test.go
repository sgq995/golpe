// Package parserlab holds executable laboratories: small tests that pin down
// assumptions the parser design depends on. If one of these starts failing
// after a Go upgrade, PLAN-001 needs revisiting.
//
// See docs/knowledge/plans/001-template-language.md#laboratories.
package parserlab

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"regexp"
	"strings"
	"testing"
)

// LAB-001: masking `template` as `func    ` and a markup region as `;` plus
// spaces (newlines kept) yields a file go/parser accepts, with an explicit
// EmptyStmt at the exact line:col where the markup started.
func TestLab001MaskedMarkupBecomesEmptyStmt(t *testing.T) {
	golpe := "package p\n\ntemplate Button(props ButtonProps) {\n\tx := 1\n\t<button>{props.Children}</button>\n\t_ = x\n}\n"
	start := strings.Index(golpe, "<button>")
	end := strings.Index(golpe, "</button>") + len("</button>")

	masked := []byte(golpe)
	copy(masked[strings.Index(golpe, "template"):], "func    ")
	for i := start; i < end; i++ {
		if masked[i] != '\n' {
			masked[i] = ' '
		}
	}
	masked[start] = ';'

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "button.golpe", masked, 0)
	if err != nil {
		t.Fatal(err)
	}
	body := f.Decls[0].(*ast.FuncDecl).Body
	var found *ast.EmptyStmt
	for _, s := range body.List {
		if e, ok := s.(*ast.EmptyStmt); ok && !e.Implicit {
			found = e
		}
	}
	if found == nil {
		t.Fatal("no explicit EmptyStmt placeholder in body")
	}
	if got := fset.Position(found.Pos()); got.Line != 5 || got.Column != 2 {
		t.Fatalf("placeholder at %v, want line 5 column 2", got)
	}
	if got := fset.Position(found.Pos()).Offset; got != start {
		t.Fatalf("placeholder offset %d, want %d", got, start)
	}
}

type templateDecl struct{ *ast.FuncDecl }
type markupStmt struct{ *ast.EmptyStmt }

// LAB-002: embedding a go/ast node promotes its unexported marker method,
// so our own node types satisfy go/ast's sealed interfaces.
func TestLab002EmbeddingSatisfiesSealedInterfaces(t *testing.T) {
	var _ ast.Decl = &templateDecl{}
	var _ ast.Stmt = &markupStmt{}
}

// LAB-003: ast.Walk/ast.Inspect panic on foreign node types, so Golpe needs
// its own walker and must lower to pure go/ast before using go/printer.
func TestLab003AstWalkPanicsOnForeignNodes(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected ast.Inspect to panic on a foreign node")
		}
	}()
	ast.Inspect(&ast.BlockStmt{List: []ast.Stmt{&markupStmt{&ast.EmptyStmt{}}}}, func(ast.Node) bool { return true })
}

// LAB-004: at statement start, go/scanner tells a receive statement (`<-ch`,
// token.ARROW) apart from a markup start (`<li`, token.LSS followed by an
// identifier), so markup detection can be done on Go tokens.
func TestLab004ScannerDistinguishesReceiveFromMarkup(t *testing.T) {
	scan := func(src string) []token.Token {
		fset := token.NewFileSet()
		var s scanner.Scanner
		s.Init(fset.AddFile("x", -1, len(src)), []byte(src), nil, 0)
		var toks []token.Token
		for {
			_, tok, _ := s.Scan()
			if tok == token.EOF {
				return toks
			}
			toks = append(toks, tok)
		}
	}
	if got := scan("<-ch"); got[0] != token.ARROW {
		t.Fatalf("<-ch scanned as %v, want ARROW", got)
	}
	if got := scan("<li"); got[0] != token.LSS || got[1] != token.IDENT {
		t.Fatalf("<li scanned as %v, want LSS IDENT", got)
	}
}

// LAB-005: the "skeleton mask". Inside a markup region, tags and text are
// blanked, embedded statement headers (`if c {`, `} else {`) are kept
// verbatim, and an interpolation `{expr}` becomes ` expr;`. One go/parser
// call then parses every Go fragment of the file with exact positions.
func TestLab005SkeletonMaskParsesEmbeddedGo(t *testing.T) {
	golpe := `package p

template NavItem(props NavItemProps) {
	<li>
	if props.Selected {
		<span>{props.Children}</span>
	} else {
		<a href={props.Route}>{props.Children}{"!"}</a> // trailing comment
	}
	</li>
}
`
	masked := labMask(golpe)
	if len(masked) != len(golpe) {
		t.Fatalf("mask must preserve length: %d != %d", len(masked), len(golpe))
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "nav.golpe", masked, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	var exprs []string
	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.ExprStmt); ok {
			off := fset.Position(s.X.Pos()).Offset
			end := fset.Position(s.X.End()).Offset
			exprs = append(exprs, golpe[off:end]) // slice the ORIGINAL source
		}
		return true
	})
	want := []string{"props.Children", "props.Route", "props.Children", `"!"`}
	if strings.Join(exprs, "|") != strings.Join(want, "|") {
		t.Fatalf("interpolations = %q, want %q", exprs, want)
	}
}

// LAB-006: `/*line file:line:col*/` comments in generated Go survive
// go/format and make go/token report positions in the .golpe file, which is
// what the compiler, go vet and gopls diagnostics use.
func TestLab006LineDirectivesMapToGolpeSource(t *testing.T) {
	gen := "package p\n\nfunc f() any {\n\treturn /*line nav.golpe:8:13*/props.Route\n}\n"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "nav_golpe.go", gen, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	ret := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ReturnStmt)
	got := fset.Position(ret.Results[0].Pos())
	if got.Filename != "nav.golpe" || got.Line != 8 || got.Column != 13 {
		t.Fatalf("position = %v, want nav.golpe:8:13", got)
	}
}

// labMask is a deliberately naive stand-in for internal/parser's masker,
// good enough for LAB-005's input: interpolations directly follow `>`, `=`
// or `}` and never contain braces.
func labMask(src string) string {
	b := []byte(strings.Replace(src, "template", "func    ", 1))
	interp := regexp.MustCompile(`\{[^{}\n]*\}`)
	keep := make([]bool, len(b))
	for _, m := range interp.FindAllStringIndex(string(b), -1) {
		lo, hi := m[0], m[1]
		if !strings.ContainsRune(">=}", rune(src[lo-1])) {
			continue
		}
		b[lo], b[hi-1] = ' ', ';'
		for i := lo; i < hi; i++ {
			keep[i] = true
		}
	}
	tag := regexp.MustCompile(`</?[a-z][^<>]*>`)
	for i, m := range tag.FindAllStringIndex(string(b), -1) {
		for j := m[0]; j < m[1]; j++ {
			if !keep[j] {
				b[j] = ' '
			}
		}
		if i == 0 {
			b[m[0]] = ';' // root markup start becomes the EmptyStmt placeholder
		}
	}
	return string(b)
}

func TestLabMaskShape(t *testing.T) {
	got := labMask("package p\ntemplate X() {\n\t<p>{a}{b}</p>\n}\n")
	want := "package p\nfunc     X() {\n\t;   a; b;    \n}\n"
	if got != want {
		t.Fatalf("labMask = %q, want %q", got, want)
	}
}

// LAB-007: go/format.Source puts a space between a block comment and the
// next token, so `/*line f:8:13*/props.Route` drifts to column 14 after
// gofmt. The generator therefore emits the directive for column C-1 followed
// by an explicit space: `/*line f:8:12*/ props.Route` maps to 8:13 and is
// already gofmt-stable. Code that starts at column 1 (top-level declarations)
// uses a `//line f:L:1` line instead, which gofmt does not indent.
func TestLab007LineDirectivesSurviveGofmt(t *testing.T) {
	cases := []struct {
		name, gen string
		line, col int
	}{
		{"naive inline drifts", "package p\nfunc f() any {\nif true {\n      return /*line nav.golpe:8:13*/props.Route }\nreturn nil}\n", 8, 14},
		{"C-1 plus space", "package p\nfunc f() any {\nif true {\n      return /*line nav.golpe:8:12*/ props.Route }\nreturn nil}\n", 8, 13},
		{"line comment at column 1", "package p\n//line nav.golpe:20:1\nvar x = a.b\n", 20, 9},
	}
	for _, c := range cases {
		out, err := format.Source([]byte(c.gen))
		if err != nil {
			t.Fatal(err)
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "nav_golpe.go", out, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		var pos token.Position
		ast.Inspect(f, func(n ast.Node) bool {
			if s, ok := n.(*ast.SelectorExpr); ok {
				pos = fset.Position(s.Pos())
			}
			return true
		})
		if pos.Filename != "nav.golpe" || pos.Line != c.line || pos.Column != c.col {
			t.Errorf("%s: after gofmt selector is at %v, want nav.golpe:%d:%d\n%s", c.name, pos, c.line, c.col, out)
		}
	}
}
