package parser

import (
	goast "go/ast"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/sgq995/golpe/internal/ast"
)

func TestMarkupIf(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "if, else if, else", src: "<div>\nif a {\n\t<b>x</b>\n} else if b {\n\ty\n} else {\n\t{z}\n}\n</div>", dump: `Element 1:1-9:7 div close=9:1
  If 2:1-8:2 if ...
    Block 2:6-4:2
      Element 3:2-3:10 b close=3:6
        Text 3:5-3:6 "x"
    If 4:8-8:2 if ...
      Block 4:13-6:2
        Text 4:14-6:1 "y"
      Block 6:8-8:2
        Interp 7:2-7:5 {...}
`},
		{name: "without else", src: "<p>\nif ok {\n\tyes\n}\n</p>", dump: `Element 1:1-5:5 p close=5:1
  If 2:1-4:2 if ...
    Block 2:7-4:2
      Text 2:8-4:1 "yes"
`},
		{name: "empty body, else on the same line", src: "<p>\nif a {} else {\n\tb\n}\n</p>", dump: `Element 1:1-5:5 p close=5:1
  If 2:1-4:2 if ...
    Block 2:6-2:8
    Block 2:14-4:2
      Text 2:15-4:1 "b"
`},
		{name: "between text, like an element", src: "<p>\nHello\nif x {\n\t<b>world</b>\n}\n!\n</p>", dump: `Element 1:1-7:5 p close=7:1
  Text 1:4-3:1 "Hello "
  If 3:1-5:2 if ...
    Block 3:6-5:2
      Element 4:2-4:14 b close=4:10
        Text 4:5-4:10 "world"
  Text 5:2-7:1 " !"
`},
		{name: "CRLF line endings", src: "<p>\r\nif a {\r\n\tx\r\n}else{\r\n\ty\r\n}\r\n</p>", dump: `Element 1:1-7:5 p close=7:1
  If 2:1-6:2 if ...
    Block 2:6-4:2
      Text 2:7-4:1 "x"
    Block 4:6-6:2
      Text 4:7-6:1 "y"
`},
		{name: "in a fragment", src: "<>\nif ok {\n\tyes\n}\n</>", dump: `Fragment 1:1-5:4 close=5:1
  If 2:1-4:2 if ...
    Block 2:7-4:2
      Text 2:8-4:1 "yes"
`},
	})
}

func TestMarkupFor(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "three clauses", src: "<ul>\nfor i := 0; i < n; i++ {\n\t<li>{i}</li>\n}\n</ul>", dump: `Element 1:1-5:6 ul close=5:1
  For 2:1-4:2 for ...
    Block 2:24-4:2
      Element 3:2-3:14 li close=3:9
        Interp 3:6-3:9 {...}
`},
		{name: "condition only", src: "<ul>\nfor i < n {\n}\n</ul>", dump: `Element 1:1-4:6 ul close=4:1
  For 2:1-3:2 for ...
    Block 2:11-3:2
`},
		{name: "infinite", src: "<ul>\nfor {\n}\n</ul>", dump: `Element 1:1-4:6 ul close=4:1
  For 2:1-3:2 for ...
    Block 2:5-3:2
`},
		{name: "range", src: "<ul>\nfor _, x := range xs {\n\t<li>{x}</li>\n}\n</ul>", dump: `Element 1:1-5:6 ul close=5:1
  For 2:1-4:2 for ...
    Block 2:22-4:2
      Element 3:2-3:14 li close=3:9
        Interp 3:6-3:9 {...}
`},
		{name: "range over an int", src: "<ul>\nfor range 10 {\n\t<li/>\n}\n</ul>", dump: `Element 1:1-5:6 ul close=5:1
  For 2:1-4:2 for ...
    Block 2:14-4:2
      Element 3:2-3:7 li self-closing
`},
	})
}

func TestMarkupSwitch(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "expression switch", src: "<div>\nswitch x {\ncase 1:\n\t<b>one</b>\ncase 2, 3:\n\ttwo\ndefault:\n}\n</div>", dump: `Element 1:1-9:7 div close=9:1
  Switch 2:1-8:2 switch ...
    Case 3:1-4:12 case ...:
      Element 4:2-4:12 b close=4:8
        Text 4:5-4:8 "one"
    Case 5:1-7:1 case ...:
      Text 5:11-7:1 "two"
    Case 7:1-7:9 default:
`},
		{name: "type switch", src: "<div>\nswitch v := x.(type) {\ncase int, string:\n\t{v}\ncase nil:\n}\n</div>", dump: `Element 1:1-7:7 div close=7:1
  Switch 2:1-6:2 switch ...
    Case 3:1-4:5 case ...:
      Interp 4:2-4:5 {...}
    Case 5:1-5:10 case ...:
`},
		{name: "with init", src: "<div>\nswitch x := f(); x {\ncase 1:\n\tone\n}\n</div>", dump: `Element 1:1-6:7 div close=6:1
  Switch 2:1-5:2 switch ...
    Case 3:1-5:1 case ...:
      Text 3:8-5:1 "one"
`},
		{name: "without tag", src: "<div>\nswitch {\ncase a > b:\n\tbig\ndefault:\n\tsmall\n}\n</div>", dump: `Element 1:1-8:7 div close=8:1
  Switch 2:1-7:2 switch ...
    Case 3:1-5:1 case ...:
      Text 3:12-5:1 "big"
    Case 5:1-7:1 default:
      Text 5:9-7:1 "small"
`},
		{name: "a clause body on the clause's line", src: "<div>\nswitch x {\ncase 1: <b>one</b>\ndefault: many }\n</div>", dump: `Element 1:1-5:7 div close=5:1
  Switch 2:1-4:16 switch ...
    Case 3:1-3:19 case ...:
      Element 3:9-3:19 b close=3:15
        Text 3:12-3:15 "one"
    Case 4:1-4:15 default:
      Text 4:9-4:15 "many"
`},
		{name: "empty", src: "<div>\nswitch x {}\n</div>", dump: `Element 1:1-3:7 div close=3:1
  Switch 2:1-2:12 switch ...
`},
		{name: "Go comments before the first clause", src: "<div>\nswitch x {\n// none yet\ncase 1:\n}\n</div>", dump: `Element 1:1-6:7 div close=6:1
  Switch 2:1-5:2 switch ...
    Case 4:1-4:8 case ...:
`},
		{name: "case and default as text in a nested element", src: "<div>\nswitch x {\ncase 1:\n\t<p>\n\tcase closed\n\tdefault:\n\t</p>\n}\n</div>", dump: `Element 1:1-9:7 div close=9:1
  Switch 2:1-8:2 switch ...
    Case 3:1-7:6 case ...:
      Element 4:2-7:6 p close=7:2
        Text 4:5-7:2 "case closed default:"
`},
	})
}

