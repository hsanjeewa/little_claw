package llm

import (
	"testing"
)

func TestPlanStep_Validate_ValidStepPasses(t *testing.T) {
	step := PlanStep{
		Description: "Check nginx status",
		Command:     "systemctl status nginx",
		IsMutative:  false,
	}

	err := step.Validate()
	if err != nil {
		t.Fatalf("expected valid step to pass validation, got: %v", err)
	}
}

func TestPlanStep_Validate_EmptyDescriptionFails(t *testing.T) {
	step := PlanStep{
		Description: "",
		Command:     "systemctl status nginx",
		IsMutative:  false,
	}

	err := step.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty description")
	}
}

func TestPlanStep_Validate_EmptyCommandFails(t *testing.T) {
	step := PlanStep{
		Description: "Check nginx",
		Command:     "",
		IsMutative:  false,
	}

	err := step.Validate()
	if err == nil {
		t.Fatal("expected validation error for empty command")
	}
}

func TestPlanStep_MutativeStepRequiresVerifyCommand(t *testing.T) {
	step := PlanStep{
		Description: "Restart nginx",
		Command:     "systemctl restart nginx",
		IsMutative:  true,
	}

	err := step.Validate()
	if err != nil {
		t.Fatalf("mutative step without verify_cmd should pass: %v", err)
	}
}

func TestParsePlan_ValidJSONReturnsSteps(t *testing.T) {
	jsonPlan := `[
  {
    "description": "Check nginx",
    "command": "systemctl status nginx",
    "is_mutative": false
  }
]`

	steps, err := ParsePlan(jsonPlan)
	if err != nil {
		t.Fatalf("expected valid JSON to parse, got: %v", err)
	}

	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}

	if steps[0].Description != "Check nginx" {
		t.Fatalf("unexpected description: %s", steps[0].Description)
	}
}

func TestParsePlan_InvalidJSONReturnsError(t *testing.T) {
	jsonPlan := `not valid json`

	_, err := ParsePlan(jsonPlan)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParsePlan_InvalidSchemaReturnsError(t *testing.T) {
	jsonPlan := `[{"description": "step"}]`

	_, err := ParsePlan(jsonPlan)
	if err == nil {
		t.Fatal("expected validation error for missing required fields")
	}
}