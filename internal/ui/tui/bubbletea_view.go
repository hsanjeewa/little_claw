package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/devops/agent/internal/domain/agent"
)

type TaskUpdatedMsg struct {
	Task agent.Task
}

type LogReceivedMsg struct {
	Log agent.ExecutionLog
}

type HitlRequestMsg struct {
	Request agent.HitlRequest
}

type Model struct {
	tasks      []agent.Task
	logs       []agent.ExecutionLog
	cursor     int
	taskChan   chan agent.Task
	logChan    chan agent.ExecutionLog
	hitlChan   chan agent.HitlRequest
	activeHitl *agent.HitlRequest
	width      int
	height     int
	ready      bool
	viewport   viewport.Model
}

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("#383838")).
		Width(100)

	panelStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#383838")).
		Padding(1)

	activePanelStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1)

	hitlStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#F5A623")).
		Foreground(lipgloss.Color("#F5A623")).
		Padding(1)

	statusOkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	statusFailStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF3333"))
	statusWaitStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5A623"))
)

func NewModel(taskChan chan agent.Task, logChan chan agent.ExecutionLog, hitlChan chan agent.HitlRequest, initialTasks []agent.Task) Model {
	return Model{
		tasks:    initialTasks,
		logs:     make([]agent.ExecutionLog, 0),
		cursor:   0,
		taskChan: taskChan,
		logChan:  logChan,
		hitlChan: hitlChan,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		listenForTask(m.taskChan),
		listenForLog(m.logChan),
		listenForHitl(m.hitlChan),
	)
}

func listenForTask(sub chan agent.Task) tea.Cmd {
	return func() tea.Msg {
		task, ok := <-sub
		if !ok {
			return nil
		}
		return TaskUpdatedMsg{Task: task}
	}
}

func listenForLog(sub chan agent.ExecutionLog) tea.Cmd {
	return func() tea.Msg {
		log, ok := <-sub
		if !ok {
			return nil
		}
		return LogReceivedMsg{Log: log}
	}
}