func TestMarkupSelect(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "receive, send, default", src: "<div>\nselect {\ncase v := <-ch:\n\t{v}\ncase out <- 1:\n\tsent\ndefault:\n}\n</div>", dump: `Element 1:1-9:7 div close=9:1
  Select 2:1-8:2 select
    Case 3:1-4:5 case ...:
      Interp 4:2-4:5 {...}
    Case 5:1-7:1 case ...:
      Text 5:15-7:1 "sent"
    Case 7:1-7:9 default:
`},
	})
}

func TestMarkupNestedStatements(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "for, element, if, for", src: "<ul>\nfor _, g := range groups {\n\t<li>\n\tif g.Open {\n\t\tfor _, x := range g.Items {\n\t\t\t{x}\n\t\t}\n\t}\n\t</li>\n}\n</ul>", dump: `Element 1:1-11:6 ul close=11:1
  For 2:1-10:2 for ...
    Block 2:26-10:2
      Element 3:2-9:7 li close=9:2
        If 4:2-8:3 if ...
          Block 4:12-8:3
            For 5:3-7:4 for ...
              Block 5:29-7:4
                Interp 6:4-6:7 {...}
`},
		{name: "switch in a select clause, if in a switch clause", src: "<div>\nselect {\ncase v := <-ch:\n\tswitch v {\n\tcase 0:\n\t\tif ok {\n\t\t\tzero\n\t\t}\n\t}\n}\n</div>", dump: `Element 1:1-11:7 div close=11:1
  Select 2:1-10:2 select
    Case 3:1-9:3 case ...:
      Switch 4:2-9:3 switch ...
        Case 5:2-8:4 case ...:
          If 6:3-8:4 if ...
            Block 6:9-8:4
              Text 6:10-8:3 "zero"
`},
	})
}

func TestMarkupOneLineBodies(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "if", src: "<p>\nif ok { <b>yes</b> }\n</p>", dump: `Element 1:1-3:5 p close=3:1
  If 2:1-2:21 if ...
    Block 2:7-2:21
      Element 2:9-2:19 b close=2:15
        Text 2:12-2:15 "yes"
`},
		{name: "if and else", src: "<p>\nif ok { yes } else { no }\n</p>", dump: `Element 1:1-3:5 p close=3:1
  If 2:1-2:26 if ...
    Block 2:7-2:14
      Text 2:8-2:13 "yes"
    Block 2:20-2:26
      Text 2:21-2:25 "no"
`},
		{name: "for with an interpolation", src: "<ul>\nfor _, x := range xs { <li>{x}</li> }\n</ul>", dump: `Element 1:1-3:6 ul close=3:1
  For 2:1-2:38 for ...
    Block 2:22-2:38
      Element 2:24-2:36 li close=2:31
        Interp 2:28-2:31 {...}
`},
		{name: "content after the closing brace", src: "<p>\nif ok { yes } and more\n</p>", dump: `Element 1:1-3:5 p close=3:1
  If 2:1-2:14 if ...
    Block 2:7-2:14
      Text 2:8-2:13 "yes"
  Text 2:14-3:1 " and more"
`},
		{name: "a body may start on the header line and go on", src: "<p>\nif ok { <b>yes</b>\n\tand more }\n</p>", dump: `Element 1:1-4:5 p close=4:1
  If 2:1-3:12 if ...
    Block 2:7-3:12
      Element 2:9-2:19 b close=2:15
        Text 2:12-2:15 "yes"
      Text 2:19-3:11 " and more"
`},
		{name: "mid-line if in a body is text", src: "<p>\nif a { if b {x} }\n</p>", dump: `Element 1:1-3:5 p close=3:1
  If 2:1-2:18 if ...
    Block 2:6-2:18
      Text 2:7-2:13 "if b "
      Interp 2:13-2:16 {...}
`},
	})
}

