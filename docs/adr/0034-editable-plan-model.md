# ADR 0034: Editable plan model

## Status

Accepted

## Context

Operators need to trust the Autopilot's plan, but a rigid, non-editable plan prevents them from adjusting operational details, disabling risky steps, or tightening tool parameters before approval. The product needs a transparent, hierarchical plan structure that supports deliberate operator refinement.

## Decision

The initial Autopilot Run uses a **hierarchical, node-based Plan View**.

- The operator sees the full structure: Goal, Execution Branches, and the individual steps within each branch.
- Before approval, the operator may perform **Plan Amendments**: explicitly disabling a step, or adjusting tool parameters.
- Plan Amendment does not involve regenerating the entire plan; it is a direct adjustment of the planned node structure.

## Consequences

- Operator trust increases when they have direct control over operational details before execution.
- The planning layer and UI must support persistence of operator amendments alongside agent-generated intent.
- Future versions may support more complex plan edits without changing the hierarchical node structure.
