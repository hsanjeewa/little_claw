# Autopilot Agentic Loop - Grilling Session Summary

## Session Outcome

✅ **Complete design agreement** for Phase 1 Autopilot agentic loop implementation

## Architectural Decisions

### 1. Execution Model
- **Host-first, sequential**: Complete all steps on Host A → Host B → Host C
- **Abort on first failure**: Stop immediately, don't continue to next host
- **LLM-driven recovery**: Send structured failure transcript → LLM generates recovery sub-plan → operator approves → execute → resume

### 2. Planning System
- **Prompt composition order** (optimized for Qwen):
  1. JSON schema
  2. Target Scope
  3. Host Capability Profiles
  4. Constraints
  5. User Goal
  6. SOUL.md
  7. IDENTITY.md
  8. Watchtower summary

- **Plan format**: Strict JSON with per-host step template
  ```json
  {
    "steps": [
      {
        "description": "human-readable",
        "command": "static bash command",
        "is_mutative": true,
        "verify_command": "optional verification"
      }
    ]
  }
  ```

### 3. Approval Policy (Phase 1)
- **Plan-level approval**: Operator reviews and approves entire plan once
- **After approval**: All steps execute automatically (no per-step HITL)
- **Recovery sub-plans**: Require explicit operator approval before execution

### 4. Target Selection
- **Slash command**: `/targets` opens overlay modal with multi-select checkboxes
- **Scope**: Updates Shell-level Target Scope (shared across modes)
- **Actions**: Enter (confirm), Esc (cancel), Ctrl+a (select all), Space (toggle)

### 5. Safety Guardrails
- **Maximum steps**: 50 per run (configurable)
- **Timeout per step**: From `config.toml` (`Agent.SSHTimeoutSeconds`)
- **Stop conditions**: Cancel/complete/max-steps/unrecoverable failure
- **Capability discovery**: Collected once at run start (read-only SSH)

### 6. Configuration
- **Context files**: `config/SOUL.md` and `config/IDENTITY.md`
- **Default templates**: Shipped as fallback (soft warning if files missing)
- **LLM**: Qwen model via OpenRouter (from `config.toml`)

### 7. Persistence
- **Storage**: SQLite via existing `AuditRepository`
- **Content**: Plan JSON, execution transcript, step statuses, recovery history

### 8. Recovery Transcript Format
```json
{
  "goal": "original user goal",
  "failed_host": "host alias",
  "failed_step_index": 3,
  "failed_step": {step object},
  "completed_steps": [
    {step, output, status, timestamp}
  ],
  "error": "error message"
}
```

## Documentation Created

1. **ADR 0035**: `docs/adr/0035-autopilot-agentic-loop-phase-1.md`
   - Full architectural specification
   - Supersedes ADR 0003 and ADR 0028 for Phase 1 plan-level approval

2. **Updated CONTEXT.md**: Added glossary terms
   - Recovery Sub-plan
   - Per-host Step Template
   - Agentic Loop
   - Plan Approval

3. **Default Templates**:
   - `config/SOUL.md` - Agent identity and purpose
   - `config/IDENTITY.md` - System persona and behavioral guidelines

## Known Issues to Address

1. **HITL Channel Bug**: `internal/domain/agent/autopilot.go:48` passes `nil` `hitlChan` to `runner.RunTask()` - needs fix
2. **PlanTasks() Stub**: `internal/infrastructure/llm/openai_client.go:64` returns error - needs implementation
3. **Autopilot Model Integration**: TUI needs plan approval UX, recovery approval UX, run state visualization

## Next Implementation Steps

### Phase 1 - Core Loop
1. Fix HITL channel bug in autopilot runner
2. Implement `PlanTasks()` with prompt composition and JSON schema validation
3. Add host capability discovery (SSH commands: `uname -a`, `echo $SHELL`, `cat /etc/os-release`)
4. Build AutopilotRun state machine (Drafting → Ready → Executing → Blocked → Completed/Failed)
5. Implement SQLite schema extensions for run persistence

### Phase 1 - Execution Engine
1. Enhance `AutopilotRunner` for host-first sequential execution
2. Implement failure detection and transcript generation
3. Build LLM recovery sub-plan generation
4. Add recovery approval gating

### Phase 1 - UI Integration
1. **Redesign AutopilotModel** to conversational split-pane layout:
   - Left pane: Conversation (user goals, agent responses, execution updates)
   - Right pane: Active plan/status (steps, host progress, current step highlighting)
2. Add `/targets` overlay modal for target selection
3. Build plan approval UX in conversation flow
4. Add recovery approval UX in conversation flow
5. Implement run state badges and progress visualization in right pane
6. Update conversation pane with structured execution messages (not raw logs)

## Testing Strategy

- Unit tests for prompt composition and JSON schema validation
- Integration tests for host capability discovery
- E2E tests for full agentic loop with simulated SSH
- Manual testing with real fleet (start with 1-2 hosts)

## Dependencies

- Existing: `Agent.Engine`, `Task` entity, `AuditRepository`
- New: `AutopilotRun` entity, `PlanTasks()` implementation, UI state components
- External: Qwen model via OpenRouter (configured)

## Success Criteria

Phase 1 is complete when:
1. Operator can select targets via `/targets`
2. Operator can enter goal, see generated plan, and approve it
3. Autopilot executes plan host-first, sequentially
4. Failures trigger LLM recovery sub-plan generation
5. Operator approves recovery and run resumes
6. Run state persists to SQLite
7. All guardrails (50 steps, timeouts, stop conditions) enforced

---

**Session complete.** All critical design decisions captured in ADR 0035 and CONTEXT.md. Ready for implementation.