// TestMarkupLineRule covers the line rule itself (SPEC-001 R18): only the
// first word of a line counts, whole words only, and errors hint at quoting.
func TestMarkupLineRule(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "prose starting with a keyword", src: "<p>\n\tif you can\n</p>", dump: `Element 1:1-3:5 p close=3:1
  Text 1:4-3:1 "if you can"
`, errs: `2:2: invalid if header: a line starting with 'if' is an embedded statement; write {"if"} to start text with it`},
		{name: "prose with an interpolation", src: "<p>\nif you can {name}\n</p>", dump: `Element 1:1-3:5 p close=3:1
  Text 1:4-2:12 "if you can "
  Interp 2:12-2:18 {...}
`, errs: `2:1: invalid if header: a line starting with 'if' is an embedded statement; write {"if"} to start text with it`},
		{name: "every keyword", src: "<p>\nif(x)\n  for\n\tswitch x\nselect\n</p>", dump: "-", errs: `2:1: invalid if header: a line starting with 'if' is an embedded statement; write {"if"} to start text with it
3:3: invalid for header: a line starting with 'for' is an embedded statement; write {"for"} to start text with it
4:2: invalid switch header: a line starting with 'switch' is an embedded statement; write {"switch"} to start text with it
5:1: invalid select header: a line starting with 'select' is an embedded statement; write {"select"} to start text with it`},
		{name: "whole words only", src: "<p>\nformat\niffy\nif_x\nIf\nifé\nfor2\nselected\nswitches\n</p>", dump: `Element 1:1-10:5 p close=10:1
  Text 1:4-10:1 "format iffy if_x If ifé for2 selected switches"
`},
		{name: "not at the start of a line", src: `<p>a if b {c} else {d}</p>`, dump: `Element 1:1-1:27 p close=1:23
  Text 1:4-1:11 "a if b "
  Interp 1:11-1:14 {...}
  Text 1:14-1:20 " else "
  Interp 1:20-1:23 {...}
`},
		{name: "quoted", src: "<p>\n{\"if \"}you can\n</p>", dump: `Element 1:1-3:5 p close=3:1
  Interp 2:1-2:8 {...}
  Text 2:8-3:1 "you can"
`},
		{name: "after a Go comment at the start of a line", src: "<p>\n/* c */ for\n</p>", dump: "-",
			errs: `2:9: invalid for header: a line starting with 'for' is an embedded statement; write {"for"} to start text with it`},
		{name: "in pre", src: "<pre>\nfor x\n</pre>", dump: "-",
			errs: `2:1: invalid for header: a line starting with 'for' is an embedded statement; write {"for"} to start text with it`},
		{name: "text-only elements take keywords as text", src: "<title>\nelse if x\n</title>", dump: `Element 1:1-3:9 title close=3:1
  Text 1:8-3:1 "else if x"
`},
		{name: "case and default outside a switch body are text", src: "<p>\ncase 1:\ndefault:\n</p>", dump: `Element 1:1-4:5 p close=4:1
  Text 1:4-4:1 "case 1: default:"
`},
		{name: "else on its own line", src: "<div>\nif a {\n\tx\n}\nelse {\n\ty\n}\n</div>", dump: `Element 1:1-8:7 div close=8:1
  If 2:1-7:2 if ...
    Block 2:6-4:2
      Text 2:7-4:1 "x"
    Block 5:6-7:2
      Text 5:7-7:1 "y"
`, errs: `5:1: else must follow '}' on the same line; write {"else"} to start text with it`},
		{name: "else if on its own line", src: "<div>\nif a {\n}\nelse if b {\n\ty\n}\n</div>", dump: `Element 1:1-7:7 div close=7:1
  If 2:1-6:2 if ...
    Block 2:6-3:2
    If 4:6-6:2 if ...
      Block 4:11-6:2
        Text 4:12-6:1 "y"
`, errs: `4:1: else must follow '}' on the same line; write {"else"} to start text with it`},
		{name: "else on its own line after an else if", src: "<div>\nif a {\n} else if b {\n}\nelse {\n\tc\n}\n</div>", dump: `Element 1:1-8:7 div close=8:1
  If 2:1-7:2 if ...
    Block 2:6-3:2
    If 3:8-7:2 if ...
      Block 3:13-4:2
      Block 5:6-7:2
        Text 5:7-7:1 "c"
`, errs: `5:1: else must follow '}' on the same line; write {"else"} to start text with it`},
		{name: "else on its own line after an else", src: "<div>\nif a {\n} else {\n}\nelse {\n\tc\n}\n</div>", dump: `Element 1:1-8:7 div close=8:1
  If 2:1-4:2 if ...
    Block 2:6-3:2
    Block 3:8-4:2
`, errs: `5:1: else must follow '}' on the same line; write {"else"} to start text with it`},
		{name: "else on the first line", src: "<div>\nelse {\n\tc\n}\n</div>", dump: "Element 1:1-5:7 div close=5:1\n",
			errs: `2:1: else must follow '}' on the same line; write {"else"} to start text with it`},
		{name: "else on its own line after text", src: "<p>\nhi\nelse {\n\ty\n}\n</p>", dump: `Element 1:1-6:5 p close=6:1
  Text 1:4-3:1 "hi"
`, errs: `3:1: else must follow '}' on the same line; write {"else"} to start text with it`},
		{name: "prose starting with else", src: "<p>\nelse, we can\n</p>", dump: `Element 1:1-3:5 p close=3:1
  Text 1:4-3:1 "else, we can"
`, errs: `2:1: else must follow '}' on the same line; write {"else"} to start text with it`},
		{name: "a } line ends unclosed markup", src: "<p>hello\n}", dump: `Element 1:1-2:1 p
  Text 1:4-2:1 "hello"
`, errs: `2:1: expected </p>, found '}'`},
		{name: "a } line in a text-only element", src: "<title>x\n  }", dump: "-", errs: `2:3: expected </title>, found '}'`},
	})
}

