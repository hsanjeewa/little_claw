package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestAutopilotRun_InitializesInDraftingState(t *testing.T) {
	m := NewAutopilotModel()

	if m.run.State != RunStateDrafting {
		t.Fatalf("expected initial state to be Drafting, got %s", m.run.State)
	}
}

func TestAutopilotRun_TransitionsToExecutingAfterApproval(t *testing.T) {
	m := NewAutopilotModel()

	planSteps := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: planSteps}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	if m.run.State != RunStateReady {
		t.Fatalf("expected state Ready after plan generation, got %s", m.run.State)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(AutopilotModel)

	approveCmd := cmd()
	updated, _ = m.Update(approveCmd)
	m = updated.(AutopilotModel)

	if m.run.State != RunStateExecuting {
		t.Fatalf("expected state Executing after approval, got %s", m.run.State)
	}

	if !m.run.Approved {
		t.Fatal("expected plan to be marked approved")
	}
}

func TestAutopilotRun_RejectionReturnsToDrafting(t *testing.T) {
	m := NewAutopilotModel()

	planSteps := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}

	planMsg := PlanGeneratedMsg{Plan: planSteps}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(AutopilotModel)

	rejectCmd := cmd()
	updated, _ = m.Update(rejectCmd)
	m = updated.(AutopilotModel)

	if m.run.State != RunStateDrafting {
		t.Fatalf("expected state Drafting after rejection, got %s", m.run.State)
	}

	if len(m.run.Plan) != 0 {
		t.Fatal("expected plan to be cleared after rejection")
	}

	if m.run.Approved {
		t.Fatal("expected plan to not be approved")
	}
}