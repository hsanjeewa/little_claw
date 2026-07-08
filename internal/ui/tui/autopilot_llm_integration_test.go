package tui

import (
	"testing"

	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestAutopilotModel_AcceptsLLMClient(t *testing.T) {
	llmClient := &LLMClient{
		client: llm.NewLocalOpenAIClient("https://test.url", "test-key", "test-model"),
		config: LLMConfig{
			BaseURL: "https://test.url",
			APIKey:  "test-key",
			Model:   "test-model",
		},
	}

	model := NewAutopilotModelWithLLMClient(llmClient)

	if model.llmClient == nil {
		t.Fatal("expected AutopilotModel to have LLMClient field, got nil")
	}

	if model.llmClient != llmClient {
		t.Fatal("expected AutopilotModel to store the provided LLMClient")
	}
}

func TestAutopilotModel_NewAutopilotModelBackwardsCompatible(t *testing.T) {
	// Ensure we don't break existing callers
	model := NewAutopilotModel()

	if model.llmClient != nil {
		t.Fatal("expected NewAutopilotModel to have nil LLMClient")
	}
}