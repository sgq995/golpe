---
okf_version: "0.2"
---
# Golpe knowledge base

An [Open Knowledge Format](https://okf.md) v0.2 bundle holding everything we
know about Golpe: who uses it, what its parts are, what we are building, why,
how, and what is left to do. Agents read this file first and descend
(progressive disclosure); they never bulk-read the tree.

# Start here
* [Process](process/) - how features go from a chat message to spec, plan, tasks and code
* [Conventions](process/conventions.md) - concept types, required frontmatter, IDs and link style

# Domain model
* [Actors](actors/) - people and external systems that interact with Golpe
* [Actuators](actuators/) - Golpe's own components that perform work
* [Entities](entities/) - the things Golpe reads, produces and reasons about

# Delivery
* [Specs](specs/) - what we build and why
* [Plans](plans/) - how we build it: architecture, snippets, tests, labs
* [Tasks](tasks/) - plan slices as todo lists, one file per slice
* [Decisions](decisions/) - architecture decision records
* [Backlog](backlog.md) - deferred and rejected ideas, each with its reason

# History
* [Log](log.md) - newest-first change history of this bundle
