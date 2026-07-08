package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAutopilotCommandBar_SendsTargetSelectionMessage(t *testing.T) {
	m := NewAutopilotModel()
	m.width = 80
	m.height = 24

	for _, r := range "/targets" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(AutopilotModel)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(AutopilotModel)

	if cmd == nil {
		t.Fatal("expected Enter on /targets to return a command")
	}

	msg := cmd()
	_, ok := msg.(OpenTargetSelectionMsg)
	if !ok {
		t.Fatal("expected /targets command to send OpenTargetSelectionMsg")
	}
}