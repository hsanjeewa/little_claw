# PRD: Tabbed TUI Modes

## Summary

Expand the DevOps Agent TUI into a shell with three independent modes presented as tabs:

1. **Watchtower** — monitoring and fleet visibility
2. **Autopilot** — agent-led execution with approvals
3. **Copilot** — human-led terminal workflow with advisory assistance

The shell should feel similar in spirit to Zellij for navigation and contextual hotkeys. Watchtower is the first mode to receive end-to-end functionality; Autopilot and Copilot ship first as structural placeholders. The current Watchtower target is the redesigned three-View experience defined by the Watchtower redesign ADRs and spec.

## Problem

The current TUI does not yet provide a coherent multi-mode operator workflow for:

- fleet-wide monitoring
- agent-led operational execution
- human-led incident/CLI assistance

Without a shell-level model, these capabilities risk becoming disconnected features instead of one product.

## Goals

- Introduce a shell with three long-lived independent modes
- Preserve mode state while switching tabs
- Add shell-level Target Scope shared across modes
- Ship a cohesive redesigned Watchtower with real metric collection and cross-view drill flow
- Establish Autopilot and Copilot layout/focus/input contracts early
- Support contextual hotkeys and slash commands where appropriate

## Non-Goals

- Full multi-run Autopilot concurrency in v1
- Full multi-session Copilot concurrency in v1
- Resident per-host metrics agents in v1
- Dynamic host groups, saved groups, or label-query scopes in v1
- Copilot observation of arbitrary external local terminals
- Slash commands as shell aliases

## Users

- DevOps engineers
- Sysadmins
- Operators managing one or more remote servers over SSH

## Product Model

### Shell

- Tabs are the UI presentation; the canonical model is **independent modes**
- Modes persist in the background when unfocused
- Shell tab bar shows **Status Badges**
- Shell shows **Contextual Hotkeys** for the active mode/focus
- Shell owns **Target Scope**

### Target Scope

- Shell-level primitive
- v1 supports:
  - **Entire Inventory**
  - **Selected Hosts**
- Watchtower honors shell Target Scope directly
- Active Autopilot/Copilot work does not silently mutate when shell scope changes

## Mode Requirements

### 1. Watchtower

Purpose: monitoring mode for fleet and host health.

#### Views

- **Fleet Aggregate**
  - default landing Watchtower View
  - shows all major metric families together using fleet-wide **Aggregate Bundles**
  - acts as the primary triage surface
- **Fleet Matrix**
  - compares one **Metric Family** across the current Target Scope
  - renders scoped hosts as paginated cards
  - remains the only Watchtower View with an active Metric Family selector
- **Host Detail**
  - deep single-host dashboard with a strong btop homage
  - shows CPU, memory, disk, and network simultaneously

#### Navigation

- Direct Watchtower View hotkeys:
  - `g` → Fleet Aggregate
  - `m` → Fleet Matrix
  - `d` → Host Detail
- Watchtower maintains a short internal drill history with `b` as local back action
- Fleet Aggregate drills into Fleet Matrix by focused module or direct family hotkey
- Fleet Matrix focus determines the global **Selected Host**
- Host Detail can drill into Fleet Matrix for the focused family and switch scoped hosts directly with `[` / `]`

#### Data Model

- Metrics are collected through **central SSH polling**
- Data is represented as **Metrics Snapshots**
- Watchtower retains a short in-memory **Trend Window** for trend visuals
- Each snapshot shows **Freshness**:
  - timestamp
  - host collection status
  - age
- Watchtower distinguishes:
  - healthy
  - elevated
  - critical
  - stale
  - missing
- Refresh Cycles use **partial success**
  - successful hosts render normally
  - failed/unreachable hosts remain visible as failed

#### Presentation Rules

- Palette is severity-driven rather than decorative
- Thresholds are fixed built-in thresholds per Metric Family in v1
- Host Detail and Fleet Aggregate use bar-first modules with supporting numbers
- Fleet Matrix uses mini bar-first cards with explicit status badges
- Host Detail prefers one-screen completeness and degrades density before scrolling
- Disk and network may collapse earlier than CPU and memory in smaller terminals

#### Escalation

Watchtower supports **explicit escalation** into:

- Copilot Session
- Autopilot Run

Escalation carries current context such as selected host, scope, metric family, and recent observations.

#### Release Gate

- The redesigned Watchtower is not release-ready until **CPU**, **memory**, **disk**, and **network** are all present across:
  - Host Detail
  - Fleet Aggregate
  - Fleet Matrix
- Release requires complete cross-view family coverage, but not identical richness in every Watchtower View

### 2. Autopilot

Purpose: agent-led execution mode for DevOps/SysAdmin work.

#### Core Model

