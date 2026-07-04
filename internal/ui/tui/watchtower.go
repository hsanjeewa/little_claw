package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/inventory"
)

type MemorySnapshotCollector func(context.Context, []inventory.TargetHost) ([]agent.MemorySnapshot, error)
type CPUSnapshotCollector func(context.Context, []inventory.TargetHost) ([]agent.CPUSnapshot, error)
type StorageSnapshotCollector func(context.Context, []inventory.TargetHost) ([]agent.StorageSnapshot, error)
type NetworkSnapshotCollector func(context.Context, []inventory.TargetHost) ([]agent.NetworkSnapshot, error)

type watchtowerViewMode int

const (
	watchtowerViewFleet watchtowerViewMode = iota
	watchtowerViewHostDetail
)

const (
	watchtowerMinSplitWidth = 40
	watchtowerMinPaneHeight = 5
)

type watchtowerSnapshotsMsg struct {
	snapshots []agent.MemorySnapshot
}

type watchtowerCPUSnapshotsMsg struct {
	snapshots []agent.CPUSnapshot
}

type watchtowerStorageSnapshotsMsg struct {
	snapshots []agent.StorageSnapshot
}

type watchtowerNetworkSnapshotsMsg struct {
	snapshots []agent.NetworkSnapshot
}

type WatchtowerModel struct {
	width            int
	height           int
	inventory        []inventory.TargetHost
	scope            TargetScope
	memoryCollector  MemorySnapshotCollector
	cpuCollector     CPUSnapshotCollector
	storageCollector StorageSnapshotCollector
	networkCollector NetworkSnapshotCollector
	metricFamily     agent.MetricFamily
	memorySnapshots  []agent.MemorySnapshot
	cpuSnapshots     []agent.CPUSnapshot
	storageSnapshots []agent.StorageSnapshot
	networkSnapshots []agent.NetworkSnapshot
	selected         int
	viewMode         watchtowerViewMode
	refreshing       bool
	legacy           Model
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
	return NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, initialTasks, inv, scope, memoryCollector, cpuCollector, nil, nil)
}

func NewWatchtowerModelWithAllCollectors(
	taskChan chan agent.Task,
	logChan chan agent.ExecutionLog,
	hitlChan chan agent.HitlRequest,
	initialTasks []agent.Task,
	inv []inventory.TargetHost,
	scope TargetScope,
	memoryCollector MemorySnapshotCollector,
	cpuCollector CPUSnapshotCollector,
	storageCollector StorageSnapshotCollector,
	networkCollector NetworkSnapshotCollector,
) WatchtowerModel {
	if scope.Kind == 0 && len(scope.Hosts) == 0 {
		scope = TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(inv)}
	}

	return WatchtowerModel{
		inventory:        cloneHosts(inv),
		scope:            scope.Clone(),
		memoryCollector:  memoryCollector,
		cpuCollector:     cpuCollector,
		storageCollector: storageCollector,
		networkCollector: networkCollector,
		metricFamily:     agent.MetricFamilyMemory,
		viewMode:         watchtowerViewFleet,
		legacy:           NewModel(taskChan, logChan, hitlChan, initialTasks),
	}
}

func (m WatchtowerModel) AttachCollectors(
	inv []inventory.TargetHost,
	memoryCollector MemorySnapshotCollector,
	cpuCollector CPUSnapshotCollector,
	storageCollector StorageSnapshotCollector,
	networkCollector NetworkSnapshotCollector,
) WatchtowerModel {
	m.inventory = cloneHosts(inv)
	m.scope = TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(inv)}
	m.memoryCollector = memoryCollector
	m.cpuCollector = cpuCollector
	m.storageCollector = storageCollector
	m.networkCollector = networkCollector
	m.memorySnapshots = nil
	m.cpuSnapshots = nil
	m.storageSnapshots = nil
	m.networkSnapshots = nil
	m.selected = 0
	m.viewMode = watchtowerViewFleet
	m.refreshing = false
	return m
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
	case watchtowerStorageSnapshotsMsg:
		m.storageSnapshots = msg.snapshots
		m.refreshing = false
		m.clampSelection()
		return m, nil
	case watchtowerNetworkSnapshotsMsg:
		m.networkSnapshots = msg.snapshots
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
		case "3":
			m.metricFamily = agent.MetricFamilyStorage
			m.clampSelection()
			return m, nil
		case "4":
			m.metricFamily = agent.MetricFamilyNetwork
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
	width := m.width
	if width == 0 {
		width = 80
	}
	height := m.height
	if height == 0 {
		height = 24
	}

	return constrainSurfaceContent(m.renderDashboard(width, height-1), width, height-1)
}

