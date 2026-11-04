package golpe_test

import (
	"testing"

	"github.com/sgq995/golpe"
)

func TestClass(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{"no parts", nil, ""},
		{"only empty parts", []string{"", ""}, ""},
		{"one part", []string{"btn"}, "btn"},
		{"two parts", []string{"btn", "primary"}, "btn primary"},
		{"skips empty parts", []string{"", "a", "", "b"}, "a b"},
		{"leading empty part", []string{"", "btn"}, "btn"},
		// SPEC-006 example: Class("btn", "primary", "") is "btn primary".
		{"trailing empty part", []string{"btn", "primary", ""}, "btn primary"},
		{"with If", []string{"btn", golpe.If(true, "active"), golpe.If(false, "disabled")}, "btn active"},
		// Parts are used exactly as given.
		{"parts are not trimmed", []string{" a ", "b"}, " a  b"},
		{"blank parts are not empty", []string{"a", " ", "b"}, "a   b"},
		{"parts are not split", []string{"a  b", "c"}, "a  b c"},
		{"duplicates are kept", []string{"a", "a"}, "a a"},
		{"parts are not escaped", []string{"<a>", `"&"`}, `<a> "&"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := golpe.Class(tt.parts...); got != tt.want {
				t.Errorf("Class(%q) = %q, want %q", tt.parts, got, tt.want)
			}
		})
	}
}

func TestIf(t *testing.T) {
	tests := []struct {
		cond bool
		s    string
		want string
	}{
		{true, "x", "x"},
		{false, "x", ""},
		{true, "", ""},
		{false, "", ""},
		{true, " a b ", " a b "},
	}
	for _, tt := range tests {
		if got := golpe.If(tt.cond, tt.s); got != tt.want {
			t.Errorf("If(%t, %q) = %q, want %q", tt.cond, tt.s, got, tt.want)
		}
	}
}
