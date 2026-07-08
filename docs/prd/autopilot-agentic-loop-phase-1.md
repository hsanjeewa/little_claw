# PRD: Autopilot Agentic Loop - Phase 1

## Problem Statement

Autopilot currently lacks an autonomous execution loop. The system can execute individual tasks but cannot:
- Generate operational plans from natural-language goals
- Execute plans across multiple hosts in sequence
- Handle failures with intelligent recovery
- Provide conversational interaction familiar to operators

Operators must manually plan and coordinate fleet operations, which reduces efficiency and increases error risk for complex multi-host workflows.

## Solution

Implement an autonomous agentic loop inspired by [charmbracelet/crush](https://github.com/charmbracelet/crush) but adapted for fleet operations:

1. **Conversational Interface**: Chat-like interaction where operators express goals and the agent proposes plans, reports progress, and handles recoveries
2. **Host-First Sequential Execution**: Complete all steps on Host A before proceeding to Host B
3. **LLM-Driven Planning**: Generate structured JSON plans from natural-language goals with injected context (SOUL, IDENTITY, fleet info, capabilities)
4. **Intelligent Failure Recovery**: Abort on first failure, send structured transcript to LLM, generate recovery sub-plan, require operator approval, then resume
5. **Split-Pane UI**: Conversation pane (left) for interaction + Active Plan/Status pane (right) for fleet progress visibility

## User Stories

### Target Selection
1. As an operator, I want to open a target selection modal via `/targets` slash command, so that I can quickly select which hosts to operate on
2. As an operator, I want to see all inventory hosts in a multi-select list with checkboxes, so that I can efficiently build a Target Scope
3. As an operator, I want to use keyboard shortcuts (Enter to confirm, Esc to cancel, Ctrl+a to select all, Space to toggle), so that I can select targets rapidly
4. As an operator, I want the Target Scope to update across all modes (Watchtower/Autopilot/Copilot), so that my selection persists when switching modes
5. As an operator, I want to see the current Target Scope in the shell chrome, so that I always know which hosts are in scope

### Planning & Approval
6. As an operator, I want to express goals in natural language (e.g., "Check nginx status on webservers"), so that I can communicate intent naturally
7. As an operator, I want to see the agent's proposed plan with clear step descriptions and commands, so that I can understand what will execute
8. As an operator, I want to approve the entire plan once (rather than approving each step), so that execution proceeds efficiently for multi-host operations
9. As an operator, I want the plan to show which steps are mutative vs. read-only, so that I can understand the safety profile
10. As an operator, I want to see the agent's reasoning in the conversation before the plan, so that I understand why certain steps were proposed

### Execution & Progress
11. As an operator, I want to see which host and which step is currently executing, so that I can monitor progress in real-time
12. As an operator, I want to see per-host progress bars in the status pane, so that I can track fleet-wide completion
13. As an operator, I want execution updates interleaved in the conversation (e.g., "Executing step 2 on web-prod-01..."), so that I maintain awareness without switching panes
14. As an operator, I want to see step completion with summarized output (not raw logs), so that the conversation remains readable
15. As an operator, I want to cancel execution at any time with `q` or `Ctrl+c`, so that I can abort if something goes wrong

### Failure Recovery
16. As an operator, I want the agent to abort on first failure rather than continuing to other hosts, so that failures are contained and visible
17. As an operator, I want to see a structured failure report showing which host/step failed and why, so that I can understand what went wrong
18. As an operator, I want the LLM to propose a recovery sub-plan for the failed host, so that we can intelligently recover without manual replanning
19. As an operator, I want to explicitly approve the recovery sub-plan before execution, so that I maintain oversight even during automated recovery
20. As an operator, I want the agent to resume the original plan from the failed step after recovery succeeds, so that the run completes without restarting

### Context & Discovery
21. As an operator, I want the agent to collect host capability profiles (OS, shell, versions) at run start, so that plans are accurate for each host
22. As an operator, I want Watchtower metrics to be included in the planning context, so that the agent considers current fleet state
23. As an operator, I want to customize the agent's identity via `config/SOUL.md`, so that the agent behavior matches our operational philosophy
24. As an operator, I want to customize the system persona via `config/IDENTITY.md`, so that the agent communicates in our preferred style
25. As an operator, I want default templates shipped for SOUL/IDENTITY, so that the system works even without custom configuration

### Safety & Guardrails
26. As an operator, I want a maximum step limit (50 steps) per run, so that runaway plans don't consume excessive resources
27. As an operator, I want per-step SSH timeouts from configuration, so that hung commands don't block indefinitely
28. As an operator, I want the agent to respect stop conditions (cancel, completion, max steps, unrecoverable failure), so that runs terminate cleanly
29. As an operator, I want read-only steps to run automatically after plan approval, so that diagnostics don't require repeated approvals
30. As an operator, I want mutative steps to run automatically after plan approval (not per-step HITL), so that fleet operations are efficient

### UI & Visualization
31. As an operator, I want a conversational interface on the left (like a chat app), so that interaction feels familiar
32. As an operator, I want an Active Plan/Status pane on the right, so that I can see structured fleet progress
33. As an operator, I want to see JSON plans with syntax highlighting, so that I can quickly parse plan structure
34. As an operator, I want bash commands with syntax highlighting, so that I can verify command correctness
35. As an operator, I want the LLM output to use code fences when helpful (e.g., ```json), so that syntax highlighting works reliably
36. As an operator, I want the system to auto-detect syntax when code fences aren't used, so that I still get highlighting even if the LLM forgets

### Persistence & History
37. As an operator, I want runs to persist to SQLite, so that I can review past executions and results
38. As an operator, I want to see run history including plans, transcripts, and outcomes, so that I can audit operations
39. As an operator, I want the transcript to include step-by-step execution with timestamps, so that I can reconstruct what happened
40. As an operator, I want recovery attempts to be recorded in run history, so that I can understand what recovery actions were taken

### Syntax Highlighting
41. As an operator, I want JSON plans to be highlighted, so that I can quickly identify step fields (description, command, is_mutative)
42. As an operator, I want bash commands to be highlighted, so that I can verify command syntax and flags
43. As an operator, I want YAML snippets to be highlighted, so that configuration examples are readable
44. As an operator, I want code fences to be detected explicitly (```json, ```bash), so that I get the right highlighting
45. As an operator, I want auto-detection when code fences aren't used, so that I still get some highlighting even with messy output

## Implementation Decisions

### Architecture
- **Agentic Loop**: Inspired by crush but adapted for SSH-based fleet operations
- **Host-First Sequential**: Complete all steps on Host A before Host B (simpler state tracking, easier failure reasoning)
- **Plan-Level Approval**: Single approval gate for entire plan (supersedes ADR 0003/0028 for Phase 1)
- **Recovery Loop**: Failure → Abort → LLM recovery sub-plan → Approval → Execute → Resume
- **Conversational UI**: Split-pane (conversation left, plan/status right) - different from crush's single-pane chat

### LLM Integration
- **Model**: Qwen via OpenRouter (from `config.toml`)
- **Prompt Composition Order** (optimized for Qwen): JSON schema → Target Scope → Capabilities → Constraints → Goal → SOUL → IDENTITY → Watchtower
- **Context Injection**: SOUL.md, IDENTITY.md, fleet targets, Watchtower metrics, host capabilities
- **Output Format**: Strict JSON with per-host step template (single `steps[]` applied to all hosts)
- **Fallback**: Default SOUL/IDENTITY templates shipped (soft warning if config files missing)

### Plan Schema
```json
{
  "steps": [
    {
      "description": "human-readable description",
      "command": "static bash command",
      "is_mutative": true,
      "verify_command": "optional verification command"
    }
  ]
}
```

### Recovery Transcript Format
```json
{
  "goal": "original user goal",
  "failed_host": "host alias",
  "failed_step_index": 3,
  "failed_step": {step object},
  "completed_steps": [{step, output, status, timestamp}],
  "error": "error message"
}
```

### Safety Guardrails
- **Maximum Steps**: 50 per run (configurable)
- **Timeout Per Step**: Respect `Agent.SSHTimeoutSeconds` from `config.toml`
- **Stop Conditions**: User cancel (q/Ctrl+c), completion, max steps reached, unrecoverable failure

### UI Components
- **AutopilotModel**: Redesigned from 3-pane (Command/Plan/Transcript) to 2-pane (Conversation/Status)
- **Conversation Pane**: Scrollable messages with role indicators (👤 Operator, 🤖 Agent)
- **Status Pane**: Plan steps list, per-host progress bars, current host/step highlighting
- **Message Types**: User goals, agent plans, execution updates, recovery attempts, completion summaries
- **Syntax Highlighting**: Hybrid (code fences → explicit lexer; no fences → auto-detect; fallback to plain text)

### Persistence
- **Storage**: SQLite via existing `AuditRepository` interface
- **Schema Extensions**: New tables for AutopilotRun, Plan, ExecutionStep, RecoveryAttempt
- **Content**: Plan JSON, execution transcript, step statuses, recovery history

### Domain Layer
- **AutopilotRun Entity**: Track run state (Drafting → Ready → Executing → Blocked → Completed/Failed)
- **ConversationMessage Entity**: Role, content, timestamp, type, actions
- **HostProgress Entity**: Track per-host execution state

### Infrastructure Layer
- **PlanTasks() Implementation**: Extend `AIAnalyzer` interface with prompt composition and JSON schema validation
- **Capability Discovery**: SSH commands (`uname -a`, `echo $SHELL`, `cat /etc/os-release`) collected once at run start
- **SSH Executor**: Reuse existing `Engine` with modifications for plan-level approval

### State Machine
```
Drafting (collecting context) 
→ Ready (plan generated, awaiting approval) 
→ Executing (running steps) 
→ Blocked (awaiting recovery approval) 
→ Completed/Failed
```

### Key Integration Seams for Testing
1. **AutopilotRunner.ExecutePlan()** - Highest level seam for testing full orchestration flow
2. **AIAnalyzer.PlanTasks()** - LLM planning interface with context injection
3. **Shell.SelectHosts()** - Target selection scope propagation
4. **AutopilotModel.Update()** - UI state machine with conversational flow

## Testing Decisions

### What Makes a Good Test
- **External Behavior**: Test plan generation, execution flow, failure recovery - not internal implementation details
- **Mock Dependencies**: Mock SSH executor, LLM client, and Vault for isolated testing
- **Deterministic Outcomes**: Use fixed LLM responses or mock completions for reproducible tests

### Modules to Test
- **AutopilotRunner.ExecutePlan()** - Full orchestration flow (success, failure, recovery)
- **AIAnalyzer.PlanTasks()** - Prompt composition, JSON schema validation
- **AutopilotModel** - Conversational flow, message rendering, approval gating
- **Capability Discovery** - Host profile collection and parsing
- **SQLite Persistence** - Run state storage and retrieval

### Prior Art
- **Watchtower Tests** (`internal/ui/tui/watchtower_test.go`) - Testing snapshot collection and state transitions
- **Shell Tests** (`internal/ui/tui/shell_test.go`) - Testing scope changes and mode switching
- **Engine Tests** (`internal/domain/agent/engine_test.go`) - Testing task execution and HITL flows

### Test Scenarios
1. **Happy Path**: Goal → Plan → Approve → Execute on 2 hosts → Complete
2. **Failure Recovery**: Goal → Plan → Approve → Execute → Fail → Recovery → Approve → Resume → Complete
3. **Max Steps**: Goal → Plan with 51 steps → Reject with error
4. **Capability Discovery**: Run start → Collect profiles → Verify OS/shell detection
5. **Conversational Flow**: Goal → Plan → Reject → New Goal → New Plan → Approve
6. **Syntax Highlighting**: JSON plan with/without code fences → Verify highlighting

## Out of Scope

### Phase 2 Features (Deferred)
- Per-host step branching (different commands per host based on capabilities)
- Parallel execution across hosts
- Automatic capability discovery updates (cached until run start)
- Plan editing (disable steps, adjust parameters) - operator reads but doesn't edit
- Skill system integration
- Run history search and filtering
- Export runs to artifacts
- Multi-run concurrency

### Infrastructure Complexity
- Distributed lock management (single AutopilotRun per shell instance)
- Real-time fleet metrics streaming (use Watchtower snapshots)
- Advanced error classification (beyond success/failure)
- Rollback capabilities for mutative operations

### UI Polish
- Theme customization for syntax highlighting
- Message archival and search
- Export conversation to transcript file
- Multi-run comparison views

## Further Notes

### ADR Supersedions
- **ADR 0035** supersedes **ADR 0003** and **ADR 0028** for Phase 1 plan-level approval
- **ADR 0035** defines new conversational UI layout (replaces 3-pane design)

### Configuration Files
- `config/SOUL.md`: Agent identity and purpose (default template shipped)
- `config/IDENTITY.md`: System persona and behavioral guidelines (default template shipped)
- `config/config.toml`: LLM configuration (base_url, model, timeout), agent settings (ssh_timeout, inventory_path)

### Documentation Created
- `docs/adr/0035-autopilot-agentic-loop-phase-1.md` - Full architectural specification
- `docs/autopilot-conversational-ui-layout.md` - Detailed UI mockup and implementation notes
- `config/SOUL.md` - Default agent identity template
- `config/IDENTITY.md` - Default system persona template
- `docs/autopilot-agentic-loop-grilling-summary.md` - Design decisions and implementation roadmap

### Success Criteria
Phase 1 is complete when:
1. Operator can select targets via `/targets` overlay modal
2. Operator can enter goal, see generated plan with syntax highlighting, and approve it
3. Autopilot executes plan host-first, sequentially across selected hosts
4. Failures trigger LLM recovery sub-plan generation with approval gating
5. Operator approves recovery and run resumes from failed step
6. Run state persists to SQLite with full history
7. All guardrails (50 steps, timeouts, stop conditions) enforced
8. Conversational UI shows messages with syntax highlighting (hybrid code fence/auto-detect)
9. Status pane shows real-time per-host progress with current step highlighting
10. Target Scope updates propagate across Watchtower/Autopilot/Copilot modes

### Known Issues to Address
1. **HITL Channel Bug**: `AutopilotRunner` passes `nil` `hitlChan` to `runner.RunTask()` - needs fix to pass real channel or disable HITL gating for approved runs
2. **PlanTasks() Stub**: `LocalOpenAIClient.PlanTasks()` returns error - needs implementation
3. **AutopilotModel Integration**: Needs plan approval UX, recovery approval UX, run state visualization
4. **Context File Loading**: Need to implement config/SOUL.md and config/IDENTITY.md loading with fallback to defaults

### Dependencies
- **Existing**: `Agent.Engine`, `Task` entity, `AuditRepository`, Watchtower snapshots
- **New**: `AutopilotRun` entity, `PlanTasks()` implementation, UI state components, SQLite schema extensions
- **External**: Qwen model via OpenRouter (configured in config.toml)
- **Bubble Tea**: Conversational UI with split-pane layout, message rendering, syntax highlighting
- **Chroma**: Syntax highlighting for JSON, bash, YAML (existing infrastructure reused)