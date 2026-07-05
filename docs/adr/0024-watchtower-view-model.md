# ADR 0024: Watchtower view model

## Status

Accepted

## Context

Watchtower is being redesigned toward a more btop-like presentation. The redesign introduces richer information density, stronger color semantics, multiple internal Watchtower Views, and keyboard-driven drill paths across fleet and host inspection.

These choices affect navigation, state ownership, scope behavior, and presentation contracts across the TUI, so the Watchtower model needs to be explicit before implementation.

## Decision

Watchtower remains a top-level **Mode**, but it now contains three internal **Watchtower Views**:

- **Host Detail** — a close btop-style per-host dashboard
- **Fleet Aggregate** — a fleet-wide summary using the same visual language as Host Detail
- **Fleet Matrix** — one Metric Family across many hosts as paginated cards

The Watchtower interaction model is:

- Watchtower Views switch through explicit hotkeys rather than cycling:
  - `d` for Host Detail
  - `g` for Fleet Aggregate
  - `m` for Fleet Matrix
- Watchtower also keeps a short internal **Watchtower View History** so operators can step back through drill paths.
- In **Host Detail**, `[` and `]` move to the previous or next scoped host.
- In **Fleet Matrix**, `[` and `]` move to the previous or next page in addition to edge-triggered arrow paging.
- **Host Detail** and **Fleet Aggregate** show all major metric families simultaneously.
- **Host Detail** uses one focused metric module at a time.
- In **Host Detail**, direct family hotkeys (`1`-`4`) move the visible module focus and also retarget refresh and escalation to that focused family.
- **Fleet Matrix** remains the only Watchtower View with an active **Metric Family** selector.
- **Fleet Matrix** paginates **Host Cards**, not metric families.
- Watchtower keeps one global **Selected Host** across all Watchtower Views.
- Watchtower chrome always shows the current **Selected Host**.
- If scope changes remove the current Selected Host, Watchtower re-homes selection to the first visible scoped host.
- If drill history restores a selection that is no longer valid after a scope or visibility change, Watchtower reclamps the selected host and matrix page before rendering.
- **Fleet Aggregate** and **Fleet Matrix** both honor the current shell **Target Scope**.
- **Fleet Aggregate** uses **Aggregate Bundles** that include both roll-up values and outlier visibility, including the named worst or peak host when relevant.
- **Fleet Aggregate** shows a dual severity signal: the main module reflects aggregate state while explicit outlier indication surfaces worst-host pressure.
- **Fleet Matrix** uses one focused card at a time, and the focused card becomes the Selected Host.
- **Fleet Aggregate** uses one focused aggregate module at a time.
- Watchtower keeps a persistent **Fleet Rail** visible across all Watchtower Views so the operator always has a stable navigation spine.
- The Watchtower palette is **severity-driven**, with explicit per-Metric-Family **Severity Thresholds**.
- **Fleet Aggregate** supports drill-down into **Fleet Matrix** for a chosen Metric Family, with both focusable modules and direct family hotkeys.
- Wide-screen **Fleet Aggregate** and **Host Detail** layouts may emphasize CPU more strongly than the other families, as long as all four families remain visible and actionable.
- The focused module in **Host Detail** can also jump directly to **Fleet Matrix** for its Metric Family.
- **Host Detail** and **Fleet Aggregate** include short trend visuals in addition to current snapshot values.
- Watchtower keeps a small in-memory **Trend Window** per relevant surface for those visuals, with a medium rolling depth of roughly 20–30 samples.
- Watchtower exposes a visible view bar/header and context-sensitive footer hotkeys.
- **Host Detail** prefers one-screen completeness. If space is constrained, Watchtower first hides secondary details, then compresses or collapses lower-priority content, and only then falls back to vertical scrolling.
- In small terminals, **Host Detail** and **Fleet Aggregate** keep all four major metric families visible, but lower-priority families such as disk and network may collapse into more compact summaries before CPU and memory do.
- Each Watchtower View resumes its own last internal state when revisited.
- Missing data and stale data are visually distinct states across Watchtower.
- **Fleet Matrix** uses spatial 2D arrow navigation across visible cards, and crossing the visible edge turns pages automatically.
- **Host Detail** and **Fleet Aggregate** also move focus according to spatial layout rather than fixed reading order.
- Watchtower v1 is keyboard-first.
- Watchtower takes a strong homage approach to btop's visual language without becoming a pixel-faithful clone.
- **Fleet Matrix** cards include a small explicit status badge.
- **Host Detail** and **Fleet Aggregate** use selective status badges for notable states such as missing, stale, or critical conditions.
- The redesigned Watchtower is released only when CPU, memory, disk, and network all satisfy this new Watchtower model across Host Detail, Fleet Aggregate, and Fleet Matrix.

## Consequences

- Watchtower state now includes both a current Watchtower View and a global Selected Host.
- The former metric-family-centered model becomes view-specific rather than global.
- Host Detail becomes a dense multi-family dashboard instead of a single-family drill surface.
- Fleet Aggregate becomes an operational triage surface, not just a passive summary.
- Fleet Matrix needs explicit card focus and host-card pagination behavior.
- Host Detail and Fleet Aggregate layouts need explicit density-priority rules for small terminals.
- Trend rendering now depends on retained in-memory history rather than single snapshots alone.
- Shell and Watchtower hotkeys must stay legible because Watchtower now contains multiple drill layers.
- Future metric families must define both presentation anatomy and severity thresholds.
- Release planning must treat the redesigned Watchtower as a cohesive multi-family surface rather than a memory-first slice.
