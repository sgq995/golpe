package token

import (
	gotoken "go/token"
	"math"
	"strconv"
	"testing"
)

// unnamed is how go/token prints a value it has no name for.
func unnamed(i int) string { return "token(" + strconv.Itoa(i) + ")" }

// TestGoValues checks the whole range below golpeBase, plus negative values
// (also below the base): conversion is a no-op both ways and String is
// go/token's, including for the unnamed markers go/token prints as token(N).
func TestGoValues(t *testing.T) {
	values := []int{math.MinInt, -2, -1}
	for i := 0; i < int(golpeBase); i++ {
		values = append(values, i)
	}
	for _, i := range values {
		g := gotoken.Token(i)
		tok := FromGo(g)
		if int(tok) != i {
			t.Fatalf("FromGo(%d) = %d, want %d", i, int(tok), i)
		}
		back, ok := tok.Go()
		if !ok || back != g {
			t.Fatalf("Token(%d).Go() = %d, %v; want %d, true", i, int(back), ok, i)
		}
		if got, want := tok.String(), g.String(); got != want {
			t.Fatalf("Token(%d).String() = %q, want go/token's %q", i, got, want)
		}
	}
}

// TestGoTokenEndsBelowGolpeBase guards the layout: go/token must name no
// value at or above golpeBase, or a Golpe token would shadow a Go one.
func TestGoTokenEndsBelowGolpeBase(t *testing.T) {
	last := -1
	for i := 0; i < int(golpeEnd); i++ {
		if gotoken.Token(i).String() != unnamed(i) {
			last = i
		}
	}
	if last < 0 || last >= int(golpeBase) {
		t.Fatalf("last value go/token names is %d; want it in [0, %d)", last, golpeBase)
	}
	t.Logf("go/token names values up to %d; Golpe tokens start at %d", last, golpeBase)
}

func TestGolpeTokens(t *testing.T) {
	// String follows go/token's convention: keywords and fixed delimiters
	// print as their spelling, token classes as their constant name.
	tests := []struct {
		tok  Token
		name string
	}{
		{TEMPLATE, "template"},
		{TAG_OPEN, "TAG_OPEN"},
		{TAG_END, ">"},
		{TAG_SELF_CLOSE, "/>"},
		{TAG_CLOSE, "TAG_CLOSE"},
		{TEXT, "TEXT"},
		{INTERP_OPEN, "{"},
		{INTERP_CLOSE, "}"},
		{HTML_COMMENT, "HTML_COMMENT"},
		{DOCTYPE, "DOCTYPE"},
	}
	if len(tests) != int(golpeEnd-golpeBase) {
		t.Fatalf("table lists %d Golpe tokens, the package defines %d", len(tests), int(golpeEnd-golpeBase))
	}
	seen := map[Token]bool{}
	for _, tt := range tests {
		if tt.tok < golpeBase || tt.tok >= golpeEnd {
			t.Errorf("%s = %d, outside the Golpe range [%d, %d)", tt.name, int(tt.tok), golpeBase, golpeEnd)
		}
		if seen[tt.tok] {
			t.Errorf("%s = %d is listed twice", tt.name, int(tt.tok))
		}
		seen[tt.tok] = true
		if got := tt.tok.String(); got != tt.name {
			t.Errorf("Token(%d).String() = %q, want %q", int(tt.tok), got, tt.name)
		}
		if g, ok := tt.tok.Go(); ok || g != gotoken.ILLEGAL {
			t.Errorf("%s.Go() = %v, %v; want ILLEGAL, false", tt.name, g, ok)
		}
	}
}

// TestUnknownValues: values past the last Golpe token have no name and no
// go/token equivalent; they print like go/token's unnamed values.
func TestUnknownValues(t *testing.T) {
	for _, i := range []int{int(golpeEnd), int(golpeEnd) + 1, 1 << 20, math.MaxInt} {
		tok := Token(i)
		if got, want := tok.String(), unnamed(i); got != want {
			t.Errorf("Token(%d).String() = %q, want %q", i, got, want)
		}
		if g, ok := tok.Go(); ok || g != gotoken.ILLEGAL {
			t.Errorf("Token(%d).Go() = %v, %v; want ILLEGAL, false", i, g, ok)
		}
	}
}

func TestLookup(t *testing.T) {
	// Every Go keyword maps to its go/token token.
	keywords := 0
	for i := 0; i < int(golpeBase); i++ {
		g := gotoken.Token(i)
		if !g.IsKeyword() {
			continue
		}
		keywords++
		if got := Lookup(g.String()); got != FromGo(g) {
			t.Errorf("Lookup(%q) = %v, want %v", g.String(), got, g)
		}
	}
	if keywords == 0 {
		t.Fatal("found no Go keywords to look up")
	}
	// Everything else is IDENT, including "template": it is a contextual
	// keyword (DEC-003), not a reserved word.
	for _, s := range []string{"template", "Template", "TEMPLATE", "element", "component", "x", "_", "", "Func", "if "} {
		if got := Lookup(s); got != FromGo(gotoken.IDENT) {
			t.Errorf("Lookup(%q) = %v, want IDENT", s, got)
		}
	}
}

func TestIsTemplateKeyword(t *testing.T) {
	g := FromGo
	tests := []struct {
		lit  string
		next Token
		want bool
	}{
		{"template", g(gotoken.IDENT), true},   // template Card(…), template List[T any](…)
		{"template", g(gotoken.LPAREN), true},  // template (c Card) View()
		{"template", g(gotoken.PERIOD), false}, // template.Must(…)
		{"template", g(gotoken.ASSIGN), false}, // template = …
		{"template", g(gotoken.DEFINE), false}, // template := …
		{"template", g(gotoken.LBRACK), false},
		{"template", g(gotoken.LBRACE), false},
		{"template", g(gotoken.COMMA), false},
		{"template", g(gotoken.SEMICOLON), false},
		{"template", g(gotoken.EOF), false},
		{"template", g(gotoken.FUNC), false},
		{"template", g(gotoken.STRING), false},
		{"template", TEMPLATE, false},
		{"template", TAG_OPEN, false},
		{"Template", g(gotoken.IDENT), false}, // keywords are case-sensitive
		{"TEMPLATE", g(gotoken.LPAREN), false},
		{"templates", g(gotoken.IDENT), false},
		{"func", g(gotoken.IDENT), false},
		{"", g(gotoken.IDENT), false},
	}
	for _, tt := range tests {
		if got := IsTemplateKeyword(tt.lit, tt.next); got != tt.want {
			t.Errorf("IsTemplateKeyword(%q, %v) = %v, want %v", tt.lit, tt.next, got, tt.want)
		}
	}
}
