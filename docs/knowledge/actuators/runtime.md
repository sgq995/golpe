---
type: Actuator
id: ACT-runtime
title: Optional helpers
description: The public root package github.com/sgq995/golpe: optional helpers (Element, Class, If) that developers may import; generated code never does.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Responsibilities
- `type Element = func(io.Writer) error` ([DEC-002](/decisions/002-element-is-an-alias.md)).
- `Class` and `If` for building attribute values ([SPEC-006](/specs/006-optional-helpers.md)).

Nothing here is required: generated code imports only the standard library ([DEC-005](/decisions/005-standard-library-only-output.md)). The runtime writers once planned for generated code were removed.

# Code
Repo root (`golpe.go`).

# Involved in
[SPEC-006](/specs/006-optional-helpers.md).
