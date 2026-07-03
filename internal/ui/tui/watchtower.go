package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/inventory"
)

type MemorySnapshotCollector func(context.Context, []inventory.TargetHost) ([]agent.MemorySnapshot, error)
type CPUSnapshotCollector func(context.Context, []inventory.TargetHost) ([]agent.CPUSnapshot, error)

type watchtowerViewMode int

const (
	watchtowerViewFleet watchtowerViewMode = iota
	watchtowerViewHostDetail
)

type watchtowerSnapshotsMsg struct {
	snapshots []agent.MemorySnapshot
}

type watchtowerCPUSnapshotsMsg struct {
	snapshots []agent.CPUSnapshot
}

type WatchtowerModel struct {
	width      int
	height     int
	inventory  []inventory.TargetHost
	scope      TargetScope
	memoryCollector MemorySnapshotCollector
	cpuCollector    CPUSnapshotCollector
	metricFamily    agent.MetricFamily
	memorySnapshots []agent.MemorySnapshot
	cpuSnapshots    []agent.CPUSnapshot
	selected   int
	viewMode   watchtowerViewMode
	refreshing bool
	legacy     Model
}

func NewWatchtowerModel(
	taskChan chan agent.Task,
	logChan chan agent.ExecutionLog,
	hitlChan chan agent.HitlRequest,
	initialTasks []agent.Task,
	inv []inventory.TargetHost,
	scope TargetScope,
	memoryCollector MemorySnapshotCollector,
) WatchtowerModel {
	return NewWatchtowerModelWithCollectors(taskChan, logChan, hitlChan, initialTasks, inv, scope, memoryCollector, nil)
}

func NewWatchtowerModelWithCollectors(
	taskChan chan agent.Task,
	logChan chan agent.ExecutionLog,
	hitlChan chan agent.HitlRequest,
	initialTasks []agent.Task,
	inv []inventory.TargetHost,
	scope TargetScope,
	memoryCollector MemorySnapshotCollector,
	cpuCollector CPUSnapshotCollector,
) WatchtowerModel {
	if scope.Kind == 0 && len(scope.Hosts) == 0 {
		scope = TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(inv)}
	}

	return WatchtowerModel{
		inventory:        cloneHosts(inv),
		scope:            scope.Clone(),
		memoryCollector:  memoryCollector,
		cpuCollector:     cpuCollector,
		metricFamily:     agent.MetricFamilyMemory,
		viewMode:         watchtowerViewFleet,
		legacy:           NewModel(taskChan, logChan, hitlChan, initialTasks),
	}
}

