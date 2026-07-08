package agent

import (
	"context"
	"testing"
)

type MockTool struct {
	Called bool
}

func (m *MockTool) Name() string { return "test-tool" }
func (m *MockTool) Execute(ctx context.Context, args []string) (string, error) {
	m.Called = true
	return "tool-output", nil
}

func TestToolExecution(t *testing.T) {
	mockTool := &MockTool{}
	ctx := context.Background()

	output, err := mockTool.Execute(ctx, []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output != "tool-output" {
		t.Errorf("Expected tool-output, got %s", output)
	}
	if !mockTool.Called {
		t.Error("Expected Execute to be called")
	}
}
