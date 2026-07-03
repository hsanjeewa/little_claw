# ADR 0007: Fleet Matrix metric family model

## Status

Accepted

## Context

Watchtower's Fleet Matrix must compare server health across the inventory without becoming visually overloaded. The product needs a consistent way to decide whether the matrix shows one metric, several fixed metrics, or an operator-defined bundle.

## Decision

Fleet Matrix will show **one Metric Family at a time**, with operator-controlled switching between families such as CPU, memory, disk, and network.

The interaction model should feel similar to `btop`: focused on one resource family at a time while making it easy to switch to another.

## Consequences

- Fleet-wide comparisons remain readable even with larger inventories.
- Watchtower needs a clear metric-family switcher and a stable normalized model for each family.
- The UI avoids the density and layout complexity of showing too many unrelated metrics at once.
