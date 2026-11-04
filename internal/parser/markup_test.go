package parser

import (
	"bytes"
	"errors"
	goast "go/ast"
	goparser "go/parser"
	"go/scanner"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/sgq995/golpe/internal/ast"
)

// markupCase is a markup source, from its first '<' (anything before is
// context), with the expected ast.Fprint dump of the region's root and the
// expected errors, one "line:col: message" per line. A dump of "-" is not
// checked.
type markupCase struct {
	name, src, dump, errs string
}

// markupParse parses the markup statement at the first '<' of src.
func markupParse(t *testing.T, src string) (fset *token.FileSet, r *region, errs error) {
	t.Helper()
	fset = token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	file.SetLinesForContent([]byte(src))
	var l errorList
	r = parseMarkup(file, []byte(src), strings.Index(src, "<"), &l)
	return fset, r, l.err()
}

// markupDump returns the dump of n, or "" for nil.
func markupDump(t *testing.T, fset *token.FileSet, n goast.Node) string {
	t.Helper()
	if n == nil {
		return ""
	}
	var b strings.Builder
	if err := ast.Fprint(&b, fset, n); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// markupErrors returns err, a scanner.ErrorList or nil, one error per line.
func markupErrors(err error) string {
	var list scanner.ErrorList
	if !errors.As(err, &list) {
		return ""
	}
	lines := make([]string, len(list))
	for i, e := range list {
		lines[i] = e.Error()
	}
	return strings.Join(lines, "\n")
}

func runMarkupCases(t *testing.T, cases []markupCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fset, r, err := markupParse(t, c.src)
			if got := markupErrors(err); got != c.errs {
				t.Errorf("%q: errors\n%s\nwant\n%s", c.src, got, c.errs)
			}
			if c.dump == "-" {
				return
			}
			if got := markupDump(t, fset, r.root); got != c.dump {
				t.Errorf("%q: dump\n%s\nwant\n%s", c.src, got, c.dump)
			}
		})
	}
}

func TestMarkupElements(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "nested", src: `<div><p>hi</p><span>x</span></div>`, dump: `Element 1:1-1:35 div close=1:29
  Element 1:6-1:15 p close=1:11
    Text 1:9-1:11 "hi"
  Element 1:15-1:29 span close=1:22
    Text 1:21-1:22 "x"
`},
		{name: "void", src: `<p>a<br>b<img src="x.png"></p>`, dump: `Element 1:1-1:31 p close=1:27
  Text 1:4-1:5 "a"
  Element 1:5-1:9 br
  Text 1:9-1:10 "b"
  Element 1:10-1:27 img
    Attr 1:15-1:26 src
      StaticValue 1:19-1:26 "x.png"
`},
		{name: "void root, any case", src: `<BR>`, dump: "Element 1:1-1:5 BR\n"},
		{name: "self-closing", src: `<div><x/><my-el a="1"/><br/></div>`, dump: `Element 1:1-1:35 div close=1:29
  Element 1:6-1:10 x self-closing
  Element 1:10-1:24 my-el self-closing
    Attr 1:17-1:22 a
      StaticValue 1:19-1:22 "1"
  Element 1:24-1:29 br self-closing
`},
		{name: "self-closing after a space", src: `<x />`, dump: "Element 1:1-1:6 x self-closing\n"},
		{name: "custom element", src: `<my-widget>hi</my-widget>`, dump: `Element 1:1-1:26 my-widget close=1:14
  Text 1:12-1:14 "hi"
`},
		{name: "close tag ignores ASCII case", src: `<DIV>x</div>`, dump: `Element 1:1-1:13 DIV close=1:7
  Text 1:6-1:7 "x"
`},
		{name: "close tag with a space", src: `<p>x</p >`, dump: `Element 1:1-1:10 p close=1:5
  Text 1:4-1:5 "x"
`},
		{name: "start tag on several lines", src: "<a\n\thref=\"/\"\n\tclass=\"b\">x</a>", dump: `Element 1:1-3:17 a close=3:13
  Attr 2:2-2:10 href
    StaticValue 2:7-2:10 "/"
  Attr 3:2-3:11 class
    StaticValue 3:8-3:11 "b"
  Text 3:12-3:13 "x"
`},
		{name: "names as authored", src: `<svg:path/>`, dump: "Element 1:1-1:12 svg:path self-closing\n"},
	})
}

func TestMarkupFragments(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "nested with text", src: `<>a<>b</>c</>`, dump: `Fragment 1:1-1:14 close=1:11
  Text 1:3-1:4 "a"
  Fragment 1:4-1:10 close=1:7
    Text 1:6-1:7 "b"
  Text 1:10-1:11 "c"
`},
		{name: "empty", src: `<></>`, dump: "Fragment 1:1-1:6 close=1:3\n"},
		{name: "in an element", src: `<ul><><li>x</li></></ul>`, dump: `Element 1:1-1:25 ul close=1:20
  Fragment 1:5-1:20 close=1:17
    Element 1:7-1:17 li close=1:12
      Text 1:11-1:12 "x"
`},
		{name: "whitespace at both ends dropped", src: "<>\n\tHello\n</>", dump: `Fragment 1:1-3:4 close=3:1
  Text 1:3-3:1 "Hello"
`},
		{name: "bare interpolation", src: `<>{p.Children(writer())}</>`, dump: `Fragment 1:1-1:28 close=1:25
  Interp 1:3-1:25 {...}
`},
	})
}

