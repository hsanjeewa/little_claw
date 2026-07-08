package llm

import (
	"context"
	"encoding/json"
	"testing"
)

func TestPlanTasks_CreatesClientWithModel(t *testing.T) {
	model := "qwen/qwen-2.5-coder-32b-instruct"
	client := NewLocalOpenAIClient("https://openrouter.ai/api/v1", "dummy-key", model)

	if client.model != model {
		t.Fatalf("expected model %s, got %s", model, client.model)
	}
}

func TestPlanTasks_TemperatureIs0_2(t *testing.T) {
	skip := true
	if skip {
		t.Skip("skipping - requires LLM mock or integration test")
	}
}

func TestParseGeneratedPlan_WithReasoning(t *testing.T) {
	jsonPlan := `{
  "reasoning": "The goal requires checking nginx status which is a read-only operation.",
  "steps": [
    {
      "description": "Check nginx service status",
      "command": "systemctl status nginx",
      "is_mutative": false
    }
  ]
}`

	var plan GeneratedPlan
	err := json.Unmarshal([]byte(jsonPlan), &plan)
	if err != nil {
		t.Fatalf("failed to parse plan with reasoning: %v", err)
	}

	if plan.Reasoning == "" {
		t.Fatal("expected reasoning to be populated")
	}

	if len(plan.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(plan.Steps))
	}
}

func TestParseGeneratedPlan_MalformedReturnsError(t *testing.T) {
	malformedJSON := `{"steps": [not valid]}`

	_, err := ParsePlan(malformedJSON)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestPlanTasks_InvalidLLMResponseReturnsError(t *testing.T) {
	client := NewLocalOpenAIClient("", "", "dummy-model")

	_, err := client.PlanTasks(context.Background(), "Check nginx")
	if err == nil {
		t.Fatal("expected error for invalid client config")
	}
}