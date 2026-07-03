# ADR 0002: Watchtower metrics collection model

## Status

Accepted

## Context

Watchtower needs to present fleet-wide and host-specific health data inside the TUI. The product needs an initial collection model that works with the current architecture and avoids unnecessary operational overhead.

## Decision

The first version of Watchtower will collect metrics through **central SSH polling** from the app/engine.

The system will periodically connect to each server in the inventory, collect a normalized metrics snapshot, and refresh the Watchtower views from those snapshots.

Each snapshot must carry visible freshness information, including collection timestamp, host collection status, and age.

Refresh Cycles use a partial-success model: hosts with successful collection render normally, while failed or unreachable hosts are shown explicitly without blocking the overall view.

We are explicitly not introducing a resident collector agent on each server for the initial version.

## Consequences

- Watchtower can ship without deploying extra software to managed hosts.
- The collection layer must normalize heterogeneous command output into a stable snapshot model.
- Freshness is bounded by the refresh interval and SSH round-trip cost.
- The UI must surface per-host freshness so operators can judge whether data is current or stale.
- The system must represent host-level collection failure as first-class state rather than collapsing the entire cycle into one error.
- Large inventories may require concurrency limits, degraded modes, or future architectural evolution.
