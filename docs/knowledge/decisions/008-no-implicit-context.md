---
type: Decision
id: DEC-008
title: No implicit context
description: Elements stay func(io.Writer) error; a context.Context, when needed, is an ordinary template parameter, and cancellation can be done by wrapping the writer.
status: stable
decision: accepted
tags: [language, api]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
verified: { by: human:sgq995, at: 2026-11-03T03:12:42Z }
---
# Context

templ passes a `context.Context` to every component and exposes an implicit
`ctx`. The agent recommended the same, for cancelling slow non-write work and
reaching deep components without prop drilling.

# Options

1. `func(ctx context.Context, w io.Writer) error` with an implicit `ctx`.
2. An interface `Render(ctx, w) error` (templ style).
3. Keep `func(io.Writer) error`.

# Decision

Option 3, without even a `ContextWriter` helper. The maintainer's reasons:
an implicit `ctx` needs special syntax; a writer can be wrapped (closure, or a
writer that fails once a context is done); explicit prop drilling lets
developers use their data however they want, which is better than a hidden
shared state.

# Consequences

- Templates that need a context take it as a parameter:
  `template Page(ctx context.Context, p PageProps)`.
- A slow call inside a template body is not interrupted by a client
  disconnect until the next write fails.
