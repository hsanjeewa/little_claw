# ADR 0005: Copilot Session to Autopilot Run promotion

## Status

Accepted

## Context

Copilot and Autopilot have distinct authority boundaries, but real operator workflows often begin as a human-led investigation and later need agent-led execution. The product needs a clear relationship between these two modes without collapsing them into one.

## Decision

The system will support **explicit promotion** from a Copilot Session to a new Autopilot Run.

The operator may deliberately promote the current Session context into Autopilot. That promotion may carry forward relevant goal framing, recent evidence, terminal context, and attached Skills for explicit review, but it creates a new Run rather than mutating Copilot into Autopilot.

## Consequences

- Copilot remains advisory by default.
- Autopilot remains an explicitly entered execution mode.
- The UI needs a clear promotion affordance and review step.
- Session-attached Skills must appear as proposed carry-forward attachments during promotion rather than silently becoming Run behavior.
- Session history and Run history remain related but distinct.