func TestMarkupAttributes(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "quoted values are literal", src: `<div x-data="{ open: false }" title='it "is"' data-x=""/>`, dump: `Element 1:1-1:58 div self-closing
  Attr 1:6-1:30 x-data
    StaticValue 1:13-1:30 "{ open: false }"
  Attr 1:31-1:46 title
    StaticValue 1:37-1:46 "it \"is\""
  Attr 1:47-1:56 data-x
    StaticValue 1:54-1:56 ""
`},
		{name: "boolean and dynamic", src: `<input disabled value={v} checked={c.On()}>`, dump: `Element 1:1-1:44 input
  Attr 1:8-1:16 disabled
  Attr 1:17-1:26 value
    Interp 1:23-1:26 {...}
  Attr 1:27-1:43 checked
    Interp 1:35-1:43 {...}
`},
		{name: "names as authored", src: `<svg viewBox="0 0 1 1" xlink:href="#a" hx-get="/x" @click="go()" :class="c"/>`, dump: `Element 1:1-1:78 svg self-closing
  Attr 1:6-1:23 viewBox
    StaticValue 1:14-1:23 "0 0 1 1"
  Attr 1:24-1:39 xlink:href
    StaticValue 1:35-1:39 "#a"
  Attr 1:40-1:51 hx-get
    StaticValue 1:47-1:51 "/x"
  Attr 1:52-1:65 @click
    StaticValue 1:59-1:65 "go()"
  Attr 1:66-1:76 :class
    StaticValue 1:73-1:76 "c"
`},
		{name: "> inside quotes", src: `<a title="a > b">x</a>`, dump: `Element 1:1-1:23 a close=1:19
  Attr 1:4-1:17 title
    StaticValue 1:10-1:17 "a > b"
  Text 1:18-1:19 "x"
`},
		{name: "> inside a dynamic value", src: `<a b={x > 1}/>`, dump: `Element 1:1-1:15 a self-closing
  Attr 1:4-1:13 b
    Interp 1:6-1:13 {...}
`},
		{name: "spaces around =", src: `<a href = "x"/>`, dump: `Element 1:1-1:16 a self-closing
  Attr 1:4-1:14 href
    StaticValue 1:11-1:14 "x"
`},
		{name: "value on two lines", src: "<p title=\"a\nb\"/>", dump: `Element 1:1-2:5 p self-closing
  Attr 1:4-2:3 title
    StaticValue 1:10-2:3 "a\nb"
`},
		{name: "no space after a value", src: `<a b="1"c/>`, dump: `Element 1:1-1:12 a self-closing
  Attr 1:4-1:9 b
    StaticValue 1:6-1:9 "1"
  Attr 1:9-1:10 c
`},
		{name: "duplicate ignoring ASCII case", src: `<a class="x" CLASS="y"/>`, dump: "-",
			errs: `1:14: duplicate attribute "CLASS"`},
		{name: "duplicate dynamic", src: `<a href="x" href={y}/>`, dump: "-",
			errs: `1:13: duplicate attribute "href"`},
		{name: "unquoted", src: `<a class=btn>x</a>`, dump: `Element 1:1-1:19 a close=1:15
  Attr 1:4-1:9 class
  Text 1:14-1:15 "x"
`, errs: `1:10: unquoted value for attribute "class"; write class="btn"`},
		{name: "unquoted before />", src: `<img src=a.png/>`, dump: `Element 1:1-1:17 img self-closing
  Attr 1:6-1:9 src
`, errs: `1:10: unquoted value for attribute "src"; write src="a.png"`},
		{name: "unquoted with a double quote", src: `<a b=x"y/>`, dump: "-", errs: `1:6: unquoted value for attribute "b"; write b='x"y'`},
		{name: "missing value", src: `<a href=>x</a>`, dump: "-", errs: `1:8: missing value for attribute "href"`},
		{name: "missing value at the end", src: `<a href=`, dump: "-", errs: `1:1: start tag <a> not terminated`},
		{name: "spread", src: `<a {props} b/>`, dump: `Element 1:1-1:15 a self-closing
  Attr 1:12-1:13 b
`, errs: `1:4: attribute spread is not supported; write name={expr}`},
		{name: "spread not terminated", src: `<a {x`, dump: "Element 1:1-1:3 a\n",
			errs: `1:4: attribute spread is not supported; write name={expr}`},
		{name: "empty interpolation", src: `<a href={}/>`, dump: "-", errs: `1:9: empty interpolation`},
		{name: "unexpected quote", src: `<a "x"/>`, dump: "-", errs: `1:4: unexpected '"' in start tag <a>`},
		{name: "unexpected slash", src: `<a b/c/>`, dump: `Element 1:1-1:9 a self-closing
  Attr 1:4-1:5 b
  Attr 1:6-1:7 c
`, errs: `1:5: unexpected '/' in start tag <a>`},
		{name: "value not terminated", src: `<p class="x`, dump: "-", errs: `1:10: attribute value not terminated`},
		{name: "interpolation not terminated", src: `<a href={x/>`, dump: "-", errs: `1:9: interpolation not terminated`},
	})
}