func constrainSurfaceContent(content string, width, maxLines int) string {
	if width < 0 {
		width = 0
	}
	if maxLines < 0 {
		maxLines = 0
	}

	lines := strings.Split(content, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}

	return strings.Join(lines, "\n")
}

func (m WatchtowerModel) renderDashboard(width, height int) string {
	header := ansi.Truncate(m.renderDashboardHeader(width), width, "")
	footer := ansi.Truncate(m.renderDashboardFooter(), width, "")
	mainHeight := max(height-lipgloss.Height(header)-lipgloss.Height(footer), 1)

	var body string
	if width < watchtowerMinSplitWidth || mainHeight < watchtowerMinPaneHeight*2 {
		body = m.renderCompactDashboard(width, mainHeight)
	} else {
		body = m.renderSplitDashboard(width, mainHeight)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m WatchtowerModel) renderDashboardHeader(width int) string {
	state := "OVERVIEW"
	if m.viewMode == watchtowerViewHostDetail {
		state = "DETAIL"
	}
	refresh := ""
	if m.refreshing {
		refresh = " | REFRESHING"
	}

	header := fmt.Sprintf(
		"WATCHTOWER %s | %s | %s | %s%s",
		m.renderMetricTabs(),
		m.scope.String(),
		m.renderFleetHealthSummary(),
		state,
		refresh,
	)

	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(ansi.Truncate(header, width, ""))
}

func (m WatchtowerModel) renderMetricTabs() string {
	tabs := []struct {
		family agent.MetricFamily
		label  string
		key    string
	}{
		{family: agent.MetricFamilyMemory, label: "MEMORY", key: "1"},
		{family: agent.MetricFamilyCPU, label: "CPU", key: "2"},
		{family: agent.MetricFamilyStorage, label: "STORAGE", key: "3"},
		{family: agent.MetricFamilyNetwork, label: "NETWORK", key: "4"},
	}
	parts := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		label := strings.ToLower(tab.label)
		if m.metricFamily == tab.family {
			label = tab.label
		}
		parts = append(parts, fmt.Sprintf("[%s:%s]", tab.key, label))
	}
	return strings.Join(parts, " ")
}

func (m WatchtowerModel) renderFleetHealthSummary() string {
	return m.currentFamilyHealthSummary()
}

func (m WatchtowerModel) renderDashboardFooter() string {
	if m.viewMode == watchtowerViewHostDetail {
		return "j/k move • b back • 1 memory • 2 cpu • 3 storage • 4 network • r refresh • a autopilot • c copilot"
	}
	return "j/k move • Enter detail • 1 memory • 2 cpu • 3 storage • 4 network • r refresh • a autopilot • c copilot"
}

func (m WatchtowerModel) renderCompactDashboard(width, height int) string {
	if height <= 1 {
		return ansi.Truncate(m.renderCompactSelectionBody(), width, "")
	}

	upperHeight := max(height/2, 1)
	lowerHeight := max(height-upperHeight, 1)
	hosts := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("HOSTS"),
		constrainSurfaceContent(m.renderHostsBody(), width, max(upperHeight-1, 1)),
	)
	title := "FLEET"
	body := m.renderFleetOverviewBody()
	if m.viewMode == watchtowerViewHostDetail {
		title = "HOST DETAIL"
		body = m.renderSelectedHostBody(true)
	} else {
		body = m.renderCompactSelectionBody()
	}
	secondary := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render(title),
		constrainSurfaceContent(body, width, max(lowerHeight-1, 1)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, hosts, secondary)
}

