# ADR 0013: Copilot guidance presentation and preference control

## Status

Accepted

## Context

Copilot needs to remain visible enough to help without polluting the terminal stream or forcing one interaction style on every operator. The product needs both a presentation model and an operator control over that guidance.

## Decision

Copilot guidance will appear in a dedicated advisory pane rather than inline in the terminal stream.

Operators will also have a **Guidance Preference** that enables or disables Copilot guidance according to their working style.

The preference model is:

- a global default set by the operator
- an optional per-Session override

## Consequences

- The terminal remains a clean execution surface.
- Copilot assistance stays visible but non-invasive when enabled.
- The UI and session model must remember operator preference and react to guidance being turned on or off.
