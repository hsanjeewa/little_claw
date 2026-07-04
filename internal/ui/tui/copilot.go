package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	copilotPaneCommand int = iota
	copilotPaneTerminal
	copilotPaneAdvisory
	copilotPaneGuidance
	copilotPaneCount
)

type CopilotModel struct {
	width        int
	height       int
	focusedPane  int
	commandInput textinput.Model
	terminal     []string
	advisory     []string
	guidance     []string
	handoff      *WatchtowerEscalationPayload
}

func NewCopilotModel() CopilotModel {
	input := textinput.New()
	input.Placeholder = "Enter command..."
	input.Focus()
	input.CharLimit = 512
	input.Width = 60

	return CopilotModel{
		focusedPane:  copilotPaneCommand,
		commandInput: input,
		terminal: []string{
			"$ _",
		},
		advisory: []string{
			"No advisories yet.",
		},
		guidance: []string{
			"Type a command and press Enter.",
			"Tab cycles focus.",
		},
	}
}

func (m CopilotModel) Init() tea.Cmd {
	return nil
}

func (m CopilotModel) ApplyWatchtowerEscalation(payload WatchtowerEscalationPayload) CopilotModel {
	m.handoff = &payload
	return m
}

func (m CopilotModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.focusedPane == copilotPaneCommand {
				cmd := strings.TrimSpace(m.commandInput.Value())
				if cmd != "" {
					m.terminal = append(m.terminal, "$ "+cmd)
					m.commandInput.Reset()
				}
				return m, nil
			}
		}
	}

	if m.focusedPane == copilotPaneCommand {
		updated, cmd := m.commandInput.Update(msg)
		m.commandInput = updated
		return m, cmd
	}

	return m, nil
}

func (m CopilotModel) View() string {
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

	terminalWidth := max(width*2/3, 1)
	sideWidth := max(width-terminalWidth, 1)

	sidePaneHeight := max(panesHeight/2, 1)

	terminal := m.renderPane("TERMINAL", m.terminal, terminalWidth, panesHeight, m.focusedPane == copilotPaneTerminal)
	advisory := m.renderPane("ADVISORY", m.advisory, sideWidth, sidePaneHeight, m.focusedPane == copilotPaneAdvisory)
	guidance := m.renderPane("GUIDANCE", m.guidance, sideWidth, max(panesHeight-sidePaneHeight, 1), m.focusedPane == copilotPaneGuidance)

	sideStack := lipgloss.JoinVertical(lipgloss.Left, advisory, guidance)
	panes := lipgloss.JoinHorizontal(lipgloss.Top, terminal, sideStack)
	sections := []string{panes, commandBar}
	if m.handoff != nil {
		sections = append([]string{m.renderHandoff()}, sections...)
	}
	return constrainSurfaceContent(lipgloss.JoinVertical(lipgloss.Left, sections...), width, height-1)
}

func (m *CopilotModel) cycleFocus() {
	if m.focusedPane == copilotPaneCommand {
		m.commandInput.Blur()
	}
	m.focusedPane = (m.focusedPane + 1) % copilotPaneCount
	if m.focusedPane == copilotPaneCommand {
		m.commandInput.Focus()
	}
}

func (m CopilotModel) renderHandoff() string {
	if m.handoff == nil {
		return ""
	}
	style := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#7D56F4")).Padding(0, 1)
	return style.Render("WATCHTOWER HANDOFF\n" + m.handoff.Summary() + "\n" + m.handoff.Observation)
}

func (m CopilotModel) commandWidth() int {
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

func (m CopilotModel) renderCommandBar() string {
	label := lipgloss.NewStyle().Bold(true).Render("COMMAND")
	bar := lipgloss.JoinHorizontal(lipgloss.Center, label, " ", m.commandInput.View())
	return bar
}

func (m CopilotModel) renderPane(title string, lines []string, width, height int, active bool) string {
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
