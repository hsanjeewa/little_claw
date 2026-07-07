# ADR 0031: Skill attachment model

## Status

Accepted

## Context

Skills are first-class in the product, but silent planner attachment would weaken operator awareness and make behavior harder to review in a supervised system.

## Decision

The initial version supports two ways to attach a **Skill**:

- the operator explicitly chooses a Skill
- the agent suggests a Skill and the operator confirms it

The initial version does **not** allow silent automatic Skill attachment.

## Consequences

- Skill usage remains visible and reviewable.
- The planner may still guide the operator toward relevant skills without hiding that choice.
- Skills remain compatible with the supervised Autopilot and Copilot models.