// TestMarkupStmtErrors covers the errors of embedded statements and the
// recovery after them.
func TestMarkupStmtErrors(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "body not closed", src: "<div>\nif a {\n\tx", dump: `Element 1:1-3:3 div
  If 2:1-3:3 if ...
    Block 2:6-3:3
      Text 2:7-3:3 "x"
`, errs: `3:3: expected '}', found 'EOF'`},
		{name: "close tag of an enclosing element", src: "<ul>\nfor x := range xs {\n\t<li>{x}</li>\n</ul>", dump: `Element 1:1-4:6 ul close=4:1
  For 2:1-3:14 for ...
    Block 2:19-3:14
      Element 3:2-3:14 li close=3:9
        Interp 3:6-3:9 {...}
`, errs: `4:1: expected '}', found </ul>`},
		{name: "stray close tag in a body", src: "<div>\nif a {\n\tx</b>\n}\n</div>", dump: `Element 1:1-5:7 div close=5:1
  If 2:1-4:2 if ...
    Block 2:6-4:2
      Text 2:7-3:3 "x"
`, errs: `3:3: expected '}', found </b>`},
		{name: "void close tag in a body", src: "<div>\nif a {\n\t</br>\n}\n</div>", dump: "-", errs: `3:2: void element <br> cannot have content or a close tag`},
		{name: "a } line closes the body around an unclosed element", src: "<ul>\nfor x := range xs {\n\t<li>{x}\n}\n</ul>", dump: `Element 1:1-5:6 ul close=5:1
  For 2:1-4:2 for ...
    Block 2:19-4:2
      Element 3:2-3:9 li
        Interp 3:6-3:9 {...}
`, errs: `4:1: expected </li>, found '}'`},
		{name: "a stray } in an element in a one-line body", src: "<p>\nif ok { <b>yes }\n</b>\n}\n</p>", dump: "-", errs: `2:16: unexpected '}' in text; write {"}"} for a literal '}'`},
		{name: "content before the first clause", src: "<div>\nswitch x {\n\tjunk\ncase 1:\n}\n</div>", dump: `Element 1:1-6:7 div close=6:1
  Switch 2:1-5:2 switch ...
    Case 4:1-4:8 case ...:
`, errs: `3:2: expected case or default`},
		{name: "a one-line switch has content before its first clause", src: "<div>\nswitch x { <b/> }\n</div>", dump: `Element 1:1-3:7 div close=3:1
  Switch 2:1-2:18 switch ...
`, errs: `2:12: expected case or default`},
		{name: "an error before the first clause is enough", src: "<div>\nselect {\n\t<p>x</b></p>\ncase <-ch:\n}\n</div>", dump: "-", errs: `3:6: expected </p>, found </b>`},
		{name: "invalid case clause", src: "<div>\nswitch x {\ncase 1:\n\tcase in point: text\n}\n</div>", dump: `Element 1:1-6:7 div close=6:1
  Switch 2:1-5:2 switch ...
    Case 3:1-5:1 case ...:
      Text 3:8-5:1 "case in point: text"
`, errs: `4:2: invalid case clause: a line starting with 'case' opens a clause; write {"case"} to start text with it`},
		{name: "invalid default clause", src: "<div>\nswitch x {\ndefault:\n\tdefault value is 0\n}\n</div>", dump: "-",
			errs: `4:2: invalid default clause: a line starting with 'default' opens a clause; write {"default"} to start text with it`},
		{name: "invalid first clause", src: "<div>\nswitch x {\ncase:\n}\n</div>", dump: `Element 1:1-5:7 div close=5:1
  Switch 2:1-4:2 switch ...
`, errs: `3:1: invalid case clause: a line starting with 'case' opens a clause; write {"case"} to start text with it`},
		{name: "invalid select clause", src: "<div>\nselect {\ncase a, b:\n}\n</div>", dump: "-",
			errs: `3:1: invalid case clause: a line starting with 'case' opens a clause; write {"case"} to start text with it`},
		{name: "clause not terminated", src: "<div>\nswitch x {\ncase (1", dump: "-",
			errs: `3:1: invalid case clause: a line starting with 'case' opens a clause; write {"case"} to start text with it`},
		{name: "clause with an unmatched bracket", src: "<div>\nswitch x {\ncase 1):\n}\n</div>", dump: "-",
			errs: `3:1: invalid case clause: a line starting with 'case' opens a clause; write {"case"} to start text with it`},
		{name: "else followed by neither if nor {", src: "<p>\nif a {\n} else b\n</p>", dump: `Element 1:1-4:5 p close=4:1
  If 2:1-3:2 if ...
    Block 2:6-3:2
  Text 3:2-4:1 " else b"
`, errs: `3:8: expected if statement or block after else`},
		{name: "else after a for", src: "<p>\nfor {\n} else {\n\tx\n}\n</p>", dump: `Element 1:1-6:5 p close=6:1
  For 2:1-3:2 for ...
    Block 2:5-3:2
`, errs: `3:3: else without if`},
		{name: "else if after a switch", src: "<p>\nswitch x {\n} else if y {\n}\n</p>", dump: `Element 1:1-5:5 p close=5:1
  Switch 2:1-3:2 switch ...
`, errs: `3:3: else without if`},
		{name: "a second else", src: "<p>\nif a {\n} else {\n} else {\n}\n</p>", dump: `Element 1:1-6:5 p close=6:1
  If 2:1-4:2 if ...
    Block 2:6-3:2
    Block 3:8-4:2
`, errs: `4:3: else without if`},
		{name: "invalid else if header", src: "<p>\nif a {\n} else if b c\n</p>", dump: `Element 1:1-4:5 p close=4:1
  If 2:1-3:2 if ...
    Block 2:6-3:2
  Text 3:2-4:1 " else if b c"
`, errs: `3:8: invalid if header`},
		{name: "invalid else if header, parsed for recovery", src: "<p>\nif a {\n} else if b c {\n\td\n}\n</p>", dump: `Element 1:1-6:5 p close=6:1
  If 2:1-5:2 if ...
    Block 2:6-3:2
    If 3:8-5:2 if ...
      Block 3:15-5:2
        Text 3:16-5:1 "d"
`, errs: `3:8: invalid if header`},
		{name: "invalid header, parsed for recovery", src: "<ul>\nfor x in xs {\n\t<li>{x}</li>\n}\n</ul>", dump: `Element 1:1-5:6 ul close=5:1
  For 2:1-4:2 for ...
    Block 2:13-4:2
      Element 3:2-3:14 li close=3:9
        Interp 3:6-3:9 {...}
`, errs: `2:1: invalid for header: a line starting with 'for' is an embedded statement; write {"for"} to start text with it`},
		{name: "invalid switch header, parsed for recovery", src: "<div>\nswitch x y {\ncase 1:\n\tone\n}\n</div>", dump: `Element 1:1-6:7 div close=6:1
  Switch 2:1-5:2 switch ...
    Case 3:1-5:1 case ...:
      Text 3:8-5:1 "one"
`, errs: `2:1: invalid switch header: a line starting with 'switch' is an embedded statement; write {"switch"} to start text with it`},
		{name: "invalid select header, parsed for recovery", src: "<div>\nselect x { // c\n}\n</div>", dump: `Element 1:1-4:7 div close=4:1
  Select 2:1-3:2 select
`, errs: `2:1: invalid select header: a line starting with 'select' is an embedded statement; write {"select"} to start text with it`},
		{name: "header search stops at a close tag", src: "<div><p>\nif (you\n</p><b>{x}</b></div>", dump: `Element 1:1-3:21 div close=3:15
  Element 1:6-3:5 p close=3:1
    Text 1:9-3:1 "if (you"
  Element 3:5-3:15 b close=3:11
    Interp 3:8-3:11 {...}
`, errs: `2:1: invalid if header: a line starting with 'if' is an embedded statement; write {"if"} to start text with it`},
		{name: "a header may compare with a comment after <", src: "<p>\nif a </* c */ b {\n}\n</p>", dump: `Element 1:1-4:5 p close=4:1
  If 2:1-3:2 if ...
    Block 2:17-3:2
`},
		{name: "clause search stops at a close tag", src: "<div>\nswitch {\ncase (a\n</div>", dump: "-",
			errs: `3:1: invalid case clause: a line starting with 'case' opens a clause; write {"case"} to start text with it
4:1: expected '}', found </div>`},
		{name: "header cut by the end of the source", src: "<p>\nif (x", dump: "-",
			errs: `2:1: invalid if header: a line starting with 'if' is an embedded statement; write {"if"} to start text with it`},
		{name: "header search stops after too many semicolons", src: "<p>\nif a; b; c\n</p>", dump: "-",
			errs: `2:1: invalid if header: a line starting with 'if' is an embedded statement; write {"if"} to start text with it`},
		{name: "invalid header ending its line with {, parsed for recovery", src: "<p>\nif f(x { // c\n\tyes\n}\nif a; b; c {\n}\nif x) {\n}\n</p>", dump: `Element 1:1-9:5 p close=9:1
  If 2:1-4:2 if ...
    Block 2:8-4:2
      Text 2:14-4:1 "yes"
  Text 4:2-5:1 " "
  If 5:1-6:2 if ...
    Block 5:12-6:2
  Text 6:2-7:1 " "
  If 7:1-8:2 if ...
    Block 7:7-8:2
`, errs: `2:1: invalid if header: a line starting with 'if' is an embedded statement; write {"if"} to start text with it
5:1: invalid if header: a line starting with 'if' is an embedded statement; write {"if"} to start text with it
7:1: invalid if header: a line starting with 'if' is an embedded statement; write {"if"} to start text with it`},
	})
}

