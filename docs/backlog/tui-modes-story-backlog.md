# Story Backlog: Tabbed TUI Modes

## Epic 1: Shell and Mode Framework

### STORY 1.1 — Create shell root model
**Goal**: Introduce a root Bubble Tea model that owns active mode, shell chrome, and shared state.

**Acceptance Criteria**
- three modes exist in the shell
- one mode is active at a time
- switching modes preserves each mode's internal state
- shell rendering does not corrupt TUI output

### STORY 1.2 — Add tab bar with mode status badges
**Goal**: Render shell tabs for Watchtower, Autopilot, and Copilot with compact background status.

**Acceptance Criteria**
- active tab is visually distinct
- each mode can provide a status badge value
- tab bar remains stable across resize events

### STORY 1.3 — Add contextual hotkey bar
**Goal**: Show contextual hotkeys for the active mode and current focus.

**Acceptance Criteria**
- global navigation keys remain minimal
- mode-local hotkeys are visible
- hotkey bar updates when focus changes

## Epic 2: Shell-Level Target Scope

### STORY 2.1 — Model Target Scope in the shell
**Goal**: Define Entire Inventory and Selected Hosts as shell-owned scope types.

**Acceptance Criteria**
- shell exposes current Target Scope
- modes can read scope from shell state
- scope model is not duplicated per mode

### STORY 2.2 — Build scope selector UI
**Goal**: Let the operator switch between Entire Inventory and explicit host selection.

**Acceptance Criteria**
- operator can choose Entire Inventory
- operator can choose one or more explicit hosts
- current scope is visible in shell chrome

### STORY 2.3 — Protect active work from silent scope mutation
**Goal**: Keep active Runs and Sessions stable when shell scope changes.

**Acceptance Criteria**
- Watchtower updates immediately on shell scope change
- active Autopilot/Copilot work is not silently rewritten
- divergence between shell scope and active-work scope is visible

## Epic 3: Structural Placeholders

### STORY 3.1 — Build Autopilot placeholder layout
**Goal**: Create the real shell-integrated Autopilot layout without full behavior.

**Acceptance Criteria**
- Command Bar frame is visible
- plan pane is visible
- transcript pane is visible
- mode survives tab switching and resize

### STORY 3.2 — Build Copilot placeholder layout
**Goal**: Create the real shell-integrated Copilot layout without full behavior.

**Acceptance Criteria**
- Command Bar frame is visible
- terminal pane frame is visible
- advisory pane frame is visible
- guidance controls are visible

## Epic 4: Watchtower Memory Vertical Slice

### STORY 4.1 — Define Metrics Snapshot domain model
**Goal**: Add a domain-level snapshot model for Watchtower metrics.

**Acceptance Criteria**
- snapshot model represents host identity, metric payload, freshness, and host status
- domain layer stays dependency-free
- tests cover validation and normalization semantics

### STORY 4.2 — Implement SSH memory collector
**Goal**: Collect memory metrics from target hosts using centralized SSH polling.

**Acceptance Criteria**
- collector can poll hosts from current inventory
- normalized memory data is returned per host
- host collection failures are preserved per host

### STORY 4.3 — Add refresh cycle orchestration
**Goal**: Periodically refresh Watchtower snapshots with partial-success handling.

**Acceptance Criteria**
- refresh cycle updates successful hosts even when some fail
- each host result includes freshness metadata
- collection interval is controllable by config or mode state

### STORY 4.4 — Render Fleet Matrix for memory
**Goal**: Show memory usage across the current Target Scope.

**Acceptance Criteria**
- Fleet Matrix honors Entire Inventory and Selected Hosts
- host rows/cards are selectable
- failed or stale hosts are visually distinguishable

### STORY 4.5 — Render Host Detail for memory
**Goal**: Drill from Fleet Matrix into single-host memory detail.

**Acceptance Criteria**
- selecting a host opens Host Detail
- a back action returns to Fleet Matrix
- host freshness and failure state remain visible

## Epic 5: Watchtower-to-Action Hooks

### STORY 5.1 — Define escalation payload
**Goal**: Create the cross-mode payload for Watchtower escalation.

**Acceptance Criteria**
- payload includes selected host or scope
- payload includes current metric family
- payload includes recent observations summary

### STORY 5.2 — Add escalation actions in Watchtower UI
**Goal**: Let the operator explicitly escalate to Copilot or Autopilot.

**Acceptance Criteria**
- UI exposes both escalation targets
- escalation does not auto-execute anything
- target mode receives the payload or stub equivalent

## Epic 6: CPU Follow-Up

### STORY 6.1 — Add CPU metric family to collection model
**Goal**: Extend Watchtower from memory-first to CPU-next.

**Acceptance Criteria**
- CPU data is collected and normalized per host
- memory implementation remains unchanged for existing flows

### STORY 6.2 — Add metric-family switching in Watchtower
**Goal**: Support switching between memory and CPU in Fleet Matrix and Host Detail.

**Acceptance Criteria**
- current metric family is visible
- switching updates the matrix without breaking focus/navigation
- contextual hotkeys reflect family switching actions

## Suggested Order

1. STORY 1.1
2. STORY 1.2
3. STORY 1.3
4. STORY 2.1
5. STORY 2.2
6. STORY 3.1
7. STORY 3.2
8. STORY 4.1
9. STORY 4.2
10. STORY 4.3
11. STORY 4.4
12. STORY 4.5
13. STORY 5.1
14. STORY 5.2
15. STORY 2.3
16. STORY 6.1
17. STORY 6.2