func TestMarkupInterpolations(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "nested braces", src: `<p>{T{A: 1}.String()}</p>`, dump: `Element 1:1-1:26 p close=1:22
  Interp 1:4-1:22 {...}
`},
		{name: "braces in strings and runes", src: "<p>{\"}\"}{'{'}{`}{`}</p>", dump: `Element 1:1-1:24 p close=1:20
  Interp 1:4-1:9 {...}
  Interp 1:9-1:14 {...}
  Interp 1:14-1:20 {...}
`},
		{name: "function literal", src: `<p>{func() string { return "}" }()}</p>`, dump: `Element 1:1-1:40 p close=1:36
  Interp 1:4-1:36 {...}
`},
		{name: "brace in a comment", src: `<p>{x /* } */}</p>`, dump: `Element 1:1-1:19 p close=1:15
  Interp 1:4-1:15 {...}
`},
		{name: "several lines", src: "<p>{f(\n\ta,\n)}</p>", dump: `Element 1:1-3:7 p close=3:3
  Interp 1:4-3:3 {...}
`},
		{name: "between text", src: `<p>a {b} c</p>`, dump: `Element 1:1-1:15 p close=1:11
  Text 1:4-1:6 "a "
  Interp 1:6-1:9 {...}
  Text 1:9-1:11 " c"
`},
		{name: "empty", src: `<p>{}</p>`, dump: "Element 1:1-1:10 p close=1:6\n",
			errs: `1:4: empty interpolation; write {"{}"} for literal braces`},
		{name: "only a comment", src: "<p>{ /* x */\n}</p>", dump: "-",
			errs: `1:4: empty interpolation; write {"{}"} for literal braces`},
		{name: "not terminated", src: `<p>{x</p>`, dump: "Element 1:1-1:3 p\n", errs: `1:4: interpolation not terminated`},
	})
}

func TestMarkupRawAndTextOnly(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "script is raw", src: `<script>if (a < b) { c() } // x</script>`, dump: `Element 1:1-1:41 script close=1:32
  Text 1:9-1:32 "if (a < b) { c() } // x"
`},
		{name: "style ends at its close tag in any case", src: `<style>a{b:c}</s </STYLE>`, dump: `Element 1:1-1:26 style close=1:18
  Text 1:8-1:18 "a{b:c}</s "
`},
		{name: "a longer name does not close", src: `<script>"</scripts>"</script>`, dump: `Element 1:1-1:30 script close=1:21
  Text 1:9-1:21 "\"</scripts>\""
`},
		{name: "raw keeps whitespace and lines", src: "<script>\nif (a) {\n}\n</script>", dump: `Element 1:1-4:10 script close=4:1
  Text 1:9-4:1 "\nif (a) {\n}\n"
`},
		{name: "empty script", src: `<script></script>`, dump: "Element 1:1-1:18 script close=1:9\n"},
		{name: "close tag with a space", src: `<SCRIPT>x</script >`, dump: `Element 1:1-1:20 SCRIPT close=1:10
  Text 1:9-1:10 "x"
`},
		{name: "self-closing script", src: `<script src="a.js"/>`, dump: `Element 1:1-1:21 script self-closing
  Attr 1:9-1:19 src
    StaticValue 1:13-1:19 "a.js"
`},
		{name: "script not terminated", src: `<script>x`, dump: `Element 1:1-1:10 script
  Text 1:9-1:10 "x"
`, errs: `1:10: expected </script>, found 'EOF'`},
		{name: "script close tag not terminated", src: `<script>x</script`, dump: `Element 1:1-1:18 script close=1:10
  Text 1:9-1:10 "x"
`, errs: `1:10: close tag </script> not terminated`},
		{name: "script close tag followed by /", src: `<script>x</script/>`, dump: `Element 1:1-1:20 script close=1:10
  Text 1:9-1:10 "x"
`, errs: `1:10: close tag </script> not terminated`},
		{name: "textarea keeps whitespace", src: "<textarea>\n  {v}  x\n</textarea>", dump: `Element 1:1-3:12 textarea close=3:1
  Text 1:11-2:3 "\n  "
  Interp 2:3-2:6 {...}
  Text 2:6-3:1 "  x\n"
`},
		{name: "textarea text is text", src: "<textarea>\n// x\nif y\n</textarea>", dump: `Element 1:1-4:12 textarea close=4:1
  Text 1:11-4:1 "\n// x\nif y\n"
`},
		{name: "title collapses whitespace", src: `<title> A  {t} </title>`, dump: `Element 1:1-1:24 title close=1:16
  Text 1:8-1:12 "A "
  Interp 1:12-1:15 {...}
`},
		{name: "title text is text", src: "<title>\nfor sale\n</title>", dump: `Element 1:1-3:9 title close=3:1
  Text 1:8-3:1 "for sale"
`},
		{name: "title rejects tags", src: `<title>a <b>c</b></title>`, dump: `Element 1:1-1:26 title close=1:18
  Text 1:8-1:10 "a "
  Element 1:10-1:18 b close=1:14
    Text 1:13-1:14 "c"
`, errs: `1:10: <title> can contain only text and interpolations`},
		{name: "textarea rejects comments", src: `<textarea><!-- x --></textarea>`, dump: "-",
			errs: `1:11: <textarea> can contain only text and interpolations`},
		{name: "title rejects fragments", src: "<title>\n<>x</>\n</title>", dump: "-",
			errs: `2:1: <title> can contain only text and interpolations`},
		{name: "pre keeps whitespace, also in nested elements", src: "<pre>  a\n <b> b  c </b>\n</pre>", dump: `Element 1:1-3:7 pre close=3:1
  Text 1:6-2:2 "  a\n "
  Element 2:2-2:15 b close=2:11
    Text 2:5-2:11 " b  c "
  Text 2:15-3:1 "\n"
`},
		{name: "pre in any case", src: `<PRE> a </PRE>`, dump: `Element 1:1-1:15 PRE close=1:9
  Text 1:6-1:9 " a "
`},
		{name: "pre drops Go comments, keeps the whitespace around", src: "<pre>a\n// c\nb</pre>", dump: `Element 1:1-3:8 pre close=3:2
  Text 1:6-2:1 "a\n"
  Text 2:5-3:2 "\nb"
`},
	})
}

