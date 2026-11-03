package main

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

// TestKnowledgeBundle lints the real knowledge base, so `go test ./...`
// fails whenever an agent leaves docs/knowledge inconsistent.
func TestKnowledgeBundle(t *testing.T) {
	problems, err := Lint(os.DirFS("../../docs/knowledge"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range problems {
		t.Error(p)
	}
}

func TestLint(t *testing.T) {
	bundle := fstest.MapFS{
		"index.md":       {Data: []byte("---\nokf_version: \"0.2\"\nextra: x\n---\n* [A](/a/)\n")},
		"a/index.md":     {Data: []byte("* [One](one.md)\n")},
		"a/one.md":       {Data: []byte("---\ntype: Spec\nid: SPEC-1\ntitle: One\ndescription: d\nstatus: wip\n---\nSee [missing](/a/missing.md) and [code](../../go.mod).\n")},
		"a/two.md":       {Data: []byte("---\ntype: Entity\nid: SPEC-1\ntitle: Two\ndescription: d\n---\n")},
		"a/three.md":     {Data: []byte("no frontmatter\n")},
		"b/sub/index.md": {Data: []byte("---\ntype: x\n---\n")},
		"a/sub/log.md":   {Data: []byte("# Log\n")},
	}
	problems, err := Lint(bundle)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		`index.md: root index.md may only carry okf_version, found "extra"`,
		`a/one.md: status "wip" is not one of`,
		`a/one.md: broken link "/a/missing.md"`,
		`a/two.md: duplicate id "SPEC-1"`,
		`a/two.md: not listed in a/index.md`,
		`a/three.md: concept has no YAML frontmatter`,
		`b/sub/index.md: index.md files must not carry frontmatter`,
	}
	got := strings.Join(problems, "\n")
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("missing problem %q in:\n%s", w, got)
		}
	}
	if len(problems) != len(want)+1 { // +1: a/three.md is also not listed in a/index.md
		t.Errorf("got %d problems, want %d:\n%s", len(problems), len(want)+1, got)
	}
}
