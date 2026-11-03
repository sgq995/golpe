---
type: Plan
id: PLAN-006
title: Optional helpers
description: Add Class and If next to the Element alias in the root package, standard library only, with table tests and runnable examples.
status: draft
spec: SPEC-006
tags: [api]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Approach

Implements [SPEC-006](/specs/006-optional-helpers.md): two pure functions in
the root package next to `Element`.

# Snippets

```go
// Class joins the non-empty parts with single spaces.
func Class(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(p)
	}
	return b.String()
}

// If returns s when cond is true, and "" otherwise.
func If(cond bool, s string) string {
	if cond {
		return s
	}
	return ""
}
```

# Testing

Table tests for both, runnable `Example` functions, 100% coverage.

# Laboratories

None.

# Risks

None.

# Slices

1. [TASK-006-01 Class and If](/tasks/006-optional-helpers/01-class-and-if.md)