func (m WatchtowerModel) renderSplitDashboard(width, height int) string {
	leftWidth := max(min(width/3, width-24), 20)
	rightWidth := max(width-leftWidth, 20)
	left := m.renderPanel("FLEET MATRIX", m.renderHostsBody(), leftWidth, height, true)

	topHeight := max(height/2, watchtowerMinPaneHeight)
	bottomHeight := max(height-topHeight, watchtowerMinPaneHeight)
	if topHeight+bottomHeight > height {
		bottomHeight = max(height-topHeight, 1)
	}

	top := m.renderPanel("FLEET", m.renderFleetOverviewBody(), rightWidth, topHeight, false)
	bottomTitle := "SELECTION"
	bottomBody := m.renderSelectedHostBody(false)
	if m.viewMode == watchtowerViewHostDetail {
		bottomTitle = "HOST DETAIL"
		bottomBody = m.renderSelectedHostBody(true)
	}
	bottom := m.renderPanel(bottomTitle, bottomBody, rightWidth, bottomHeight, m.viewMode == watchtowerViewHostDetail)

	right := lipgloss.JoinVertical(lipgloss.Left, top, bottom)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m WatchtowerModel) renderPanel(title, body string, width, height int, active bool) string {
	style := panelStyle
	if active {
		style = activePanelStyle
	}

	frameWidth := style.GetHorizontalFrameSize()
	frameHeight := style.GetVerticalFrameSize()
	contentWidth := max(width-frameWidth, 0)
	contentHeight := max(height-frameHeight, 0)
	innerWidth := max(contentWidth-style.GetHorizontalPadding(), 1)
	innerHeight := max(contentHeight-style.GetVerticalPadding(), 1)
	bodyHeight := max(innerHeight-1, 0)

	header := lipgloss.NewStyle().Bold(true).Render(ansi.Truncate(title, innerWidth, ""))
	parts := []string{header}
	if bodyHeight > 0 {
		fitted := lipgloss.NewStyle().Width(innerWidth).Height(bodyHeight).Render(constrainSurfaceContent(body, innerWidth, bodyHeight))
		parts = append(parts, fitted)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return style.Width(contentWidth).Height(contentHeight).Render(content)
}

func (m WatchtowerModel) renderHostsBody() string {
	switch m.metricFamily {
	case agent.MetricFamilyCPU:
		visible := m.visibleCPUSnapshots()
		if len(visible) == 0 {
			return "No CPU snapshots yet.\nPress r to refresh."
		}
		lines := make([]string, 0, len(visible))
		for i, snapshot := range visible {
			marker := " "
			if i == m.selected {
				marker = ">"
			}
			status := cpuFreshnessLabel(snapshot)
			if snapshot.Status == agent.SnapshotStatusFailed {
				status = "FAILED"
			}
			lines = append(lines, fmt.Sprintf("%s %-11s %3.0f%% %s", marker, snapshot.HostAlias, snapshot.UsagePercent, status))
		}
		return strings.Join(lines, "\n")
	case agent.MetricFamilyStorage:
		visible := m.visibleStorageSnapshots()
		if len(visible) == 0 {
			return "No storage snapshots yet.\nPress r to refresh."
		}
		lines := make([]string, 0, len(visible))
		for i, snapshot := range visible {
			marker := " "
			if i == m.selected {
				marker = ">"
			}
			status := storageFreshnessLabel(snapshot)
			if snapshot.Status == agent.SnapshotStatusFailed {
				status = "FAILED"
			}
			lines = append(lines, fmt.Sprintf("%s %-11s %3.0f%% %s", marker, snapshot.HostAlias, snapshot.UsedPercent, status))
		}
		return strings.Join(lines, "\n")
	case agent.MetricFamilyNetwork:
		visible := m.visibleNetworkSnapshots()
		if len(visible) == 0 {
			return "No network snapshots yet.\nPress r to refresh."
		}
		lines := make([]string, 0, len(visible))
		for i, snapshot := range visible {
			marker := " "
			if i == m.selected {
				marker = ">"
			}
			status := networkFreshnessLabel(snapshot)
			if snapshot.Status == agent.SnapshotStatusFailed {
				status = "FAILED"
			}
			lines = append(lines, fmt.Sprintf("%s %-11s %s %s", marker, snapshot.HostAlias, formatBytes(snapshot.RxBytesPerSec+snapshot.TxBytesPerSec)+"/s", status))
		}
		return strings.Join(lines, "\n")
	default:
		visible := m.visibleMemorySnapshots()
		if len(visible) == 0 {
			return "No memory snapshots yet.\nPress r to refresh."
		}
		lines := make([]string, 0, len(visible))
		for i, snapshot := range visible {
			marker := " "
			if i == m.selected {
				marker = ">"
			}
			status := freshnessLabel(snapshot)
			if snapshot.Status == agent.SnapshotStatusFailed {
				status = "FAILED"
			}
			lines = append(lines, fmt.Sprintf("%s %-11s %3.0f%% %s", marker, snapshot.HostAlias, snapshot.UsedPercent, status))
		}
		return strings.Join(lines, "\n")
	}
}

func (m WatchtowerModel) renderFleetOverviewBody() string {
	switch m.metricFamily {
	case agent.MetricFamilyCPU:
		visible := m.visibleCPUSnapshots()
		if len(visible) == 0 {
			return "No CPU fleet metrics available."
		}
		var sum float64
		peak := visible[0]
		failed := 0
		fresh := cpuFreshnessLabel(visible[0])
		for _, snapshot := range visible {
			sum += snapshot.UsagePercent
			if snapshot.UsagePercent > peak.UsagePercent {
				peak = snapshot
			}
			if snapshot.Status == agent.SnapshotStatusFailed {
				failed++
			}
			if fresh == "FAILED" && cpuFreshnessLabel(snapshot) != "FAILED" {
				fresh = cpuFreshnessLabel(snapshot)
			}
		}
		avg := sum / float64(len(visible))
		return strings.Join([]string{
			fmt.Sprintf("Average CPU: %.1f%%", avg),
			fmt.Sprintf("Peak host: %s (%.1f%%)", peak.HostAlias, peak.UsagePercent),
			fmt.Sprintf("Fleet freshness: %s", fresh),
			fmt.Sprintf("Failed collectors: %d", failed),
		}, "\n")
	case agent.MetricFamilyStorage:
		visible := m.visibleStorageSnapshots()
		if len(visible) == 0 {
			return "No storage fleet metrics available."
		}
		var usedTotal uint64
		var capacityTotal uint64
		var percentSum float64
		failed := 0
		peak := visible[0]
		fresh := storageFreshnessLabel(visible[0])
		for _, snapshot := range visible {
			usedTotal += snapshot.UsedBytes
			capacityTotal += snapshot.TotalBytes
			percentSum += snapshot.UsedPercent
			if snapshot.UsedPercent > peak.UsedPercent {
				peak = snapshot
			}
			if snapshot.Status == agent.SnapshotStatusFailed {
				failed++
			}
			if fresh == "FAILED" && storageFreshnessLabel(snapshot) != "FAILED" {
				fresh = storageFreshnessLabel(snapshot)
			}
		}
		avg := percentSum / float64(len(visible))
		return strings.Join([]string{
			fmt.Sprintf("Used storage: %s / %s", formatBytes(usedTotal), formatBytes(capacityTotal)),
			fmt.Sprintf("Average usage: %.1f%%", avg),
			fmt.Sprintf("Peak host: %s (%.1f%%)", peak.HostAlias, peak.UsedPercent),
			fmt.Sprintf("Fleet freshness: %s", fresh),
			fmt.Sprintf("Failed collectors: %d", failed),
		}, "\n")
	case agent.MetricFamilyNetwork:
		visible := m.visibleNetworkSnapshots()
		if len(visible) == 0 {
			return "No network fleet metrics available."
		}
		var rxTotal uint64
		var txTotal uint64
		failed := 0
		peak := visible[0]
		fresh := networkFreshnessLabel(visible[0])
		for _, snapshot := range visible {
			rxTotal += snapshot.RxBytesPerSec
			txTotal += snapshot.TxBytesPerSec
			if snapshot.RxBytesPerSec+snapshot.TxBytesPerSec > peak.RxBytesPerSec+peak.TxBytesPerSec {
				peak = snapshot
			}
			if snapshot.Status == agent.SnapshotStatusFailed {
				failed++
			}
			if fresh == "FAILED" && networkFreshnessLabel(snapshot) != "FAILED" {
				fresh = networkFreshnessLabel(snapshot)
			}
		}
		return strings.Join([]string{
			fmt.Sprintf("Ingress: %s/s", formatBytes(rxTotal)),
			fmt.Sprintf("Egress: %s/s", formatBytes(txTotal)),
			fmt.Sprintf("Peak host: %s (%s/s)", peak.HostAlias, formatBytes(peak.RxBytesPerSec+peak.TxBytesPerSec)),
			fmt.Sprintf("Fleet freshness: %s", fresh),
			fmt.Sprintf("Failed collectors: %d", failed),
		}, "\n")
	default:
		visible := m.visibleMemorySnapshots()
		if len(visible) == 0 {
			return "No memory fleet metrics available."
		}
		var usedTotal uint64
		var capacityTotal uint64
		var percentSum float64
		failed := 0
		peak := visible[0]
		fresh := freshnessLabel(visible[0])
		for _, snapshot := range visible {
			usedTotal += snapshot.UsedBytes
			capacityTotal += snapshot.TotalBytes
			percentSum += snapshot.UsedPercent
			if snapshot.UsedPercent > peak.UsedPercent {
				peak = snapshot
			}
			if snapshot.Status == agent.SnapshotStatusFailed {
				failed++
			}
			if fresh == "FAILED" && freshnessLabel(snapshot) != "FAILED" {
				fresh = freshnessLabel(snapshot)
			}
		}
		avg := percentSum / float64(len(visible))
		return strings.Join([]string{
			fmt.Sprintf("Used memory: %s / %s", formatBytes(usedTotal), formatBytes(capacityTotal)),
			fmt.Sprintf("Average usage: %.1f%%", avg),
			fmt.Sprintf("Peak host: %s (%.1f%%)", peak.HostAlias, peak.UsedPercent),
			fmt.Sprintf("Fleet freshness: %s", fresh),
			fmt.Sprintf("Failed collectors: %d", failed),
		}, "\n")
	}
}

func (m WatchtowerModel) renderSelectedHostBody(detailFocus bool) string {
	switch m.metricFamily {
	case agent.MetricFamilyCPU:
		snapshot, ok := m.selectedCPUSnapshot()
		if !ok {
			return "No host selected."
		}
		lines := []string{
			fmt.Sprintf("Host: %s (%s)", snapshot.HostAlias, snapshot.HostIP),
			fmt.Sprintf("CPU Usage: %.1f%% %s", snapshot.UsagePercent, renderUsageBar(snapshot.UsagePercent, 12)),
			fmt.Sprintf("Freshness: %s", cpuFreshnessLabel(snapshot)),
		}
		if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
			lines = append(lines, fmt.Sprintf("Error: %s", snapshot.Error))
		}
		if detailFocus {
			lines = append(lines, "", "Focused host detail. Press b to return.")
		} else {
			lines = append(lines, "", "Press Enter for focused detail.")
		}
		return strings.Join(lines, "\n")
	case agent.MetricFamilyStorage:
		snapshot, ok := m.selectedStorageSnapshot()
		if !ok {
			return "No host selected."
		}
		lines := []string{
			fmt.Sprintf("Host: %s (%s)", snapshot.HostAlias, snapshot.HostIP),
			fmt.Sprintf("Used: %s / %s", formatBytes(snapshot.UsedBytes), formatBytes(snapshot.TotalBytes)),
			fmt.Sprintf("Usage: %.1f%% %s", snapshot.UsedPercent, renderUsageBar(snapshot.UsedPercent, 12)),
			fmt.Sprintf("Freshness: %s", storageFreshnessLabel(snapshot)),
		}
		if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
			lines = append(lines, fmt.Sprintf("Error: %s", snapshot.Error))
		}
		if detailFocus {
			lines = append(lines, "", "Focused host detail. Press b to return.")
		} else {
			lines = append(lines, "", "Press Enter for focused detail.")
		}
		return strings.Join(lines, "\n")
	case agent.MetricFamilyNetwork:
		snapshot, ok := m.selectedNetworkSnapshot()
		if !ok {
			return "No host selected."
		}
		lines := []string{
			fmt.Sprintf("Host: %s (%s)", snapshot.HostAlias, snapshot.HostIP),
			fmt.Sprintf("Ingress: %s/s", formatBytes(snapshot.RxBytesPerSec)),
			fmt.Sprintf("Egress: %s/s", formatBytes(snapshot.TxBytesPerSec)),
			fmt.Sprintf("Throughput: %s/s", formatBytes(snapshot.RxBytesPerSec+snapshot.TxBytesPerSec)),
			fmt.Sprintf("Freshness: %s", networkFreshnessLabel(snapshot)),
		}
		if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
			lines = append(lines, fmt.Sprintf("Error: %s", snapshot.Error))
		}
		if detailFocus {
			lines = append(lines, "", "Focused host detail. Press b to return.")
		} else {
			lines = append(lines, "", "Press Enter for focused detail.")
		}
		return strings.Join(lines, "\n")
	default:
		snapshot, ok := m.selectedMemorySnapshot()
		if !ok {
			return "No host selected."
		}
		lines := []string{
			fmt.Sprintf("Host: %s (%s)", snapshot.HostAlias, snapshot.HostIP),
			fmt.Sprintf("Used: %s / %s", formatBytes(snapshot.UsedBytes), formatBytes(snapshot.TotalBytes)),
			fmt.Sprintf("Usage: %.1f%% %s", snapshot.UsedPercent, renderUsageBar(snapshot.UsedPercent, 12)),
			fmt.Sprintf("Freshness: %s", freshnessLabel(snapshot)),
		}
		if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
			lines = append(lines, fmt.Sprintf("Error: %s", snapshot.Error))
		}
		if detailFocus {
			lines = append(lines, "", "Focused host detail. Press b to return.")
		} else {
			lines = append(lines, "", "Press Enter for focused detail.")
		}
		return strings.Join(lines, "\n")
	}
}

