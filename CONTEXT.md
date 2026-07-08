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
The agent-led execution mode where the system drives an operational loop under explicit approval policy. For fleet operations, this includes agentic planning, host-first sequential execution, and LLM-driven failure recovery.
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

**Scheduled Runbook**:
A saved operational procedure that can start on a time-based schedule against a defined target scope.
_Avoid_: Cron line, background script, job template

**Schedule Trigger**:
The time-based rule that starts a Scheduled Runbook.
_Avoid_: Tick, timer loop, cron daemon

**Schedule Approval Mode**:
The policy that determines whether a Scheduled Runbook executes unattended or produces a run that waits for operator approval.
_Avoid_: Trust level, job permission

**Operational Tool**:
A named agent capability with a defined operational purpose, such as inspecting a host, managing a service, or editing a file.
_Avoid_: Arbitrary command, plugin, shell alias

**Guarded Shell Action**:
An exceptional shell-based action used when no suitable Operational Tool exists and still subject to approval and safety checks.
_Avoid_: Default execution path, raw SSH mode

**Host Capability Profile**:
A discovered description of a target host's operating environment, including operating system, distribution details, available tools, service manager, package manager, and relevant versions.
_Avoid_: Metrics snapshot, inventory record, static host config

**Capability Discovery**:
The process of inspecting a target host to build or refresh its Host Capability Profile before planning or execution.
_Avoid_: Health check, metrics scrape, static inventory parsing

**Execution Branch**:
A capability-compatible subset of the target scope that follows a shared planned step sequence within one Autopilot Run.
_Avoid_: Separate run, host tag, shell fan-out

**Session Memory**:
The resumable working context carried within a Copilot Session.
_Avoid_: Global agent memory, run history, host facts

**Run History**:
The persisted record of prior Autopilot Runs, including plans, approvals, and outcomes.
_Avoid_: Chat transcript, free-form notes, host inventory

**Plan View**:
The hierarchical, node-based representation of an Autopilot Run that shows the Goal, all Execution Branches, and the sequence of steps for each branch.
_Avoid_: Task list, static script, command stream

**Plan Amendment**:
An explicit operator change to a planned step sequence—such as disabling a step or adjusting tool parameters—performed while the Run is in a drafting or blocked state.
_Avoid_: Ad hoc shell override, silent agent rewrite, plan regeneration from scratch

**Skill**:
A reusable instruction package that shapes how the agent interprets a class of operational goals, gathers evidence, selects tools, and structures plans.
_Avoid_: Arbitrary code plugin, shell script, slash command

**Skill Attachment**:
The explicit decision to apply a Skill to a Session or Run, whether chosen directly by the operator or accepted from an agent suggestion.
_Avoid_: Silent auto-load, hidden planner behavior

**Prompt Reference**:
A structured `@` reference in agent input that points to a local file or script so the agent can use that artifact as part of the current Session or Run context.
_Avoid_: Shell redirection, raw path text, terminal attachment

**Prompt Attachment**:
The structured artifact created from a Prompt Reference and carried alongside natural-language agent input.
_Avoid_: Inlined prompt text, shell argument, ad hoc file read

**Scope Hint**:
A natural-language targeting cue in agent input that may restate or narrow the current Target Scope without replacing the shell-level scope model.
_Avoid_: Canonical scope selector, hidden retargeting, saved host group

**Verification Step**:
A planned read-only confirmation step that checks whether a preceding mutating step produced the intended operational outcome.
_Avoid_: Exit code check only, ad hoc spot check, implicit success

**Blocked Subset**:
The hosts within an Execution Branch that cannot continue after a failed verification or unresolved execution problem, while other hosts from the same branch may still proceed.
_Avoid_: Whole-run failure, silent skip, retry queue

**Recovery Action**:
An explicit operator intervention on a Blocked Subset, such as inspecting evidence, retrying the failed step, skipping the affected hosts, or aborting the remaining work for that subset.
_Avoid_: Automatic replanning, silent retry loop, ghost failure

**Recovery Sub-plan**:
A focused plan generated by the LLM in response to a failure, containing diagnostic and remediation steps targeted specifically at the failed host. Requires explicit operator approval before execution.
_Avoid_: Full plan regeneration, automatic retry without analysis, ghost recovery

**Per-host Step Template**:
A single `steps[]` array that Autopilot applies to each host in the Target Scope sequentially. The LLM generates one template rather than per-host step lists.
_Avoid_: Per-host branch repetition, explicit host-specific command lists, ad hoc host branching

**Agentic Loop**:
The autonomous execution cycle where Autopilot generates a plan from context, executes steps, reports failures to the LLM, and recovers with operator approval until the goal is achieved.
_Avoid_: Fixed script execution, manual operator intervention at every step, single-pass planning

**Plan Approval**:
The explicit operator gate where the reviewed plan must be confirmed before Autopilot begins execution. Once approved, all steps (including mutative) execute automatically.
_Avoid_: Per-step approval gating, silent execution without review, delayed approval after start
