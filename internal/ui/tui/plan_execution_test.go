package tui

import (
	"testing"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestAutopilot_PlanApprovalSubmitsTasks(t *testing.T) {
	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)

	model := NewAutopilotModelWithChannels(taskChan, logChan, hitlChan)
	model.run = AutopilotRun{
		State: RunStateReady,
		Plan: []llm.PlanStep{
			{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
		},
		Approved: false,
	}

	updated, _ := model.Update(PlanApprovedMsg{})
	m := updated.(AutopilotModel)

	if m.run.State != RunStateExecuting {
		t.Fatalf("expected Executing state after approval, got %s", m.run.State)
	}

	select {
	case task := <-taskChan:
		if task.Command != "systemctl status nginx" {
			t.Fatalf("expected task command %q, got %q", "systemctl status nginx", task.Command)
		}
		if task.IsMutative != false {
			t.Fatalf("expected task IsMutative=false, got %v", task.IsMutative)
		}
	default:
		t.Fatal("expected task to be submitted to taskChan")
	}
}

func TestAutopilot_PlanApprovalWithLLMClientSubmitsTasks(t *testing.T) {
	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)

	llmClient := &LLMClient{
		client: llm.NewLocalOpenAIClient("", "", ""),
		config: LLMConfig{},
	}

	model := NewAutopilotModelWithAllDependencies(llmClient, taskChan, logChan, hitlChan)
	model.run = AutopilotRun{
		State: RunStateReady,
		Plan: []llm.PlanStep{
			{Description: "Check disk", Command: "df -h", IsMutative: false},
		},
		Approved: false,
	}

	updated, _ := model.Update(PlanApprovedMsg{})
	m := updated.(AutopilotModel)

	if m.run.State != RunStateExecuting {
		t.Fatalf("expected Executing state after approval, got %s", m.run.State)
	}

	select {
	case task := <-taskChan:
		if task.Command != "df -h" {
			t.Fatalf("expected task command %q, got %q", "df -h", task.Command)
		}
	default:
		t.Fatal("expected task to be submitted to taskChan")
	}
}