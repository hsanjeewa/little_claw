package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/domain/agent"
)

type TaskUpdatedMsg struct {
	Task agent.Task
}

type LogReceivedMsg struct {
	Log agent.ExecutionLog
}

type HitlRequest struct {
	Task         agent.Task
	ResponseChan chan bool
}

type HitlRequestMsg struct {
	Request HitlRequest
}

type Model struct {
	tasks      []agent.Task
	logs       []agent.ExecutionLog
	cursor     int
	taskChan   chan agent.Task
	logChan    chan agent.ExecutionLog
	hitlChan   chan HitlRequest
	activeHitl *HitlRequest
	width      int
	height     int
}

func NewModel(taskChan chan agent.Task, logChan chan agent.ExecutionLog, hitlChan chan HitlRequest, initialTasks []agent.Task) Model {
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

func listenForHitl(sub chan HitlRequest) tea.Cmd {
	return func() tea.Msg {
		req, ok := <-sub
		if !ok {
			return nil
		}
		return HitlRequestMsg{Request: req}
	}
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			}
		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

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
		return m, listenForTask(m.taskChan)

	case LogReceivedMsg:
		m.logs = append([]agent.ExecutionLog{msg.Log}, m.logs...)
		if len(m.logs) > 100 {
			m.logs = m.logs[:100]
		}
		return m, listenForLog(m.logChan)
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	leftPanelWidth := m.width / 2
	rightPanelWidth := m.width - leftPanelWidth

	var statusBar string
	var panelHeight int
	if m.activeHitl != nil {
		statusBar = fmt.Sprintf("⚠️  [HITL GATE] Review target [%s] -> Command: %s\nApprove Execution? (y/N): ", m.activeHitl.Task.HostAlias, m.activeHitl.Task.Command)
		panelHeight = m.height - 3
	} else {
		statusBar = "  j/k: navigate, q: quit"
		panelHeight = m.height - 2
	}

	leftPanel := renderTaskList(m.tasks, m.cursor, leftPanelWidth, panelHeight)
	rightPanel := renderLogList(m.logs, rightPanelWidth, panelHeight)

	splitViews := sideBySide(leftPanel, rightPanel)
	
	return fmt.Sprintf("%s\n%s", splitViews, statusBar)
}

func renderTaskList(tasks []agent.Task, cursor int, width, height int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-*s\n", width, "--- TASKS ---"))
	
	lines := 1
	for i, t := range tasks {
		if lines >= height {
			break
		}
		
		cursorChar := " "
		if cursor == i {
			cursorChar = ">"
		}
		
		line := fmt.Sprintf("%s [%s] %s: %s", cursorChar, t.Status, t.HostAlias, t.Command)
		if len(line) > width {
			line = line[:width-3] + "..."
		}
		
		b.WriteString(fmt.Sprintf("%-*s\n", width, line))
		lines++
	}
	
	for lines < height {
		b.WriteString(fmt.Sprintf("%-*s\n", width, ""))
		lines++
	}
	
	return b.String()
}

func renderLogList(logs []agent.ExecutionLog, width, height int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-*s\n", width, "--- RECENT LOGS ---"))
	
	lines := 1
	for _, l := range logs {
		if lines >= height {
			break
		}
		
		timeStr := l.Timestamp.Format("15:04:05")
		line := fmt.Sprintf("[%s] %s - %s: %s", timeStr, l.Host, l.Status, l.Command)
		if len(line) > width {
			line = line[:width-3] + "..."
		}
		
		b.WriteString(fmt.Sprintf("%-*s\n", width, line))
		lines++
	}
	
	for lines < height {
		b.WriteString(fmt.Sprintf("%-*s\n", width, ""))
		lines++
	}
	
	return b.String()
}

func sideBySide(left, right string) string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")
	
	var b strings.Builder
	for i := 0; i < len(leftLines) && i < len(rightLines); i++ {
		b.WriteString(leftLines[i])
		b.WriteString(rightLines[i])
		if i < len(leftLines)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
