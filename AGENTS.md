# OpenCode Agent Instructions

This repository contains a DevOps Agent implemented in Go using Clean Architecture, Domain-Driven Design (DDD), and Bubble Tea for the Terminal UI.

## Build and Run

- **Build**: `CGO_ENABLED=0 go build -o agent ./cmd/agent/main.go`
- **Run**: `go run cmd/agent/main.go`
- **Database**: SQLite database is automatically created at `./agent.db`. It uses `modernc.org/sqlite`, a pure Go SQLite driver, so CGO is **not** required.
- **Docker**: A multi-stage Dockerfile is provided. Build with `docker build -t devops-agent .`.
- **Just**: A `Justfile` is provided for common commands (e.g., `just build`, `just run`).

## Architecture & Code Boundaries

The project strictly follows Clean Architecture:

- **Domain Layer** (`internal/domain/agent`):
  - Contains entities, value objects, and repository/service interfaces.
  - **CRITICAL**: The domain layer MUST NOT import any external dependencies or infrastructure-specific packages.
- **Infrastructure Layer** (`internal/infrastructure`):
  - Implements interfaces defined in the domain layer (e.g., SQLite DB, SSH Client).
- **UI Layer** (`internal/ui/tui`):
  - Implements the TUI using `github.com/charmbracelet/bubbletea`.

## Coding Conventions

- **Error Handling**: Always wrap errors with context using the pattern: `fmt.Errorf("context: %w", err)`.
- **TUI Constraints**: Do not use standard `log.Print` or `fmt.Println` to `stdout`/`stderr` inside goroutines or handlers when the TUI is running, as it will corrupt the Bubble Tea interface. Route logs to a file or use the provided `logChan` to render logs inside the TUI.
- **SSH Client**: SSH execution runs in a pseudo-terminal (`vt100`). Always close SSH sessions and connections explicitly using `defer` blocks.

## Testing

- Standard Go testing applies: `go test ./...`
- Mock domain interfaces in `internal/domain` when testing infrastructure or UI layers.
