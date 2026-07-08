# ADR 0003: Autopilot approval policy

## Status

Accepted

## Context

Autopilot is intended to be agent-led, but operational safety depends on a clear execution boundary. The product needs a default policy that is useful for real work without allowing surprising state changes.

## Decision

The initial Autopilot approval policy is:

- **Read-only actions may run automatically**.
- **State-changing actions require explicit operator approval**.
- **Approval is requested per state-changing step, not once for the full Run**.

Autopilot may inspect systems, gather context, reason about the problem, and prepare proposed command sequences without prior approval. Any action that changes remote system state must be confirmed by the operator before execution.

## Consequences

- Autopilot remains useful for investigation and planning without becoming unsafe by default.
- The execution layer must classify actions as read-only or state-changing.
- The execution layer must pause at each mutating step boundary until the operator approves or rejects that step.
- The UI must make pending approvals explicit and easy to review.
- Future policy expansion can add environment-specific or role-specific approval rules without changing the core operating model.
