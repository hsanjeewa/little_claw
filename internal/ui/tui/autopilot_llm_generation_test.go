package tui

import (
	"os"
	"testing"

	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestAutopilotModel_handlePlanGeneration_UsesLLMClient(t *testing.T) {
	if os.Getenv("OPENAI_API_KEY") == "" || os.Getenv("LLM_MODEL") == "" {
		t.Skip("Skipping LLM integration test - OPENAI_API_KEY or LLM_MODEL not set")
	}

	llmClient := &LLMClient{
		client: llm.NewLocalOpenAIClient(
			os.Getenv("OPENAI_BASE_URL"),
			os.Getenv("OPENAI_API_KEY"),
			os.Getenv("LLM_MODEL"),
		),
		config: LLMConfig{
			BaseURL: os.Getenv("OPENAI_BASE_URL"),
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			Model:   os.Getenv("LLM_MODEL"),
		},
	}

	model := NewAutopilotModelWithLLMClient(llmClient)

	cmd := model.handlePlanGeneration("test goal")

	if cmd == nil {
		t.Fatal("expected handlePlanGeneration to return a command")
	}

	msg := cmd()
	planMsg, ok := msg.(PlanGeneratedMsg)
	if !ok {
		t.Fatal("expected PlanGeneratedMsg message")
	}

	if planMsg.Error != nil {
		t.Fatalf("expected no error, got %v", planMsg.Error)
	}

	if len(planMsg.Plan) == 0 {
		t.Fatal("expected non-empty plan from LLM")
	}
}