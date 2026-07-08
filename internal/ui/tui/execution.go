package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/llm"
)

type TaskCompletedMsg struct {
	StepIndex int
	Output    string
}

type TaskFailedMsg struct {
	StepIndex int
	Error     string
}

type ExecutionProgress struct {
	CurrentHost string
	CurrentStep int
	TotalSteps  int
	TotalHosts  int
}

type ExecutionProgressMsg struct {
	Progress ExecutionProgress
}

func generateExecutionOrder(plan []llm.PlanStep, hosts []string) []string {
	var order []string
	for _, host := range hosts {
		for _, step := range plan {
			order = append(order, fmt.Sprintf("%s: %s", host, step.Description))
		}
	}
	return order
}

func RunPlanStep(ctx context.Context, stepIndex int, task agent.Task, executor agent.CommandExecutor) tea.Cmd {
	return func() tea.Msg {
		output, err := executor.Execute(ctx, task)
		if err != nil {
			return TaskFailedMsg{StepIndex: stepIndex, Error: err.Error()}
		}
		return TaskCompletedMsg{StepIndex: stepIndex, Output: output}
	}
}

func (m AutopilotModel) executePlan(plan []llm.PlanStep, hosts []string) tea.Cmd {
	if m.run.State != RunStateExecuting {
		return nil
	}

	executionOrder := generateExecutionOrder(plan, hosts)

	progressChan := make(chan ExecutionProgress)
	go func() {
		for i, taskDesc := range executionOrder {
			parts := strings.SplitN(taskDesc, ": ", 2)
			host := parts[0]
			stepDesc := parts[1]

			progressChan <- ExecutionProgress{
				CurrentHost: host,
				CurrentStep: i,
				TotalSteps:  len(executionOrder),
				TotalHosts:  len(hosts),
			}

			if i%len(plan) == 0 {
				m.transcript = append(m.transcript, TranscriptEntry{
					Kind: TranscriptSystem,
					Text: fmt.Sprintf("Starting execution on %s", host),
				}.String())
			}
			m.transcript = append(m.transcript, TranscriptEntry{
				Kind: TranscriptSystem,
				Text: fmt.Sprintf("Step completed: %s on %s", stepDesc, host),
			}.String())

		}

		m.run.State = RunStateCompleted
	}()

	return func() tea.Msg {
		return ExecutionProgressMsg{Progress: <-progressChan}
	}
}
