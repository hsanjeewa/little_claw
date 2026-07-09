package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/inventory"
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

func (llmClient *LLMClient) GeneratePlanWithLLM(ctx context.Context, goal string, selectedHosts []inventory.TargetHost, watchtowerState WatchtowerStateSnapshot) ([]llm.PlanStep, string, error) {
	scope := llmClient.getScope(selectedHosts)
	capabilities := llmClient.getCapabilities(selectedHosts, watchtowerState)
	constraints := llmClient.getConstraints()
	watchtowerContext := llmClient.getWatchtowerContext(selectedHosts, watchtowerState)

	plan, reasoning, err := llmClient.client.GeneratePlan(ctx, goal, scope, capabilities, constraints, watchtowerContext)
	if err != nil {
		return nil, "", fmt.Errorf("LLM plan generation failed: %w", err)
	}

	return plan, reasoning, nil
}

func (llmClient *LLMClient) getScope(selectedHosts []inventory.TargetHost) string {
	if len(selectedHosts) == 0 {
		return "Available hosts: localhost"
	}

	var hosts []string
	for _, host := range selectedHosts {
		hosts = append(hosts, fmt.Sprintf("%s (%s:%d as %s)", host.Alias, host.IP, host.Port, host.User))
	}
	return "Selected hosts: " + strings.Join(hosts, ", ")
}

func (llmClient *LLMClient) getWatchtowerContext(selectedHosts []inventory.TargetHost, state WatchtowerStateSnapshot) string {
	if len(selectedHosts) == 0 || len(state.MemorySnapshots) == 0 {
		return "No system state available"
	}

	var contextParts []string
	for _, host := range selectedHosts {
		hostContext := fmt.Sprintf("\nHost: %s", host.Alias)
		
		if memorySnapshots, ok := state.MemorySnapshots[host.Alias]; ok && len(memorySnapshots) > 0 {
			latest := memorySnapshots[len(memorySnapshots)-1]
			usedGB := float64(latest.UsedBytes) / (1024 * 1024 * 1024)
			totalGB := float64(latest.TotalBytes) / (1024 * 1024 * 1024)
			hostContext += fmt.Sprintf("\n  Memory: %.1f%% used (%.1fGB/%.1fGB)", 
				latest.UsedPercent, usedGB, totalGB)
		}
		
		if cpuSnapshots, ok := state.CPUSnapshots[host.Alias]; ok && len(cpuSnapshots) > 0 {
			latest := cpuSnapshots[len(cpuSnapshots)-1]
			hostContext += fmt.Sprintf("\n  CPU: %.1f%% used", latest.UsagePercent)
		}
		
		if storageSnapshots, ok := state.StorageSnapshots[host.Alias]; ok && len(storageSnapshots) > 0 {
			latest := storageSnapshots[len(storageSnapshots)-1]
			usedGB := float64(latest.UsedBytes) / (1024 * 1024 * 1024)
			totalGB := float64(latest.TotalBytes) / (1024 * 1024 * 1024)
			hostContext += fmt.Sprintf("\n  Storage /: %.1f%% used (%.1fGB/%.1fGB)", 
				latest.UsedPercent, usedGB, totalGB)
		}
		
		if networkSnapshots, ok := state.NetworkSnapshots[host.Alias]; ok && len(networkSnapshots) > 0 {
			latest := networkSnapshots[len(networkSnapshots)-1]
			rxKbps := float64(latest.RxBytesPerSec) * 8 / 1024
			txKbps := float64(latest.TxBytesPerSec) * 8 / 1024
			hostContext += fmt.Sprintf("\n  Network RX: %.1fKB/s, TX: %.1fKB/s", 
				rxKbps, txKbps)
		}
		
		contextParts = append(contextParts, hostContext)
	}
	
	return "Current system state:" + strings.Join(contextParts, "\n")
}

func (llmClient *LLMClient) getCapabilities(selectedHosts []inventory.TargetHost, watchtowerState WatchtowerStateSnapshot) string {
	if len(selectedHosts) == 0 {
		return "Linux systems, SSH access, shell commands"
	}

	var capabilities []string
	capabilities = append(capabilities, "SSH access")
	capabilities = append(capabilities, "Standard Linux shell commands")

	for _, host := range selectedHosts {
		if host.User != "root" {
			capabilities = append(capabilities, "Sudo access for "+host.Alias)
		} else {
			capabilities = append(capabilities, "Root access for "+host.Alias)
		}
	}

	return strings.Join(capabilities, ", ")
}

func (llmClient *LLMClient) getConstraints() string {
	return "Operations must be safe, non-destructive, and approved by operator before execution. Use read-only commands for diagnostics first. Document all mutative operations."
}

func (m AutopilotModel) handlePlanGeneration(goal string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if m.llmClient != nil {
			plan, _, err := m.llmClient.GeneratePlanWithLLM(ctx, goal, m.selectedHosts, m.watchtowerState)
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
