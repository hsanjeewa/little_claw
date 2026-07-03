# ADR 0018: Watchtower target scope behavior

## Status

Accepted

## Context

Watchtower's Fleet Matrix is fleet-oriented, but the product also introduces a shell-level Target Scope that may represent either the Entire Inventory or Selected Hosts. The relationship between these concepts must be explicit.

## Decision

Watchtower honors the shell **Target Scope** directly.

- If the Target Scope is **Entire Inventory**, Fleet Matrix compares all hosts.
- If the Target Scope is **Selected Hosts**, Fleet Matrix compares only that selection.

## Consequences

- Scope behavior remains consistent across modes.
- Host selection becomes useful before escalation or deep inspection.
- Watchtower stays fleet-oriented without forcing a separate scope model.