func TestMarkupComments(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "HTML comment", src: `<div><!-- a {b} <c> --></div>`, dump: `Element 1:1-1:30 div close=1:24
  HTMLComment 1:6-1:24 "<!-- a {b} <c> -->"
`},
		{name: "HTML comment root", src: `<!-- c -->`, dump: "HTMLComment 1:1-1:11 \"<!-- c -->\"\n"},
		{name: "HTML comment on lines", src: "<!--\n x\n-->", dump: "HTMLComment 1:1-3:4 \"<!--\\n x\\n-->\"\n"},
		{name: "HTML comment not terminated", src: `<p><!-- x</p>`, dump: "Element 1:1-1:3 p\n", errs: `1:4: HTML comment not terminated`},
		{name: "doctype", src: `<!DOCTYPE html>`, dump: "Doctype 1:1-1:16 \"<!DOCTYPE html>\"\n"},
		{name: "doctype in any case", src: `<!doctype html>`, dump: "Doctype 1:1-1:16 \"<!doctype html>\"\n"},
		{name: "doctype not terminated", src: `<!DOCTYPE html`, dump: "", errs: `1:1: doctype not terminated`},
		{name: "doctype inside markup", src: `<div><!DOCTYPE html></div>`, dump: "Element 1:1-1:27 div close=1:21\n",
			errs: `1:6: misplaced <!DOCTYPE>: it must be a markup statement of its own`},
		{name: "other <! markup", src: `<div><!x></div>`, dump: "Element 1:1-1:16 div close=1:10\n",
			errs: `1:6: invalid "<!" markup: only <!-- … --> and <!DOCTYPE …> are allowed`},
		{name: "other <! markup as root", src: `<![CDATA[x]]>`, dump: "",
			errs: `1:1: invalid "<!" markup: only <!-- … --> and <!DOCTYPE …> are allowed`},
		{name: "other <! markup without >", src: "<p><!x\n</p>", dump: `Element 1:1-2:5 p close=2:1
  Text 1:6-2:1 "x"
`, errs: `1:4: invalid "<!" markup: only <!-- … --> and <!DOCTYPE …> are allowed`},
		{name: "Go comments at node boundaries", src: "<ul>\n\t// a line comment\n\t<li>a</li> // trailing\n\t/* block */ <li>b</li>\n\t<li>c</li> /* after */\n</ul>", dump: `Element 1:1-6:6 ul close=6:1
  Element 3:2-3:12 li close=3:7
    Text 3:6-3:7 "a"
  Text 3:12-3:13 " "
  Element 4:14-4:24 li close=4:19
    Text 4:18-4:19 "b"
  Text 4:24-5:2 " "
  Element 5:2-5:12 li close=5:7
    Text 5:6-5:7 "c"
`},
		{name: "URLs are text", src: `<a href="https://x.dev">see https://x.dev</a>`, dump: `Element 1:1-1:46 a close=1:42
  Attr 1:4-1:24 href
    StaticValue 1:9-1:24 "https://x.dev"
  Text 1:25-1:42 "see https://x.dev"
`},
		{name: "// right after a tag is text", src: `<p><b>a</b>// x</p>`, dump: `Element 1:1-1:20 p close=1:16
  Element 1:4-1:12 b close=1:8
    Text 1:7-1:8 "a"
  Text 1:12-1:16 "// x"
`},
		{name: "// inside text is text", src: `<p>a // b</p>`, dump: `Element 1:1-1:14 p close=1:10
  Text 1:4-1:10 "a // b"
`},
		{name: "after an interpolation", src: "<p>{x} // c\n</p>", dump: `Element 1:1-2:5 p close=2:1
  Interp 1:4-1:7 {...}
`},
		{name: "after an HTML comment", src: `<p><!-- x --> /* c */</p>`, dump: `Element 1:1-1:26 p close=1:22
  HTMLComment 1:4-1:14 "<!-- x -->"
`},
		{name: "at the start of a line in text", src: "<p>a\n/* c */ b</p>", dump: `Element 1:1-2:14 p close=2:10
  Text 1:4-2:1 "a "
  Text 2:8-2:10 "b"
`},
		{name: "Go comment not terminated", src: "<p>\n\t/* x</p>", dump: "Element 1:1-1:3 p\n", errs: `2:2: comment not terminated`},
	})
}

