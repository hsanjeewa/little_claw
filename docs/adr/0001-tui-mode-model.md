# ADR 0001: TUI mode model

## Status

Accepted

## Context

The TUI is being expanded beyond a single workflow into three top-level product areas: Watchtower, Autopilot, and Copilot. The product needs a stable mental model for navigation, state ownership, and long-running behavior before UI implementation begins.

## Decision

The TUI will present these areas as tabs in the shell, but the canonical domain model is **independent modes**, not lightweight views.

Each mode owns its own purpose and state boundary:

- **Watchtower** is the monitoring mode.
- **Autopilot** is the agent-led execution mode.
- **Copilot** is the human-led assistance mode.

Switching between tabs does not stop the underlying mode. Modes continue running in the background while unfocused, and the shell should expose their status.

Execution authority is explicitly separated:

- **Autopilot** may plan and execute operational work according to an approval policy.
- **Copilot** advises the operator but does not take control by default.

## Consequences

- Navigation must preserve per-mode state instead of rebuilding screens on every switch.
- The app shell needs visible background-status indicators for long-running modes.
- Agent sessions and monitoring refresh loops must be modeled as resumable, long-lived mode processes.
- Copilot and Autopilot must remain distinct in both UI language and permission model.
