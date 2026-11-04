package golpe_test

import (
	"cmp"
	"fmt"

	"github.com/sgq995/golpe"
)

func ExampleClass() {
	fmt.Printf("%q\n", golpe.Class("btn", "primary", ""))
	fmt.Printf("%q\n", golpe.Class("", "a", "", "b"))
	fmt.Printf("%q\n", golpe.Class())
	// Output:
	// "btn primary"
	// "a b"
	// ""
}

// A class value built from optional parts, as in the template
//
//	<button class={golpe.Class("btn", cmp.Or(p.Variant, "primary"), golpe.If(p.Active, "active"))}>
//
// cmp.Or falls back to a default variant, If keeps "active" only for an
// active button, and Class skips whatever came out empty.
func ExampleClass_optionalParts() {
	type ButtonProps struct {
		Variant string
		Active  bool
	}
	for _, p := range []ButtonProps{
		{},
		{Variant: "danger"},
		{Variant: "ghost", Active: true},
	} {
		fmt.Println(golpe.Class("btn", cmp.Or(p.Variant, "primary"), golpe.If(p.Active, "active")))
	}
	// Output:
	// btn primary
	// btn danger
	// btn ghost active
}

func ExampleIf() {
	for _, active := range []bool{true, false} {
		fmt.Printf("%q\n", golpe.If(active, "active"))
	}
	// Output:
	// "active"
	// ""
}
