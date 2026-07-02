# PROJECT KNOWLEDGE BASE

**Generated:** 2026-07-02

## OVERVIEW
This is an autonomous DevOps Agent implemented in Go. It uses Clean Architecture, Domain-Driven Design (DDD), and Bubble Tea for an interactive Terminal UI (TUI). It leverages Qwen 2.5 (via OpenAI SDK) for LLM-based bash log analysis and a local encrypted AES-256 vault for SSH secret management.

## STRUCTURE
```text
./
├── cmd/agent/          # Entry point
├── internal/
│   ├── domain/         # Pure domain entities/interfaces (ZERO external dependencies)
│   ├── infrastructure/ # External services (SSH, SQLite, LLM, Config)
│   └── ui/tui/         # Bubble Tea UI
└── hosts.yaml          # Ansible-style inventory for target servers
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| Adding a new TUI feature | `internal/ui/tui/bubbletea_view.go` | Must manage window size constraints perfectly |
| Orchestration logic | `internal/domain/agent/engine.go` | Where `RunTask` lives |
| LLM / OpenAI config | `internal/infrastructure/llm/` | Model configured via `.env` |
| SSH / Agent checks | `internal/infrastructure/ssh/` | Where idempotency and execution happen |

## BUBBLE TEA SKILL
A comprehensive Bubble Tea UI skill is installed at `.opencode/skills/bubbletea.md`. It covers:
- Critical rules (no stdout, non-blocking commands, deterministic sizing)
- Lipgloss sizing rules (padding/border frame calculations)
- Common patterns (channels, side-by-side panels, HITL gates)
- Error handling (errors as messages, context cancellation)
- Component patterns (viewport, adaptive colors, dynamic styles)
- Feature checklist for every Bubble Tea component

Load with: `skill(name="bubbletea")` or reference `.opencode/skills/bubbletea.md`.

## CONVENTIONS
- **Error Handling**: Always wrap errors with context using `fmt.Errorf("context: %w", err)`.
- **TUI Constraints**: Do not use `log.Print` or `fmt.Println` to `stdout`/`stderr` during agent operation; it corrupts Bubble Tea. Use the provided `logChan` or write to a file.
- **Dependency Rule**: The `internal/domain` layer MUST NOT import any external dependencies or infrastructure-specific packages.

## ANTI-PATTERNS (THIS PROJECT)
- **Hardcoding Wait Times**: Never use static `time.Sleep` for network/LLM calls. Context timeouts are configured via `SSH_TIMEOUT_SECONDS` and `LLM_TIMEOUT_SECONDS` in `.env`.
- **Horizontal Overflow in UI**: When building Lipgloss UI components, border width (2 or 4 chars) MUST be explicitly subtracted when assigning widths to child views to prevent terminal wrap-around bugs.

## COMMANDS
```bash
# Build binary without CGO requirements (Pure Go)
just build
# Alternatively: CGO_ENABLED=0 go build -o agent ./cmd/agent/main.go

# Run the TUI application
just run
```

## NOTES
- **CGO is NOT required** because this project uses `modernc.org/sqlite`.
- Ensure `.env` is populated with `OPENAI_API_KEY`. The app falls back to OpenRouter.ai and `qwen/qwen-2.5-coder-32b-instruct` if variables are missing.

## CONFIGURATION
The agent reads structural configuration from `config/config.toml` (timeouts, endpoints, inventory paths).
Sensitive keys (like `OPENAI_API_KEY`) should remain in the `.env` file.
