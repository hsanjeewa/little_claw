# ADR 0010: Slash command boundary

## Status

Accepted

## Context

Autopilot and Copilot need a fast text-driven control surface. Without a clear boundary, slash commands could blur into ordinary shell execution and weaken the distinction between application control and terminal commands.

## Decision

Slash Commands are **in-app control actions only**.

They are used to drive structured product behavior such as planning, approvals, scope changes, handoff, or summarization. They are not shell aliases and do not replace ordinary terminal commands.

Examples of Slash Commands include actions like:

- `/plan`
- `/approve`
- `/scope selected`
- `/handoff`
- `/summarize`

Operational commands such as `systemctl`, `kubectl`, `ssh`, or package-manager commands remain normal terminal input or agent-proposed execution steps.

## Consequences

- The UI and parser can distinguish product intent from shell intent cleanly.
- Autopilot and Copilot keep a stable interaction model even when terminal execution is embedded nearby.
- Slash command discovery, help text, and validation become part of the application layer rather than the shell layer.
