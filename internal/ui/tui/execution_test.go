package tui

import (
	"context"
	"testing"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/llm"
)

type fakeExecutor struct {
	output string
	err    error
	calls  []agent.Task
}

func (f *fakeExecutor) Execute(ctx context.Context, task agent.Task) (string, error) {
	f.calls = append(f.calls, task)
	return f.output, f.err
}

func TestExecution_HostFirstSequentialOrder(t *testing.T) {
	plan := []llm.PlanStep{
		{Description: "Step 1", Command: "cmd1", IsMutative: false},
		{Description: "Step 2", Command: "cmd2", IsMutative: false},
	}

	hosts := []string{"host-a", "host-b"}

	expectedOrder := []string{
		"host-a: Step 1",
		"host-a: Step 2",
		"host-b: Step 1",
		"host-b: Step 2",
	}

	actualOrder := generateExecutionOrder(plan, hosts)

	for i, expected := range expectedOrder {
		if i >= len(actualOrder) {
			t.Fatalf("expected at least %d execution steps, got %d", len(expectedOrder), len(actualOrder))
		}
		if actualOrder[i] != expected {
			t.Fatalf("step %d: expected %s, got %s", i, expected, actualOrder[i])
		}
	}
}

func TestExecution_AbortsOnFirstFailure(t *testing.T) {
	m := NewAutopilotModel()

	plan := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: plan}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	updated, _ = m.Update(PlanApprovedMsg{})
	m = updated.(AutopilotModel)

	if m.run.State != RunStateExecuting {
		t.Fatal("expected Executing state after approval")
	}

	failureMsg := TaskFailedMsg{
		StepIndex: 0,
		Error:     "Connection refused",
	}

	updated, _ = m.Update(failureMsg)
	m = updated.(AutopilotModel)

	if m.run.State != RunStateDrafting {
		t.Fatalf("expected Drafting state (generating recovery), got %s", m.run.State)
	}
}

func TestExecution_TransitionsToCompletedOnSuccess(t *testing.T) {
	m := NewAutopilotModel()

	plan := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: plan}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	updated, _ = m.Update(PlanApprovedMsg{})
	m = updated.(AutopilotModel)

	if m.run.State != RunStateExecuting {
		t.Fatal("expected Executing state after approval")
	}

	successMsg := TaskCompletedMsg{
		StepIndex: 0,
		Output:    "nginx is running",
	}

	updated, _ = m.Update(successMsg)
	m = updated.(AutopilotModel)

	if m.run.State != RunStateCompleted {
		t.Fatalf("expected Completed state after all steps succeed, got %s", m.run.State)
	}
}

func TestExecution_ProgressUpdatesDuringExecution(t *testing.T) {
	m := NewAutopilotModel()

	plan := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: plan}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	updated, _ = m.Update(PlanApprovedMsg{})
	m = updated.(AutopilotModel)

	if m.run.State != RunStateExecuting {
		t.Fatal("expected Executing state after approval")
	}

	progressMsg := ExecutionProgressMsg{
		Progress: ExecutionProgress{
			CurrentHost: "host-a",
			CurrentStep: 0,
			TotalSteps:  2,
			TotalHosts:  2,
		},
	}

	updated, _ = m.Update(progressMsg)
	m = updated.(AutopilotModel)

	if m.run.CurrentHost != "host-a" {
		t.Fatalf("expected CurrentHost to be 'host-a', got %s", m.run.CurrentHost)
	}

	if m.run.CurrentStep != 0 {
		t.Fatalf("expected CurrentStep to be 0, got %d", m.run.CurrentStep)
	}
}

func TestRunPlanStep_ExecutesTaskAndReturnsCompletedMessage(t *testing.T) {
	ctx := context.Background()
	task := agent.Task{
		ID:        "task-1",
		HostAlias: "host-a",
		Command:   "systemctl status nginx",
	}
	executor := &fakeExecutor{output: "active"}

	cmd := RunPlanStep(ctx, 0, task, executor)
	msg := cmd()

	if len(executor.calls) != 1 {
		t.Fatalf("expected 1 executor call, got %d", len(executor.calls))
	}
	if executor.calls[0].Command != "systemctl status nginx" {
		t.Fatalf("expected command %q, got %q", "systemctl status nginx", executor.calls[0].Command)
	}

	completed, ok := msg.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}
	if completed.StepIndex != 0 {
		t.Fatalf("expected StepIndex 0, got %d", completed.StepIndex)
	}
	if completed.Output != "active" {
		t.Fatalf("expected output %q, got %q", "active", completed.Output)
	}
}
