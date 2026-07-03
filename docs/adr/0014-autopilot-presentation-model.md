# ADR 0014: Autopilot presentation model

## Status

Accepted

## Context

Autopilot is a Run-oriented, agent-led mode inspired by products like Crush. Operators need both visibility into what the agent is doing and a practical structure for approvals, progress tracking, and recovery.

## Decision

Autopilot will use a **plan-first presentation model with transcript alongside**.

The plan is the primary operational structure. The transcript remains available as supporting evidence, reasoning trace, and execution context.

For v1, Autopilot manages one active Run at a time, while preserving prior Runs as resumable history.

## Consequences

- Approvals can attach naturally to plan steps.
- Operators can understand progress without reading the full transcript.
- The UI must preserve a useful relationship between structured plan state and the underlying interaction trace.
