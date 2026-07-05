# Watchtower Redesign Spec

## Purpose

Define the concrete interaction, layout, navigation, and presentation rules for the redesigned Watchtower experience.

This document is implementation-facing. It translates the accepted Watchtower glossary and ADR decisions into a build-ready product spec.

## References

- `CONTEXT.md`
- `docs/adr/0018-watchtower-scope-behavior.md`
- `docs/adr/0024-watchtower-view-model.md`
- `docs/adr/0025-watchtower-redesign-release-gate.md`

## Product Summary

Watchtower remains a top-level TUI **Mode** for monitoring and fleet visibility.

Inside Watchtower, the redesign introduces three internal **Watchtower Views**:

1. **Host Detail** — close btop-style single-host dashboard
2. **Fleet Aggregate** — compressed fleet-wide summary using the same visual language
3. **Fleet Matrix** — one **Metric Family** shown across scoped hosts as paginated cards

The redesign is keyboard-first, strongly inspired by btop’s density and energy, but is not intended to be a pixel-faithful clone.

## Core State Model

Watchtower state must include:

- current **Watchtower View**
- current shell **Target Scope** projection
- global **Selected Host**
- per-view remembered internal state
- short bounded **Watchtower View History** for drill-back
- in-memory **Trend Window** data for short trend visuals

### Global Rules

- **Selected Host** is global across all Watchtower Views.
- Watchtower chrome always shows the current Selected Host.
- If Target Scope changes and the Selected Host is no longer visible, Watchtower re-homes to the first visible scoped host.
- Per-view state is resumed when revisiting a Watchtower View.
- Watchtower history tracks drill transitions only, not manual `d/g/m` switches, host switching, page switching, or local focus movement.
- Watchtower history restores full prior view state on back.
- Watchtower history uses bounded depth and best-effort restore if scope changes invalidate parts of historical state.

## View Model

### 1. Host Detail

Purpose: inspect one host deeply.

#### Content

- show all four major metric families simultaneously:
  - CPU
  - Memory
  - Disk
  - Network
- use a close btop-style module layout
- prefer bars as the primary visual anchor
- include supporting numeric detail
- include short subdued trend strips

#### Focus

- one metric module is focused at all times
- focus follows spatial layout, not just reading order
- focused module is actionable, not decorative

#### Actions

- `Enter` on focused module drills into **Fleet Matrix** for that module’s Metric Family
- direct family hotkeys move focus to that module and retarget refresh/escalation to the focused family
- `[` / `]` switch to previous/next scoped host

#### Density Rules

- prefer one-screen completeness
- if constrained:
  1. hide secondary details
  2. compress or collapse lower-priority content
  3. collapse disk/network before CPU/memory when needed
  4. only then fall back to vertical scroll

### 2. Fleet Aggregate

Purpose: provide a fast fleet-wide operational summary and triage entry point.

#### Content

- show all four major metric families simultaneously
- use the same visual language as Host Detail, but in a compressed sibling layout
- preserve a persistent **Fleet Rail** beside the aggregate canvas so the operator can keep host context while triaging fleet-wide pressure
- each family renders as an **Aggregate Bundle**
- each Aggregate Bundle includes:
  - fleet-level roll-up
  - outlier visibility
  - named worst/peak host when relevant
  - short subdued trend strip

#### Severity

- use dual signal:
  - main module reflects aggregate state
  - explicit outlier indication surfaces worst-host pressure

#### Focus

- one aggregate module is focused at all times
- focus follows spatial layout
- wide layouts may lead with CPU and place memory, storage, and network into supporting regions instead of a strict equal-card grid

#### Actions

- `Enter` on focused module drills into **Fleet Matrix** for that family
- direct family hotkeys set the active family before drill-down

### 3. Fleet Matrix

Purpose: compare one Metric Family across the current Target Scope.

#### Content

- active family only
- one card per visible scoped host
- cards use a consistent outer shell and family-specific internal anatomy
- cards are mini bar-first, with supporting numbers
- the main matrix canvas pairs a compact family summary strip with a larger focused-host inspection region
- cards include:
  - current value emphasis
  - subdued short trend strip
  - small explicit status badge

#### Scope and Selection

- operates on the current Target Scope only
- single-focus only in v1
- no in-matrix multi-select in v1
- no in-matrix filtering in v1
- focused card becomes the global Selected Host

#### Pagination and Navigation

