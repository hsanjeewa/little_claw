# DevOps Agent

This context defines the domain language for the terminal-based DevOps agent. It focuses on the operator experience, agent roles, and server-oriented workflows rather than implementation details.

## Language

**Mode**:
A top-level operating context within the TUI with its own purpose, lifecycle, and state boundary.
_Avoid_: Tab, screen, page

**Watchtower**:
The monitoring mode focused on fleet and server health visibility.
_Avoid_: Dashboard, metrics page

**Watchtower View**:
A switchable presentation within Watchtower, such as Host Detail, Fleet Aggregate, or Fleet Matrix.
_Avoid_: Mode, tab, page

**Watchtower View History**:
The short internal navigation trail that lets an operator return to the previous Watchtower View in a drill flow.
_Avoid_: Browser history, shell back stack, global undo

**Fleet Matrix**:
The Watchtower view that compares one selected metric across the full inventory.
_Avoid_: Overview table, summary grid

**Fleet Aggregate**:
The Watchtower view that summarizes fleet-wide resource health using the same visual language as Host Detail.
_Avoid_: Overview dashboard, rolled-up btop, global summary

**Aggregate Bundle**:
A Watchtower summary that combines a fleet-level aggregate with outlier visibility, such as average, peak or worst host, and affected host count.
_Avoid_: Single stat, plain rollup, one-number summary

**Host Card Pagination**:
The Fleet Matrix interaction that moves through additional per-host cards when the visible terminal space cannot show the full selected host set at once.
_Avoid_: Metric paging, tab flip, scrolling the whole Watchtower

**Metric Family**:
A switchable category of related resource measurements, such as CPU, memory, disk, or network.
_Avoid_: Column pack, stat group

**Host Detail**:
The Watchtower view that shows a broad set of metrics for one server at a time.
_Avoid_: btop clone, server page

**Autopilot**:
The agent-led execution mode where the system drives an operational loop under explicit approval policy.
_Avoid_: Assistant, chatbot, shell helper

**Approval Policy**:
The rule set that determines which Autopilot actions may run automatically and which require operator confirmation.
_Avoid_: Permissions, trust level

**Copilot**:
The human-led assistance mode where the operator drives the session and the agent advises without taking control by default.
_Avoid_: Autopilot, executor

**Inventory**:
The managed set of target servers known to the agent.
_Avoid_: Cluster, fleet list

**Metrics Snapshot**:
A normalized point-in-time reading of server health and resource usage collected for display in Watchtower.
_Avoid_: Raw stats, telemetry blob

**Selected Host**:
The globally active host within Watchtower that anchors Host Detail and remains stable while the operator switches between Watchtower Views.
_Avoid_: Cursor host, temporary card focus, local row selection

**Freshness**:
The visible age and collection status of a Metrics Snapshot.
_Avoid_: Cache health, polling lag

**Severity Palette**:
The Watchtower color system where visual intensity communicates operational state such as healthy, elevated, or critical pressure rather than decoration alone.
_Avoid_: Random accents, purely cosmetic theme, rainbow styling

**Severity Threshold**:
An explicit per-Metric-Family boundary that determines when Watchtower escalates visual state from healthy to elevated or critical.
_Avoid_: Hidden heuristics, one universal cutoff, decorative coloring

**Trend Window**:
A small rolling in-memory history of recent Metrics Snapshots used to render short Watchtower trend visuals.
_Avoid_: Full telemetry store, single snapshot, unbounded history

**Run**:
An objective-driven Autopilot session that owns its plan, evidence, approvals, actions, and resumable history.
_Avoid_: Chat, prompt, one-off command

**Goal**:
The natural-language operator intent that starts or redirects an Autopilot Run.
_Avoid_: Ticket, raw prompt, command list

**Run State**:
The current lifecycle status of an Autopilot Run.
_Avoid_: Agent mood, task phase

**Session**:
A human-led Copilot working thread that captures terminal exchanges, agent guidance, and resumable operator context.
_Avoid_: Run, chat prompt, shell buffer

**Terminal Session**:
The in-app command and output stream inside the TUI that Copilot is allowed to observe.
_Avoid_: System shell, external terminal

**Task Context**:
The human-led working situation, question, or incident framing that starts a Copilot Session.
_Avoid_: Goal, ticket, raw terminal buffer

**Target Scope**:
The set of servers an action or view applies to, such as the full inventory or an explicit selection.
_Avoid_: Split mode, filter mode

**Entire Inventory**:
The target scope containing every server currently in the inventory.
_Avoid_: All nodes, whole fleet

**Selected Hosts**:
The target scope containing only the servers explicitly chosen by the operator.
_Avoid_: Partial fleet, picked items

**Refresh Cycle**:
One scheduled pass that collects current metrics from the inventory.
_Avoid_: Tick, scrape

**Operator**:
The technical user running the TUI to observe systems and perform operational work.
_Avoid_: End user, customer

**Promotion**:
An explicit operator action that turns a Copilot Session into a new Autopilot Run using carried-forward context.
_Avoid_: Auto-handoff, silent escalation

**Escalation**:
An explicit operator action that turns Watchtower context into a Copilot Session or Autopilot Run.
_Avoid_: Auto-open, implicit drill-through

**Status Badge**:
A compact shell-level indicator that summarizes the background state of a mode.
_Avoid_: Notification, full status panel

**Contextual Hotkeys**:
The currently available key actions shown for the active mode and current focus.
_Avoid_: Global shortcuts list, help page

**Slash Command**:
A typed command prefixed with `/` that triggers a structured in-app action rather than raw terminal execution.
_Avoid_: Shell command, prompt shortcut

**Command Bar**:
The unified text input for product-level intent and Slash Commands within an interactive mode.
_Avoid_: Chat box, shell prompt

**Guidance Preference**:
The operator-controlled setting that enables or disables Copilot guidance visibility and behavior.
_Avoid_: Mute flag, AI mode
