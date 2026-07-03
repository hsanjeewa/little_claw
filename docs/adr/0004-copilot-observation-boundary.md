# ADR 0004: Copilot observation boundary

## Status

Accepted

## Context

Copilot is intended to assist a human operator during terminal-based operational work. The phrase "listen to communication between servers and the DevOps engineer" needs a precise product boundary so the feature is implementable, comprehensible, and safe.

## Decision

The first version of Copilot will observe only the **in-app terminal session inside the TUI**.

Copilot may analyze commands and output that pass through its own terminal pane, then provide guidance, warnings, summaries, and suggestions. It will not observe arbitrary shell activity elsewhere on the operator machine.

## Consequences

- Copilot has a clear and enforceable observation scope.
- Users get predictable behavior tied to the TUI instead of implicit background monitoring.
- Integration complexity is reduced because observation is limited to the application-managed terminal stream.
- Broader workstation or multi-terminal observation remains a future product decision rather than an accidental v1 commitment.