// TestMarkupPreStatements covers statements in preformatted content: a line
// of an embedded statement (header, clause, closing brace) adds no
// whitespace of its own.
func TestMarkupPreStatements(t *testing.T) {
	runMarkupCases(t, []markupCase{
		{name: "for", src: "<pre>\nfor _, l := range ls {\n{l}\n}\n</pre>", dump: `Element 1:1-5:7 pre close=5:1
  Text 1:6-2:1 "\n"
  For 2:1-4:2 for ...
    Block 2:22-4:2
      Interp 3:1-3:4 {...}
      Text 3:4-4:1 "\n"
`},
		{name: "indented if and else", src: "<pre>\n  if a {\n    x\n  } else {\n    y\n  }\n</pre>", dump: `Element 1:1-7:7 pre close=7:1
  Text 1:6-2:1 "\n"
  If 2:3-6:4 if ...
    Block 2:8-4:4
      Text 3:1-4:1 "    x\n"
    Block 4:10-6:4
      Text 5:1-6:1 "    y\n"
`},
		{name: "switch", src: "<pre>\nswitch n {\ncase 1:\none\ndefault:\nmany\n}\n</pre>", dump: `Element 1:1-8:7 pre close=8:1
  Text 1:6-2:1 "\n"
  Switch 2:1-7:2 switch ...
    Case 3:1-5:1 case ...:
      Text 4:1-5:1 "one\n"
    Case 5:1-7:1 default:
      Text 6:1-7:1 "many\n"
`},
		{name: "blank lines before the first clause", src: "<pre>\nswitch n {\n\n  case 1:\n}\n</pre>", dump: `Element 1:1-6:7 pre close=6:1
  Text 1:6-2:1 "\n"
  Switch 2:1-5:2 switch ...
    Case 4:3-4:10 case ...:
`},
		{name: "a one-line body keeps its spaces", src: "<pre>\nif ok { yes }\n</pre>", dump: `Element 1:1-3:7 pre close=3:1
  Text 1:6-2:1 "\n"
  If 2:1-2:14 if ...
    Block 2:7-2:14
      Text 2:8-2:13 " yes "
`},
		{name: "content after the closing brace", src: "<pre>\nif ok {yes} and\n</pre>", dump: `Element 1:1-3:7 pre close=3:1
  Text 1:6-2:1 "\n"
  If 2:1-2:12 if ...
    Block 2:7-2:12
      Text 2:8-2:11 "yes"
  Text 2:12-3:1 " and\n"
`},
		{name: "Go comments on statement lines", src: "<pre>\nfor x := range xs { // each\n{x}\n} // end\n</pre>", dump: `Element 1:1-5:7 pre close=5:1
  Text 1:6-2:1 "\n"
  For 2:1-4:2 for ...
    Block 2:19-4:2
      Interp 3:1-3:4 {...}
      Text 3:4-4:1 "\n"
`},
		{name: "a Go comment before the keyword", src: "<pre>\na\n/* c */ if ok {\nb\n}\n</pre>", dump: `Element 1:1-6:7 pre close=6:1
  Text 1:6-3:1 "\na\n"
  If 3:9-5:2 if ...
    Block 3:15-5:2
      Text 4:1-5:1 "b\n"
`},
		{name: "a statement inherits pre from its element", src: "<pre><code>\nif ok {\nx\n}\n</code></pre>", dump: `Element 1:1-5:14 pre close=5:8
  Element 1:6-5:8 code close=5:1
    Text 1:12-2:1 "\n"
    If 2:1-4:2 if ...
      Block 2:7-4:2
        Text 3:1-4:1 "x\n"
`},
		{name: "else on its own line", src: "<pre>\nif a {\n}\nelse {\n}\n</pre>", dump: `Element 1:1-6:7 pre close=6:1
  Text 1:6-2:1 "\n"
  If 2:1-5:2 if ...
    Block 2:6-3:2
    Block 4:6-5:2
`, errs: `4:1: else must follow '}' on the same line; write {"else"} to start text with it`},
	})
}

// stmtNode returns the first embedded statement in the tree of n.
func stmtNode(n goast.Node) goast.Node {
	var found goast.Node
	ast.Inspect(n, func(c goast.Node) bool {
		switch c.(type) {
		case *ast.If, *ast.For, *ast.Switch, *ast.Select:
			if found == nil {
				found = c
			}
		}
		return found == nil
	})
	return found
}

