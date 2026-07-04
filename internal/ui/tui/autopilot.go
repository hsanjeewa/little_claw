package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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
	width := m.width
	if width == 0 {
		width = 80
	}
	height := m.height
	if height == 0 {
		height = 24
	}

	contentHeight := max(height-1, 1)

	commandBar := ansi.Truncate(m.renderCommandBar(), width, "")
	commandHeight := lipgloss.Height(commandBar)
	panesHeight := max(contentHeight-commandHeight, 1)

	leftWidth := max(width/2, 1)
	rightWidth := max(width-leftWidth, 1)

	plan := m.renderPane("PLAN", m.plan, leftWidth, panesHeight, m.focusedPane == autopilotPanePlan)
	transcript := m.renderPane("TRANSCRIPT", m.transcript, rightWidth, panesHeight, m.focusedPane == autopilotPaneTranscript)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, plan, transcript)
	sections := []string{panes, commandBar}
	if m.handoff != nil {
		sections = append([]string{m.renderHandoff()}, sections...)
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

func (m AutopilotModel) renderPane(title string, lines []string, width, height int, active bool) string {
	style := panelStyle
	if active {
		style = activePanelStyle
	}

	frameWidth := style.GetHorizontalFrameSize()
	frameHeight := style.GetVerticalFrameSize()
	styleWidth := max(width-frameWidth, 0)
	styleHeight := max(height-frameHeight, 0)
	innerWidth := max(styleWidth-style.GetHorizontalPadding(), 1)
	innerHeight := max(styleHeight-style.GetVerticalPadding()-1, 0)

	header := lipgloss.NewStyle().Bold(true).Render(title)
	body := strings.Join(lines, "\n")
	if body == "" {
		body = " "
	}

	bodyStyle := lipgloss.NewStyle().Width(innerWidth).Height(innerHeight)
	content := lipgloss.JoinVertical(lipgloss.Left, header, bodyStyle.Render(body))
	return style.Width(styleWidth).Height(styleHeight).Render(content)
}
