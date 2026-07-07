# ADR 0026: Scheduled Runbook approval model

## Status

Accepted

## Context

The product needs CRON-like scheduled work in addition to interactive Autopilot runs. A purely interactive approval model does not fit scheduled execution, because some jobs must run unattended while others must wait for a human decision when they fire.

## Decision

The initial scheduled-work model supports two **Schedule Approval Modes**:

- **Pre-approved**: the operator approves the Scheduled Runbook definition when creating or updating it, and future scheduled executions may run unattended according to that definition.
- **Approval-gated**: the schedule may create a new run when it fires, but execution waits for operator approval before any mutating work proceeds.

These modes apply to **Scheduled Runbooks**, which are separate from ad hoc interactive Autopilot Runs even if they share execution machinery.

## Consequences

- The product can support unattended recurring maintenance without weakening the supervised Autopilot model.
- The scheduler must persist both the runbook definition and its approval mode.
- The UI must clearly distinguish between a scheduled run that is executing unattended and a scheduled run that is waiting for approval.
- Future versions may add stricter environment-specific policy without changing the core distinction between pre-approved and approval-gated schedules.
