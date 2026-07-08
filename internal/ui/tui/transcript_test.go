package tui

import (
	"testing"
)

func TestTranscriptEntry_OperatorTextFormatting(t *testing.T) {
	entry := TranscriptEntry{
		Kind: TranscriptOperator,
		Text: "Check nginx status",
	}

	formatted := entry.String()
	expected := "> Check nginx status"
	if formatted != expected {
		t.Fatalf("expected %q, got %q", expected, formatted)
	}

	if !entry.OperatorText() {
		t.Fatal("expected OperatorText() to return true")
	}
	if entry.AgentText() {
		t.Fatal("expected AgentText() to return false")
	}
	if entry.SystemText() {
		t.Fatal("expected SystemText() to return false")
	}
}

func TestTranscriptEntry_AgentTextFormatting(t *testing.T) {
	entry := TranscriptEntry{
		Kind: TranscriptAgent,
		Text: "I'll check nginx on all webservers",
	}

	formatted := entry.String()
	expected := "🤖 I'll check nginx on all webservers"
	if formatted != expected {
		t.Fatalf("expected %q, got %q", expected, formatted)
	}

	if entry.OperatorText() {
		t.Fatal("expected OperatorText() to return false")
	}
	if !entry.AgentText() {
		t.Fatal("expected AgentText() to return true")
	}
	if entry.SystemText() {
		t.Fatal("expected SystemText() to return false")
	}
}

func TestTranscriptEntry_SystemTextFormatting(t *testing.T) {
	entry := TranscriptEntry{
		Kind: TranscriptSystem,
		Text: "Connecting to servers...",
	}

	formatted := entry.String()
	expected := "ℹ️  Connecting to servers..."
	if formatted != expected {
		t.Fatalf("expected %q, got %q", expected, formatted)
	}

	if entry.OperatorText() {
		t.Fatal("expected OperatorText() to return false")
	}
	if entry.AgentText() {
		t.Fatal("expected AgentText() to return false")
	}
	if !entry.SystemText() {
		t.Fatal("expected SystemText() to return true")
	}
}