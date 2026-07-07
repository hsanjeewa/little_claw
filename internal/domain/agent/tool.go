package agent

import "context"

type Tool interface {
	Name() string
	Execute(ctx context.Context, args []string) (string, error)
}
