package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/inventory"
)

type WatchtowerStateSnapshot struct {
	MemorySnapshots map[string][]agent.MemorySnapshot
	CPUSnapshots    map[string][]agent.CPUSnapshot
	StorageSnapshots map[string][]agent.StorageSnapshot
	NetworkSnapshots map[string][]agent.NetworkSnapshot
}

const (
	autopilotPaneCommand int = iota
	autopilotPanePlan
	autopilotPaneTranscript
	autopilotPaneCount
)

type AutopilotModel struct {
	width        int
	height       int
	focusedPane  int
	commandInput textinput.Model
	plan         []string
	transcript   []string
	handoff      *WatchtowerEscalationPayload
	run          AutopilotRun
	llmClient    *LLMClient
	executor     agent.CommandExecutor
	taskChan     chan agent.Task
	logChan      chan agent.ExecutionLog
	hitlChan     chan agent.HitlRequest
	taskHostAlias string
	taskHostIP    string
	taskHostUser  string
	taskHostPort  int
	selectedHosts []inventory.TargetHost
	watchtowerState WatchtowerStateSnapshot
}

func NewAutopilotModel() AutopilotModel {
	input := textinput.New()
	input.Placeholder = "Enter command..."
	input.Focus()
	input.CharLimit = 512
	input.Width = 60

	return AutopilotModel{
		focusedPane:  autopilotPaneCommand,
		commandInput: input,
		plan:         []string{},
		transcript:   []string{},
		run: AutopilotRun{
			State:      RunStateDrafting,
			Approved:   false,
			MaxRetries: defaultMaxRetries,
		},
	}
}

func NewAutopilotModelWithLLMClient(llmClient *LLMClient) AutopilotModel {
	model := NewAutopilotModel()
	model.llmClient = llmClient
	return model
}

func NewAutopilotModelWithChannels(
	taskChan chan agent.Task,
	logChan chan agent.ExecutionLog,
	hitlChan chan agent.HitlRequest,
) AutopilotModel {
	model := NewAutopilotModel()
	model.taskChan = taskChan
	model.logChan = logChan
	model.hitlChan = hitlChan
	return model
}

func NewAutopilotModelWithAllDependencies(
	llmClient *LLMClient,
	taskChan chan agent.Task,
	logChan chan agent.ExecutionLog,
	hitlChan chan agent.HitlRequest,
) AutopilotModel {
	model := NewAutopilotModel()
	model.llmClient = llmClient
	model.taskChan = taskChan
	model.logChan = logChan
	model.hitlChan = hitlChan
	return model
}

func (m AutopilotModel) WithExecutor(executor agent.CommandExecutor) AutopilotModel {
	m.executor = executor
	return m
}

func (m AutopilotModel) WithTargetHost(alias, ip, user string, port int) AutopilotModel {
	m.taskHostAlias = alias
	m.taskHostIP = ip
	m.taskHostUser = user
	m.taskHostPort = port
	return m
}

func (m AutopilotModel) WithSelectedHosts(hosts []inventory.TargetHost) AutopilotModel {
	m.selectedHosts = hosts
	return m
}

func (m AutopilotModel) WithWatchtowerState(state WatchtowerStateSnapshot) AutopilotModel {
	m.watchtowerState = state
	return m
}

func (m AutopilotModel) Init() tea.Cmd {
	return nil
}

func (m AutopilotModel) ApplyWatchtowerEscalation(payload WatchtowerEscalationPayload, selectedHosts []inventory.TargetHost, watchtowerState WatchtowerStateSnapshot) AutopilotModel {
	m.handoff = &payload
	m.selectedHosts = selectedHosts
	m.watchtowerState = watchtowerState
	return m
}

