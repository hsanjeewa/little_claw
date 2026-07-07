# ADR 0032: Prompt reference model

## Status

Accepted

## Context

Operators need a lightweight way to ground agent intent in existing files and scripts while staying inside the Command Bar interaction model. Plain text paths are ambiguous, and shell-style attachment behavior would blur the line between agent input and terminal execution.

## Decision

The initial version supports **Prompt References** using `@` syntax in agent input.

A Prompt Reference points to a local file or script and makes that artifact part of the current Session or Run context for agent reasoning.

Each Prompt Reference becomes a separate **Prompt Attachment** rather than being inlined into the natural-language prompt text.

Prompt References are part of product-level agent input, not shell execution syntax.

## Consequences

- Operators can refer to concrete artifacts directly when shaping goals, task context, or agent guidance.
- Multiple referenced artifacts remain individually visible and auditable instead of disappearing into a single expanded prompt string.
- The parser must distinguish Prompt References from Slash Commands and ordinary text.
- Future versions may expand Prompt References beyond local files and scripts without changing the basic interaction model.
