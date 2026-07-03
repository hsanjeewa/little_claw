# ADR 0006: Target scope defaults

## Status

Accepted

## Context

Multiple modes need a shared way to describe whether a view or action applies to all managed servers or only a subset. The existing phrase "split mode" is ambiguous because it mixes layout concerns with execution/view scope.

## Decision

The product will use **Target Scope** as the canonical concept for server applicability.

The initial target scopes are:

- **Entire Inventory**
- **Selected Hosts**

Default behavior:

- Watchtower enters with **Entire Inventory** as the default scope.
- Action-oriented modes may require the operator to narrow scope before state-changing execution.

## Consequences

- UI and domain language become clearer by separating scope from layout.
- Cross-layer APIs can model scope explicitly instead of relying on ad hoc selection flags.
- Risky operations can enforce narrower scope without changing Watchtower's broad observability model.
