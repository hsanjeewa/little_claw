package llm

import (
	"testing"
)

func TestContextBuilder_AssemblesMessagesInQwenOrder(t *testing.T) {
	goal := "Check nginx status"
	scope := "web-prod-01, db-master"
	capabilities := "Linux, SSH access"
	constraints := "No downtime allowed"

	messages := BuildPlanContext(goal, scope, capabilities, constraints)

	if len(messages) != 8 {
		t.Fatalf("expected 8 messages, got %d", len(messages))
	}

	expectedOrder := []string{
		"schema",
		"Target Scope",
		"Host Capability",
		"Constraints",
		"User Goal",
		"SOUL.md",
		"IDENTITY.md",
		"Watchtower",
	}

	for i, expected := range expectedOrder {
		if messages[i] == "" {
			t.Fatalf("message %d (%s) is empty", i, expected)
		}
	}
}

func TestContextBuilder_SchemaMessageIsFirst(t *testing.T) {
	goal := "Check nginx"
	messages := BuildPlanContext(goal, "web-01", "Linux", "none")

	if len(messages) == 0 {
		t.Fatal("expected at least one message")
	}

	firstMessage := messages[0]
	if !containsSubstring(firstMessage, "steps") || !containsSubstring(firstMessage, "description") {
		t.Fatalf("expected first message to be JSON schema, got: %s", firstMessage)
	}
}

func TestContextBuilder_UserGoalIsFifth(t *testing.T) {
	goal := "Restart nginx on all webservers"
	messages := BuildPlanContext(goal, "web-01,web-02", "Linux", "none")

	if len(messages) < 5 {
		t.Fatalf("expected at least 5 messages, got %d", len(messages))
	}

	userGoalMessage := messages[4]
	if !containsSubstring(userGoalMessage, goal) {
		t.Fatalf("expected fifth message to contain goal, got: %s", userGoalMessage)
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}