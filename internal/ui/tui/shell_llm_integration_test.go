package tui

import (
	"testing"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestShell_ProvidesLLMClientToAutopilot(t *testing.T) {
	llmClient := &LLMClient{
		client: llm.NewLocalOpenAIClient("", "", ""),
		config: LLMConfig{},
	}

	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)
	initialTasks := []agent.Task{}

	shell := NewShellWithLLMClient(taskChan, logChan, hitlChan, initialTasks, llmClient)

	autopilot, ok := shell.autopilot.(AutopilotModel)
	if !ok {
		t.Fatal("expected shell.autopilot to be AutopilotModel")
	}

	if autopilot.llmClient == nil {
		t.Fatal("expected AutopilotModel to have LLMClient from Shell")
	}

	if autopilot.llmClient != llmClient {
		t.Fatal("expected AutopilotModel to have the same LLMClient instance")
	}
}

func TestShell_ProvidesChannelsToAutopilot(t *testing.T) {
	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)
	initialTasks := []agent.Task{}

	shell := NewShellWithLLMClient(taskChan, logChan, hitlChan, initialTasks, nil)

	autopilot, ok := shell.autopilot.(AutopilotModel)
	if !ok {
		t.Fatal("expected shell.autopilot to be AutopilotModel")
	}

	if autopilot.taskChan != taskChan {
		t.Fatal("expected AutopilotModel to receive taskChan from Shell")
	}
	if autopilot.logChan != logChan {
		t.Fatal("expected AutopilotModel to receive logChan from Shell")
	}
	if autopilot.hitlChan != hitlChan {
		t.Fatal("expected AutopilotModel to receive hitlChan from Shell")
	}
}

func TestShell_NewShellBackwardsCompatible(t *testing.T) {
	taskChan := make(chan agent.Task, 10)
	logChan := make(chan agent.ExecutionLog, 10)
	hitlChan := make(chan agent.HitlRequest, 10)
	initialTasks := []agent.Task{}

	shell := NewShell(taskChan, logChan, hitlChan, initialTasks)

	if shell.autopilot == nil {
		t.Fatal("expected shell to have autopilot")
	}

	autopilot, ok := shell.autopilot.(AutopilotModel)
	if !ok {
		t.Fatal("expected shell.autopilot to be AutopilotModel")
	}

	if autopilot.llmClient != nil {
		t.Fatal("expected NewShell to create AutopilotModel with nil LLMClient")
	}
}