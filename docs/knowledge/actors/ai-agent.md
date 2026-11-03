---
type: Actor
id: ACTOR-ai-agent
title: AI agent
description: A coding agent (for example Claude Code) that writes the knowledge base and implements task slices, alone or in parallel with other agents.
generated: { by: claude-code/agent, at: 2026-11-03T00:00:00Z }
---
# Responsibilities
- Follows the [feature workflow](/process/feature-workflow.md) to write [specs](/specs/), [plans](/plans/) and [task slices](/tasks/).
- Follows the [implementation workflow](/process/implementation-workflow.md) and the [git workflow](/process/git-workflow.md).
- Keeps the bundle valid with [okflint](/actuators/okflint.md).

# Identity
Recorded in frontmatter as `generated.by: claude-code/agent`.
