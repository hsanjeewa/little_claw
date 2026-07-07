# ADR 0028: Execution branch approval model

## Status

Accepted

## Context

An Autopilot Run may target multiple hosts with different operating environments. The product needs a practical approval unit that remains understandable to the operator without creating one approval prompt per host.

## Decision

Within one Autopilot Run, planning may create multiple **Execution Branches** based on compatible host capability profiles.

Execution Branch membership is determined **per step** by the minimum capability set required for that step, rather than by every difference in the full host profile.

For mutating work, approval is requested **per branch step**, not per host and not once for the full run branch.

This means a state-changing step may be approved for a capability-compatible host group that shares the same step semantics.

## Consequences

- The operator can review mutating intent at a meaningful unit larger than a single host.
- Approval noise stays lower than per-host prompting while remaining more precise than whole-run approval.
- Branching remains driven by operational relevance instead of fragmenting runs around unrelated host differences.
- The planner and UI must make branch membership and step semantics visible before approval.
