# ADR 0011: Unified command bar model

## Status

Accepted

## Context

Autopilot and Copilot both need a text-driven interaction surface for agent intent and slash-command control. Copilot also embeds a terminal experience, so the product needs a clean separation between application control input and shell execution input.

## Decision

Autopilot and Copilot will use a **Command Bar** as the unified input for product-level interaction.

- Plain text in the Command Bar expresses natural-language intent to the agent.
- Slash-prefixed input invokes Slash Commands.
- `@` references in the Command Bar attach local files or scripts as structured Prompt References for the current interaction.
- In Copilot, ordinary shell commands are entered in the terminal pane, not in the Command Bar.

## Consequences

- Product control remains consistent across the agent-driven modes.
- Operators can ground agent intent in concrete local artifacts without collapsing the Command Bar into raw shell input.
- Terminal execution remains a distinct interaction surface instead of being mixed into agent intent input.
- Focus management becomes an important shell concern because users must be able to tell whether they are typing into the Command Bar or the terminal pane.
