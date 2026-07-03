package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCopilotModel_ViewRendersCommandBarAndPanes(t *testing.T) {
	m := NewCopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(CopilotModel)

	view := m.View()
	if !strings.Contains(view, "COMMAND") {
		t.Fatalf("expected Copilot view to render command bar label, got:\n%s", view)
	}
	if !strings.Contains(view, "TERMINAL") {
		t.Fatalf("expected Copilot view to render terminal pane label, got:\n%s", view)
	}
	if !strings.Contains(view, "ADVISORY") {
		t.Fatalf("expected Copilot view to render advisory pane label, got:\n%s", view)
	}
	if !strings.Contains(view, "GUIDANCE") {
		t.Fatalf("expected Copilot view to render guidance control surface label, got:\n%s", view)
	}
}

func TestCopilotModel_CommandBarAcceptsInput(t *testing.T) {
	m := NewCopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(CopilotModel)

	for _, r := range "check disk" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(CopilotModel)
	}

	view := m.View()
	if !strings.Contains(view, "check disk") {
		t.Fatalf("expected Copilot command bar to display typed input, got:\n%s", view)
	}
}

func TestCopilotModel_StoresWindowSize(t *testing.T) {
	m := NewCopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(CopilotModel)

	if m.width != 80 {
		t.Fatalf("expected width 80, got %d", m.width)
	}
	if m.height != 24 {
		t.Fatalf("expected height 24, got %d", m.height)
	}
}

func TestCopilotModel_FocusCyclesOnTab(t *testing.T) {
	m := NewCopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(CopilotModel)

	initialFocus := m.focusedPane
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(CopilotModel)

	if m.focusedPane == initialFocus {
		t.Fatalf("expected focus to change after Tab, still at %d", m.focusedPane)
	}
}
