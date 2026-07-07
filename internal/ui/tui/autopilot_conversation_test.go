package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAutopilotCommandBar_AcceptsNaturalLanguageGoals(t *testing.T) {
	m := NewAutopilotModel()
	m.width = 80
	m.height = 24

	goal := "Check nginx status on webservers"
	for _, r := range goal {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(AutopilotModel)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(AutopilotModel)

	if cmd == nil {
		t.Fatal("expected Enter to return command")
	}

	msg := cmd()
	if _, ok := msg.(GeneratePlanMsg); !ok {
		t.Fatal("expected GeneratePlanMsg")
	}

	updated, _ = m.Update(msg)
	m = updated.(AutopilotModel)

	if len(m.transcript) != 1 {
		t.Fatalf("expected 1 transcript entry, got %d", len(m.transcript))
	}

	lastMessage := m.transcript[len(m.transcript)-1]
	expectedPrefix := "> " + goal
	if lastMessage != expectedPrefix {
		t.Fatalf("expected transcript entry %q, got %q", expectedPrefix, lastMessage)
	}

	if m.commandInput.Value() != "" {
		t.Fatalf("expected command input to be reset after submission, got %q", m.commandInput.Value())
	}
}

func TestAutopilotConversationPane_DisplaysMessageHistory(t *testing.T) {
	m := NewAutopilotModel()
	m.width = 80
	m.height = 24

	goals := []string{"Check nginx status", "Verify disk space on DB servers"}
	for _, goal := range goals {
		for _, r := range goal {
			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			m = updated.(AutopilotModel)
		}
		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = updated.(AutopilotModel)

		if cmd == nil {
			t.Fatal("expected Enter to return command")
		}

		msg := cmd()
		updated, _ = m.Update(msg)
		m = updated.(AutopilotModel)

		planMsg := m.GeneratePlan(goal)
		planResult := planMsg()
		updated, _ = m.Update(planResult)
		m = updated.(AutopilotModel)
	}

	if len(m.transcript) < len(goals)*2 {
		t.Fatalf("expected at least %d messages in transcript (including reasoning), got %d", len(goals)*2, len(m.transcript))
	}

	view := m.View()
	for _, goal := range goals {
		words := strings.Fields(goal)
		foundCount := 0
		for _, word := range words {
			if strings.Contains(view, word) {
				foundCount++
			}
		}
		if foundCount < len(words)/2 {
			t.Fatalf("expected view to contain content from goal %q (found %d/%d words), got:\n%s", goal, foundCount, len(words), view)
		}
	}
}

func TestAutopilotSlashCommands_AreNotAddedToTranscript(t *testing.T) {
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
		t.Fatal("expected slash command to return a command")
	}

	if strings.Contains(strings.Join(m.transcript, "\n"), "/targets") {
		t.Fatal("expected slash command not to be added to transcript")
	}
}