func listenForHitl(sub chan agent.HitlRequest) tea.Cmd {
	return func() tea.Msg {
		req, ok := <-sub
		if !ok {
			return nil
		}
		return HitlRequestMsg{Request: req}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.activeHitl != nil {
			switch msg.String() {
			case "y", "Y":
				m.activeHitl.ResponseChan <- true
				m.activeHitl = nil
			case "n", "N":
				m.activeHitl.ResponseChan <- false
				m.activeHitl = nil
			case "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.updateViewportContent()
			}
		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
				m.updateViewportContent()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		headerHeight := 3 
		footerHeight := 3 
		if m.activeHitl != nil {
			footerHeight = 6 
		}

		mainAreaHeight := m.height - headerHeight - footerHeight
		if mainAreaHeight < 0 {
			mainAreaHeight = 0
		}

		innerPanelHeight := mainAreaHeight - 4
		if innerPanelHeight < 0 {
			innerPanelHeight = 0
		}

		leftWidth := (m.width / 3) - 2
		if leftWidth < 0 {
			leftWidth = 0
		}
		rightWidth := m.width - leftWidth - 4
		if rightWidth < 0 {
			rightWidth = 0
		}

		viewportHeight := innerPanelHeight - 2
		if viewportHeight < 0 {
			viewportHeight = 0
		}
		
		vpWidth := rightWidth - 4
		if vpWidth < 0 {
			vpWidth = 0
		}

		if !m.ready {
			m.viewport = viewport.New(vpWidth, viewportHeight)
			m.ready = true
			m.updateViewportContent()
		} else {
			m.viewport.Width = vpWidth
			m.viewport.Height = viewportHeight
		}

	case HitlRequestMsg:
		m.activeHitl = &msg.Request
		return m, listenForHitl(m.hitlChan)

	case TaskUpdatedMsg:
		for i, t := range m.tasks {
			if t.ID == msg.Task.ID {
				m.tasks[i] = msg.Task
				break
			}
		}
		m.updateViewportContent()
		return m, listenForTask(m.taskChan)

	case LogReceivedMsg:
		m.logs = append([]agent.ExecutionLog{msg.Log}, m.logs...)
		if len(m.logs) > 100 {
			m.logs = m.logs[:100]
		}
		m.updateViewportContent()
		return m, listenForLog(m.logChan)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) updateViewportContent() {
	if len(m.tasks) == 0 {
		return
	}
	
	currentTask := m.tasks[m.cursor]
	
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Command: %s\n", currentTask.Command))
	content.WriteString(fmt.Sprintf("Status: %s\n\n", currentTask.Status))
	
	foundLog := false
	for _, l := range m.logs {
		if l.Command == currentTask.Command && l.Host == currentTask.HostIP {
			content.WriteString(l.Output)
			foundLog = true
			break
		}
	}
	
	if !foundLog {
		content.WriteString("Awaiting execution output...")
	}
	
	m.viewport.SetContent(content.String())
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	okCount, failCount, waitCount, changedCount := 0, 0, 0, 0
	for _, t := range m.tasks {
		switch t.Status {
		case agent.StatusSuccess, agent.StatusIdempotent:
			okCount++
		case agent.StatusFailed:
			failCount++
		case "WAITING":
			waitCount++
		case agent.StatusChanged:
			changedCount++
		}
	}

	headerLine1 := ansi.Truncate("🤖 DevOps Agent | 🟢 Status: Active", m.width, "")
	headerLine2 := ansi.Truncate(fmt.Sprintf("📊 Recap: %d Ok | %d Changed | %d Failed | %d Waiting", okCount, changedCount, failCount, waitCount), m.width, "")
	header := headerStyle.Width(m.width).Render(headerLine1 + "\n" + headerLine2)

	headerHeight := 3 
	footerHeight := 3 
	if m.activeHitl != nil {
		footerHeight = 6 
	}
	mainAreaHeight := m.height - headerHeight - footerHeight
	if mainAreaHeight < 0 {
		mainAreaHeight = 0
	}
	innerPanelHeight := mainAreaHeight - 4
	if innerPanelHeight < 0 {
		innerPanelHeight = 0
	}

	leftWidth := (m.width / 3) - 2
	if leftWidth < 0 {
		leftWidth = 0
	}
	rightWidth := m.width - leftWidth - 4
	if rightWidth < 0 {
		rightWidth = 0
	}

	var taskList strings.Builder
	for i, t := range m.tasks {
		if i >= innerPanelHeight {
			break
		}
		
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		
		statusStr := string(t.Status)
		switch t.Status {
		case agent.StatusSuccess, agent.StatusIdempotent:
			statusStr = statusOkStyle.Render(statusStr)
		case agent.StatusFailed:
			statusStr = statusFailStyle.Render(statusStr)
		case "WAITING", agent.StatusRunning:
			statusStr = statusWaitStyle.Render(statusStr)
		}
		
		line := fmt.Sprintf("%s [%s] %s: %s", cursor, statusStr, t.HostAlias, t.Command)
		
		targetWidth := leftWidth - 4
		if targetWidth < 0 {
			targetWidth = 0
		}
		truncatedLine := ansi.Truncate(line, targetWidth, "")
		cleanLine := lipgloss.NewStyle().Width(targetWidth).Render(truncatedLine)
		taskList.WriteString(cleanLine + "\n")
	}
	
	leftPanel := activePanelStyle.Width(leftWidth).Height(mainAreaHeight - 2).Render(taskList.String())
	
	vpTitle := ansi.Truncate("AI ROOT CAUSE ANALYSIS & LOGS", rightWidth-4, "")
	vp := panelStyle.Width(rightWidth).Height(mainAreaHeight - 2).Render(
		lipgloss.NewStyle().Bold(true).Render(vpTitle) + "\n\n" + m.viewport.View(),
	)
	
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, vp)

	var footer string
	if m.activeHitl != nil {
		promptLine1 := ansi.Truncate(fmt.Sprintf("⚠️  [HITL GATE] Target [%s] -> Command: %s", m.activeHitl.Task.HostAlias, m.activeHitl.Task.Command), m.width-4, "")
		promptLine2 := "Approve Execution? (y/N): "
		footer = hitlStyle.Width(m.width - 2).Render(promptLine1 + "\n" + promptLine2)
	} else {
		footerText := ansi.Truncate("↑/↓: scroll output | j/k: navigate list | q: quit", m.width-2, "")
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Padding(1).Render(footerText)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, mainArea, footer)
}