func TestMarkupWhitespace(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "spec paragraph", src: "<p>\n  Hello,\n  {name}\n</p>", dump: `Element 1:1-4:5 p close=4:1
  Text 1:4-3:3 "Hello, "
  Interp 3:3-3:9 {...}
`},
		{name: "spec bold", src: "<div><b>Hi</b>\n  there</div>", dump: `Element 1:1-2:14 div close=2:8
  Element 1:6-1:15 b close=1:11
    Text 1:9-1:11 "Hi"
  Text 1:15-2:8 " there"
`},
		{name: "runs between elements", src: "<ul>\n\t<li>a</li>\n\t<li>b</li>\n</ul>", dump: `Element 1:1-4:6 ul close=4:1
  Element 2:2-2:12 li close=2:7
    Text 2:6-2:7 "a"
  Text 2:12-3:2 " "
  Element 3:2-3:12 li close=3:7
    Text 3:6-3:7 "b"
`},
		{name: "a run becomes one space", src: "<p>a \t\n b</p>", dump: `Element 1:1-2:7 p close=2:3
  Text 1:4-2:3 "a b"
`},
		{name: "only whitespace", src: "<p> \n </p>", dump: "Element 1:1-2:6 p close=2:2\n"},
		{name: "around a single child", src: "<p>\n\t<b>x</b>\n</p>", dump: `Element 1:1-3:5 p close=3:1
  Element 2:2-2:10 b close=2:6
    Text 2:5-2:6 "x"
`},
		{name: "Go comment at the end", src: "<li><a>x</a> // note\n</li>", dump: `Element 1:1-2:6 li close=2:1
  Element 1:5-1:13 a close=1:9
    Text 1:8-1:9 "x"
`},
		{name: "one run across a Go comment", src: `<p><b>a</b> /* c */ <i>b</i></p>`, dump: `Element 1:1-1:33 p close=1:29
  Element 1:4-1:12 b close=1:8
    Text 1:7-1:8 "a"
  Text 1:12-1:13 " "
  Element 1:21-1:29 i close=1:25
    Text 1:24-1:25 "b"
`},
		{name: "entities and > are text", src: `<p>&nbsp;a&amp; > b</p>`, dump: `Element 1:1-1:24 p close=1:20
  Text 1:4-1:20 "&nbsp;a&amp; > b"
`},
	})
}

