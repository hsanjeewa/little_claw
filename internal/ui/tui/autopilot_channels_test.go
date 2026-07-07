package tui

import (
	"testing"

	"github.com/devops/agent/internal/domain/agent"
)

func TestAutopilotModel_AcceptsExecutionChannels(t *testing.T) {
	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)

	model := NewAutopilotModelWithChannels(taskChan, logChan, hitlChan)

	if model.taskChan == nil {
		t.Fatal("expected AutopilotModel to have taskChan")
	}
	if model.logChan == nil {
		t.Fatal("expected AutopilotModel to have logChan")
	}
	if model.hitlChan == nil {
		t.Fatal("expected AutopilotModel to have hitlChan")
	}

	if model.taskChan != taskChan {
		t.Fatal("expected taskChan to be same instance")
	}
}

func TestAutopilotModel_NewAutopilotModelChannelsBackwardsCompatible(t *testing.T) {
	model := NewAutopilotModel()

	if model.taskChan != nil {
		t.Fatal("expected NewAutopilotModel to have nil taskChan")
	}
	if model.logChan != nil {
		t.Fatal("expected NewAutopilotModel to have nil logChan")
	}
	if model.hitlChan != nil {
		t.Fatal("expected NewAutopilotModel to have nil hitlChan")
	}
}