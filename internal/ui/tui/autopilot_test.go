package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/domain/agent"
)

func TestAutopilotModel_ViewRendersCommandBarAndPanes(t *testing.T) {
	m := NewAutopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AutopilotModel)

	view := m.View()
	if !strings.Contains(view, commandBarLabel) {
		t.Fatalf("expected Autopilot view to render command bar label, got:\n%s", view)
	}
	if !strings.Contains(view, "PLAN") {
		t.Fatalf("expected Autopilot view to render plan pane label, got:\n%s", view)
	}
	if !strings.Contains(view, "TRANSCRIPT") {
		t.Fatalf("expected Autopilot view to render transcript pane label, got:\n%s", view)
	}
}

func TestAutopilotModel_CommandBarAcceptsInput(t *testing.T) {
	m := NewAutopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AutopilotModel)

	for _, r := range "deploy web" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(AutopilotModel)
	}

	view := m.View()
	if !strings.Contains(view, "deploy web") {
		t.Fatalf("expected Autopilot command bar to display typed input, got:\n%s", view)
	}
}

func TestAutopilotModel_StoresWindowSize(t *testing.T) {
	m := NewAutopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AutopilotModel)

	if m.width != 80 {
		t.Fatalf("expected width 80, got %d", m.width)
	}
	if m.height != 24 {
		t.Fatalf("expected height 24, got %d", m.height)
	}
}

func TestAutopilotModel_HandoffCompactViewFitsBoundsAndShowsLabels(t *testing.T) {
	m := NewAutopilotModel()
	m = m.ApplyWatchtowerEscalation(WatchtowerEscalationPayload{
		Target:       ModeAutopilot,
		MetricFamily: agent.MetricFamilyMemory,
		SelectedHost: "db-master",
		Observation:  "Memory 85.0% used",
	})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 15})
	m = updated.(AutopilotModel)

	view := m.View()
	assertRenderedWithinBounds(t, view, 80, 14)
	for _, label := range []string{"WATCHTOWER HANDOFF", "db-master", "PLAN", "TRANSCRIPT", commandBarLabel} {
		if !strings.Contains(view, label) {
			t.Fatalf("expected compact escalated Autopilot view to contain %q, got:\n%s", label, view)
		}
	}
}

func TestAutopilotModel_PaneLabelsVisibleAfterCompactResize(t *testing.T) {
	m := NewAutopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AutopilotModel)
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 15})
	m = updated.(AutopilotModel)

	view := m.View()
	assertRenderedWithinBounds(t, view, 80, 14)
	for _, label := range []string{"PLAN", "TRANSCRIPT", commandBarLabel} {
		if !strings.Contains(view, label) {
			t.Fatalf("expected resized compact Autopilot view to contain %q, got:\n%s", label, view)
		}
	}
}

func TestAutopilotModel_FocusCyclesOnTab(t *testing.T) {
	m := NewAutopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AutopilotModel)

	initialFocus := m.focusedPane
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(AutopilotModel)

	if m.focusedPane == initialFocus {
		t.Fatalf("expected focus to change after Tab, still at %d", m.focusedPane)
	}
}

func TestAutopilotModel_DoesNotQuitOnQ(t *testing.T) {
	m := NewAutopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AutopilotModel)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(AutopilotModel)

	if isQuitCmd(cmd) {
		t.Fatal("expected 'q' not to quit in Autopilot")
	}
}

func TestAutopilotModel_QuitsOnSlashExit(t *testing.T) {
	m := NewAutopilotModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AutopilotModel)

	for _, r := range "/exit" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(AutopilotModel)
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !isQuitCmd(cmd) {
		t.Fatal("expected '/exit' command to quit the app")
	}
}