func (m WatchtowerModel) renderCompactSelectionBody() string {
	if m.viewMode == watchtowerViewHostDetail {
		return m.renderSelectedHostBody(true)
	}
	return strings.Join([]string{m.renderFleetOverviewBody(), "", "Press Enter for focused detail."}, "\n")
}

func (m WatchtowerModel) currentFamilyHealthSummary() string {
	count, failed := 0, 0
	switch m.metricFamily {
	case agent.MetricFamilyCPU:
		count = len(m.visibleCPUSnapshots())
		for _, snapshot := range m.visibleCPUSnapshots() {
			if snapshot.Status == agent.SnapshotStatusFailed {
				failed++
			}
		}
	case agent.MetricFamilyStorage:
		count = len(m.visibleStorageSnapshots())
		for _, snapshot := range m.visibleStorageSnapshots() {
			if snapshot.Status == agent.SnapshotStatusFailed {
				failed++
			}
		}
	case agent.MetricFamilyNetwork:
		count = len(m.visibleNetworkSnapshots())
		for _, snapshot := range m.visibleNetworkSnapshots() {
			if snapshot.Status == agent.SnapshotStatusFailed {
				failed++
			}
		}
	default:
		count = len(m.visibleMemorySnapshots())
		for _, snapshot := range m.visibleMemorySnapshots() {
			if snapshot.Status == agent.SnapshotStatusFailed {
				failed++
			}
		}
	}
	if count == 0 {
		return fmt.Sprintf("NO %s DATA", m.metricFamily)
	}
	return fmt.Sprintf("%d hosts • %d ok • %d failed", count, count-failed, failed)
}

