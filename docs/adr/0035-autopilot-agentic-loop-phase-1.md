# ADR 0035: Autopilot Agentic Loop - Phase 1

## Status

Accepted

## Context

Autopilot needs an autonomous execution loop that can:
- Generate operational plans from natural-language goals
- Execute plans across remote hosts via SSH
- Handle failures with LLM-driven recovery
- Maintain operator oversight through explicit approvals

The design is inspired by [charmbracelet/crush](https://github.com/charmbracelet/crush) but adapted for remote fleet operations instead of local machine work.

## Decision

### Execution Model: Host-First, Sequential

- **Ordering**: Complete all steps for Host A → Host B → Host C (not step-parallel across hosts)
- **Rationale**: Easier to reason about failures, simpler state tracking, matches fleet operations mental model

### Planning and Approval

#### Plan Generation
- **Trigger**: Operator enters natural-language Goal in Autopilot Command Bar
- **Context Injection** (ordered for Qwen optimization):
  1. JSON schema (structure constraints)
  2. Target Scope (selected hosts with IPs/ports/users)
  3. Host Capability Profiles (OS/shell per host, collected at run start)
  4. Constraints (max 50 steps, host-first sequential, abort on failure)
  5. User Goal (the operator's intent)
  6. `config/SOUL.md` (agent identity/purpose)
  7. `config/IDENTITY.md` (system persona)
  8. Watchtower summary (optional metrics context)
- **Output**: Strict JSON plan with `steps[]` template
- **Step Template**: Single `steps[]` list applied to each host sequentially

#### Plan Schema
```json
{
  "steps": [
    {
      "description": "human-readable description",
      "command": "bash command string",
      "is_mutative": true/false,
      "verify_command": "optional verification command"
    }
  ]
}
```

#### Approval Policy (Supersedes ADR 0003 and ADR 0028 for Phase 1)
- **Plan-level approval**: Operator reviews and approves the entire plan once
- **After approval**: All steps (including mutative) execute automatically without per-step HITL
- **Rationale**: Balances safety with automation for fleet operations

### Failure Recovery Loop

1. **Failure Detection**: Step fails on a host
2. **Abort Run**: Stop execution immediately
3. **Generate Failure Report**: Send structured JSON to LLM:
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
4. **LLM Recovery Plan**: LLM generates recovery sub-plan (same JSON schema)
5. **Recovery Approval**: Operator must explicitly approve recovery sub-plan
6. **Execute Recovery**: Run recovery steps on failed host
7. **Resume**: If recovery succeeds, continue with original plan from failed step

### Target Selection

- **Slash Command**: `/targets` opens overlay modal with multi-select checkboxes
- **Scope**: Updates Shell-level Target Scope (shared across Watchtower/Autopilot/Copilot)
- **UX**: Enter to confirm, Esc to cancel, Ctrl+a to select all, Space to toggle

### Safety Guardrails

- **Maximum steps**: 50 per run (configurable)
- **Timeout per step**: Respect `Agent.SSHTimeoutSeconds` from `config.toml`
- **Stop conditions**: User cancel (`q`/`Ctrl+c`), completion, max steps reached, unrecoverable failure
- **Discovery**: Host capability profiles collected once at run start (read-only SSH commands)

### Configuration

- **Context Files**: 
  - `config/SOUL.md` - agent identity/purpose
  - `config/IDENTITY.md` - system persona
  - Default templates shipped as fallback if files missing (soft warning)
- **LLM Configuration**: Qwen model via OpenRouter (config.toml: `llm.base_url`, `llm.model`)

### Persistence

- **Storage**: SQLite via existing `AuditRepository` interface
- **Content**: Plan JSON, execution transcript, step statuses, recovery history

### UI Layout: Conversational Split Pane (Inspired by Crush)

Inspired by [charmbracelet/crush](https://github.com/charmbracelet/crush) conversational interface.

**Left Pane (Conversation):**
- User goals as user messages
- Agent plan proposals as agent messages
- Execution updates as messages ("Executing on web-prod-01...")
- Recovery attempts as agent messages
- Final completion/summary messages
- Natural language interaction flow

**Right Pane (Active Plan/Status):**
- Structured plan steps list with per-host status
- Current host indicator
- Current step highlighting
- Progress bar or completion percentage
- Quick summary: "3/5 hosts done, step 2 of 4"
- Fleet-level progress visibility

**Rationale:**
- Conversational left pane provides familiar chat-like experience
- Structured right pane maintains fleet operation visibility
- Better than pure chat for multi-host execution tracking
- Preserves operator trust through explicit plan structure

### UI State Machine

```
Drafting (collecting context) → Ready (plan generated, awaiting approval) → 
Executing (running steps) → Blocked (awaiting recovery approval) → 
Completed/Failed
```

### Known Issues

- **HITL Channel Bug**: `internal/domain/agent/autopilot.go` passes `nil` `hitlChan` to `runner.RunTask()`; needs fix to pass real channel or disable HITL gating for approved runs

## Consequences

- **Planning Layer**: Must implement `PlanTasks()` with strict JSON schema and prompt composition
- **Execution Layer**: Must support host-first sequential execution with recovery resumption
- **UI Layer**: 
  - Conversational split-pane layout (conversation left, plan/status right)
  - Plan approval UX, recovery approval UX, `/targets` modal, run state visualization
  - Bubble Tea model changes to support dual-pane layout
- **Domain Layer**: Need `AutopilotRun` entity for state tracking
- **Infrastructure**: Extend SQLite schema for run persistence
- **Breaking Changes**: Phase 1 supersedes per-step approval (ADR 0003, ADR 0028) for plan-level approval
- **UI Breaking Changes**: Phase 1 requires conversational UI layout (replaces 3-pane Command/Plan/Transcript design)