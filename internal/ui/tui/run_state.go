package tui

import (
	"github.com/devops/agent/internal/infrastructure/llm"
)

type RunState string

const (
	RunStateDrafting  RunState = "Drafting"
	RunStateReady     RunState = "Ready"
	RunStateExecuting RunState = "Executing"
	RunStateBlocked   RunState = "Blocked"
	RunStateCompleted RunState = "Completed"
	RunStateFailed    RunState = "Failed"
)

type AutopilotRun struct {
	State            RunState
	Plan             []llm.PlanStep
	Approved         bool
	LastCompletedStep int
	CurrentHost      string
	CurrentStep      int
	RetryCount       int
	MaxRetries       int
	OriginalError    string
	RecoveryPlan     []llm.PlanStep
}

type PlanApprovedMsg struct{}

type PlanRejectedMsg struct{}