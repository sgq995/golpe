---
type: Spec
id: SPEC-004
title: Language server
description: golpe lsp gives editors Go-quality diagnostics, completion, hover and navigation in .golpe files by proxying gopls over generated Go.
status: draft
plan: PLAN-004
tags: [lsp, tooling]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
---
# Summary

A language server for `.golpe` files that reuses gopls for everything Go,
so Golpe gets diagnostics, completion, hover, go-to-definition, references
and rename in Go code and interpolations without reimplementing Go tooling.

# Motivation

Without editor support a superset language is painful: Go code inside
templates would get no completion and errors would only show on build.
gopls already does all of this for Go; it cannot be imported as a library
(its packages are `internal`), but it can be driven as a subprocess over
generated Go, which is how templ's LSP works.

# Actors

[Developer](/actors/developer.md) · [Editor](/actors/editor.md) ·
[gopls](/actors/gopls.md) · [Language server](/actuators/language-server.md)

# Entities

[Generated file](/entities/generated-file.md) ·
[Source map](/entities/source-map.md) ·
[Golpe source file](/entities/golpe-file.md)

# Requirements

- **R1** `golpe lsp`, a subcommand of the single CLI, MUST speak LSP 3.17 over
  stdio.
- **R2** On open/change of a `.golpe` file it MUST regenerate the Go in
  memory and present it to gopls as the overlay of `x_gen.go`.
- **R3** Diagnostics from gopls for generated files MUST be mapped to `.golpe`
  positions; Golpe syntax errors MUST be published too.
- **R4** Completion, hover, signature help, definition, references and rename
  MUST work inside Go regions, embedded statement headers and
  interpolations, with positions mapped both ways.
- **R5** Requests on `.go` files MUST be forwarded to gopls untouched, so one
  server can serve a mixed workspace.
- **R6** Requests at positions inside pure markup (tags, text) SHOULD return
  empty results, never errors.

# Acceptance criteria

- [ ] An integration test drives the server with a scripted LSP client:
      a type error in `{props.Nmae}` is published at the right range; hover on
      `props` shows its type; definition on `Button` jumps into the `.golpe`
      file.

# Out of scope

HTML-aware features (tag/attribute completion, auto-closing tags), formatting.

# Open questions

1. Discover gopls on `PATH`, or pin and `go run` a version?
2. Ship editor extensions (VS Code) in this repo or separately?

# Status

Confirmed in the design interview (2026-11-03): the LSP proxies **gopls**,
and work starts after the parser and generator ship; this spec stays a draft
until then.
