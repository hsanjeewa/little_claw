# ADR 0020: Copilot session entry

## Status

Accepted

## Context

Copilot is human-led and advisory, so its entry model should differ from Autopilot's execution-oriented Goal model. The product needs a canonical way to describe what starts a Copilot Session.

## Decision

A Copilot Session starts from a **Task Context**.

Task Context is the human-led working situation, question, or incident framing that anchors the Session. It may include current host or scope, recent terminal activity, and the operator's current objective, but it does not need to be a clean executable Goal.

In v1, Session entry may begin lazily from terminal activity, with Task Context added, inferred, or edited as the Session becomes clearer.

## Consequences

- Copilot can support ambiguous or exploratory work without pretending every interaction is an execution plan.
- Copilot remains distinct from Autopilot in both language and operating model.
- Promotion to Autopilot becomes a meaningful transition from Task Context into Goal-driven execution.
- The UI should let operators review and correct inferred Task Context instead of treating inference as authoritative.