func (m WatchtowerModel) Init() tea.Cmd {
	cmds := []tea.Cmd{m.legacy.Init()}
	if len(m.inventory) > 0 {
		if cmd := m.refreshCurrentFamilyCmd(m.inventory); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (m WatchtowerModel) SetScope(scope TargetScope) WatchtowerModel {
	m.scope = scope.Clone()
	visible := m.visibleRowCount()
	if visible == 0 {
		m.selected = 0
		m.viewMode = watchtowerViewFleet
		return m
	}
	if m.selected >= visible {
		m.selected = visible - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
	return m
}

func (m WatchtowerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		updated, _ := m.legacy.Update(msg)
		m.legacy = updated.(Model)
		return m, nil
	case watchtowerSnapshotsMsg:
		m.memorySnapshots = msg.snapshots
		m.refreshing = false
		m.clampSelection()
		return m, nil
	case watchtowerCPUSnapshotsMsg:
		m.cpuSnapshots = msg.snapshots
		m.refreshing = false
		m.clampSelection()
		return m, nil
	case TaskUpdatedMsg, LogReceivedMsg, HitlRequestMsg:
		updated, cmd := m.legacy.Update(msg)
		m.legacy = updated.(Model)
		return m, cmd
	case tea.KeyMsg:
		visibleCount := m.visibleRowCount()
		switch strings.ToLower(msg.String()) {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			if m.currentCollectorMissing() {
				return m, nil
			}
			hosts := m.hostsForRefresh()
			m.refreshing = true
			return m, m.refreshCurrentFamilyCmd(hosts)
		case "1":
			m.metricFamily = agent.MetricFamilyMemory
			m.clampSelection()
			return m, nil
		case "2":
			m.metricFamily = agent.MetricFamilyCPU
			m.clampSelection()
			return m, nil
		case "a":
			payload, ok := m.escalationPayload(ModeAutopilot)
			if !ok {
				return m, nil
			}
			return m, func() tea.Msg { return watchtowerEscalationMsg{Payload: payload} }
		case "c":
			payload, ok := m.escalationPayload(ModeCopilot)
			if !ok {
				return m, nil
			}
			return m, func() tea.Msg { return watchtowerEscalationMsg{Payload: payload} }
		case "enter":
			if m.viewMode == watchtowerViewFleet && visibleCount > 0 {
				m.viewMode = watchtowerViewHostDetail
			}
			return m, nil
		case "b", "esc":
			if m.viewMode == watchtowerViewHostDetail {
				m.viewMode = watchtowerViewFleet
			}
			return m, nil
		case "j", "down":
			if m.viewMode == watchtowerViewFleet && m.selected < visibleCount-1 {
				m.selected++
			}
			return m, nil
		case "k", "up":
			if m.viewMode == watchtowerViewFleet && m.selected > 0 {
				m.selected--
			}
			return m, nil
		}
	}

	return m, nil
}

func (m WatchtowerModel) View() string {
	if m.width == 0 || m.height == 0 {
		m.width = 100
		m.height = 24
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(string(m.metricFamily))
	if m.refreshing {
		title += "  [REFRESHING]"
	}

	scopeLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#A0A0A0")).Render(m.scope.String())

	if m.viewMode == watchtowerViewHostDetail {
		return lipgloss.JoinVertical(lipgloss.Left, title, scopeLine, m.renderHostDetail())
	}

	return lipgloss.JoinVertical(lipgloss.Left, title, scopeLine, m.renderFleetMatrix())
}

func (m WatchtowerModel) refreshCurrentFamilyCmd(hosts []inventory.TargetHost) tea.Cmd {
	if m.metricFamily == agent.MetricFamilyCPU {
		return m.refreshCPUCmd(hosts)
	}
	return m.refreshMemoryCmd(hosts)
}

func (m WatchtowerModel) refreshMemoryCmd(hosts []inventory.TargetHost) tea.Cmd {
	if m.memoryCollector == nil {
		return nil
	}
	return func() tea.Msg {
		snapshots, err := m.memoryCollector(context.Background(), hosts)
		if err != nil {
			failed := make([]agent.MemorySnapshot, 0, len(hosts))
			for _, host := range hosts {
				failed = append(failed, agent.MemorySnapshot{
					HostAlias:   host.Alias,
					HostIP:      host.IP,
					Status:      agent.SnapshotStatusFailed,
					CollectedAt: time.Now(),
					Error:       err.Error(),
				})
			}
			return watchtowerSnapshotsMsg{snapshots: failed}
		}
		return watchtowerSnapshotsMsg{snapshots: snapshots}
	}
}

func (m WatchtowerModel) refreshCPUCmd(hosts []inventory.TargetHost) tea.Cmd {
	if m.cpuCollector == nil {
		return nil
	}
	return func() tea.Msg {
		snapshots, err := m.cpuCollector(context.Background(), hosts)
		if err != nil {
			failed := make([]agent.CPUSnapshot, 0, len(hosts))
			for _, host := range hosts {
				failed = append(failed, agent.CPUSnapshot{
					HostAlias:   host.Alias,
					HostIP:      host.IP,
					Status:      agent.SnapshotStatusFailed,
					CollectedAt: time.Now(),
					Error:       err.Error(),
				})
			}
			return watchtowerCPUSnapshotsMsg{snapshots: failed}
		}
		return watchtowerCPUSnapshotsMsg{snapshots: snapshots}
	}
}

func (m WatchtowerModel) hostsForRefresh() []inventory.TargetHost {
	if m.scope.Kind == ScopeSelectedHosts && len(m.scope.Hosts) > 0 {
		return cloneHosts(m.scope.Hosts)
	}
	return cloneHosts(m.inventory)
}

func (m WatchtowerModel) visibleMemorySnapshots() []agent.MemorySnapshot {
	if len(m.memorySnapshots) == 0 {
		return nil
	}
	visible := make([]agent.MemorySnapshot, 0, len(m.memorySnapshots))
	for _, snapshot := range m.memorySnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		visible = append(visible, snapshot)
	}
	sort.Slice(visible, func(i, j int) bool {
		return visible[i].HostAlias < visible[j].HostAlias
	})
	return visible
}

func (m WatchtowerModel) visibleCPUSnapshots() []agent.CPUSnapshot {
	if len(m.cpuSnapshots) == 0 {
		return nil
	}
	visible := make([]agent.CPUSnapshot, 0, len(m.cpuSnapshots))
	for _, snapshot := range m.cpuSnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		visible = append(visible, snapshot)
	}
	sort.Slice(visible, func(i, j int) bool {
		return visible[i].HostAlias < visible[j].HostAlias
	})
	return visible
}

func (m WatchtowerModel) visibleRowCount() int {
	if m.metricFamily == agent.MetricFamilyCPU {
		return len(m.visibleCPUSnapshots())
	}
	return len(m.visibleMemorySnapshots())
}

func (m *WatchtowerModel) clampSelection() {
	visible := m.visibleRowCount()
	if visible == 0 {
		m.selected = 0
		m.viewMode = watchtowerViewFleet
		return
	}
	if m.selected >= visible {
		m.selected = visible - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

func (m WatchtowerModel) currentCollectorMissing() bool {
	if m.metricFamily == agent.MetricFamilyCPU {
		return m.cpuCollector == nil
	}
	return m.memoryCollector == nil
}

func (m WatchtowerModel) renderFleetMatrix() string {
	if m.metricFamily == agent.MetricFamilyCPU {
		return m.renderCPUFleetMatrix()
	}
	visible := m.visibleMemorySnapshots()
	if len(visible) == 0 {
		return "No memory snapshots available yet. Press r to refresh."
	}

	lines := []string{"FLEET MATRIX", ""}
	for i, snapshot := range visible {
		marker := " "
		if i == m.selected {
			marker = ">"
		}
		lines = append(lines, fmt.Sprintf("%s %-16s %8s / %-8s %6.1f%%  %s",
			marker,
			snapshot.HostAlias,
			formatBytes(snapshot.UsedBytes),
			formatBytes(snapshot.TotalBytes),
			snapshot.UsedPercent,
			freshnessLabel(snapshot),
		))
		if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
			lines = append(lines, fmt.Sprintf("  error: %s", snapshot.Error))
		}
	}
	lines = append(lines, "", "j/k move • Enter detail • 1 memory • 2 cpu • r refresh • a autopilot • c copilot")
	return strings.Join(lines, "\n")
}

func (m WatchtowerModel) renderCPUFleetMatrix() string {
	visible := m.visibleCPUSnapshots()
	if len(visible) == 0 {
		return "No CPU snapshots available yet. Press r to refresh."
	}
	lines := []string{"FLEET MATRIX", ""}
	for i, snapshot := range visible {
		marker := " "
		if i == m.selected {
			marker = ">"
		}
		lines = append(lines, fmt.Sprintf("%s %-16s %6.1f%%  %s",
			marker,
			snapshot.HostAlias,
			snapshot.UsagePercent,
			cpuFreshnessLabel(snapshot),
		))
		if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
			lines = append(lines, fmt.Sprintf("  error: %s", snapshot.Error))
		}
	}
	lines = append(lines, "", "j/k move • Enter detail • 1 memory • 2 cpu • r refresh • a autopilot • c copilot")
	return strings.Join(lines, "\n")
}

func (m WatchtowerModel) renderHostDetail() string {
	if m.metricFamily == agent.MetricFamilyCPU {
		return m.renderCPUHostDetail()
	}
	visible := m.visibleMemorySnapshots()
	if len(visible) == 0 {
		return "HOST DETAIL\n\nNo host selected."
	}
	snapshot := visible[m.selected]
	lines := []string{
		"HOST DETAIL",
		"",
		fmt.Sprintf("Host: %s (%s)", snapshot.HostAlias, snapshot.HostIP),
		fmt.Sprintf("Used: %s", formatBytes(snapshot.UsedBytes)),
		fmt.Sprintf("Total: %s", formatBytes(snapshot.TotalBytes)),
		fmt.Sprintf("Usage: %.1f%%", snapshot.UsedPercent),
		fmt.Sprintf("Freshness: %s", freshnessLabel(snapshot)),
	}
	if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
		lines = append(lines, fmt.Sprintf("Error: %s", snapshot.Error))
	}
	lines = append(lines, "", "b back • 1 memory • 2 cpu • r refresh • a autopilot • c copilot")
	return strings.Join(lines, "\n")
}

