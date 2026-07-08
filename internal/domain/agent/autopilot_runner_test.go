package agent

import (
	"context"
	"testing"
)

type MockEngine struct {
	RunTaskCalled bool
	CallCount     int
}

func (m *MockEngine) RunTask(ctx context.Context, task Task, hitlChan chan<- HitlRequest) {
	m.RunTaskCalled = true
	m.CallCount++
}

func TestAutopilotRunner_Execute(t *testing.T) {
	mockEngine := &MockEngine{}
	runner := NewAutopilotRunner(mockEngine)

	task := Task{ID: "t1", Command: "cmd1"}
	root := NewPlanNode("root", NodeTask, &task)
	plan := NewAutopilotPlan("p1", root)

	err := runner.ExecutePlan(context.Background(), plan)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if mockEngine.CallCount != 1 {
		t.Errorf("Expected 1 call, got %d", mockEngine.CallCount)
	}
}

func TestAutopilotRunner_ExecuteWithVerification(t *testing.T) {
	mockEngine := &MockEngine{}
	runner := NewAutopilotRunner(mockEngine)

	vTask := Task{ID: "v1", Command: "verify"}
	task := Task{ID: "t1", Command: "cmd1", VerificationTask: &vTask}
	root := NewPlanNode("root", NodeTask, &task)
	plan := NewAutopilotPlan("p1", root)

	err := runner.ExecutePlan(context.Background(), plan)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if mockEngine.CallCount != 2 {
		t.Errorf("Expected 2 calls, got %d", mockEngine.CallCount)
	}
}
