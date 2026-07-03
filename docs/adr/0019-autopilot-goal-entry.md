# ADR 0019: Autopilot goal entry

## Status

Accepted

## Context

Autopilot Runs are objective-driven, but the product needs a clear and simple way for operators to start them. A rigid form would slow down exploratory operational work, while an unstructured model still needs a canonical concept for what anchors a Run.

## Decision

An Autopilot Run starts from a natural-language **Goal** entered through the Command Bar.

Examples include:

- "Investigate high memory on selected hosts"
- "Patch nginx and verify service health on these servers"

The Goal is the canonical operator intent that anchors planning, evidence gathering, approvals, and execution.

Goal changes follow this rule:

- refinement of the same underlying objective stays within the current Run
- a materially different objective starts a new Run

## Consequences

- Starting a Run remains lightweight and terminal-native.
- The planning layer must interpret freeform operator intent into structured operational steps.
- The system must distinguish Goal refinement from Goal replacement.
- Future structured enrichment can be added without replacing the Goal as the primary entry concept.