func TestMarkupErrors(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "mismatch", src: `<div><span></div>`, dump: `Element 1:1-1:18 div close=1:12
  Element 1:6-1:11 span
`, errs: `1:12: expected </span>, found </div>`},
		{name: "mismatch reported once while unwinding", src: `<ul><li><b>x</ul>`, dump: `Element 1:1-1:18 ul close=1:13
  Element 1:5-1:13 li
    Element 1:9-1:13 b
      Text 1:12-1:13 "x"
`, errs: `1:13: expected </b>, found </ul>`},
		{name: "close tag closing nothing", src: `<p>x</b></p>`, dump: `Element 1:1-1:13 p close=1:9
  Text 1:4-1:5 "x"
`, errs: `1:5: expected </p>, found </b>`},
		{name: "void element with content", src: `<div><br>x</br></div>`, dump: `Element 1:1-1:22 div close=1:16
  Element 1:6-1:10 br
  Text 1:10-1:11 "x"
`, errs: `1:11: void element <br> cannot have content or a close tag`},
		{name: "fragment mismatch", src: `<><b></>`, dump: `Fragment 1:1-1:9 close=1:6
  Element 1:3-1:5 b
`, errs: `1:6: expected </b>, found </>`},
		{name: "element not closed", src: `<p>x`, dump: `Element 1:1-1:5 p
  Text 1:4-1:5 "x"
`, errs: `1:5: expected </p>, found 'EOF'`},
		{name: "fragment not closed", src: `<>x`, dump: `Fragment 1:1-1:4
  Text 1:3-1:4 "x"
`, errs: `1:4: expected </>, found 'EOF'`},
		{name: "start tag not terminated", src: `<p class="x"`, dump: `Element 1:1-1:13 p
  Attr 1:4-1:13 class
    StaticValue 1:10-1:13 "x"
`, errs: `1:1: start tag <p> not terminated`},
		{name: "start tag cut by a tag", src: "<div\n<p>x</p></div>", dump: `Element 1:1-2:15 div close=2:9
  Element 2:1-2:9 p close=2:5
    Text 2:4-2:5 "x"
`, errs: `1:1: start tag <div> not terminated`},
		{name: "close tag not terminated", src: `<p>x</p`, dump: `Element 1:1-1:8 p close=1:5
  Text 1:4-1:5 "x"
`, errs: `1:5: close tag </p> not terminated`},
		{name: "close tag with junk", src: `<p>x</p class="y"> z`, dump: `Element 1:1-1:19 p close=1:5
  Text 1:4-1:5 "x"
`, errs: `1:5: close tag </p> not terminated`},
		{name: "close tag cut by a tag", src: "<div><p>x</p\n</div>", dump: `Element 1:1-2:7 div close=2:1
  Element 1:6-1:13 p close=1:10
    Text 1:9-1:10 "x"
`, errs: `1:10: close tag </p> not terminated`},
		{name: "stray close tag between texts", src: `<p>a </b> b</p>`, dump: `Element 1:1-1:16 p close=1:12
  Text 1:4-1:6 "a "
  Text 1:10-1:12 "b"
`, errs: `1:6: expected </p>, found </b>`},
		{name: "stray } after an interpolation", src: `<p>{x}}</p>`, dump: "-", errs: `1:7: unexpected '}' in text; write {"}"} for a literal '}'`},
		{name: "stray }", src: `<p>a } b</p>`, dump: `Element 1:1-1:13 p close=1:9
  Text 1:4-1:9 "a } b"
`, errs: `1:6: unexpected '}' in text; write {"}"} for a literal '}'`},
		{name: "stray <", src: `<p>a < b</p>`, dump: `Element 1:1-1:13 p close=1:9
  Text 1:4-1:9 "a < b"
`, errs: `1:6: unexpected '<' in text; write {"<"} for a literal '<'`},
		{name: "< before a digit", src: `<p><3</p>`, dump: "-", errs: `1:4: unexpected '<' in text; write {"<"} for a literal '<'`},
		{name: "</ before a digit", src: `<p></3</p>`, dump: "-", errs: `1:4: unexpected '<' in text; write {"<"} for a literal '<'`},
		{name: "one error per line", src: "<p>{}{}\n}</p>", dump: "-", errs: `1:4: empty interpolation; write {"{}"} for literal braces
2:1: expected </p>, found '}'`},
		{name: "not markup", src: `<3`, dump: "", errs: `1:1: expected markup`},
	})
}

// TestMarkupLineRule pins the hook TASK-001-05 replaces: the first word of
// every line of content is checked for an embedded statement keyword, whole
// words only, except in text-only and raw text elements.
func TestMarkupLineRule(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "keyword", src: "<p>\n\tif you can\n</p>", dump: `Element 1:1-3:5 p close=3:1
  Text 1:4-3:1 "if you can"
`, errs: `2:2: embedded statements are not supported yet`},
		{name: "every keyword", src: "<p>\nif(x)\n  for\n\tswitch x\nselect\n</p>", dump: "-", errs: `2:1: embedded statements are not supported yet
3:3: embedded statements are not supported yet
4:2: embedded statements are not supported yet
5:1: embedded statements are not supported yet`},
		{name: "whole words only", src: "<p>\nformat\niffy\nif_x\nIf\nifé\nfor2\n</p>", dump: "-"},
		{name: "not at the start of a line", src: `<p>if x</p>`, dump: "-"},
		{name: "after a Go comment at the start of a line", src: "<p>\n/* c */ for\n</p>", dump: "-", errs: `2:9: embedded statements are not supported yet`},
		{name: "in pre", src: "<pre>\nfor x\n</pre>", dump: "-", errs: `2:1: embedded statements are not supported yet`},
		{name: "a } line ends unclosed markup", src: "<p>hello\n}", dump: `Element 1:1-2:1 p
  Text 1:4-2:1 "hello"
`, errs: `2:1: expected </p>, found '}'`},
		{name: "a } line in a text-only element", src: "<title>x\n  }", dump: "-", errs: `2:3: expected </title>, found '}'`},
	})
}

// markupMask is a test-only version of pass 3 (TASK-001-06), following the
// region documentation: it blanks every region except its kept spans and
// newlines, writes ';' at its start and its semis, and turns the first
// "template" into "func    ".
func markupMask(src []byte, regions ...*region) []byte {
	out := bytes.Clone(src)
	for _, r := range regions {
		k := 0
		for i := r.start; i < r.end; i++ {
			for k < len(r.keep) && r.keep[k].to <= i {
				k++
			}
			if out[i] != '\n' && !(k < len(r.keep) && r.keep[k].from <= i) {
				out[i] = ' '
			}
		}
		out[r.start] = ';'
		for _, s := range r.semis {
			out[s] = ';'
		}
	}
	if i := bytes.Index(out, []byte("template")); i >= 0 {
		copy(out[i:], "func    ")
	}
	return out
}

