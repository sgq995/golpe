---
type: Decision
id: DEC-009
title: Escape interpolated values by default
description: Value interpolations and dynamic attribute values are HTML-escaped; static text the author writes is verbatim; unescaped output goes explicitly through writer().
status: stable
decision: accepted
tags: [security, generator]
generated: { by: claude-code/agent, at: 2026-11-03T03:12:42Z }
verified: { by: human:sgq995, at: 2026-11-03T03:12:42Z }
---
# Context

During the interview the maintainer considered writing every `{expr}` as
trusted HTML and leaving escaping to the developer, partly to avoid
formatting overhead.

# Decision

Escape by default: `html.EscapeString(fmt.Sprint(x))` for values and dynamic
attributes. Static text and static attribute values are written verbatim, so
entities such as `&nbsp;` work. Trusted HTML is written explicitly through
`writer()`.

Reason: unescaped-by-default turns any user-supplied string (a name, a search
query) into an XSS hole, and mainstream engines (html/template, templ, React,
Svelte) escape by default.

# Consequences

- Escaping is not sanitizing: `href={u}` with `javascript:…` still runs. URL
  and JavaScript safety is a later security spec ([backlog](/backlog.md)).
