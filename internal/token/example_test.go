package token_test

import (
	"fmt"
	"go/scanner"
	gotoken "go/token"

	"github.com/sgq995/golpe/internal/token"
)

// Go's tokens and Golpe's share one Token type and print the go/token way:
// keywords and delimiters as their spelling, token classes by name.
func ExampleToken_String() {
	fmt.Println(token.FromGo(gotoken.FUNC), token.FromGo(gotoken.LBRACE), token.FromGo(gotoken.IDENT))
	fmt.Println(token.TEMPLATE, token.INTERP_OPEN, token.TAG_SELF_CLOSE, token.TAG_OPEN, token.TEXT)
	// Output:
	// func { IDENT
	// template { /> TAG_OPEN TEXT
}

func ExampleFromGo() {
	tok := token.FromGo(gotoken.ARROW)
	g, ok := tok.Go()
	fmt.Println(tok, g == gotoken.ARROW, ok)
	// Output: <- true true
}

func ExampleToken_Go() {
	for _, tok := range []token.Token{token.FromGo(gotoken.IDENT), token.TEMPLATE, token.TEXT} {
		g, ok := tok.Go()
		fmt.Println(tok, "->", g, ok)
	}
	// Output:
	// IDENT -> IDENT true
	// template -> ILLEGAL false
	// TEXT -> ILLEGAL false
}

// Lookup is go/token.Lookup: "template" is not reserved, so it is an
// identifier.
func ExampleLookup() {
	fmt.Println(token.Lookup("func"), token.Lookup("select"), token.Lookup("template"), token.Lookup("element"))
	// Output: func select IDENT IDENT
}

// The parser scans .golpe source with go/scanner, which reports "template"
// as an identifier. At the start of each top-level declaration it asks
// IsTemplateKeyword about that identifier and the token after it.
func ExampleIsTemplateKeyword() {
	for _, src := range []string{
		"template Card(title string)",
		"template (c Card) View()",
		"template.Must(t, err)",
	} {
		var s scanner.Scanner
		file := gotoken.NewFileSet().AddFile("example.golpe", -1, len(src))
		s.Init(file, []byte(src), nil, 0)
		_, _, lit := s.Scan()
		_, next, _ := s.Scan()
		fmt.Printf("%-28s %v\n", src, token.IsTemplateKeyword(lit, token.FromGo(next)))
	}
	// Output:
	// template Card(title string)  true
	// template (c Card) View()     true
	// template.Must(t, err)        false
}

func ExampleIsVoid() {
	for _, tag := range []string{"br", "IMG", "div", "my-input"} {
		fmt.Println(tag, token.IsVoid(tag))
	}
	// Output:
	// br true
	// IMG true
	// div false
	// my-input false
}

func ExampleIsRawText() {
	for _, tag := range []string{"script", "STYLE", "textarea"} {
		fmt.Println(tag, token.IsRawText(tag))
	}
	// Output:
	// script true
	// STYLE true
	// textarea false
}

func ExampleIsEmbeddedKeyword() {
	for _, word := range []string{"if", "select", "If", "else"} {
		fmt.Println(word, token.IsEmbeddedKeyword(word))
	}
	// Output:
	// if true
	// select true
	// If false
	// else false
}

// A tag is markup when its name is an HTML, SVG or MathML element, ignoring
// ASCII case, or a custom element; any other name is a component tag
// (SPEC-005 R1). So "Button" is the HTML button element, not a component.
func ExampleIsElement() {
	for _, tag := range []string{"div", "DIV", "linearGradient", "mrow", "Button", "Card", "ui.Button", "my-widget"} {
		fmt.Println(tag, token.IsElement(tag))
	}
	// Output:
	// div true
	// DIV true
	// linearGradient true
	// mrow true
	// Button true
	// Card false
	// ui.Button false
	// my-widget false
}

// Any name with a '-' is a custom element, and so markup, even where the
// HTML standard would reject it as a custom element name.
func ExampleIsCustomElement() {
	for _, tag := range []string{"my-widget", "My-Widget", "annotation-xml", "myWidget", "my_widget"} {
		fmt.Println(tag, token.IsCustomElement(tag))
	}
	// Output:
	// my-widget true
	// My-Widget true
	// annotation-xml true
	// myWidget false
	// my_widget false
}
