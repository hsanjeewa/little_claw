# ADR 0017: Shell-level target scope

## Status

Accepted

## Context

Target Scope applies across Watchtower, Autopilot, and Copilot. The product needs a consistent place for inventory and host selection so scope does not fragment across layers or require each mode to reinvent selection behavior.

## Decision

**Target Scope** is a shell-level primitive.

The shell owns the primary scope model and exposes it to modes. Individual modes may narrow or reinterpret that scope for their local workflow, but they do not redefine the cross-product concept.

Natural-language agent input may include a **Scope Hint** that restates or narrows the intended target set for the current interaction, but shell-level Target Scope remains canonical. If agent input conflicts with the current shell scope, the product must surface that conflict for explicit operator review instead of silently retargeting work.

In v1, Scope Hints may refer only to:

- the **Entire Inventory**
- the current **Selected Hosts**
- explicit host names already known in the inventory

They do not support saved groups, role labels, or query-defined host sets.

In v1, the shell-level scope selector supports:

- **Entire Inventory**
- **Explicit host selection** from the inventory

It does not yet support saved groups or query-defined dynamic groups.

## Consequences

- Cross-layer APIs can depend on one shared scope concept.
- Mode transitions and escalations can carry scope cleanly.
- Prompt-driven targeting can help shape intent without creating a second hidden source of truth for scope.
- The shell chrome must expose current scope clearly enough that operators understand what a mode is acting on.
