# ADR 0025: Watchtower redesign release gate

## Status

Accepted

## Context

ADR 0023 established an initial Watchtower delivery sequence of memory first, then CPU. The Watchtower redesign now introduces a tightly connected three-view model, shared visual language, severity semantics, trend windows, cross-view drill paths, and richer per-family module anatomy.

Releasing the redesigned Watchtower with only one family fully realized would create an uneven operator experience and weaken the coherence of the new presentation model.

## Decision

This decision supersedes ADR 0023 for the redesigned Watchtower release.

The redesigned Watchtower will not be considered release-ready until all major metric families required by the new Watchtower experience are implemented to the new quality bar.

For the redesign, the expected major metric families are:

- **CPU**
- **Memory**
- **Disk**
- **Network**

The release gate applies to the integrated redesign across:

- **Host Detail**
- **Fleet Aggregate**
- **Fleet Matrix**

All four major metric families must be present across those Watchtower Views at release, but they do not need identical richness in every view. View-specific anatomy and intentionally lighter presentation in some views are acceptable as long as the overall Watchtower experience remains coherent.

## Consequences

- The redesigned Watchtower ships as a cohesive product surface rather than a memory-first tracer bullet.
- Implementation planning becomes broader because all major metric families must satisfy the new presentation and interaction contracts.
- Release quality depends on complete cross-view family coverage, not strict identical module density everywhere.
- Intermediate development states may exist behind incomplete work, but they do not qualify as the redesign release target.
- Metric-family consistency becomes part of release quality, not just future follow-up work.
