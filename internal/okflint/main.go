// Command okflint validates the project's knowledge base as an OKF v0.2
// bundle, plus the Golpe-specific conventions described in
// docs/knowledge/process/conventions.md.
//
// Usage:
//
//	go run ./internal/okflint [bundle-dir]   # default: docs/knowledge
//
// It is deterministic, has no dependencies outside the standard library and
// exits non-zero when any rule is violated. `go test ./...` also lints the
// real bundle (see TestKnowledgeBundle).
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

func main() {
	dir := "docs/knowledge"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	problems, err := Lint(os.DirFS(dir))
	if err != nil {
		fmt.Fprintln(os.Stderr, "okflint:", err)
		os.Exit(2)
	}
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, p)
	}
	if len(problems) > 0 {
		fmt.Fprintf(os.Stderr, "okflint: %d problem(s) in %s\n", len(problems), dir)
		os.Exit(1)
	}
	fmt.Printf("okflint: %s is a valid OKF bundle\n", dir)
}

// schema lists, per concept type, the frontmatter keys this project requires
// on top of OKF's `type`. Unknown types are allowed (OKF §4.1) and only need
// `type`, `title` and `description`.
var schema = map[string][]string{
	"Actor":      {"id", "title", "description"},
	"Actuator":   {"id", "title", "description"},
	"Entity":     {"id", "title", "description"},
	"Spec":       {"id", "title", "description", "status"},
	"Plan":       {"id", "title", "description", "status", "spec"},
	"Task Slice": {"id", "title", "description", "spec", "plan", "state"},
	"Decision":   {"id", "title", "description", "status", "decision"},
	"Playbook":   {"id", "title", "description"},
}

var allowed = map[string][]string{
	"status":   {"draft", "stable", "deprecated"},
	"state":    {"todo", "doing", "done", "blocked"},
	"decision": {"proposed", "accepted", "superseded", "rejected"},
}

var mdLink = regexp.MustCompile(`\]\(([^)\s]+)\)`)

type doc struct {
	path string
	fm   map[string]string
	has  bool // has a frontmatter block
	body string
}

// Lint returns one human-readable line per problem found in the bundle.
func Lint(bundle fs.FS) ([]string, error) {
	var docs []doc
	err := fs.WalkDir(bundle, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		b, err := fs.ReadFile(bundle, p)
		if err != nil {
			return err
		}
		fm, has, body, err := splitFrontmatter(b)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		docs = append(docs, doc{path: p, fm: fm, has: has, body: body})
		return nil
	})
	if err != nil {
		return nil, err
	}

	exists := map[string]bool{}
	for _, d := range docs {
		exists[d.path] = true
	}
	ids := map[string]string{}
	var problems []string
	report := func(p, format string, args ...any) {
		problems = append(problems, p+": "+fmt.Sprintf(format, args...))
	}

	for _, d := range docs {
		base := path.Base(d.path)
		switch {
		case base == "index.md" && d.path == "index.md":
			for k := range d.fm {
				if k != "okf_version" {
					report(d.path, "root index.md may only carry okf_version, found %q", k)
				}
			}
		case base == "index.md":
			if d.has {
				report(d.path, "index.md files must not carry frontmatter")
			}
		case base == "log.md":
			// Free-form; dates are checked by convention, not by the linter.
		default:
			lintConcept(d, ids, report)
			idx := path.Join(path.Dir(d.path), "index.md")
			if !exists[idx] {
				report(d.path, "directory has no index.md listing this concept")
			} else if !listedIn(docs, idx, d.path) {
				report(d.path, "not listed in %s", idx)
			}
		}
		for _, m := range mdLink.FindAllStringSubmatch(d.body, -1) {
			target, ok := resolve(d.path, m[1])
			if ok && !exists[target] && !exists[path.Join(target, "index.md")] {
				report(d.path, "broken link %q", m[1])
			}
		}
	}
	sort.Strings(problems)
	return problems, nil
}

func lintConcept(d doc, ids map[string]string, report func(string, string, ...any)) {
	if !d.has {
		report(d.path, "concept has no YAML frontmatter")
		return
	}
	typ := d.fm["type"]
	if typ == "" {
		report(d.path, "frontmatter has no non-empty type")
		return
	}
	required, ok := schema[typ]
	if !ok {
		required = []string{"title", "description"}
	}
	for _, k := range required {
		if d.fm[k] == "" {
			report(d.path, "%s is missing required key %q", typ, k)
		}
	}
	for k, vals := range allowed {
		if v, ok := d.fm[k]; ok && v != "" && !contains(vals, v) {
			report(d.path, "%s %q is not one of %v", k, v, vals)
		}
	}
	if id := d.fm["id"]; id != "" {
		if other, dup := ids[id]; dup {
			report(d.path, "duplicate id %q (also in %s)", id, other)
		}
		ids[id] = d.path
	}
}

func listedIn(docs []doc, index, concept string) bool {
	for _, d := range docs {
		if d.path != index {
			continue
		}
		for _, m := range mdLink.FindAllStringSubmatch(d.body, -1) {
			if t, ok := resolve(index, m[1]); ok && t == concept {
				return true
			}
		}
	}
	return false
}

// resolve maps a link target to a bundle path. It reports false for external
// links, anchors and non-markdown targets.
func resolve(from, target string) (string, bool) {
	if i := strings.IndexAny(target, "#?"); i >= 0 {
		target = target[:i]
	}
	if target == "" || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
		return "", false
	}
	dir := strings.HasSuffix(target, "/")
	if !dir && !strings.HasSuffix(target, ".md") {
		return "", false
	}
	var p string
	if strings.HasPrefix(target, "/") {
		p = strings.TrimPrefix(target, "/")
	} else {
		p = path.Join(path.Dir(from), target)
	}
	p = path.Clean(p)
	if strings.HasPrefix(p, "../") || p == ".." {
		return "", false // points outside the bundle (e.g. source code)
	}
	if dir {
		p = path.Join(p, "index.md")
	}
	return p, true
}

// splitFrontmatter extracts top-level `key: value` pairs. Nested YAML is kept
// opaque: a key whose value is on following indented lines maps to "…".
func splitFrontmatter(b []byte) (map[string]string, bool, string, error) {
	fm := map[string]string{}
	if !bytes.HasPrefix(b, []byte("---\n")) {
		return fm, false, string(b), nil
	}
	rest := b[len("---\n"):]
	end := bytes.Index(rest, []byte("\n---\n"))
	if end < 0 {
		if bytes.HasSuffix(rest, []byte("\n---")) {
			end = len(rest) - len("\n---")
		} else {
			return nil, false, "", fmt.Errorf("unterminated frontmatter")
		}
	}
	body := ""
	if end+len("\n---\n") <= len(rest) {
		body = string(rest[end+len("\n---\n"):])
	}
	sc := bufio.NewScanner(bytes.NewReader(rest[:end]))
	for sc.Scan() {
		line := sc.Text()
		if line == "" || line[0] == ' ' || line[0] == '\t' || line[0] == '-' || line[0] == '#' {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			return nil, false, "", fmt.Errorf("frontmatter line %q is not key: value", line)
		}
		v = strings.TrimSpace(v)
		if i := strings.Index(v, " #"); i >= 0 {
			v = strings.TrimSpace(v[:i])
		}
		v = strings.Trim(v, `"'`)
		if v == "" {
			v = "…"
		}
		fm[strings.TrimSpace(k)] = v
	}
	return fm, true, body, nil
}

func contains(xs []string, x string) bool {
	for _, s := range xs {
		if s == x {
			return true
		}
	}
	return false
}