// TestMarkupHeaders checks where headers end: at the '{' Go's grammar
// gives the body, after composite literals, function literals, strings and
// comments that hold braces.
func TestMarkupHeaders(t *testing.T) {
	cases := []struct{ line, header string }{
		{"if f := func() bool { return true }; f() {", "if f := func() bool { return true }; f()"},
		{"for _, x := range []int{1, 2} {", "for _, x := range []int{1, 2}"},
		{"if x == (T{}) {", "if x == (T{})"},
		{"if m := map[string]int{\"a\": 1}; m[\"a\"] > 0 {", "if m := map[string]int{\"a\": 1}; m[\"a\"] > 0"},
		{"if p := (struct{ X int }{1}); p.X > 0 {", "if p := (struct{ X int }{1}); p.X > 0"},
		{"if s := \"{\"; s != \"}\" {", "if s := \"{\"; s != \"}\""},
		{"if x /* { */ {", "if x /* { */"},
		{"if ok { // a { in a comment", "if ok"},
		{"if v, ok := m[k]; ok {", "if v, ok := m[k]; ok"},
		{"if a <b {", "if a <b"},
		{"if c == '{' {", "if c == '{'"},
		{"if(x) {", "if(x)"},
		{"for{", "for"},
		{"if a &&\n\tb {", "if a &&\n\tb"},
		{"for i := 0; i < len(s); i++ {", "for i := 0; i < len(s); i++"},
		{"for ;; {", "for ;;"},
		{"for i := range 10 {", "for i := range 10"},
		{"for {", "for"},
		{"switch f := func() int { return 1 }; f() {", "switch f := func() int { return 1 }; f()"},
		{"switch x := v.(type) {", "switch x := v.(type)"},
		{"switch {", "switch"},
		{"select {", "select"},
	}
	for _, c := range cases {
		src := "<div>\n" + c.line + "\n}\n</div>"
		fset, r, err := markupParse(t, src)
		if err != nil {
			t.Errorf("%q: %v", c.line, err)
			continue
		}
		var lbrace token.Pos
		switch n := stmtNode(r.root).(type) {
		case *ast.If:
			lbrace = n.Body.Lbrace
		case *ast.For:
			lbrace = n.Body.Lbrace
		case *ast.Switch:
			lbrace = n.Lbrace
		case *ast.Select:
			lbrace = n.Lbrace
		default:
			t.Errorf("%q: no statement", c.line)
			continue
		}
		kw := len("<div>\n")
		if got := strings.TrimRight(src[kw:fset.Position(lbrace).Offset], " "); got != c.header {
			t.Errorf("%q: header %q, want %q", c.line, got, c.header)
		}
	}
}

// TestMarkupClauses checks where case and default clause headers end: at
// the first ':' outside parentheses, brackets and braces.
func TestMarkupClauses(t *testing.T) {
	cases := []struct{ stmt, line, header string }{
		{"switch x", "case a[1:2] == b:", "case a[1:2] == b:"},
		{"switch x", "case \"x:y\", ':':", "case \"x:y\", ':':"},
		{"switch x", "case T{A: 1}.B:", "case T{A: 1}.B:"},
		{"switch x", "case func() int { a: return 1 }():", "case func() int { a: return 1 }():"},
		{"switch x", "case x /* : */ :", "case x /* : */ :"},
		{"switch x", "case a,\n\tb: // c", "case a,\n\tb:"},
		{"switch x", "default :", "default :"},
		{"switch v := x.(type)", "case int, []int, map[string]int, func(int) error, *T:", "case int, []int, map[string]int, func(int) error, *T:"},
		{"select", "case v, ok := <-ch:", "case v, ok := <-ch:"},
		{"select", "case ch <- f(func() { a: for {} }):", "case ch <- f(func() { a: for {} }):"},
		{"select", "case <-ch:", "case <-ch:"},
		{"select", "default:", "default:"},
	}
	for _, c := range cases {
		src := "<div>\n" + c.stmt + " {\n" + c.line + "\n}\n</div>"
		fset, r, err := markupParse(t, src)
		if err != nil {
			t.Errorf("%q: %v", c.line, err)
			continue
		}
		var cases []*ast.Case
		switch n := stmtNode(r.root).(type) {
		case *ast.Switch:
			cases = n.Cases
		case *ast.Select:
			cases = n.Cases
		}
		if len(cases) != 1 {
			t.Errorf("%q: %d clauses, want 1", c.line, len(cases))
			continue
		}
		cl := cases[0]
		if got := src[fset.Position(cl.Keyword).Offset : fset.Position(cl.Colon).Offset+1]; got != c.header {
			t.Errorf("%q: clause header %q, want %q", c.line, got, c.header)
		}
		if cl.Default != strings.HasPrefix(c.line, "default") {
			t.Errorf("%q: Default = %v", c.line, cl.Default)
		}
	}
}

// TestValidHeaderAndClause covers the go/parser checks of header extents
// directly, including inputs the markup parser does not produce.
func TestValidHeaderAndClause(t *testing.T) {
	headers := []struct {
		header string
		valid  bool
	}{
		{"if x ", true},
		{"for _, x := range xs ", true},
		{"switch v := x.(type) ", true},
		{"select ", true},
		{"if x = 1 ", false},
		{"if a {} else ", false}, // parses, with the body at the first '{'
		{"L: for ", false},       // parses, as a labeled statement
	}
	for _, h := range headers {
		if got := validHeader([]byte(h.header)); got != h.valid {
			t.Errorf("validHeader(%q) = %v, want %v", h.header, got, h.valid)
		}
	}
	clauses := []struct {
		sel    bool
		clause string
		valid  bool
	}{
		{false, "case 1, 2:", true},
		{false, "default:", true},
		{true, "case v := <-ch:", true},
		{true, "case a, b:", false},
		{false, "case:", false},
	}
	for _, c := range clauses {
		if got := validClause(c.sel, []byte(c.clause)); got != c.valid {
			t.Errorf("validClause(%v, %q) = %v, want %v", c.sel, c.clause, got, c.valid)
		}
	}
	for _, src := range []string{
		"package p;func _(){a();b()}",
		"package p;func _(){}\nfunc g(){}",
		"package p;var x int",
		"package p;func _()",
	} {
		if _, ok := parseOneStmt(src); ok {
			t.Errorf("parseOneStmt(%q) reports one statement", src)
		}
	}
}

