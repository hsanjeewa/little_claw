package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestAutopilotPlanApproval_DisplaysInPlanPane(t *testing.T) {
	m := NewAutopilotModel()
	m.width = 80
	m.height = 24

	planSteps := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
		{Description: "Restart nginx", Command: "systemctl restart nginx", IsMutative: true},
	}

	planMsg := PlanGeneratedMsg{Plan: planSteps}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	view := m.View()
	if !strings.Contains(view, "[APPROVE]") {
		t.Fatal("expected approval UI to show [APPROVE] option")
	}
	if !strings.Contains(view, "[REJECT]") {
		t.Fatal("expected approval UI to show [REJECT] option")
	}
}

func TestAutopilotPlanApproval_KeyboardShortcuts(t *testing.T) {
	m := NewAutopilotModel()
	m.width = 80
	m.height = 24

	planSteps := []llm.PlanStep{
		{Description: "Check nginx", Command: "systemctl status nginx", IsMutative: false},
	}
	planMsg := PlanGeneratedMsg{Plan: planSteps}
	updated, _ := m.Update(planMsg)
	m = updated.(AutopilotModel)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(AutopilotModel)

	if cmd == nil {
		t.Fatal("expected 'a' key to return command for approve action")
	}

	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = updated.(AutopilotModel)

	if cmd == nil {
		t.Fatal("expected 'r' key to return command for reject action")
	}
}