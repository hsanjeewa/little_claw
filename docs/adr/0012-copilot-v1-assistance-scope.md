# ADR 0012: Copilot v1 assistance scope

## Status

Accepted

## Context

Copilot observes the in-app Terminal Session, but its concrete behavior must be narrow enough to be predictable and useful. Without a defined assistance scope, Copilot risks becoming either noisy or ambiguously agentic.

## Decision

In v1, Copilot's assistance scope is limited to four behaviors:

1. **Explain** what just happened
2. **Warn** about risky or suspicious patterns
3. **Suggest** plausible next commands or actions
4. **Summarize** the session or incident so far

Copilot does not take control by default, does not continuously interrupt the operator with unrelated guidance, and does not redefine terminal execution behavior.

## Consequences

- Copilot stays distinct from Autopilot.
- Users can form reliable expectations about what kind of help Copilot provides.
- UI affordances can be optimized around explanations, warnings, suggestions, and summaries instead of broad autonomous behaviors.