- Primary unit: **Run**
- Run starts from a natural-language **Goal** entered via **Command Bar**
- Goal may be refined within the same Run if the objective is materially the same
- Materially different objective starts a new Run

#### Approval Policy

- Read-only actions may execute automatically
- State-changing actions require explicit operator approval

#### Concurrency

- v1 supports one active Run at a time
- prior Runs remain in resumable history

#### Run Lifecycle

- Drafting
- Ready
- Executing
- Blocked
- Completed
- Failed

Background behavior:

- Drafting and Executing may continue when unfocused
- Ready and Blocked wait

#### UI Model

- Plan-first presentation with transcript alongside
- Slash commands supported as in-app control actions only
- Unified Command Bar for natural language + slash commands

### 3. Copilot

Purpose: human-led terminal workflow with agent assistance.

#### Core Model

- Primary unit: **Session**
- Session starts from **Task Context**
- Task Context may emerge lazily from terminal activity and be corrected later

#### Observation Boundary

- Copilot observes only the in-app **Terminal Session** inside the TUI
- It does not observe arbitrary external terminal activity

#### Assistance Scope (v1)

Copilot may:

1. Explain
2. Warn
3. Suggest
4. Summarize

Copilot does not take control by default.

#### Guidance UX

- Guidance appears in an advisory pane, not inline in the terminal stream
- Guidance can be enabled/disabled by **Guidance Preference**
- Preference model:
  - global default
  - per-Session override

#### Concurrency

- v1 supports one active Session at a time
- prior Sessions remain resumable in history

#### Handoff

- explicit **Promotion** from Copilot Session to new Autopilot Run

## Cross-Mode Interaction Rules

- Watchtower follows shell Target Scope live
- Autopilot/Copilot do not silently adopt shell scope changes during active work
- Promotion and Escalation are explicit operator actions
- Slash commands are product actions, not shell aliases

## Initial Delivery Sequence

1. Build shell skeleton for all three modes
2. Ship structural placeholders for Autopilot and Copilot
3. Prioritize Watchtower as first real end-to-end feature
4. Treat the redesigned Watchtower as a cohesive multi-family surface rather than a memory-first tracer bullet

### Placeholder Scope

- **Autopilot placeholder**
  - mode chrome
  - Command Bar
  - empty plan/transcript layout
- **Copilot placeholder**
  - mode chrome
  - Command Bar
  - terminal pane frame
  - advisory pane frame
  - visible guidance controls

## Success Criteria

- Operator can switch among three persistent modes in one shell
- Shell Target Scope is visible and usable across modes
- Watchtower exposes Fleet Aggregate, Fleet Matrix, and Host Detail as coherent Watchtower Views
- Watchtower displays CPU, memory, disk, and network across the redesigned surface
- Watchtower supports drill flow across Fleet Aggregate, Fleet Matrix, and Host Detail
- Watchtower surfaces freshness and partial host failures clearly
- Watchtower distinguishes missing versus stale states and preserves selected-host continuity across views
- Shell exposes contextual hotkeys and mode status badges
- Autopilot and Copilot placeholders validate layout and focus model

## References

- `CONTEXT.md`
- `docs/adr/0001-tui-mode-model.md`
- `docs/adr/0002-watchtower-metrics-collection.md`
- `docs/adr/0003-autopilot-approval-policy.md`
- `docs/adr/0004-copilot-observation-boundary.md`
- `docs/adr/0005-session-to-run-promotion.md`
- `docs/adr/0006-target-scope-defaults.md`
- `docs/adr/0007-fleet-matrix-metric-family-model.md`
- `docs/adr/0008-watchtower-escalation-model.md`
- `docs/adr/0009-shell-status-and-hotkeys.md`
- `docs/adr/0010-slash-command-boundary.md`
- `docs/adr/0011-command-bar-model.md`
- `docs/adr/0012-copilot-v1-assistance-scope.md`
- `docs/adr/0013-copilot-guidance-presentation.md`
- `docs/adr/0014-autopilot-presentation-model.md`
- `docs/adr/0015-autopilot-run-lifecycle.md`
- `docs/adr/0016-copilot-session-concurrency.md`
- `docs/adr/0017-shell-level-target-scope.md`
- `docs/adr/0018-watchtower-scope-behavior.md`
- `docs/adr/0019-autopilot-goal-entry.md`
- `docs/adr/0020-copilot-session-entry.md`
- `docs/adr/0021-active-work-scope-change-behavior.md`
- `docs/adr/0022-initial-delivery-sequence.md`
- `docs/adr/0023-watchtower-initial-metric-sequencing.md`
- `docs/adr/0024-watchtower-view-model.md`
- `docs/adr/0025-watchtower-redesign-release-gate.md`
- `docs/plans/watchtower-redesign-spec.md`
