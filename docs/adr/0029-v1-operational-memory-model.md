# ADR 0029: v1 operational memory model

## Status

Accepted

## Context

The product needs memory-like features comparable to agent systems, but broad open-ended agent memory would add ambiguity, retrieval complexity, and unclear safety boundaries to the first version.

## Decision

The initial memory model is limited to operationally explicit persisted structures:

- **Session Memory** for resumable Copilot working context
- **Run History** for prior Autopilot plans, approvals, and outcomes
- **Host Capability Profile** for discovered host environment facts
- **Scheduled Runbooks** for saved recurring operational procedures

The initial version does **not** include a general-purpose long-term agent memory layer.

## Consequences

- Memory remains auditable and closely tied to operator work.
- Retrieval behavior can stay bounded to explicit product concepts instead of open-ended semantic recall.
- Future versions may add broader memory capabilities without redefining the initial operational records.
