package agent

import (
	"testing"
)

func TestPlanNode_Creation(t *testing.T) {
	node := NewPlanNode("root", NodeGoal, nil)
	if node.ID != "root" {
		t.Errorf("Expected ID root, got %s", node.ID)
	}
	if node.Type != NodeGoal {
		t.Errorf("Expected Type GOAL, got %s", node.Type)
	}
}

func TestPlanNode_Disable(t *testing.T) {
	task := Task{ID: "t1"}
	node := NewPlanNode("t1", NodeTask, &task)
	node.Enabled = true
	node.Disable()

	if node.Enabled {
		t.Error("Expected node to be disabled")
	}
}