// markupGraft gives the embedded statements and interpolations of r the
// go/ast nodes go/parser built from the mask, as the graft (TASK-001-06)
// will: the statement at each keyword and the expression between the braces
// of each interpolation.
func markupGraft(t *testing.T, fset *token.FileSet, f *goast.File, r *region) {
	t.Helper()
	at := map[token.Pos]goast.Stmt{}
	goast.Inspect(f, func(n goast.Node) bool {
		if s, ok := n.(goast.Stmt); ok {
			if _, dup := at[s.Pos()]; !dup {
				at[s.Pos()] = s
			}
		}
		return true
	})
	for _, n := range r.stmts {
		s := at[n.Pos()]
		ok := s != nil
		switch n := n.(type) {
		case *ast.If:
			n.Stmt, ok = s.(*goast.IfStmt)
		case *ast.For:
			n.Stmt = s
		case *ast.Switch:
			n.Stmt = s
		case *ast.Select:
			n.Stmt, ok = s.(*goast.SelectStmt)
		case *ast.Case:
			n.Clause = s
		default:
			ok = false
		}
		if !ok {
			t.Fatalf("graft: no statement for %T at %v", n, fset.Position(n.Pos()))
		}
	}
	for _, in := range r.interps {
		for p := in.Lbrace + 1; p < in.Rbrace && in.X == nil; p++ {
			if es, ok := at[p].(*goast.ExprStmt); ok {
				in.X = es.X
			}
		}
		if in.X == nil {
			t.Fatalf("graft: no expression for the interpolation at %v", fset.Position(in.Lbrace))
		}
	}
}

// checkRegion checks the invariants the mask relies on: keep spans in
// source order, disjoint and inside the region; semis inside the region,
// increasing and outside the keep spans.
func checkRegion(t *testing.T, r *region) {
	t.Helper()
	last := r.start
	for _, k := range r.keep {
		if k.from < last || k.to <= k.from || k.to > r.end {
			t.Fatalf("keep %v out of order or out of [%d, %d): %v", k, r.start, r.end, r.keep)
		}
		last = k.to
	}
	for i, s := range r.semis {
		if s <= r.start || s >= r.end || i > 0 && s <= r.semis[i-1] {
			t.Fatalf("semis %v out of order or out of (%d, %d)", r.semis, r.start, r.end)
		}
		for _, k := range r.keep {
			if k.from <= s && s < k.to {
				t.Fatalf("semi %d inside keep %v", s, k)
			}
		}
	}
}

// trimLines returns the mask m with the blanks at the end of each line
// removed, so that expected masks can be written without them.
func trimLines(m []byte) string {
	lines := strings.Split(string(m), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}

// TestMarkupStmtRegion checks what embedded statements add to the region:
// their Go parts in the keep spans, a ';' after a closing '}' that the line
// goes on after, and the statements for the graft.
func TestMarkupStmtRegion(t *testing.T) {
	src := "package p\n\ntemplate T(ok bool, xs []string) {\n\t<p>\n\tif ok { yes } else { {xs[0]} }\n\tfor _, x := range xs {\n\t\t{x}\n\t} after {len(xs)}\n\t</p>\n}\n"
	fset := token.NewFileSet()
	r, f, m, err := markupGoParse(t, fset, src)
	if err != nil {
		t.Fatal(err)
	}
	checkRegion(t, r)
	wantMask := "package p\n\nfunc     T(ok bool, xs []string) {\n\t;\n if ok {     } else {  xs[0]; }\n for _, x := range xs {\n   x;\n };       len(xs);\n\n}\n"
	if got := trimLines(m); got != wantMask {
		t.Errorf("mask\n%q\nwant\n%q", got, wantMask)
	}
	markupGraft(t, fset, f, r)
	wantDump := `Element 4:2-9:6 p close=9:2
  If 5:2-5:32 if ok
    Block 5:8-5:15
      Text 5:9-5:14 "yes"
    Block 5:21-5:32
      Interp 5:23-5:30 {xs[0]}
  Text 5:32-6:2 " "
  For 6:2-8:3 for _, x := range xs
    Block 6:23-8:3
      Interp 7:3-7:6 {x}
  Text 8:3-8:10 " after "
  Interp 8:10-8:19 {len(xs)}
`
	if got := markupDump(t, fset, r.root); got != wantDump {
		t.Errorf("dump\n%s\nwant\n%s", got, wantDump)
	}
	if len(r.stmts) != 2 {
		t.Errorf("stmts %d, want 2", len(r.stmts))
	}
}

// TestMarkupStmtSkeletons masks templates with embedded statements, also
// invalid ones, and checks that go/parser parses every skeleton and that
// every statement and interpolation finds its go/ast node: an invalid
// header leaves no Go behind to cascade into more errors.
func TestMarkupStmtSkeletons(t *testing.T) {
	cases := []struct {
		name, body string
		stmts      int
		errs       string
	}{
		{"one-line bodies and content after them", "<p>\n\tif a { {1} } else if b { {2} } else { {3} } {4}\n\tfor { {5} }</p>", 3, ""},
		{"a closing brace before a closing brace", "<p>\n\tif a {\n\tif b { {1} }}\n\t</p>", 2, ""},
		{"nested switch and select", "<div>\n\tswitch x := f(); x {\n\tcase 1, 2:\n\t\tselect {\n\t\tcase v := <-ch:\n\t\t\t{v}\n\t\tcase ch <- 1:\n\t\tdefault:\n\t\t}\n\tdefault:\n\t\t{x}\n\t}\n\t</div>", 7, ""},
		{"comments", "<p>\n\tif a { // c\n\t\t{1} /* d */\n\t} // e\n\t</p>", 1, ""},
		{"invalid header with its body", "<ul>\n\tfor x in xs {\n\t\t<li>{x}</li>\n\t\tif x {\n\t\t\t{1}\n\t\t}\n\t}\n\t</ul>", 1,
			`x.golpe:5:2: invalid for header: a line starting with 'for' is an embedded statement; write {"for"} to start text with it`},
		{"invalid else if branch", "<p>\n\tif a {\n\t} else if b c {\n\t\t{1}\n\t} {2}\n\t</p>", 1, `x.golpe:6:9: invalid if header`},
		{"invalid else if header and text after it", "<p>\n\tif a {\n\t} else if b c { {1} } {2}\n\t</p>", 1, `x.golpe:6:9: invalid if header`},
		{"invalid switch with clauses", "<div>\n\tswitch a b {\n\tcase 1:\n\t\t{1}\n\t}\n\t</div>", 0,
			`x.golpe:5:2: invalid switch header: a line starting with 'switch' is an embedded statement; write {"switch"} to start text with it`},
		{"content before the first clause", "<div>\n\tswitch x {\n\t\t{1}\n\t\tif a {\n\t\t}\n\tcase 1:\n\t}\n\t</div>", 2, `x.golpe:6:3: expected case or default`},
		{"else without if", "<p>\n\tfor { {1} } else { {2} } {3}\n\t</p>", 1, `x.golpe:5:14: else without if`},
		{"else on its own line", "<div>\n\tif a {\n\t}\n\telse { {1} }\n\t</div>", 1,
			`x.golpe:7:2: else must follow '}' on the same line; write {"else"} to start text with it`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := "package p\n\ntemplate T() {\n\t" + c.body + "\n}\n"
			fset := token.NewFileSet()
			r, f, _, err := markupGoParse(t, fset, src)
			if got := markupErrors(err); got != c.errs {
				t.Errorf("errors\n%s\nwant\n%s", got, c.errs)
			}
			checkRegion(t, r)
			if len(r.stmts) != c.stmts {
				t.Errorf("stmts %d, want %d", len(r.stmts), c.stmts)
			}
			markupGraft(t, fset, f, r)
		})
	}
}

