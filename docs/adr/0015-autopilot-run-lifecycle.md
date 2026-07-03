# ADR 0015: Autopilot Run lifecycle

## Status

Accepted

## Context

Autopilot needs a stable lifecycle model for UI badges, approvals, persistence, and recovery. Without canonical states, Run behavior will be inconsistent across layers.

## Decision

The initial **Run State** lifecycle is:

- **Drafting**
- **Ready**
- **Executing**
- **Blocked**
- **Completed**
- **Failed**

State meanings:

- **Drafting**: gathering context and shaping a plan
- **Ready**: plan is prepared and waiting for approval or next instruction
- **Executing**: running approved steps
- **Blocked**: waiting on approval, missing information, or failure resolution
- **Completed**: finished successfully
- **Failed**: run cannot proceed without restart or intervention

Background behavior:

- **Drafting** and **Executing** may continue while the mode is unfocused.
- **Ready** and **Blocked** remain waiting.
- Shell status should reflect the current Run State while Autopilot is in the background.

## Consequences

- Status badges, persistence, and filtering can use a shared state vocabulary.
- Approval UX can distinguish between a merely prepared Run and a blocked one.
- Recovery flows can be designed against explicit states instead of ad hoc flags.
