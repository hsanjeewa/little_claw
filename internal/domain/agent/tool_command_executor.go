package agent

import (
	"context"
)

type ToolCommandExecutor struct {
	tools    map[string]Tool
	fallback *GuardedShell
}

func NewToolCommandExecutor(fallback *GuardedShell) *ToolCommandExecutor {
	return &ToolCommandExecutor{
		tools:    make(map[string]Tool),
		fallback: fallback,
	}
}

func (t *ToolCommandExecutor) RegisterTool(tool Tool) {
	t.tools[tool.Name()] = tool
}

func (t *ToolCommandExecutor) Execute(ctx context.Context, task Task) (string, error) {
	if tool, ok := t.tools[task.Command]; ok {
		return tool.Execute(ctx, nil)
	}
	return t.fallback.Execute(ctx, task.Command)
}
