# ADR 0021: Active work scope change behavior

## Status

Accepted

## Context

Target Scope is a shell-level primitive, but Autopilot Runs and Copilot Sessions may already be active when the operator changes that scope. Silent mutation would make active work hard to trust.

## Decision

Changing shell **Target Scope** does not silently rewrite the scope of an active Run or Session.

- Watchtower updates immediately to reflect the new shell scope.
- Autopilot and Copilot surface that shell scope has changed.
- The operator explicitly decides whether the active Run or Session should adopt the new scope.

## Consequences

- Active execution and advisory context remain stable and trustworthy.
- The UI needs a visible distinction between shell scope and active-work scope when they diverge.
- Cross-mode behavior remains predictable during long-lived workflows.
