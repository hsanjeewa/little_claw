# ADR 0008: Watchtower escalation model

## Status

Accepted

## Context

Watchtower is the monitoring mode, but operators need a fast path from observation to action when they notice anomalies. The product needs a defined relationship between Watchtower and the action-oriented modes.

## Decision

Watchtower will support **explicit escalation** into either:

- a **Copilot Session**, or
- an **Autopilot Run**

Escalation carries forward the relevant context, such as selected host, target scope, active metric family, and recent observations. The operator chooses whether the next step is human-led or agent-led.

## Consequences

- Watchtower remains an observational mode rather than becoming an execution surface.
- Operators get a faster transition from diagnosis to action.
- Cross-mode handoff data must be modeled explicitly instead of being reconstructed from scratch.
