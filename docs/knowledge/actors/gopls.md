---
type: Actor
id: ACTOR-gopls
title: gopls
description: The official Go language server, which Golpe's language server runs as a subprocess to provide Go intelligence.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Role
External system. gopls's packages are internal to golang.org/x/tools/gopls and cannot be imported as a library, so the [language server](/actuators/language-server.md) runs it as a child process and proxies requests to it, over the [generated files](/entities/generated-file.md).

# Involved in
[SPEC-004](/specs/004-language-server.md).
