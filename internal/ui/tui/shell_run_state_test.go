package tui

import (
	"testing"
)

func TestShell_StatusBadge_ReturnsNAForWatchtower(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	badge := shell.runStateBadge()
	if badge != "N/A" {
		t.Fatalf("expected N/A for Watchtower mode, got %s", badge)
	}
}

func TestShell_StatusBadge_ShowsAutopilotState(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)
	shell.mode = ModeAutopilot

	autopilot := NewAutopilotModel()
	autopilot.run.State = RunStateReady
	shell.autopilot = autopilot

	autopilotModel, ok := shell.autopilot.(AutopilotModel)
	if !ok {
		t.Fatal("failed to cast autopilot to AutopilotModel")
	}

	badge := shell.runStateBadge()
	if badge != "Ready" {
		t.Fatalf("expected Ready badge, got %s", badge)
	}

	autopilotModel.run.State = RunStateExecuting
	shell.autopilot = autopilotModel

	badge = shell.runStateBadge()
	if badge != "Executing" {
		t.Fatalf("expected Executing badge, got %s", badge)
	}
}