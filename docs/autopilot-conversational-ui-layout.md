# Autopilot UI Layout - Conversational Split Pane

## Overview

Inspired by [charmbracelet/crush](https://github.com/charmbracelet/crush) conversational interface, adapted for fleet operations.

## Layout Structure

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ AUTOPILOT | Selected Hosts: web-prod-01, db-master | [q] quit              │
├──────────────────────────────┬──────────────────────────────────────────────┤
│ CONVERSATION (60% width)      │ ACTIVE PLAN / STATUS (40% width)             │
│                              │                                              │
│ 👤 Operator:                 │ 📋 Plan: Investigate high memory          │
│ Check memory usage on        │                                              │
│ selected hosts               │ Steps:                                       │
│                              │ ✅ 1. free -h                                  │
│ 🤖 Agent:                    │ 🔄 2. Check top memory consumers             │
│ I'll investigate memory       │    (web-prod-01: Running)                    │
│ usage across 2 hosts.        │    (db-master: Pending)                       │
│                              │ ⏳ 3. Identify consuming processes             │
│ 👤 Operator:                 │ ⏳ 4. Propose cleanup actions                 │
│ Approve plan                 │                                              │
│                              │ Progress:                                     │
│ 🤖 Agent:                    │ web-prod-01: ████████░░ 2/4 steps             │
│ Executing on web-prod-01...  │ db-master:   ░░░░░░░░░░ 0/4 steps             │
│                              │                                              │
│ ✅ Step 1 complete:          │                                              │
│ Mem: 8.2GB / 16GB (51%)      │                                              │
│                              │                                              │
│ 🔄 Running step 2 on         │                                              │
│ web-prod-01...               │                                              │
│                              │                                              │
│                              │                                              │
│ [Enter message...]           │                                              │
└──────────────────────────────┴──────────────────────────────────────────────┘
```

## Left Pane: Conversation

### Message Types

1. **User Goals** (👤 Operator)
   - Natural language intent
   - Plain text input from command bar
   - Example: "Check memory usage on selected hosts"

2. **Agent Plan Proposals** (🤖 Agent)
   - Generated plan presentation
   - JSON plan formatted as readable steps
   - Approval request embedded
   - Example:
     ```
     🤖 Agent:
     I'll investigate memory usage across 2 hosts.

     Proposed Plan:
     1. free -h (check memory)
     2. Check top memory consumers
     3. Identify consuming processes
     4. Propose cleanup actions

     [Approve Plan] [Reject] [Edit]
     ```

3. **Execution Updates** (🤖 Agent)
   - Progress notifications per host/step
   - Success/failure confirmations
   - Structured output, not raw logs
   - Example:
     ```
     🤖 Agent:
     Executing on web-prod-01...

     ✅ Step 1 complete:
     Mem: 8.2GB / 16GB (51%)

     🔄 Running step 2 on web-prod-01...
     ```

4. **Recovery Attempts** (🤖 Agent)
   - Failure notification
   - Recovery plan proposal
   - Approval request
   - Example:
     ```
     🤖 Agent:
     ❌ Failed on db-master: Step 2 - "top" command timeout

     Recovery Plan:
     1. Check if top is installed
     2. Use ps aux as fallback
     3. Resume original step 2

     [Approve Recovery] [Skip Host] [Abort]
     ```

5. **Completion/Summary** (🤖 Agent)
   - Final run status
   - Per-host summary
   - Next steps or warnings
   - Example:
     ```
     🤖 Agent:
     ✅ Run complete

     web-prod-01: All 4 steps succeeded
     db-master:   All 4 steps succeeded

     Memory usage normal on both hosts. No cleanup needed.
     ```

### User Actions

- **Enter message**: Type natural language or approval responses
- **Approve/Reject buttons**: Embedded in agent messages
- **Cancel**: `q` or `Ctrl+c` to abort current run

## Right Pane: Active Plan / Status

### Plan Display

- **Plan Title**: Goal or plan description
- **Steps List**: All steps with per-host status indicators
- **Step Status Icons**:
  - ⏳ Pending
  - 🔄 Running
  - ✅ Complete
  - ❌ Failed
  - ⚠️ Blocked

### Per-Host Progress

For each host in scope:
- **Host Alias**: web-prod-01, db-master
- **Progress Bar**: Visual completion (e.g., `████░░░░ 2/4 steps`)
- **Current Step**: Which step is currently running
- **Overall Status**: Success/Failure/Blocked/Complete

### Quick Summary

- **Overall Progress**: "3/5 hosts done, step 2 of 4"
- **Time Elapsed**: "2m 34s"
- **Estimated Remaining**: "~1m remaining"

## Interaction Flow

### 1. Initial Planning
```
Left Pane:
👤 Operator: Check memory on web servers
🤖 Agent: Generated plan [Show Plan] [Approve] [Edit]

Right Pane:
📋 Plan: Check memory on web servers
Steps:
1. free -h
2. top -b -n 1 | head -20
3. ps aux --sort=-%mem | head -10
```

### 2. Plan Approval & Execution
```
Left Pane:
👤 Operator: Approve plan
🤖 Agent: Executing on web-prod-01...
✅ Step 1 complete
🔄 Running step 2...

Right Pane:
web-prod-01: 🔄 Step 2 of 3 (████░░░░)
db-master:   ⏳ Step 0 of 3 (░░░░░░░░)
```

### 3. Failure & Recovery
```
Left Pane:
🤖 Agent: ❌ Failed on web-prod-01: Step 2 timeout
Recovery plan: [Show Recovery] [Approve] [Skip Host]

Right Pane:
web-prod-01: ⚠️ Blocked - awaiting recovery
db-master:   ⏳ Pending
```

### 4. Completion
```
Left Pane:
🤖 Agent: ✅ Run complete
web-prod-01: 3/3 steps succeeded
db-master:   3/3 steps succeeded

Right Pane:
web-prod-01: ✅ Complete (████████)
db-master:   ✅ Complete (████████)
```

## Bubble Tea Model Structure

```go
type AutopilotModel struct {
    // Conversational state
    messages      []ConversationMessage
    commandInput  textinput.Model
    focusedPane   int // conversation or status

    // Plan/status state
    plan          *AutopilotPlan
    runState      RunState
    hostProgress  map[string]HostProgress

    // Metadata
    width, height int
    currentRunID  string
}

type ConversationMessage struct {
    Role      string // "operator" or "agent"
    Content   string
    Timestamp time.Time
    Type      MessageType // goal, plan, execution, recovery, completion
    Actions   []MessageAction // approve, reject, edit, etc.
}

type HostProgress struct {
    Alias           string
    TotalSteps      int
    CompletedSteps  int
    CurrentStep     int
    Status          StepStatus
    LastOutput      string
}
```

## Implementation Notes

### Rendering Conversation
- Use `bubbletea` list or viewport for scrollable messages
- **Syntax highlighting (hybrid)**:
  - **Explicit**: Code fences (` ```json`, ` ```bash`) → Chroma with specified lexer
  - **Auto-detect**: No fences → `detectLanguage()` → auto-highlight bash/json/yaml
  - **Fallback**: Plain text if detection fails
- Collapsible long outputs
- Message grouping by type (planning, execution, recovery)

### Rendering Status Pane
- Fixed height or proportional split
- Update in real-time during execution
- Use lipgloss progress bars and status colors
- Host cards with compact information

### Resizing Behavior
- Left pane: Min 40% width, expands for long messages
- Right pane: Min 25% width, shows structured info
- Responsive to window resize events
- Preserve scroll position during resize

## Comparison with Crush

| Aspect | Crush | Little Claw Autopilot |
|--------|-------|------------------------|
| **Domain** | Local coding | Remote fleet operations |
| **Conversation** | Full chat interface | Split-pane (chat + status) |
| **Progress Visibility** | Inline in chat | Structured right pane |
| **Multi-entity tracking** | Not applicable | Per-host progress bars |
| **Tool output** | Inline in chat | Summarized in chat, details in logs |
| **Scope** | Single machine | Multiple hosts |

**Why Split Pane:**
- Conversational left provides familiar chat experience
- Structured right maintains fleet visibility
- Better for multi-host execution than pure chat
- Preserves operator trust through explicit plan structure