// markupGoParse masks src around the region at the first '<' after the
// first "{", parses the skeleton with go/parser into fset (which may already
// hold files) and returns the region, the file and the mask.
func markupGoParse(t *testing.T, fset *token.FileSet, src string) (*region, *goast.File, []byte, error) {
	t.Helper()
	base := fset.Base()
	lf := lineFile("x.golpe", base, []byte(src))
	var l errorList
	open := strings.Index(src, "{")
	r := parseMarkup(lf, []byte(src), open+strings.Index(src[open:], "<"), &l)
	m := markupMask([]byte(src), r)
	if len(m) != len(src) {
		t.Fatalf("mask changed the length")
	}
	f, err := goparser.ParseFile(fset, "x.golpe", m, goparser.ParseComments|goparser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("skeleton does not parse: %v\n%s", err, m)
	}
	if f.FileStart != token.Pos(base) {
		t.Fatalf("go/parser added the file at %d, markup positions assumed %d", f.FileStart, base)
	}
	return r, f, m, l.err()
}

// exprsIn returns the source of the expression statements go/parser found
// in the spans of the interpolations, as the graft will.
func exprsIn(t *testing.T, fset *token.FileSet, f *goast.File, src string, r *region) []string {
	t.Helper()
	var got []string
	for _, in := range r.interps {
		var x goast.Expr
		goast.Inspect(f, func(n goast.Node) bool {
			if s, ok := n.(*goast.ExprStmt); ok && in.Lbrace < s.Pos() && s.End() <= in.Rbrace {
				x = s.X
			}
			return x == nil
		})
		if x == nil {
			t.Fatalf("no expression for the interpolation at %v", fset.Position(in.Lbrace))
		}
		got = append(got, src[fset.Position(x.Pos()).Offset:fset.Position(x.End()).Offset])
	}
	return got
}

// TestMarkupRegion checks what the region hands to the mask and the graft,
// and that go/parser parses the masked skeleton with the interpolation
// expressions and the Go comment at their places.
func TestMarkupRegion(t *testing.T) {
	src := "package p\n\ntemplate T(u, t string) {\n\tx := 1\n\t<a href={u} title=\"t\">{t} // c\n\t</a>\n\t_ = x\n}\n"
	fset := token.NewFileSet()
	fset.AddFile("other.go", -1, 100) // the markup file is not the first
	r, f, m, err := markupGoParse(t, fset, src)
	if err != nil {
		t.Fatal(err)
	}
	if r.start != 46 || r.end != 82 {
		t.Errorf("region [%d, %d), want [46, 82)", r.start, r.end)
	}
	if want := []span{{55, 56}, {69, 70}, {72, 76}}; !slices.Equal(r.keep, want) {
		t.Errorf("keep %v, want %v", r.keep, want)
	}
	if want := []int{56, 70}; !slices.Equal(r.semis, want) {
		t.Errorf("semis %v, want %v", r.semis, want)
	}
	wantMask := "package p\n\nfunc     T(u, t string) {\n\tx := 1\n\t;        u;            t; // c\n     \n\t_ = x\n}\n"
	if string(m) != wantMask {
		t.Errorf("mask\n%q\nwant\n%q", m, wantMask)
	}
	if got := exprsIn(t, fset, f, src, r); !slices.Equal(got, []string{"u", "t"}) {
		t.Errorf("interpolation expressions %q", got)
	}
	if len(f.Comments) != 1 || f.Comments[0].Text() != "c\n" || fset.Position(f.Comments[0].Pos()).Offset != 72 {
		t.Errorf("comments: %v", f.Comments)
	}
	wantDump := `Element 5:2-6:6 a close=6:2
  Attr 5:5-5:13 href
    Interp 5:10-5:13 {...}
  Attr 5:14-5:23 title
    StaticValue 5:20-5:23 "t"
  Interp 5:24-5:27 {...}
`
	if got := markupDump(t, fset, r.root); got != wantDump {
		t.Errorf("dump with go/parser's file set\n%s\nwant\n%s", got, wantDump)
	}
}

// TestMarkupKeepComments checks that Go comments, and only those at node
// boundaries, are kept as Go for the mask.
func TestMarkupKeepComments(t *testing.T) {
	src := "<ul>\n\t// a line comment\n\t<li>a</li> // trailing\n\t/* block */ <li>b</li>\n\t<li>c</li> /* after */\n</ul>"
	_, r, err := markupParse(t, src)
	if err != nil {
		t.Fatal(err)
	}
	want := []span{{6, 23}, {36, 47}, {49, 60}, {84, 95}}
	if !slices.Equal(r.keep, want) {
		t.Errorf("keep %v, want %v", r.keep, want)
	}
	for _, k := range r.keep {
		if c := src[k.from:k.to]; !strings.HasPrefix(c, "//") && !strings.HasPrefix(c, "/*") {
			t.Errorf("kept %q", c)
		}
	}
	if len(r.semis) != 0 || len(r.interps) != 0 || r.start != 0 || r.end != len(src) {
		t.Errorf("region %+v", r)
	}
	_, r, _ = markupParse(t, `<a href="https://x.dev">see https://x.dev // x /* y */</a>`)
	if len(r.keep) != 0 {
		t.Errorf("text kept as Go: %v", r.keep)
	}
}

