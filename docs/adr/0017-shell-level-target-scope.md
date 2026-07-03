# ADR 0017: Shell-level target scope

## Status

Accepted

## Context

Target Scope applies across Watchtower, Autopilot, and Copilot. The product needs a consistent place for inventory and host selection so scope does not fragment across layers or require each mode to reinvent selection behavior.

## Decision

**Target Scope** is a shell-level primitive.

The shell owns the primary scope model and exposes it to modes. Individual modes may narrow or reinterpret that scope for their local workflow, but they do not redefine the cross-product concept.

In v1, the shell-level scope selector supports:

- **Entire Inventory**
- **Explicit host selection** from the inventory

It does not yet support saved groups or query-defined dynamic groups.

## Consequences

- Cross-layer APIs can depend on one shared scope concept.
- Mode transitions and escalations can carry scope cleanly.
- The shell chrome must expose current scope clearly enough that operators understand what a mode is acting on.
