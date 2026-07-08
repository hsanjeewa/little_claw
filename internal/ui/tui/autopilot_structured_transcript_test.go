package tui

import (
	"strings"
	"testing"
)

func TestAutopilotTranscript_SupportsStructuredMessages(t *testing.T) {
	m := NewAutopilotModel()
	m.width = 80
	m.height = 24

	agentMsg := TranscriptEntry{
		Kind: TranscriptAgent,
		Text: "I've identified 3 webservers",
	}
	sysMsg := TranscriptEntry{
		Kind: TranscriptSystem,
		Text: "Connecting to hosts...",
	}

	m.transcript = append(m.transcript, agentMsg.String())
	m.transcript = append(m.transcript, sysMsg.String())

	view := m.View()
	if !strings.Contains(view, "🤖") {
		t.Fatal("expected view to contain agent emoji")
	}
	if !strings.Contains(view, "ℹ️") {
		t.Fatal("expected view to contain system emoji")
	}
}