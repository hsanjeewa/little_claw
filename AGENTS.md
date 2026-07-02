# PROJECT KNOWLEDGE BASE

**Generated:** 2026-07-02

## OVERVIEW
This is an autonomous DevOps Agent implemented in Go. It uses Clean Architecture, Domain-Driven Design (DDD), and Bubble Tea for an interactive Terminal UI (TUI). It leverages Qwen 2.5 (via OpenAI SDK) for LLM-based bash log analysis and a local encrypted AES-256 vault for SSH secret management.

## STRUCTURE
```text
./
├── .agent/skills/       # Agent skills (Bubble Tea, etc.)
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
A comprehensive Bubble Tea UI skill is installed at `.agent/skills/bubbletea.md`. It covers:
- Critical rules (no stdout, non-blocking commands, deterministic sizing)
- Lipgloss sizing rules (padding/border frame calculations)
- Common patterns (channels, side-by-side panels, HITL gates)
- Error handling (errors as messages, context cancellation)
- Component patterns (viewport, adaptive colors, dynamic styles)
- Feature checklist for every Bubble Tea component

Load with: `skill(name="bubbletea")` or reference `.agent/skills/bubbletea.md`.

## CONVENTIONS
- **Error Handling**: Always wrap errors with context using `fmt.Errorf("context: %w", err)`.
- **TUI Constraints**: Do not use `log.Print` or `fmt.Println` to `stdout`/`stderr` during agent operation; it corrupts Bubble Tea. Use the provided `logChan` or write to a file.
- **Dependency Rule**: The `internal/domain` layer MUST NOT import any external dependencies or infrastructure-specific packages.

## TDD (TEST-DRIVEN DEVELOPMENT)
This project follows TDD strictly. Write tests BEFORE implementation.

### Test Commands
```bash
just test                          # Run all tests
go test -v ./internal/domain/...   # Test domain layer only
go test -v -run TestSpecificName   # Run single test by name
go test -cover ./...               # Coverage report
```

### TDD Workflow
1. **Red**: Write a failing test that defines expected behavior
2. **Green**: Write minimum code to make the test pass
3. **Refactor**: Clean up while keeping tests green

### Coverage Rule
**Every code change MUST improve or maintain test coverage.** Before committing:
```bash
go test -cover ./...
```
If coverage decreases, add tests for the uncovered paths. Target: 80%+ for domain layer, 60%+ overall.

### Where Tests Live
- `internal/domain/agent/*_test.go` — Domain logic tests (fast, no external deps)
- `internal/infrastructure/ssh/*_test.go` — SSH client tests (mock the vault)
- `internal/infrastructure/database/*_test.go` — SQLite tests (use temp DB)
- `internal/infrastructure/llm/*_test.go` — LLM client tests (mock HTTP)
- `internal/ui/tui/*_test.go` — TUI model tests (test Update/View logic)

### Testing Patterns
```go
// Domain tests — pure logic, no mocks needed
func TestNewTask_Validation(t *testing.T) {
    _, err := agent.NewTask("id", "host", "", 22, "root", "cmd", false)
    if err == nil {
        t.Fatal("expected error for empty HostIP")
    }
}

// Infrastructure tests — mock interfaces
type mockVault struct{}
func (m *mockVault) GetPrivateKey(host string) (string, error) { return "key", nil }
func (m *mockVault) GetSudoPassword(host string) (string, error) { return "pass", nil }

// TUI tests — simulate messages
func TestModel_UpdateWindowSize(t *testing.T) {
    m := NewModel(nil, nil, nil, nil)
    msg := tea.WindowSizeMsg{Width: 80, Height: 24}
    updated, _ := m.Update(msg)
    if updated.(Model).width != 80 {
        t.Fatal("width not updated")
    }
}
```

## ANTI-PATTERNS (THIS PROJECT)
- **Hardcoding Wait Times**: Never use static `time.Sleep` for network/LLM calls. Context timeouts are configured via `SSH_TIMEOUT_SECONDS` and `LLM_TIMEOUT_SECONDS` in `.env`.
- **Horizontal Overflow in UI**: When building Lipgloss UI components, border width (2 or 4 chars) MUST be explicitly subtracted when assigning widths to child views to prevent terminal wrap-around bugs.
- **Skipping Tests**: Never merge code without tests. If a feature has no test, it's not done.

## COMMANDS
```bash
# Build binary without CGO requirements (Pure Go)
just build
# Alternatively: CGO_ENABLED=0 go build -o agent ./cmd/agent/main.go

# Run the TUI application
just run

# Run all tests
just test
```

## NOTES
- **CGO is NOT required** because this project uses `modernc.org/sqlite`.
- Ensure `.env` is populated with `OPENAI_API_KEY`. The app falls back to OpenRouter.ai and `qwen/qwen-2.5-coder-32b-instruct` if variables are missing.

## CONFIGURATION
The agent reads structural configuration from `config/config.toml` (timeouts, endpoints, inventory paths).
Sensitive keys (like `OPENAI_API_KEY`) should remain in the `.env` file.

## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues for this repo, and external PRs are treated as a triage surface. See `docs/agents/issue-tracker.md`.

### Triage labels

The canonical triage labels use their default names: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, and `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout: one `CONTEXT.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