- pagination is over host cards only
- Metric Family selection is separate from pagination
- support 2D spatial navigation between visible cards
- if arrow navigation crosses the visible grid edge, page automatically
- `[` / `]` explicitly move previous/next page
- drilling into Fleet Matrix should land on the page containing the current Selected Host and focus that host’s card

## View Switching and Drill Flow

### Direct Watchtower View Switching

- `d` → Host Detail
- `g` → Fleet Aggregate
- `m` → Fleet Matrix

These direct switches do not create Watchtower history entries.

### Back

- `b` → go back through **Watchtower View History**
- back is local to Watchtower, not a shell leader sequence
- back is intended to undo drill flow, not arbitrary local state changes
- if history restore would land on an out-of-scope or invisible host/page, Watchtower reclamps selection before rendering

### Drill Precedence Rules

- if a family hotkey is used, the explicit family wins
- if `Enter` is used on a focused module, the focused module’s family wins
- if the chosen family has missing data, navigation still occurs and the destination shows the missing-data state

### Intended Triage Flow

Primary flow:

1. **Fleet Aggregate**
2. choose hot family
3. **Fleet Matrix**
4. choose/focus host
5. **Host Detail**

## Default Entry Behavior

- default Watchtower landing view: **Fleet Aggregate**
- first entry default family focus: **Memory**
- subsequent revisits restore prior per-view state

## Visual System

### Palette

- palette is **severity-driven**, not merely decorative
- v1 uses fixed built-in **Severity Thresholds**
- thresholds are uniform per Metric Family across the fleet
- thresholds should be discoverable in contextual help for the focused family

### Trend Visuals

- maintain a medium rolling **Trend Window** of roughly 20–30 samples
- trends remain subdued/neutral
- current bars, values, and badges carry the main severity signal

### Status States

Watchtower must visually distinguish:

- healthy
- elevated
- critical
- stale
- missing

Rules:

- missing and stale are distinct
- Fleet Matrix cards always show explicit status badges
- Host Detail and Fleet Aggregate use selective badges for notable states such as missing, stale, or critical conditions

## Layout Rules

### Host Detail and Fleet Aggregate

- keep all four major families visible even in smaller terminals if possible
- allow disk/network to collapse earlier than CPU/memory
- spatial focus should continue to map to visible layout
- prefer recomposition over truncation: shrink chrome first, then compress secondary details, then collapse lower-priority regions
- keep the persistent **Fleet Rail** narrower and denser than the main content panes so it behaves as a navigation spine rather than a peer dashboard

### Fleet Matrix

- render as cards in a responsive grid when space allows
- when fewer cards fit, rely on host-card pagination rather than switching presentation model

## Release Gate

This redesign is not release-ready until all four major metric families are present across the redesigned Watchtower surface:

- CPU
- Memory
- Disk
- Network

This applies across:

- Host Detail
- Fleet Aggregate
- Fleet Matrix

Release requires complete cross-view family coverage, but not strict identical richness in every view. View-aware completeness is acceptable.

## v1 Non-Goals

- in-matrix multi-select
- in-matrix filtering/slicing
- operator-configurable thresholds
- mouse-first interaction model
- exact pixel-faithful btop reproduction
- requirement for identical family density in every Watchtower View

## Implementation Notes

- Existing Watchtower conventions for scope and shell integration should be preserved where still compatible.
- Existing left/right host navigation should not remain the primary Host Detail movement model once module focus becomes spatial; host switching belongs to `[` / `]`.
- Any old globally active Metric Family concept should be narrowed to Fleet Matrix only.
- Refresh/freshness plumbing must support both global visibility and local stale signaling.

## Acceptance Checklist

- three Watchtower Views exist and are directly switchable with `d/g/m`
- Watchtower opens on Fleet Aggregate
- Host Detail shows CPU/memory/disk/network at once
- Fleet Aggregate shows compressed Aggregate Bundles for all four families
- Fleet Matrix shows one family at a time as paginated cards
- global Selected Host is visible in chrome and preserved across views
- Host Detail supports `[` / `]` host switching
- Fleet Matrix supports 2D focus, edge-triggered page turns, and `[` / `]` page changes
- `b` restores prior drill state through Watchtower View History
- missing and stale are visually distinct
- trend strips use retained in-memory history
- severity thresholds are fixed, per-family, and discoverable in contextual help
- redesign is not considered release-ready until CPU/memory/disk/network all exist across the redesigned Watchtower surface
