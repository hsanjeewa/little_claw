package agent

import (
	"testing"
)

func TestRunbook_Creation(t *testing.T) {
	plan := NewAutopilotPlan("p1", NewPlanNode("root", NodeGoal, nil))
	runbook := NewRunbook("r1", "daily-backup", "0 0 * * *", ModePreApproved, plan)

	if runbook.ApprovalMode != ModePreApproved {
		t.Errorf("Expected PRE_APPROVED, got %s", runbook.ApprovalMode)
	}
}