func (m AutopilotModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case GeneratePlanMsg:
		m.transcript = append(m.transcript, "> "+msg.Goal)
		return m, m.handlePlanGeneration(msg.Goal)
	case PlanGeneratedMsg:
		if msg.Error != nil {
			errorMsg := TranscriptEntry{
				Kind: TranscriptSystem,
				Text: fmt.Sprintf("Plan generation failed: %v", msg.Error),
			}
			m.transcript = append(m.transcript, errorMsg.String())
		} else {
			reasoningMsg := TranscriptEntry{
				Kind: TranscriptAgent,
				Text: msg.Reasoning,
			}
			m.transcript = append(m.transcript, reasoningMsg.String())

			planStrings := make([]string, len(msg.Plan))
			for i, step := range msg.Plan {
				mutativeFlag := ""
				if step.IsMutative {
					mutativeFlag = "[MUTATIVE]"
				}
				planStrings[i] = fmt.Sprintf("%d. %s %s\n   %s", i+1, step.Description, mutativeFlag, step.Command)
			}
			m.plan = planStrings
			m.run = AutopilotRun{
				State:      RunStateReady,
				Plan:       msg.Plan,
				Approved:   false,
				MaxRetries: defaultMaxRetries,
			}
		}
		return m, nil
	case PlanApprovedMsg:
		m.run.Approved = true
		m.run.State = RunStateExecuting
		if m.executor != nil && len(m.run.Plan) > 0 {
			task, err := m.buildTaskFromStep(m.run.Plan[0])
			if err != nil {
				return m, nil
			}
			return m, RunPlanStep(context.Background(), 0, task, m.executor)
		}
		m.submitPlanTasks()
		return m, nil
	case PlanRejectedMsg:
		m.run.State = RunStateDrafting
		m.run.Plan = nil
		m.run.Approved = false
		return m, nil
	case TaskCompletedMsg:
		m.transcript = append(m.transcript, TranscriptEntry{
			Kind: TranscriptSystem,
			Text: fmt.Sprintf("Step completed: %s", msg.Output),
		}.String())
		m.run.LastCompletedStep = msg.StepIndex
		if msg.StepIndex+1 >= len(m.run.Plan) {
			m.run.State = RunStateCompleted
			return m, nil
		}
		nextStep := msg.StepIndex + 1
		if m.executor != nil {
			task, err := m.buildTaskFromStep(m.run.Plan[nextStep])
			if err != nil {
				return m, nil
			}
			return m, RunPlanStep(context.Background(), nextStep, task, m.executor)
		}
		return m, nil
	case TaskFailedMsg:
		m.transcript = append(m.transcript, TranscriptEntry{
			Kind: TranscriptSystem,
			Text: fmt.Sprintf("Step failed: %s", msg.Error),
		}.String())
		m.run.OriginalError = msg.Error
		return m.triggerRecovery(msg.Error)
	case ExecutionProgressMsg:
		m.run.CurrentHost = msg.Progress.CurrentHost
		m.run.CurrentStep = msg.Progress.CurrentStep
		return m, nil
	case RecoveryPlanGeneratedMsg:
		m.run.RecoveryPlan = msg.Plan
		m.run.State = RunStateReady
		recoveryPlanText := []string{
			"--- RECOVERY PLAN (Attempt " + fmt.Sprintf("%d/%d", m.run.RetryCount, m.run.MaxRetries) + ") ---",
			"Original Error: " + m.run.OriginalError,
			"",
		}
		for i, step := range msg.Plan {
			flag := ""
			if step.IsMutative {
				flag = "⚠️ MUTATIVE"
			}
			recoveryPlanText = append(recoveryPlanText, fmt.Sprintf("%d. %s %s", i+1, step.Description, flag))
		}
		recoveryPlanText = append(recoveryPlanText, "", "[APPROVE RECOVERY] Press 'a'", "[REJECT] Press 'r'")
		m.plan = recoveryPlanText
		return m, nil
	case RecoveryApprovedMsg:
		if len(m.run.RecoveryPlan) > 0 {
			m.run.Plan = m.run.RecoveryPlan
			m.run.State = RunStateExecuting
			m.transcript = append(m.transcript, TranscriptEntry{
				Kind: TranscriptSystem,
				Text: fmt.Sprintf("🔄 Executing recovery plan (attempt %d/%d)", m.run.RetryCount, m.run.MaxRetries),
			}.String())
			return m, func() tea.Msg {
				return PlanApprovedMsg{}
			}
		}
		return m, nil
	case RecoveryRejectedMsg:
		m.run.State = RunStateFailed
		m.transcript = append(m.transcript, TranscriptEntry{
			Kind: TranscriptSystem,
			Text: "❌ Recovery rejected by operator",
		}.String())
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.commandInput.Width = m.commandWidth()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			m.cycleFocus()
			return m, nil
		case tea.KeyRunes:
			if m.run.State == RunStateReady {
				switch string(msg.Runes) {
				case "a":
					if len(m.run.RecoveryPlan) > 0 {
						return m, func() tea.Msg { return RecoveryApprovedMsg{} }
					}
					return m, func() tea.Msg { return PlanApprovedMsg{} }
				case "r":
					if len(m.run.RecoveryPlan) > 0 {
						return m, func() tea.Msg { return RecoveryRejectedMsg{} }
					}
					return m, func() tea.Msg { return PlanRejectedMsg{} }
				}
			}
		case tea.KeyEnter:
			if m.focusedPane == autopilotPaneCommand {
				cmd := strings.TrimSpace(m.commandInput.Value())
				if cmd != "" {
					if cmd == "/exit" {
						return m, tea.Quit
					}
					if cmd == "/targets" {
						m.commandInput.Reset()
						return m, func() tea.Msg {
							return OpenTargetSelectionMsg{}
						}
					}
					if !strings.HasPrefix(cmd, "/") {
						m.commandInput.Reset()
						return m, m.GeneratePlan(cmd)
					}
					m.transcript = append(m.transcript, "> "+cmd)
					m.commandInput.Reset()
				}
				return m, nil
			}
		}
	}

	if m.focusedPane == autopilotPaneCommand {
		updated, cmd := m.commandInput.Update(msg)
		m.commandInput = updated
		return m, cmd
	}

	return m, nil
}

