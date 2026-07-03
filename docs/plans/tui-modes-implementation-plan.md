# Implementation Plan: Tabbed TUI Modes

## Objective

Deliver a tabbed TUI shell with three persistent modes, using Watchtower as the first functional vertical slice and structural placeholders for Autopilot and Copilot.

## Delivery Strategy

- Build the shell and shared seams first
- Validate cross-mode state, scope, and navigation early
- Implement Watchtower memory path end-to-end before CPU and agent mode behavior
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

## Phase 3: Watchtower Data Pipeline (Memory)

### Outcomes

- centralized SSH polling returns normalized memory snapshots
- refresh cycle supports per-host partial success
- freshness metadata is preserved

### Work

1. Define Watchtower snapshot domain model
2. Define collection interface in domain layer
3. Implement infrastructure collector using SSH
4. Normalize memory metrics into stable snapshot shape
5. Add refresh-cycle orchestration:
   - interval
   - concurrency control
   - partial failures
6. Add freshness metadata:
   - collected at
   - age
   - host status

### Validation

- tests for snapshot normalization
- tests for partial success behavior
- diagnostics clean on new domain/infrastructure files

## Phase 4: Watchtower UI (Memory)

### Outcomes

- Fleet Matrix displays memory across current Target Scope
- host selection drills into Host Detail
- failures and stale data are visible

### Work

1. Implement Fleet Matrix memory view
2. Implement host selection/focus model
3. Implement Host Detail memory view
4. Add back navigation to Fleet Matrix
5. Surface freshness and host failure state in both views
6. Honor shell Target Scope live

### Validation

- manual verification with Entire Inventory and Selected Hosts
- resize/layout tests for Bubble Tea model
- no horizontal overflow

## Phase 5: Watchtower Escalation Hooks

### Outcomes

- Watchtower can package context for future Copilot/Autopilot handoff
- full execution handoff may remain stubbed initially

### Work

1. Define escalation payload shape:
   - source mode
   - host/scope
   - metric family
   - recent observations
2. Add UI affordances for escalate to Copilot / Autopilot
3. Wire placeholder handoff targets

### Validation

- selected context is preserved in hook payload
- shell navigation to target mode works predictably

## Phase 6: CPU Follow-Up

### Outcomes

- second Metric Family proves Watchtower extensibility

### Work

1. Extend snapshot model for CPU
2. Add CPU collection/normalization
3. Add Fleet Matrix family switching
4. Add CPU Host Detail presentation

### Validation

- memory behavior remains stable
- family switching updates hotkeys and rendering cleanly

## Engineering Constraints

- Follow Bubble Tea sizing rules strictly
- No stdout/stderr writes during TUI runtime
- Keep domain layer free of external dependencies
- Wrap errors with context
- Prefer test-first changes for domain logic and UI model behavior

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
