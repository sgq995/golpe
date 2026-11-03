---
type: Actuator
id: ACT-runtime
title: Runtime
description: The public root package github.com/sgq995/golpe: the Element type and the small helpers generated code calls to write and escape values.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Responsibilities
- Define [Element](/entities/element.md).
- Write interpolated values: render Elements, HTML-escape everything else ([DEC-004](/decisions/004-runtime-dispatch-for-interpolation.md)).
- Escape attribute values.

Generated code depends only on the standard library and this package.

# Code
Repo root (`golpe.go`).

# Involved in
[SPEC-002](/specs/002-code-generation.md).