// TestMarkupRegionRecovery checks where a region with errors ends: the Go
// after it still parses.
func TestMarkupRegionRecovery(t *testing.T) {
	cases := []struct {
		name, src string
		end       int
		errs      string
	}{
		{"a } line ends the region", "package p\n\ntemplate T() {\n\t<div>\n\t\t<p>hello</p>\n}\n", 48,
			`x.golpe:6:1: expected </div>, found '}'`},
		{"a close tag of an enclosing element", "package p\n\ntemplate T() {\n\t<div>\n\t\t<p>hello\n\t</div>\n}\n", 51,
			`x.golpe:6:2: expected </p>, found </div>`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, _, _, err := markupGoParse(t, token.NewFileSet(), c.src)
			if r.end != c.end {
				t.Errorf("region ends at %d, want %d", r.end, c.end)
			}
			if got := markupErrors(err); got != c.errs {
				t.Errorf("errors\n%s\nwant\n%s", got, c.errs)
			}
		})
	}
}

// TestMarkupSpecExamples parses the markup of SPEC-001's Button and Layout
// examples, masks it and parses the skeleton with go/parser.
func TestMarkupSpecExamples(t *testing.T) {
	cases := []struct {
		name, src, dump string
		exprs           []string
	}{
		{"Button", `package p

type ButtonProps struct {
    Children element
}

template Button(props ButtonProps) {
    <button>{props.Children(writer())}</button>
}
`, `Element 8:5-8:48 button close=8:39
  Interp 8:13-8:39 {...}
`, []string{"props.Children(writer())"}},
		{"Layout", `package p

template Layout(p LayoutProps) {
    if p.Header != nil {
        if err := p.Header(writer()); err != nil {   // writer() in Go mode
            return err
        }
    }
    <>{p.Children(writer())}</>                      // fragment
}
`, `Fragment 9:5-9:32 close=9:29
  Interp 9:7-9:29 {...}
`, []string{"p.Children(writer())"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fset := token.NewFileSet()
			fset.AddFile("other.go", -1, 10)
			r, f, _, err := markupGoParse(t, fset, c.src)
			if err != nil {
				t.Fatal(err)
			}
			if got := markupDump(t, fset, r.root); got != c.dump {
				t.Errorf("dump\n%s\nwant\n%s", got, c.dump)
			}
			if got := exprsIn(t, fset, f, c.src, r); !slices.Equal(got, c.exprs) {
				t.Errorf("interpolation expressions %q, want %q", got, c.exprs)
			}
		})
	}
}

// TestLineFile checks that the stand-in file gives the positions, and the
// line:col, of the file go/parser creates later.
func TestLineFile(t *testing.T) {
	src := []byte("package p\n\nfunc f() {\n\t_ = 1\n}\n")
	fset := token.NewFileSet()
	fset.AddFile("a.go", -1, 7)
	lf := lineFile("b.golpe", fset.Base(), src)
	f, err := goparser.ParseFile(fset, "b.golpe", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	real := fset.File(f.Pos())
	if real.Base() != lf.Base() || real.Size() != lf.Size() || real.Name() != lf.Name() || real.LineCount() != lf.LineCount() {
		t.Fatalf("stand-in %s base %d size %d lines %d, real %s base %d size %d lines %d",
			lf.Name(), lf.Base(), lf.Size(), lf.LineCount(), real.Name(), real.Base(), real.Size(), real.LineCount())
	}
	for off := 0; off <= len(src); off++ {
		if lf.Pos(off) != real.Pos(off) || lf.Position(lf.Pos(off)) != real.Position(real.Pos(off)) {
			t.Fatalf("offset %d: stand-in %v, real %v", off, lf.Position(lf.Pos(off)), real.Position(real.Pos(off)))
		}
	}
}

func TestErrorList(t *testing.T) {
	at := func(line, col int) token.Position {
		return token.Position{Filename: "x.golpe", Line: line, Column: col}
	}
	var l errorList
	if err := l.err(); err != nil {
		t.Fatalf("empty list: %v", err)
	}
	l.add(at(2, 1), "b")
	l.add(at(1, 5), "a2")
	l.add(at(1, 3), "a1")
	l.addGo(nil)
	l.addGo(scanner.ErrorList{{Pos: at(3, 4), Msg: "go1"}, {Pos: at(4, 1), Msg: "go2"}})
	l.addGo(errors.New("no position"))
	want := "no position\nx.golpe:1:3: a1\nx.golpe:2:1: b\nx.golpe:3:4: go1\nx.golpe:4:1: go2"
	if got := markupErrors(l.err()); got != want {
		t.Errorf("errors\n%s\nwant\n%s", got, want)
	}
	if got := markupErrors(l.err()); got != want {
		t.Errorf("err changed the list: second call\n%s", got)
	}
}