func (m WatchtowerModel) renderCPUHostDetail() string {
	visible := m.visibleCPUSnapshots()
	if len(visible) == 0 {
		return "HOST DETAIL\n\nNo host selected."
	}
	snapshot := visible[m.selected]
	lines := []string{
		"HOST DETAIL",
		"",
		fmt.Sprintf("Host: %s (%s)", snapshot.HostAlias, snapshot.HostIP),
		fmt.Sprintf("CPU Usage: %.1f%%", snapshot.UsagePercent),
		fmt.Sprintf("Freshness: %s", cpuFreshnessLabel(snapshot)),
	}
	if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
		lines = append(lines, fmt.Sprintf("Error: %s", snapshot.Error))
	}
	lines = append(lines, "", "b back • 1 memory • 2 cpu • r refresh • a autopilot • c copilot")
	return strings.Join(lines, "\n")
}

func (m WatchtowerModel) escalationPayload(target Mode) (WatchtowerEscalationPayload, bool) {
	if m.metricFamily == agent.MetricFamilyCPU {
		visible := m.visibleCPUSnapshots()
		if len(visible) == 0 {
			return WatchtowerEscalationPayload{}, false
		}
		snapshot := visible[m.selected]
		observation := fmt.Sprintf("CPU %.1f%% used • %s", snapshot.UsagePercent, cpuFreshnessLabel(snapshot))
		if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
			observation = fmt.Sprintf("Collection failed: %s", snapshot.Error)
		}
		return WatchtowerEscalationPayload{
			Target:       target,
			MetricFamily: agent.MetricFamilyCPU,
			Scope:        m.scope.Clone(),
			SelectedHost: snapshot.HostAlias,
			Observation:  observation,
		}, true
	}
	visible := m.visibleMemorySnapshots()
	if len(visible) == 0 {
		return WatchtowerEscalationPayload{}, false
	}
	snapshot := visible[m.selected]
	observation := fmt.Sprintf("Memory %.1f%% used (%s / %s) • %s", snapshot.UsedPercent, formatBytes(snapshot.UsedBytes), formatBytes(snapshot.TotalBytes), freshnessLabel(snapshot))
	if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
		observation = fmt.Sprintf("Collection failed: %s", snapshot.Error)
	}
	return WatchtowerEscalationPayload{
		Target:       target,
		MetricFamily: agent.MetricFamilyMemory,
		Scope:        m.scope.Clone(),
		SelectedHost: snapshot.HostAlias,
		Observation:  observation,
	}, true
}

func formatBytes(value uint64) string {
	const gib = 1024 * 1024 * 1024
	const mib = 1024 * 1024
	if value >= gib {
		return fmt.Sprintf("%.1fGiB", float64(value)/gib)
	}
	if value >= mib {
		return fmt.Sprintf("%.0fMiB", float64(value)/mib)
	}
	return fmt.Sprintf("%dB", value)
}

func freshnessLabel(snapshot agent.MemorySnapshot) string {
	if snapshot.Status == agent.SnapshotStatusFailed {
		return "FAILED"
	}
	if snapshot.CollectedAt.IsZero() {
		return "STALE"
	}
	return fmt.Sprintf("UPDATED %s", snapshot.CollectedAt.Format("15:04:05"))
}

func cpuFreshnessLabel(snapshot agent.CPUSnapshot) string {
	if snapshot.Status == agent.SnapshotStatusFailed {
		return "FAILED"
	}
	if snapshot.CollectedAt.IsZero() {
		return "STALE"
	}
	return fmt.Sprintf("UPDATED %s", snapshot.CollectedAt.Format("15:04:05"))
}
