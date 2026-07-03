# ADR 0022: Initial delivery sequence

## Status

Accepted

## Context

The product direction now includes three independent modes in one shell: Watchtower, Autopilot, and Copilot. The first delivery needs to validate the shell architecture across all three modes without delaying the first meaningful end-user capability.

## Decision

The initial delivery sequence is:

1. Build the **shell skeleton for all three modes**.
2. Provide **minimal placeholders** for Autopilot and Copilot.
3. Give **implementation priority to Watchtower** as the first mode with real end-to-end functionality.

The initial placeholders for Autopilot and Copilot are **structural placeholders** rather than static splash screens.

- **Autopilot placeholder** includes mode chrome, Command Bar, and empty plan/transcript layout.
- **Copilot placeholder** includes mode chrome, Command Bar, terminal pane frame, advisory pane frame, and visible guidance controls.

This means the product shape is visible early across all three modes, while Watchtower becomes the first fully useful vertical slice.

## Consequences

- The shell, tab model, Target Scope, and mode persistence can be validated early across the whole product.
- Autopilot and Copilot can establish navigation and state boundaries before their deeper workflows are implemented.
- Layout, focus, hotkey presentation, and shell chrome can be exercised before full agent behavior exists.
- Watchtower becomes the first production-quality workflow and the reference slice for cross-mode escalation later.
