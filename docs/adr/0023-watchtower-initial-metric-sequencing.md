# ADR 0023: Watchtower initial metric family sequencing

## Status

Superseded by ADR 0025

## Context

Watchtower Fleet Matrix uses one Metric Family at a time. The first implemented family sets the initial collection contract, presentation pattern, and Host Detail drill-down shape.

## Decision

The initial Watchtower metric family sequence is:

1. **Memory first**
2. **CPU second**

Memory is the first production metric family for the Watchtower slice. CPU follows as the immediate next family after the memory path is established.

## Consequences

- The first slice stays tighter while still defining a clear short-term expansion path.
- Memory becomes the reference implementation for metric collection, freshness, Fleet Matrix rendering, and Host Detail drill-down.
- CPU can be added next using the same shell and Watchtower interaction model.