func (m AutopilotModel) View() string {
	width := m.width
	if width == 0 {
		width = 80
	}
	height := m.height
	if height == 0 {
		height = 24
	}

	contentHeight := max(height-1, 1)

	handoff := ""
	if m.handoff != nil {
		handoff = m.renderHandoff()
	}
	handoffHeight := lipgloss.Height(handoff)

	commandBar := ansi.Truncate(m.renderCommandBar(), width, "")
	commandHeight := lipgloss.Height(commandBar)
	panesHeight := max(contentHeight-handoffHeight-commandHeight, 1)

	leftWidth := max(width/2, 1)
	rightWidth := max(width-leftWidth, 1)

	planContent := m.plan
	if m.run.State == RunStateReady {
		planContent = append(m.plan, "", "[APPROVE] Press 'a'", "[REJECT] Press 'r'")
	}
	plan := renderPane("PLAN", planContent, leftWidth, panesHeight, m.focusedPane == autopilotPanePlan)
	transcript := renderPane("TRANSCRIPT", m.transcript, rightWidth, panesHeight, m.focusedPane == autopilotPaneTranscript)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, plan, transcript)
	sections := []string{panes, commandBar}
	if handoff != "" {
		sections = append([]string{handoff}, sections...)
	}
	return constrainSurfaceContent(lipgloss.JoinVertical(lipgloss.Left, sections...), width, height-1)
}

func (m *AutopilotModel) cycleFocus() {
	if m.focusedPane == autopilotPaneCommand {
		m.commandInput.Blur()
	}
	m.focusedPane = (m.focusedPane + 1) % autopilotPaneCount
	if m.focusedPane == autopilotPaneCommand {
		m.commandInput.Focus()
	}
}

func (m AutopilotModel) renderHandoff() string {
	if m.handoff == nil {
		return ""
	}
	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#7D56F4")).Padding(0, 1)
	return style.Render("WATCHTOWER HANDOFF\n" + m.handoff.Summary() + "\n" + m.handoff.Observation)
}

func (m AutopilotModel) commandWidth() int {
	w := m.width
	if w == 0 {
		w = 80
	}
	labelWidth := lipgloss.Width(commandBarLabel + " ")
	padding := 4
	available := w - labelWidth - padding
	available = max(available, 20)
	return available
}

func (m AutopilotModel) renderCommandBar() string {
	label := lipgloss.NewStyle().Bold(true).Render(commandBarLabel)
	bar := lipgloss.JoinHorizontal(lipgloss.Center, label, " ", m.commandInput.View())
	return bar
}
