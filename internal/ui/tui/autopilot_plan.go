package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/llm"
)

type PlanGenerator interface {
	GeneratePlanWithLLM(ctx context.Context, goal string) ([]llm.PlanStep, string, error)
}

type GeneratePlanMsg struct {
	Goal string
}

type PlanGeneratedMsg struct {
	Plan      []llm.PlanStep
	Reasoning string
	Error     error
}

type GeneratingPlan struct{}

type LLMClient struct {
	client *llm.LocalOpenAIClient
	config LLMConfig
}

type LLMConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

func NewLLMClient(config LLMConfig) *LLMClient {
	client := llm.NewLocalOpenAIClient(config.BaseURL, config.APIKey, config.Model)
	return &LLMClient{
		client: client,
		config: config,
	}
}

func (m AutopilotModel) GeneratePlan(goal string) tea.Cmd {
	return func() tea.Msg {
		return GeneratePlanMsg{Goal: goal}
	}
}

func (llmClient *LLMClient) GeneratePlanWithLLM(ctx context.Context, goal string) ([]llm.PlanStep, string, error) {
	scope := llmClient.getScope()
	capabilities := llmClient.getCapabilities()
	constraints := llmClient.getConstraints()

	plan, reasoning, err := llmClient.client.GeneratePlan(ctx, goal, scope, capabilities, constraints)
	if err != nil {
		return nil, "", fmt.Errorf("LLM plan generation failed: %w", err)
	}

	return plan, reasoning, nil
}

func (llmClient *LLMClient) getScope() string {
	return "Available hosts: localhost"
}

func (llmClient *LLMClient) getCapabilities() string {
	return "Linux systems, SSH access, shell commands"
}

func (llmClient *LLMClient) getConstraints() string {
	return "Operations must be safe, non-destructive, and approved by operator before execution"
}

func (m AutopilotModel) handlePlanGeneration(goal string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if m.llmClient != nil {
			plan, _, err := m.llmClient.GeneratePlanWithLLM(ctx, goal)
			if err != nil {
				return PlanGeneratedMsg{
					Error: fmt.Errorf("LLM request failed: %v", err),
				}
			}
			return PlanGeneratedMsg{
				Plan:      plan,
				Reasoning: "",
				Error:     nil,
			}
		}

		plan := []llm.PlanStep{
			{
				Description: "Analyze request",
				Command:     "echo 'Analyzing request: ' + '" + goal + "'",
				IsMutative:  false,
			},
		}

		return PlanGeneratedMsg{
			Plan:      plan,
			Reasoning: "Plan generated for: " + goal,
			Error:     nil,
		}
	}
}

func (m AutopilotModel) submitPlanTasks() {
	if m.taskChan == nil || len(m.run.Plan) == 0 {
		return
	}

	for _, step := range m.run.Plan {
		task, err := m.buildTaskFromStep(step)
		if err != nil {
			continue
		}
		m.taskChan <- task
	}
}

func (m AutopilotModel) buildTaskFromStep(step llm.PlanStep) (agent.Task, error) {
	hostAlias := m.taskHostAlias
	hostIP := m.taskHostIP
	hostUser := m.taskHostUser
	hostPort := m.taskHostPort
	if hostAlias == "" {
		hostAlias = "localhost"
	}
	if hostIP == "" {
		hostIP = "127.0.0.1"
	}
	if hostUser == "" {
		hostUser = "root"
	}
	if hostPort == 0 {
		hostPort = 22
	}
	return agent.NewTask(
		uuid.New().String(),
		hostAlias,
		hostIP,
		hostPort,
		hostUser,
		step.Command,
		step.IsMutative,
	)
}
