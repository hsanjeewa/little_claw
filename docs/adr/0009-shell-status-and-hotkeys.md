# ADR 0009: Shell status badges and contextual hotkeys

## Status

Accepted

## Context

The TUI shell presents multiple independent long-lived modes. Operators need lightweight awareness of background activity and a fast way to discover available actions without leaving the current workflow.

## Decision

The shell-level tab bar will show **Status Badges** for each mode.

- **Watchtower** shows health or freshness state.
- **Autopilot** shows execution state such as running, waiting for approval, or error.
- **Copilot** shows session activity state.

The shell will also show **Contextual Hotkeys** similar in spirit to Zellij: the visible key actions adapt to the active mode and current focus.

Autopilot and Copilot will support **Slash Commands** as a first-class in-app interaction mechanism.

## Consequences

- Operators can understand background mode state without switching tabs.
- The shell chrome must be modeled separately from mode content.
- Keybindings need both stable global conventions and mode-specific context-sensitive actions.
- Text-driven workflows in Autopilot and Copilot need a parser and command registry distinct from raw shell command execution.
