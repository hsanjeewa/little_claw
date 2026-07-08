package tui

import (
	"testing"

	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestRecovery_AutoTriggersOnFailure(t *testing.T) {
	m := NewAutopilotModel()

	plan := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: plan}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	updated, _ = m.Update(PlanApprovedMsg{})
	m = updated.(AutopilotModel)

	if m.run.State != RunStateExecuting {
		t.Fatal("expected Executing state before failure")
	}

	failureMsg := TaskFailedMsg{
		StepIndex: 0,
		Error:     "Connection refused",
	}

	updated, cmd := m.Update(failureMsg)
	m = updated.(AutopilotModel)

	t.Logf("After TaskFailedMsg: State=%s, RetryCount=%d, MaxRetries=%d, cmd=%v", m.run.State, m.run.RetryCount, m.run.MaxRetries, cmd != nil)

	if cmd != nil {
		recoveryMsg := cmd()
		t.Logf("Recovery command executed, msg=%T", recoveryMsg)
		if recoveryMsg != nil {
			updated, _ = m.Update(recoveryMsg)
			m = updated.(AutopilotModel)
			t.Logf("After RecoveryPlanGeneratedMsg: State=%s", m.run.State)
		}
	} else {
		t.Logf("No command returned from Update, checking if recovery was triggered")
		t.Logf("Plan length: %d, LastCompletedStep: %d", len(m.run.Plan), m.run.LastCompletedStep)
	}

	if m.run.State != RunStateReady {
		t.Fatalf("expected Ready state after recovery plan generation, got %s", m.run.State)
	}

	if m.run.RetryCount != 1 {
		t.Fatalf("expected RetryCount to be 1, got %d", m.run.RetryCount)
	}

	if m.run.OriginalError != "Connection refused" {
		t.Fatalf("expected OriginalError to be 'Connection refused', got %s", m.run.OriginalError)
	}
}

func TestRecovery_EnforcesMaxRetries(t *testing.T) {
	m := NewAutopilotModel()
	m.run.MaxRetries = 2

	plan := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: plan}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	m.run.MaxRetries = 2

	updated, _ = m.Update(PlanApprovedMsg{})
	m = updated.(AutopilotModel)

	m.run.RetryCount = 2

	failureMsg := TaskFailedMsg{
		StepIndex: 0,
		Error:     "Connection refused",
	}

	updated, _ = m.Update(failureMsg)
	m = updated.(AutopilotModel)

	if m.run.State != RunStateFailed {
		t.Fatalf("expected Failed state after max retries, got %s", m.run.State)
	}
}

func TestRecovery_GeneratesContextualPlan(t *testing.T) {
	m := NewAutopilotModel()

	plan := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: plan}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	updated, _ = m.Update(PlanApprovedMsg{})
	m = updated.(AutopilotModel)

	failureMsg := TaskFailedMsg{
		StepIndex: 0,
		Error:     "Connection refused",
	}

	updated, _ = m.Update(failureMsg)
	m = updated.(AutopilotModel)

	recoveryPlanMsg := RecoveryPlanGeneratedMsg{
		Plan: m.generateRecoveryPlan("Connection refused"),
	}

	updated, _ = m.Update(recoveryPlanMsg)
	m = updated.(AutopilotModel)

	if m.run.State != RunStateReady {
		t.Fatalf("expected Ready state after recovery plan generation, got %s", m.run.State)
	}

	if len(m.run.RecoveryPlan) == 0 {
		t.Fatal("expected non-empty recovery plan")
	}

	if len(m.plan) == 0 {
		t.Fatal("expected plan pane to be populated with recovery steps")
	}
}

func TestRecovery_PlanDisplaysInPane(t *testing.T) {
	m := NewAutopilotModel()

	plan := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: plan}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	updated, _ = m.Update(PlanApprovedMsg{})
	m = updated.(AutopilotModel)

	failureMsg := TaskFailedMsg{
		StepIndex: 0,
		Error:     "Connection refused",
	}

	updated, _ = m.Update(failureMsg)
	m = updated.(AutopilotModel)

	recoveryPlanMsg := RecoveryPlanGeneratedMsg{
		Plan: []llm.PlanStep{
			{Description: "Check network connectivity", Command: "ping -c 3 8.8.8.8", IsMutative: false},
			{Description: "Verify SSH service", Command: "systemctl status sshd", IsMutative: false},
		},
	}

	updated, _ = m.Update(recoveryPlanMsg)
	m = updated.(AutopilotModel)

	planContent := m.plan
	if len(planContent) == 0 {
		t.Fatal("expected non-empty plan content")
	}

	hasRecoveryHeader := false
	for _, line := range planContent {
		if contains(line, "RECOVERY PLAN") {
			hasRecoveryHeader = true
			break
		}
	}

	if !hasRecoveryHeader {
		t.Fatal("expected plan pane to contain 'RECOVERY PLAN' header")
	}
}

func TestRecovery_ApprovalExecutesPlan(t *testing.T) {
	m := NewAutopilotModel()

	plan := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: plan}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	updated, _ = m.Update(PlanApprovedMsg{})
	m = updated.(AutopilotModel)

	failureMsg := TaskFailedMsg{
		StepIndex: 0,
		Error:     "Connection refused",
	}

	updated, _ = m.Update(failureMsg)
	m = updated.(AutopilotModel)

	recoveryPlanMsg := RecoveryPlanGeneratedMsg{
		Plan: []llm.PlanStep{
			{Description: "Check network connectivity", Command: "ping -c 3 8.8.8.8", IsMutative: false},
		},
	}

	updated, _ = m.Update(recoveryPlanMsg)
	m = updated.(AutopilotModel)

	if m.run.State != RunStateReady {
		t.Fatalf("expected Ready state before recovery approval, got %s", m.run.State)
	}

	updated, _ = m.Update(RecoveryApprovedMsg{})
	m = updated.(AutopilotModel)

	if m.run.State != RunStateExecuting {
		t.Fatalf("expected Executing state after recovery approval, got %s", m.run.State)
	}

	if len(m.run.Plan) != 1 {
		t.Fatalf("expected 1 recovery step in Plan, got %d", len(m.run.Plan))
	}
}

func TestRecovery_RejectionFailsRun(t *testing.T) {
	m := NewAutopilotModel()

	plan := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: plan}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	updated, _ = m.Update(PlanApprovedMsg{})
	m = updated.(AutopilotModel)

	failureMsg := TaskFailedMsg{
		StepIndex: 0,
		Error:     "Connection refused",
	}

	updated, _ = m.Update(failureMsg)
	m = updated.(AutopilotModel)

	recoveryPlanMsg := RecoveryPlanGeneratedMsg{
		Plan: []llm.PlanStep{
			{Description: "Check network connectivity", Command: "ping -c 3 8.8.8.8", IsMutative: false},
		},
	}

	updated, _ = m.Update(recoveryPlanMsg)
	m = updated.(AutopilotModel)

	updated, _ = m.Update(RecoveryRejectedMsg{})
	m = updated.(AutopilotModel)

	if m.run.State != RunStateFailed {
		t.Fatalf("expected Failed state after recovery rejection, got %s", m.run.State)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}