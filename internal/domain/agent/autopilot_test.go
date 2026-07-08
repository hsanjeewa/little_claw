package agent

import (
	"testing"
)

func TestAutopilotPlan_Initialization(t *testing.T) {
	root := NewPlanNode("root", NodeGoal, nil)
	plan := NewAutopilotPlan("plan-1", root)

	if plan.Root.ID != "root" {
		t.Errorf("Expected root ID root, got %s", plan.Root.ID)
	}
	if plan.Status != StatusPending {
		t.Errorf("Expected status PENDING, got %s", plan.Status)
	}
}