// TestMarkupStmtSpecExamples parses the markup of SPEC-001's NavItem and
// List examples, masks it, parses the skeleton with go/parser and grafts
// the headers and expressions.
func TestMarkupStmtSpecExamples(t *testing.T) {
	cases := []struct{ name, src, mask, dump string }{
		{"NavItem", `package p

template NavItem(props NavItemProps) {
    <li>
    if props.Selected {
        <span>{props.Children(writer())}</span>
    } else {
        <a href={props.Route}>{props.Children(writer())}</a> // a Go comment
    }
    </li>
}
`, `package p

func     NavItem(props NavItemProps) {
    ;
    if props.Selected {
               props.Children(writer());
    } else {
                 props.Route;  props.Children(writer());     // a Go comment
    }

}
`, `Element 4:5-10:10 li close=10:5
  If 5:5-9:6 if props.Selected
    Block 5:23-7:6
      Element 6:9-6:48 span close=6:41
        Interp 6:15-6:41 {props.Children(writer())}
    Block 7:12-9:6
      Element 8:9-8:61 a close=8:57
        Attr 8:12-8:30 href
          Interp 8:17-8:30 {props.Route}
        Interp 8:31-8:57 {props.Children(writer())}
`},
		{"List", `package p

template List[T fmt.Stringer](title string, items []T, hidden bool) {
    heading := strings.ToUpper(title)    // Go mode
    <section hidden={hidden}>
        <h2>{heading}</h2>
        switch len(items) {
        case 0:
            <p>Nothing here yet.</p>
        default:
            <ul>
            for _, it := range items {
                <li>{it}</li>
            }
            </ul>
        }
    </section>
}
`, `package p

func     List[T fmt.Stringer](title string, items []T, hidden bool) {
    heading := strings.ToUpper(title)    // Go mode
    ;                hidden;
             heading;
        switch len(items) {
        case 0:

        default:

            for _, it := range items {
                     it;
            }

        }

}
`, `Element 5:5-17:15 section close=17:5
  Attr 5:14-5:29 hidden
    Interp 5:21-5:29 {hidden}
  Element 6:9-6:27 h2 close=6:22
    Interp 6:13-6:22 {heading}
  Text 6:27-7:9 " "
  Switch 7:9-16:10 switch len(items)
    Case 8:9-9:37 case 0:
      Element 9:13-9:37 p close=9:33
        Text 9:16-9:33 "Nothing here yet."
    Case 10:9-15:18 default:
      Element 11:13-15:18 ul close=15:13
        For 12:13-14:14 for _, it := range items
          Block 12:38-14:14
            Element 13:17-13:30 li close=13:25
              Interp 13:21-13:25 {it}
`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fset := token.NewFileSet()
			fset.AddFile("other.go", -1, 10)
			r, f, m, err := markupGoParse(t, fset, c.src)
			if err != nil {
				t.Fatal(err)
			}
			checkRegion(t, r)
			if got := trimLines(m); got != c.mask {
				t.Errorf("mask\n%s\nwant\n%s", got, c.mask)
			}
			markupGraft(t, fset, f, r)
			if got := markupDump(t, fset, r.root); got != c.dump {
				t.Errorf("dump\n%s\nwant\n%s", got, c.dump)
			}
		})
	}
}

// TestMarkupStmtKeep checks that a statement with an invalid header keeps
// none of its Go parts, while the statements and interpolations in its body
// keep theirs.
func TestMarkupStmtKeep(t *testing.T) {
	src := "<ul>\nfor x in xs {\n\tif x { {x} }\n}\n</ul>"
	_, r, err := markupParse(t, src)
	if err == nil {
		t.Fatal("no error")
	}
	var kept []string
	for _, k := range r.keep {
		kept = append(kept, src[k.from:k.to])
	}
	if want := []string{"if x {", "x", "}"}; !slices.Equal(kept, want) {
		t.Errorf("kept %q, want %q", kept, want)
	}
	if len(r.stmts) != 1 {
		t.Errorf("stmts %d, want 1", len(r.stmts))
	}
}
