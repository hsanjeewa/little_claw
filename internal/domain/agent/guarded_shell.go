package agent

import (
	"context"
	"fmt"
)

type GuardedShell struct{}

func (g *GuardedShell) Execute(ctx context.Context, command string) (string, error) {
	// Guarded execution logic would go here.
	return fmt.Sprintf("guarded shell executing: %s", command), nil
}
