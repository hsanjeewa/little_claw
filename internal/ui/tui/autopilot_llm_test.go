package tui

import (
	"context"
	"os"
	"testing"

	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestLLMIntegration_RealLLM(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real LLM integration test in short mode")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	apiKey := os.Getenv("OPENAI_API_KEY")
	model := os.Getenv("LLM_MODEL")

	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping real LLM test")
	}

	client := llm.NewLocalOpenAIClient(baseURL, apiKey, model)

	plan, reasoning, err := client.GeneratePlan(testCtx(), "Check nginx status", "web-server", "Linux SSH access", "Safe read-only operations")
	if err != nil {
		t.Fatalf("GeneratePlan failed: %v", err)
	}

	if len(plan) == 0 {
		t.Fatal("LLM returned empty plan")
	}

	if reasoning == "" {
		t.Error("LLM returned empty reasoning")
	}

	for i, step := range plan {
		if step.Description == "" {
			t.Errorf("Step %d has empty description", i)
		}
		if step.Command == "" {
			t.Errorf("Step %d has empty command", i)
		}
	}
}

func TestLLMIntegration_InvalidAPIKey(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real LLM integration test in short mode")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	model := os.Getenv("LLM_MODEL")

	client := llm.NewLocalOpenAIClient(baseURL, "invalid-key", model)

	_, _, err := client.GeneratePlan(testCtx(), "Check nginx status", "web-server", "Linux SSH access", "Safe read-only operations")
	if err == nil {
		t.Fatal("Expected error with invalid API key, got nil")
	}
}

func TestLLMIntegration_ConfigHandling(t *testing.T) {
	baseURL := "https://openrouter.ai/api/v1"
	apiKey := "test-key"
	model := "qwen/qwen-2.5-coder-32b-instruct"

	client := llm.NewLocalOpenAIClient(baseURL, apiKey, model)

	if client == nil {
		t.Fatal("NewLocalOpenAIClient returned nil")
	}
}

func testCtx() context.Context {
	return context.Background()
}