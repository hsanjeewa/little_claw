package agent

import (
	"testing"
)

func TestSkillRegistry_Lookup(t *testing.T) {
	registry := NewSkillRegistry()
	registry.Register(Skill{Name: "tdd", Description: "test driven dev"})
	registry.Register(Skill{Name: "git", Description: "git ops"})

	matches := registry.Lookup("t")
	if len(matches) != 1 || matches[0].Name != "tdd" {
		t.Errorf("Expected tdd, got %v", matches)
	}

	matches = registry.Lookup("z")
	if len(matches) != 0 {
		t.Errorf("Expected 0, got %d", len(matches))
	}
}
