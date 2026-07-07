package agent

import (
	"context"
	"strings"
	"testing"
)

func TestToolCommandExecutor_ExecuteTool(t *testing.T) {
	mockTool := &MockTool{}
	fallback := &GuardedShell{}
	executor := NewToolCommandExecutor(fallback)
	executor.RegisterTool(mockTool)

	task := Task{Command: "test-tool"}
	output, err := executor.Execute(context.Background(), task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output != "tool-output" {
		t.Errorf("Expected tool-output, got %s", output)
	}
	if !mockTool.Called {
		t.Error("Expected tool to be called")
	}
}

func TestToolCommandExecutor_Fallback(t *testing.T) {
	fallback := &GuardedShell{}
	executor := NewToolCommandExecutor(fallback)

	task := Task{Command: "unknown-command"}
	output, err := executor.Execute(context.Background(), task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !strings.Contains(output, "guarded shell executing") {
		t.Errorf("Expected guarded shell output, got %s", output)
	}
}
