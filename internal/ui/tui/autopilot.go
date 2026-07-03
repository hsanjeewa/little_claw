package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
		plan: []string{
			"1. Discover inventory",
			"2. Generate plan",
			"3. Execute autonomously",
		},
		transcript: []string{
			"Waiting for first task...",
		},
	}
}

func (m AutopilotModel) Init() tea.Cmd {
	return nil
}

func (m AutopilotModel) ApplyWatchtowerEscalation(payload WatchtowerEscalationPayload) AutopilotModel {
	m.handoff = &payload
	return m
}

func (m AutopilotModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.commandInput.Width = m.commandWidth()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			m.cycleFocus()
			return m, nil
		case tea.KeyEnter:
			if m.focusedPane == autopilotPaneCommand {
				cmd := strings.TrimSpace(m.commandInput.Value())
				if cmd != "" {
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
	if m.width == 0 || m.height == 0 {
		m.width = 80
		m.height = 24
	}

	contentHeight := m.height - 1
	contentHeight = max(contentHeight, 6)

	commandBar := m.renderCommandBar()
	commandHeight := lipgloss.Height(commandBar)
	panesHeight := contentHeight - commandHeight
	panesHeight = max(panesHeight, 3)

	plan := m.renderPane("PLAN", m.plan, panesHeight, m.focusedPane == autopilotPanePlan)
	transcript := m.renderPane("TRANSCRIPT", m.transcript, panesHeight, m.focusedPane == autopilotPaneTranscript)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, plan, transcript)
	sections := []string{panes, commandBar}
	if m.handoff != nil {
		sections = append([]string{m.renderHandoff()}, sections...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
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
	labelWidth := lipgloss.Width("COMMAND ")
	padding := 4
	available := w - labelWidth - padding
	available = max(available, 20)
	return available
}

func (m AutopilotModel) renderCommandBar() string {
	label := lipgloss.NewStyle().Bold(true).Render("COMMAND")
	bar := lipgloss.JoinHorizontal(lipgloss.Center, label, " ", m.commandInput.View())
	return bar
}

func (m AutopilotModel) renderPane(title string, lines []string, height int, active bool) string {
	w := max(m.width/2, 10)

	style := panelStyle
	if active {
		style = activePanelStyle
	}

	innerWidth := w - style.GetHorizontalFrameSize() - style.GetHorizontalPadding()
	innerHeight := height - style.GetVerticalFrameSize() - style.GetVerticalPadding()
	innerWidth = max(innerWidth, 1)
	innerHeight = max(innerHeight, 1)

	header := lipgloss.NewStyle().Bold(true).Render(title)
	body := strings.Join(lines, "\n")
	if body == "" {
		body = " "
	}

	bodyStyle := lipgloss.NewStyle().Width(innerWidth).Height(innerHeight)
	content := lipgloss.JoinVertical(lipgloss.Left, header, bodyStyle.Render(body))
	return style.Width(w).Height(height).Render(content)
}