func (m WatchtowerModel) selectedMemorySnapshot() (agent.MemorySnapshot, bool) {
	visible := m.visibleMemorySnapshots()
	if len(visible) == 0 || m.selected >= len(visible) {
		return agent.MemorySnapshot{}, false
	}
	return visible[m.selected], true
}

func (m WatchtowerModel) selectedCPUSnapshot() (agent.CPUSnapshot, bool) {
	visible := m.visibleCPUSnapshots()
	if len(visible) == 0 || m.selected >= len(visible) {
		return agent.CPUSnapshot{}, false
	}
	return visible[m.selected], true
}

func (m WatchtowerModel) selectedStorageSnapshot() (agent.StorageSnapshot, bool) {
	visible := m.visibleStorageSnapshots()
	if len(visible) == 0 || m.selected >= len(visible) {
		return agent.StorageSnapshot{}, false
	}
	return visible[m.selected], true
}

func (m WatchtowerModel) selectedNetworkSnapshot() (agent.NetworkSnapshot, bool) {
	visible := m.visibleNetworkSnapshots()
	if len(visible) == 0 || m.selected >= len(visible) {
		return agent.NetworkSnapshot{}, false
	}
	return visible[m.selected], true
}

func renderUsageBar(percent float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := int((percent / 100) * float64(width))
	filled = max(0, min(filled, width))
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func (m WatchtowerModel) refreshCurrentFamilyCmd(hosts []inventory.TargetHost) tea.Cmd {
	switch m.metricFamily {
	case agent.MetricFamilyCPU:
		return m.refreshCPUCmd(hosts)
	case agent.MetricFamilyStorage:
		return m.refreshStorageCmd(hosts)
	case agent.MetricFamilyNetwork:
		return m.refreshNetworkCmd(hosts)
	default:
		return m.refreshMemoryCmd(hosts)
	}
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

func (m WatchtowerModel) refreshStorageCmd(hosts []inventory.TargetHost) tea.Cmd {
	if m.storageCollector == nil {
		return nil
	}
	return func() tea.Msg {
		snapshots, err := m.storageCollector(context.Background(), hosts)
		if err != nil {
			failed := make([]agent.StorageSnapshot, 0, len(hosts))
			for _, host := range hosts {
				failed = append(failed, agent.StorageSnapshot{HostAlias: host.Alias, HostIP: host.IP, Status: agent.SnapshotStatusFailed, CollectedAt: time.Now(), Error: err.Error()})
			}
			return watchtowerStorageSnapshotsMsg{snapshots: failed}
		}
		return watchtowerStorageSnapshotsMsg{snapshots: snapshots}
	}
}

func (m WatchtowerModel) refreshNetworkCmd(hosts []inventory.TargetHost) tea.Cmd {
	if m.networkCollector == nil {
		return nil
	}
	return func() tea.Msg {
		snapshots, err := m.networkCollector(context.Background(), hosts)
		if err != nil {
			failed := make([]agent.NetworkSnapshot, 0, len(hosts))
			for _, host := range hosts {
				failed = append(failed, agent.NetworkSnapshot{HostAlias: host.Alias, HostIP: host.IP, Status: agent.SnapshotStatusFailed, CollectedAt: time.Now(), Error: err.Error()})
			}
			return watchtowerNetworkSnapshotsMsg{snapshots: failed}
		}
		return watchtowerNetworkSnapshotsMsg{snapshots: snapshots}
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

func (m WatchtowerModel) visibleStorageSnapshots() []agent.StorageSnapshot {
	if len(m.storageSnapshots) == 0 {
		return nil
	}
	visible := make([]agent.StorageSnapshot, 0, len(m.storageSnapshots))
	for _, snapshot := range m.storageSnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		visible = append(visible, snapshot)
	}
	sort.Slice(visible, func(i, j int) bool { return visible[i].HostAlias < visible[j].HostAlias })
	return visible
}

func (m WatchtowerModel) visibleNetworkSnapshots() []agent.NetworkSnapshot {
	if len(m.networkSnapshots) == 0 {
		return nil
	}
	visible := make([]agent.NetworkSnapshot, 0, len(m.networkSnapshots))
	for _, snapshot := range m.networkSnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		visible = append(visible, snapshot)
	}
	sort.Slice(visible, func(i, j int) bool { return visible[i].HostAlias < visible[j].HostAlias })
	return visible
}

func (m WatchtowerModel) visibleRowCount() int {
	switch m.metricFamily {
	case agent.MetricFamilyCPU:
		return len(m.visibleCPUSnapshots())
	case agent.MetricFamilyStorage:
		return len(m.visibleStorageSnapshots())
	case agent.MetricFamilyNetwork:
		return len(m.visibleNetworkSnapshots())
	default:
		return len(m.visibleMemorySnapshots())
	}
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
	switch m.metricFamily {
	case agent.MetricFamilyCPU:
		return m.cpuCollector == nil
	case agent.MetricFamilyStorage:
		return m.storageCollector == nil
	case agent.MetricFamilyNetwork:
		return m.networkCollector == nil
	default:
		return m.memoryCollector == nil
	}
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
	if m.metricFamily == agent.MetricFamilyStorage {
		visible := m.visibleStorageSnapshots()
		if len(visible) == 0 {
			return WatchtowerEscalationPayload{}, false
		}
		snapshot := visible[m.selected]
		observation := fmt.Sprintf("Storage %.1f%% used (%s / %s) • %s", snapshot.UsedPercent, formatBytes(snapshot.UsedBytes), formatBytes(snapshot.TotalBytes), storageFreshnessLabel(snapshot))
		if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
			observation = fmt.Sprintf("Collection failed: %s", snapshot.Error)
		}
		return WatchtowerEscalationPayload{Target: target, MetricFamily: agent.MetricFamilyStorage, Scope: m.scope.Clone(), SelectedHost: snapshot.HostAlias, Observation: observation}, true
	}
	if m.metricFamily == agent.MetricFamilyNetwork {
		visible := m.visibleNetworkSnapshots()
		if len(visible) == 0 {
			return WatchtowerEscalationPayload{}, false
		}
		snapshot := visible[m.selected]
		observation := fmt.Sprintf("Network %s/s in • %s/s out • %s", formatBytes(snapshot.RxBytesPerSec), formatBytes(snapshot.TxBytesPerSec), networkFreshnessLabel(snapshot))
		if snapshot.Status == agent.SnapshotStatusFailed && snapshot.Error != "" {
			observation = fmt.Sprintf("Collection failed: %s", snapshot.Error)
		}
		return WatchtowerEscalationPayload{Target: target, MetricFamily: agent.MetricFamilyNetwork, Scope: m.scope.Clone(), SelectedHost: snapshot.HostAlias, Observation: observation}, true
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

func storageFreshnessLabel(snapshot agent.StorageSnapshot) string {
	if snapshot.Status == agent.SnapshotStatusFailed {
		return "FAILED"
	}
	if snapshot.CollectedAt.IsZero() {
		return "STALE"
	}
	return fmt.Sprintf("UPDATED %s", snapshot.CollectedAt.Format("15:04:05"))
}

func networkFreshnessLabel(snapshot agent.NetworkSnapshot) string {
	if snapshot.Status == agent.SnapshotStatusFailed {
		return "FAILED"
	}
	if snapshot.CollectedAt.IsZero() {
		return "STALE"
	}
	return fmt.Sprintf("UPDATED %s", snapshot.CollectedAt.Format("15:04:05"))
}
