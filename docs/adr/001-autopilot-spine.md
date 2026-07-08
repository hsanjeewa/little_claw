# ADR 001: Autopilot Run Spine

## Status
Proposed

## Context
Issue #16 requires a "spine" for Autopilot, enabling the sequential execution of tasks instead of the current individual `RunTask` approach.

## Decision
We will introduce an `AutopilotPlan` domain entity to manage task sequences and an `AutopilotRunner` to orchestrate execution.

## Consequences
- Requires new domain entities: `AutopilotPlan` and `TaskState`.
- `AutopilotRunner` will depend on `Engine` for `RunTask`.
- State tracking will rely on `AuditRepository`.
