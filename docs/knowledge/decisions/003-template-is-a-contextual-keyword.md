---
type: Decision
id: DEC-003
title: template is a contextual keyword
description: 'template' is a keyword only at the start of a top-level declaration followed by a name or receiver, so existing Go code (including html/template's package name) keeps working.
status: draft
decision: proposed
tags: [language, parser]
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Context

The keyword could be `template` or `component`. Either is a common Go
identifier; `template` is the package name of `html/template` and
`text/template`, so `template.Must(...)` and `var template = …` exist in real
code.

# Options

1. Reserved keyword: breaks valid Go; Golpe would no longer be a superset.
2. Contextual keyword: `template` starts a template declaration only at
   top level, where Go expects `func`/`var`/`const`/`type`/`import`, and
   only when followed by an identifier or `(` (method receiver).
3. Use `component` instead: same problem, less common, but the maintainer's
   examples use `template`.

# Decision

Option 2 with the word `template`. At top level Go never allows a bare
identifier to start a declaration, so the rule is unambiguous, and every
valid Go file keeps its meaning. `component` is not reserved; an alias can be
added later if wanted.

# Consequences

- Inside function bodies and expressions `template` is an ordinary
  identifier (`template.HTMLEscapeString(s)` works).
- A package-level `template` *identifier* cannot be used as the first token
  of a declaration, which Go does not allow anyway.
