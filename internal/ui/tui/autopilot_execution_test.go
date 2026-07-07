package tui

import (
	"testing"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestAutopilot_ApprovedPlan_ExecutesFirstStep(t *testing.T) {
	executor := &fakeExecutor{output: "active"}
	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)

	model := NewAutopilotModelWithAllDependencies(nil, taskChan, logChan, hitlChan).WithExecutor(executor)
	model.run = AutopilotRun{
		State: RunStateReady,
		Plan: []llm.PlanStep{
			{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
		},
	}

	updated, cmd := model.Update(PlanApprovedMsg{})
	m := updated.(AutopilotModel)

	if m.run.State != RunStateExecuting {
		t.Fatalf("expected Executing state after approval, got %s", m.run.State)
	}
	if cmd == nil {
		t.Fatal("expected a command to execute the first step")
	}

	msg := cmd()
	completed, ok := msg.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}
	if completed.Output != "active" {
		t.Fatalf("expected output %q, got %q", "active", completed.Output)
	}
}

func TestAutopilot_TaskCompleted_ExecutesNextStep(t *testing.T) {
	executor := &fakeExecutor{output: "ok"}
	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)

	model := NewAutopilotModelWithAllDependencies(nil, taskChan, logChan, hitlChan).WithExecutor(executor)
	model.run = AutopilotRun{
		State: RunStateExecuting,
		Plan: []llm.PlanStep{
			{Description: "Step 1", Command: "cmd1", IsMutative: false},
			{Description: "Step 2", Command: "cmd2", IsMutative: false},
		},
		LastCompletedStep: -1,
	}

	updated, cmd := model.Update(TaskCompletedMsg{StepIndex: 0, Output: "ok"})
	m := updated.(AutopilotModel)

	if m.run.LastCompletedStep != 0 {
		t.Fatalf("expected LastCompletedStep 0, got %d", m.run.LastCompletedStep)
	}
	if cmd == nil {
		t.Fatal("expected a command to execute the next step")
	}

	msg := cmd()
	completed, ok := msg.(TaskCompletedMsg)
	if !ok {
		t.Fatalf("expected TaskCompletedMsg, got %T", msg)
	}
	if completed.StepIndex != 1 {
		t.Fatalf("expected StepIndex 1, got %d", completed.StepIndex)
	}
}

func TestAutopilot_FullExecution_RunStateCompletesAfterAllSteps(t *testing.T) {
	executor := &fakeExecutor{output: "ok"}
	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)

	model := NewAutopilotModelWithAllDependencies(nil, taskChan, logChan, hitlChan).WithExecutor(executor)
	model.run = AutopilotRun{
		State: RunStateReady,
		Plan: []llm.PlanStep{
			{Description: "Step 1", Command: "cmd1", IsMutative: false},
			{Description: "Step 2", Command: "cmd2", IsMutative: false},
		},
	}

	updated, cmd := model.Update(PlanApprovedMsg{})
	m := updated.(AutopilotModel)
	if m.run.State != RunStateExecuting {
		t.Fatalf("expected Executing state after approval, got %s", m.run.State)
	}

	msg := cmd()
	completed := msg.(TaskCompletedMsg)
	updated, cmd = m.Update(completed)
	m = updated.(AutopilotModel)
	if m.run.State != RunStateExecuting {
		t.Fatalf("expected Executing state after first step, got %s", m.run.State)
	}
	if m.run.LastCompletedStep != 0 {
		t.Fatalf("expected LastCompletedStep 0, got %d", m.run.LastCompletedStep)
	}

	msg = cmd()
	completed = msg.(TaskCompletedMsg)
	updated, _ = m.Update(completed)
	m = updated.(AutopilotModel)
	if m.run.State != RunStateCompleted {
		t.Fatalf("expected Completed state after all steps, got %s", m.run.State)
	}
	if m.run.LastCompletedStep != 1 {
		t.Fatalf("expected LastCompletedStep 1, got %d", m.run.LastCompletedStep)
	}
}
