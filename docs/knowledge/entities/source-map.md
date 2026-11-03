---
type: Entity
id: ENT-source-map
title: Source map
description: The mapping between positions in a generated file and positions in its .golpe source, carried by /*line*/ directives.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Shape
`/*line nav.golpe:8:13*/` comments emitted before every piece of user Go code in the [generated file](/entities/generated-file.md). go/token honours them, so compiler and vet errors name the `.golpe` file (LAB-006 in [PLAN-001](/plans/001-template-language.md)). The [language server](/actuators/language-server.md) rebuilds a bidirectional table from the same directives.

# Involved in
[SPEC-002](/specs/002-code-generation.md), [SPEC-004](/specs/004-language-server.md).
