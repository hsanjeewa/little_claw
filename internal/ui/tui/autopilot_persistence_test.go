package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestAutopilotRun_ApprovedPlanPersists(t *testing.T) {
	m := NewAutopilotModel()

	planSteps := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
		{Description: "Restart nginx", Command: "systemctl restart nginx", IsMutative: true},
	}

	planMsg := PlanGeneratedMsg{Plan: planSteps}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	if len(m.run.Plan) != len(planSteps) {
		t.Fatalf("expected plan to be persisted in run state, got %d steps", len(m.run.Plan))
	}

	for i, step := range m.run.Plan {
		if step.Description != planSteps[i].Description {
			t.Fatalf("step %d description mismatch: expected %s, got %s", i, planSteps[i].Description, step.Description)
		}
	}
}

func TestAutopilotRun_RejectionClearsPlan(t *testing.T) {
	m := NewAutopilotModel()

	planSteps := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: planSteps}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	if len(m.run.Plan) == 0 {
		t.Fatal("expected plan to be persisted after generation")
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(AutopilotModel)

	rejectCmd := cmd()
	updated, _ = m.Update(rejectCmd)
	m = updated.(AutopilotModel)

	if len(m.run.Plan) != 0 {
		t.Fatal("expected plan to be cleared after rejection")
	}
}