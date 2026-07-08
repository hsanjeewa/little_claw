# ADR 0030: Skill model

## Status

Accepted

## Context

The product needs first-class skills comparable to modern agent systems, but the meaning of skill must stay distinct from tools, slash commands, and general memory. A plugin-like execution model would create unclear authority boundaries in the first version.

## Decision

In the initial version, a **Skill** is a reusable instruction package that shapes planning and execution behavior for a class of operational work.

A Skill may attach to either a **Copilot Session** or an **Autopilot Run**.

A Skill may guide:

- goal interpretation
- evidence gathering
- tool selection
- approval structure
- verification expectations

A Skill is not itself an arbitrary executable plugin, and it does not replace **Operational Tools** or **Slash Commands**.

## Consequences

- Skills can remain understandable, reviewable, and composable with the supervised Autopilot model.
- Skills can guide both human-led investigation and agent-led execution without collapsing Copilot and Autopilot into one mode.
- Tools stay responsible for operational execution, while Skills shape how the agent approaches the work.
- Future versions may enrich skill packaging and discovery without redefining the core concept.
