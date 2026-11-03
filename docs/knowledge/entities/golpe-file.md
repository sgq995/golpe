---
type: Entity
id: ENT-golpe-file
title: Golpe source file
description: A .golpe file: a Go source file that may also contain template declarations.
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Shape
A Go file (package clause, imports, declarations) with the `.golpe` extension. Any valid Go file is a valid Golpe file with the same meaning. In addition it may contain [template declarations](/entities/template-declaration.md) and use the predeclared names `element` ([Element](/entities/element.md)) and [writer()](/entities/writer-function.md).

Each `x.golpe` produces one [generated file](/entities/generated-file.md), `x_gen.go`, in the same package.

# Involved in
[SPEC-001](/specs/001-template-language.md), [SPEC-002](/specs/002-code-generation.md), [SPEC-003](/specs/003-cli.md).
