package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/infrastructure/inventory"
)

type TargetSelectionConfirmedMsg struct {
	hosts []inventory.TargetHost
}

type TargetSelectionCancelledMsg struct{}

type TargetSelectionModel struct {
	hosts      []inventory.TargetHost
	selected   map[string]bool
	cursor     int
	width      int
	height     int
}

func NewTargetSelectionModel(hosts []inventory.TargetHost) TargetSelectionModel {
	selected := make(map[string]bool)
	for _, h := range hosts {
		selected[h.Alias] = false
	}
	return TargetSelectionModel{
		hosts:    hosts,
		selected: selected,
		cursor:   0,
	}
}

func (m TargetSelectionModel) Init() tea.Cmd {
	return nil
}

func (m TargetSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, func() tea.Msg {
				return TargetSelectionConfirmedMsg{hosts: m.SelectedHosts()}
			}
		case tea.KeyEsc:
			return m, func() tea.Msg {
				return TargetSelectionCancelledMsg{}
			}
		case tea.KeyCtrlA:
			allSelected := true
			for _, host := range m.hosts {
				if !m.selected[host.Alias] {
					allSelected = false
					break
				}
			}
			for _, host := range m.hosts {
				m.selected[host.Alias] = !allSelected
			}
		case tea.KeySpace:
			if len(m.hosts) > 0 {
				host := m.hosts[m.cursor]
				m.selected[host.Alias] = !m.selected[host.Alias]
			}
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.cursor < len(m.hosts)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m TargetSelectionModel) View() string {
	var sb strings.Builder
	sb.WriteString("Select Targets:\n")
	sb.WriteString(strings.Repeat("-", 40))
	sb.WriteString("\n")

	for i, host := range m.hosts {
		checkbox := "[ ]"
		if m.selected[host.Alias] {
			checkbox = "[x]"
		}
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		sb.WriteString(fmt.Sprintf("%s %s %s\n", cursor, checkbox, host.Alias))
	}

	sb.WriteString(strings.Repeat("-", 40))
	sb.WriteString("\n")
	sb.WriteString("Enter: confirm | Esc: cancel | Space: toggle | Ctrl+a: select all")

	return sb.String()
}

func (m TargetSelectionModel) SelectedHosts() []inventory.TargetHost {
	var result []inventory.TargetHost
	for _, host := range m.hosts {
		if m.selected[host.Alias] {
			result = append(result, host)
		}
	}
	return result
}