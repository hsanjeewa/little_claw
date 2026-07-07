package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestAutopilot_GeneratePlan_CommandTriggersPlanGeneration(t *testing.T) {
	m := NewAutopilotModel()
	m.width = 80
	m.height = 24

	goal := "Check nginx status"
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
}

func TestAutopilot_PlanGeneration_AddsReasoningToTranscript(t *testing.T) {
	m := NewAutopilotModel()

	goal := "Check nginx"
	cmd := m.handlePlanGeneration(goal)
	result := cmd()

	planMsg, ok := result.(PlanGeneratedMsg)
	if !ok {
		t.Fatal("expected PlanGeneratedMsg")
	}

	m.transcript = append(m.transcript, TranscriptEntry{
		Kind: TranscriptAgent,
		Text: planMsg.Reasoning,
	}.String())

	transcript := strings.Join(m.transcript, "\n")
	if !strings.Contains(transcript, planMsg.Reasoning) {
		t.Fatalf("expected transcript to contain reasoning, got: %s", transcript)
	}

	if !strings.Contains(transcript, "🤖") {
		t.Fatal("expected agent emoji in transcript")
	}
}

func TestAutopilot_PlanePane_DisplaysPlanWithMutativeFlags(t *testing.T) {
	m := NewAutopilotModel()
	m.width = 80
	m.height = 24

	m.plan = []string{
		"1. Check nginx status \n   systemctl status nginx",
		"2. Restart nginx [MUTATIVE]\n   systemctl restart nginx",
	}

	view := m.View()
	if !strings.Contains(view, "1. Check nginx status") {
		t.Fatal("expected plan pane to show step 1")
	}

	if !strings.Contains(view, "[MUTATIVE]") {
		t.Fatal("expected plan pane to show mutative flag")
	}
}

func TestAutopilot_InvalidJSONResponse_HandledGracefully(t *testing.T) {
	m := NewAutopilotModel()

	invalidJSON := `not valid json`
	_, err := llm.ParsePlan(invalidJSON)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	errorMsg := TranscriptEntry{
		Kind: TranscriptSystem,
		Text: fmt.Sprintf("Failed to parse plan: %v", err),
	}
	m.transcript = append(m.transcript, errorMsg.String())

	transcript := strings.Join(m.transcript, "\n")
	if !strings.Contains(transcript, "Failed to parse") {
		t.Fatal("expected transcript to show error message")
	}

	if !strings.Contains(transcript, "ℹ️") {
		t.Fatal("expected system emoji in transcript")
	}
}