# Implementation Plan: Tabbed TUI Modes

## Objective

Deliver a tabbed TUI shell with three persistent modes, using the redesigned Watchtower as the first fully realized product surface and structural placeholders for Autopilot and Copilot.

## Delivery Strategy

- Build the shell and shared seams first
- Validate cross-mode state, scope, and navigation early
- Implement the redesigned Watchtower architecture before polishing family-specific presentation details
- Treat Watchtower as a cohesive multi-family release target, not a memory-first tracer bullet
- Keep agent modes structurally real but behaviorally shallow until shell patterns settle

## Phase 1: Shell Foundation

### Outcomes

- TUI shell exists with three tabs/modes
- tab switching preserves per-mode state
- global hotkeys are minimal and stable
- contextual hotkeys are rendered by active mode/focus
- shell-level Target Scope is visible
- status badge plumbing exists

### Work

1. Introduce shell/root model that owns:
   - active mode
   - mode registry/state containers
   - Target Scope
   - shell chrome rendering
2. Define mode interface/contract for:
   - Init
   - Update
   - View
   - hotkey surface
   - status badge surface
3. Implement tab bar and mode switching
4. Implement contextual hotkey bar
5. Add shell-level Target Scope model:
   - Entire Inventory
   - Selected Hosts
6. Ensure background persistence semantics at mode state level

### Validation

- mode switching does not reset placeholder state
- tab bar updates active mode correctly
- hotkey bar changes with mode/focus
- shell renders current Target Scope consistently

## Phase 2: Structural Placeholders

### Outcomes

- all three modes are navigable
- Autopilot and Copilot layouts exist
- focus and shell integration are proven without full behavior

### Work

1. Watchtower placeholder upgraded into real working area shell
2. Autopilot structural placeholder:
   - Command Bar frame
   - plan panel frame
   - transcript panel frame
3. Copilot structural placeholder:
   - Command Bar frame
   - terminal pane frame
   - advisory pane frame
   - guidance toggle control frame
4. Add status badge states for stub modes

### Validation

- placeholder modes are visually stable under window resize
- focus indicators are clear
- no stdout/stderr corruption of Bubble Tea UI

## Phase 3: Watchtower Redesign Foundation

### Outcomes

- Watchtower state model supports three Watchtower Views
- global Selected Host behavior is explicit
- drill history, per-view state restore, and Target Scope integration are modeled before presentation polish

### Work

1. Refactor Watchtower state model to include:
   - current Watchtower View
   - global Selected Host
   - per-view remembered state
   - Watchtower View History
2. Wire direct view switching and local back behavior:
   - `g` for Fleet Aggregate
   - `m` for Fleet Matrix
   - `d` for Host Detail
   - `b` for Watchtower back
3. Honor shell Target Scope consistently across all Watchtower Views
4. Define focus contracts for:
   - focused aggregate module
   - focused matrix card
   - focused host-detail module
5. Define size/degradation policy for small terminals

### Validation

- tests for Watchtower state transitions and back behavior
- tests for scope change re-homing of Selected Host
- resize/layout tests cover no-overflow degradation rules

## Phase 4: Watchtower Data and Trend Pipeline

### Outcomes

- centralized SSH polling returns normalized Metrics Snapshots for all major metric families
- refresh cycle supports per-host partial success
- freshness metadata and short Trend Window retention are preserved

### Work

1. Confirm or extend Watchtower snapshot domain model for:
   - CPU
   - Memory
   - Disk
   - Network
2. Define or refine collection interfaces in the domain layer
3. Implement or extend infrastructure collectors using SSH
4. Normalize all required families into stable snapshot shapes
5. Add refresh-cycle orchestration:
   - interval
   - concurrency control
   - partial failures
6. Add freshness metadata:
   - collected at
   - age
   - host status
7. Add bounded in-memory Trend Window retention for trend strips

### Validation

- tests for snapshot normalization across all four families
- tests for partial success behavior
- tests for freshness and trend retention behavior
- diagnostics clean on new domain/infrastructure files

## Phase 5: Watchtower View Implementation

### Outcomes

- Fleet Aggregate, Fleet Matrix, and Host Detail all exist with the new interaction model
- cross-view drill flow and keyboard behavior feel coherent
- severity, missing, and stale states are legible across the redesigned surface

### Work

1. Implement Fleet Aggregate:
   - compressed Aggregate Bundles for CPU/memory/disk/network
   - focused module behavior
   - drill into Fleet Matrix by Enter or family hotkey
2. Implement Fleet Matrix:
   - one Metric Family at a time
   - paginated host cards
   - 2D spatial card navigation
   - edge-triggered page turns
   - explicit `[` / `]` paging
3. Implement Host Detail:
   - bar-first multi-family dashboard
   - focused module behavior
   - `[` / `]` host switching
   - drill into Fleet Matrix by focused family or hotkey
4. Implement Watchtower chrome:
   - view bar/header
   - Selected Host indicator
   - contextual footer hotkeys
5. Surface status treatment:
   - healthy/elevated/critical
   - missing versus stale
   - selective badges in major dashboards
   - explicit badges in Fleet Matrix cards
6. Implement one-screen completeness and fallback compression rules

### Validation

- manual verification with Entire Inventory and Selected Hosts
- resize/layout tests for Bubble Tea model
- no horizontal overflow
- keyboard navigation is spatial where layouts are spatial
- drill flow preserves state and selected-host continuity

## Phase 6: Watchtower Escalation Hooks

### Outcomes

- Watchtower can package redesigned context for future Copilot/Autopilot handoff
- full execution handoff may remain stubbed initially

### Work

1. Define escalation payload shape for redesigned Watchtower:
   - source Watchtower View
   - selected host
   - target scope
   - metric family
   - recent observations
   - freshness context
2. Add UI affordances for escalate to Copilot / Autopilot
3. Wire placeholder handoff targets

### Validation

- selected context is preserved in hook payload
- shell navigation to target mode works predictably
- redesigned Watchtower state survives escalation affordance usage

## Engineering Constraints

- Follow Bubble Tea sizing rules strictly
- No stdout/stderr writes during TUI runtime
- Keep domain layer free of external dependencies
- Wrap errors with context
- Prefer test-first changes for domain logic and UI model behavior
- Align implementation with `docs/plans/watchtower-redesign-spec.md`

## Release Readiness Gate

The redesigned Watchtower is not release-ready until:

- CPU, memory, disk, and network are all present across:
  - Host Detail
  - Fleet Aggregate
  - Fleet Matrix
- cross-view family coverage is complete
- view-aware completeness is acceptable, but missing families are not
- the redesigned keyboard model, freshness model, and trend model behave coherently across the full Watchtower surface

## Risks

### 1. Shell complexity too early
Mitigation: keep minimal globals and clear mode contracts.

### 2. Watchtower layout overflow
Mitigation: test width math aggressively and build memory-first UI with simple grids.

### 3. SSH polling latency on larger inventories
Mitigation: concurrency limits, freshness indicators, partial success model.

### 4. Agent modes feel fake
Mitigation: make placeholders structural, not decorative.

## Recommended Implementation Order

1. shell root model
2. tab bar + hotkey bar
3. Target Scope model and selector
4. placeholder modes
5. Watchtower snapshot domain model
6. SSH memory collection
7. Fleet Matrix memory UI
8. Host Detail memory UI
9. escalation hooks
10. CPU as next metric family
