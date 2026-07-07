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
	watchtowerViewAggregate watchtowerViewMode = iota
	watchtowerViewMatrix
	watchtowerViewHostDetail
)

var watchtowerMetricFamilyOrder = []agent.MetricFamily{
	agent.MetricFamilyMemory,
	agent.MetricFamilyCPU,
	agent.MetricFamilyStorage,
	agent.MetricFamilyNetwork,
}

const (
	watchtowerMinSplitWidth = 40
	watchtowerMinPaneHeight = 5
	watchtowerMemoryElevatedThreshold = 75.0
	watchtowerMemoryCriticalThreshold = 90.0
	watchtowerCPUElevatedThreshold    = 70.0
	watchtowerCPUCriticalThreshold    = 85.0
	watchtowerStorageElevatedThreshold = 80.0
	watchtowerStorageCriticalThreshold = 90.0
	watchtowerNetworkElevatedThreshold = 64 * 1024 * 1024
	watchtowerNetworkCriticalThreshold = 96 * 1024 * 1024
	watchtowerTrendWindowLimit       = 24
)

var watchtowerMemoryStaleThreshold = 10 * time.Minute

const watchtowerFleetTrendHostAlias = "__fleet__"

type watchtowerSeverity int

const (
	watchtowerSeverityNormal watchtowerSeverity = iota
	watchtowerSeverityElevated
	watchtowerSeverityCritical
)

type watchtowerFreshnessState int

const (
	watchtowerFreshnessFresh watchtowerFreshnessState = iota
	watchtowerFreshnessStale
	watchtowerFreshnessFailed
	watchtowerFreshnessMissing
)

type watchtowerTrendKey struct {
	family    agent.MetricFamily
	hostAlias string
}

type watchtowerTrendPoint struct {
	Value       float64
	CollectedAt time.Time
}

type watchtowerMemoryHostState struct {
	host       inventory.TargetHost
	snapshot   agent.MemorySnapshot
	hasSnapshot bool
	freshness  watchtowerFreshnessState
	severity   watchtowerSeverity
	trend      []watchtowerTrendPoint
}

type watchtowerCPUHostState struct {
	host       inventory.TargetHost
	snapshot   agent.CPUSnapshot
	hasSnapshot bool
	freshness  watchtowerFreshnessState
	severity   watchtowerSeverity
	trend      []watchtowerTrendPoint
}

type watchtowerStorageHostState struct {
	host       inventory.TargetHost
	snapshot   agent.StorageSnapshot
	hasSnapshot bool
	freshness  watchtowerFreshnessState
	severity   watchtowerSeverity
	trend      []watchtowerTrendPoint
}

type watchtowerNetworkHostState struct {
	host       inventory.TargetHost
	snapshot   agent.NetworkSnapshot
	hasSnapshot bool
	freshness  watchtowerFreshnessState
	severity   watchtowerSeverity
	trend      []watchtowerTrendPoint
}

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

type watchtowerRefreshTickMsg struct{}

type watchtowerViewHistoryEntry struct {
	viewMode       watchtowerViewMode
	metricFamily   agent.MetricFamily
	aggregateFocus int
	detailFocus    int
	selected       int
	matrixPage     int
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
	aggregateFocus   int
	detailFocus      int
	matrixPage       int
	matrixPageSize   int
	history          []watchtowerViewHistoryEntry
	refreshing       bool
	now              func() time.Time
	trendWindows     map[watchtowerTrendKey][]watchtowerTrendPoint
	legacy           Model
	refreshInterval  time.Duration
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
	return NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, initialTasks, inv, scope, memoryCollector, cpuCollector, nil, nil, 0)
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
	refreshInterval time.Duration,
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
		viewMode:         watchtowerViewAggregate,
		aggregateFocus:   0,
		detailFocus:      0,
		matrixPage:       0,
		matrixPageSize:   4,
		now:              time.Now,
		trendWindows:     make(map[watchtowerTrendKey][]watchtowerTrendPoint),
		legacy:           NewModel(taskChan, logChan, hitlChan, initialTasks),
		refreshInterval:  refreshInterval,
	}
}

func (m WatchtowerModel) AttachCollectors(
	inv []inventory.TargetHost,
	memoryCollector MemorySnapshotCollector,
	cpuCollector CPUSnapshotCollector,
	storageCollector StorageSnapshotCollector,
	networkCollector NetworkSnapshotCollector,
	refreshInterval time.Duration,
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
	m.viewMode = watchtowerViewAggregate
	m.aggregateFocus = 0
	m.detailFocus = 0
	m.matrixPage = 0
	m.matrixPageSize = 4
	m.history = nil
	m.refreshing = false
	m.refreshInterval = refreshInterval
	m.trendWindows = make(map[watchtowerTrendKey][]watchtowerTrendPoint)
	if m.now == nil {
		m.now = time.Now
	}
	return m
}

func (m WatchtowerModel) Init() tea.Cmd {
	cmds := []tea.Cmd{m.legacy.Init()}
	if len(m.inventory) > 0 {
		for _, cmd := range []tea.Cmd{
			m.refreshMemoryCmd(m.inventory),
			m.refreshCPUCmd(m.inventory),
			m.refreshStorageCmd(m.inventory),
			m.refreshNetworkCmd(m.inventory),
		} {
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	if m.refreshInterval > 0 {
		cmds = append(cmds, m.refreshTickCmd())
	}
	return tea.Batch(cmds...)
}

func (m WatchtowerModel) refreshTickCmd() tea.Cmd {
	return tea.Tick(m.refreshInterval, func(time.Time) tea.Msg {
		return watchtowerRefreshTickMsg{}
	})
}

func (m WatchtowerModel) SetScope(scope TargetScope) WatchtowerModel {
	m.scope = scope.Clone()
	m.matrixPage = 0
	visible := m.visibleRowCount()
	if visible == 0 {
		m.selected = 0
		m.viewMode = watchtowerViewAggregate
		return m
	}
	if m.selected >= visible || m.selected < 0 {
		m.selected = 0
	}
	m.ensureSelectedVisible()
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
		m.recordMemoryTrendWindows(msg.snapshots)
		m.refreshing = false
		m.clampSelection()
		return m, nil
	case watchtowerCPUSnapshotsMsg:
		m.cpuSnapshots = msg.snapshots
		m.recordCPUTrendWindows(msg.snapshots)
		m.refreshing = false
		m.clampSelection()
		return m, nil
	case watchtowerStorageSnapshotsMsg:
		m.storageSnapshots = msg.snapshots
		m.recordStorageTrendWindows(msg.snapshots)
		m.refreshing = false
		m.clampSelection()
		return m, nil
	case watchtowerNetworkSnapshotsMsg:
		m.networkSnapshots = msg.snapshots
		m.recordNetworkTrendWindows(msg.snapshots)
		m.refreshing = false
		m.clampSelection()
		return m, nil
	case watchtowerRefreshTickMsg:
		if m.refreshing || m.currentCollectorMissing() {
			return m, m.refreshTickCmd()
		}
		hosts := m.hostsForRefresh()
		m.refreshing = true
		return m, tea.Batch(m.refreshCurrentFamilyCmd(hosts), m.refreshTickCmd())
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
		case "1", "2", "3", "4":
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(watchtowerMetricFamilyOrder) {
				m.metricFamily = watchtowerMetricFamilyOrder[idx]
				m.aggregateFocus = idx
				if m.viewMode == watchtowerViewHostDetail {
					m.detailFocus = idx
				}
				m.clampSelection()
			}
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
		case "g":
			if m.viewMode == watchtowerViewAggregate {
				return m, nil
			}
			m.viewMode = watchtowerViewAggregate
			return m, nil
		case "m":
			if m.viewMode == watchtowerViewMatrix {
				return m, nil
			}
			m.viewMode = watchtowerViewMatrix
			return m, nil
		case "d":
			if visibleCount == 0 {
				return m, nil
			}
			if m.viewMode != watchtowerViewHostDetail {
				m.pushHistory()
			}
			m.viewMode = watchtowerViewHostDetail
			m.detailFocus = m.metricFamilyIndex()
			return m, nil
		case "enter":
			switch m.viewMode {
			case watchtowerViewMatrix:
				m.pushHistory()
				m.viewMode = watchtowerViewHostDetail
				m.detailFocus = m.metricFamilyIndex()
			case watchtowerViewAggregate:
				m.metricFamily = watchtowerMetricFamilyOrder[m.aggregateFocus]
				m.pushHistory()
				m.viewMode = watchtowerViewMatrix
			case watchtowerViewHostDetail:
				m.metricFamily = watchtowerMetricFamilyOrder[m.detailFocus]
				m.pushHistory()
				m.viewMode = watchtowerViewMatrix
			}
			return m, nil
		case "b", "esc":
			if m.viewMode == watchtowerViewHostDetail || m.viewMode == watchtowerViewMatrix {
				m.popHistory()
			}
			return m, nil
		case "j", "down":
			switch m.viewMode {
			case watchtowerViewMatrix:
				if m.selected < visibleCount-1 {
					m.selected++
				}
				m.ensureSelectedVisible()
			case watchtowerViewAggregate:
				if m.aggregateFocus < len(watchtowerMetricFamilyOrder)-1 {
					m.aggregateFocus++
					m.metricFamily = watchtowerMetricFamilyOrder[m.aggregateFocus]
				}
			case watchtowerViewHostDetail:
				if m.detailFocus < len(watchtowerMetricFamilyOrder)-1 {
					m.detailFocus++
					m.metricFamily = m.detailFocusFamily()
				}
			}
			return m, nil
		case "k", "up":
			switch m.viewMode {
			case watchtowerViewMatrix:
				if m.selected > 0 {
					m.selected--
				}
				m.ensureSelectedVisible()
			case watchtowerViewAggregate:
				if m.aggregateFocus > 0 {
					m.aggregateFocus--
					m.metricFamily = watchtowerMetricFamilyOrder[m.aggregateFocus]
				}
			case watchtowerViewHostDetail:
				if m.detailFocus > 0 {
					m.detailFocus--
					m.metricFamily = m.detailFocusFamily()
				}
			}
			return m, nil
		case "[":
			switch m.viewMode {
			case watchtowerViewMatrix:
				if m.matrixPage > 0 {
					m.matrixPage--
					m.selected = m.matrixPage * m.matrixPageSize
					m.ensureSelectedVisible()
				}
			case watchtowerViewHostDetail:
				if visibleCount > 0 {
					m.selected--
					if m.selected < 0 {
						m.selected = visibleCount - 1
					}
					m.ensureSelectedVisible()
				}
			}
			return m, nil
		case "]":
			switch m.viewMode {
			case watchtowerViewMatrix:
				pageCount := m.matrixPageCount()
				if m.matrixPage < pageCount-1 {
					m.matrixPage++
					m.selected = m.matrixPage * m.matrixPageSize
					m.ensureSelectedVisible()
				}
			case watchtowerViewHostDetail:
				if visibleCount > 0 {
					m.selected++
					if m.selected >= visibleCount {
						m.selected = 0
					}
					m.ensureSelectedVisible()
				}
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *WatchtowerModel) pushHistory() {
	if len(m.history) >= 8 {
		m.history = m.history[1:]
	}
	m.history = append(m.history, watchtowerViewHistoryEntry{
		viewMode:       m.viewMode,
		metricFamily:   m.metricFamily,
		aggregateFocus: m.aggregateFocus,
		detailFocus:    m.detailFocus,
		selected:       m.selected,
		matrixPage:     m.matrixPage,
	})
}

func (m *WatchtowerModel) popHistory() {
	if len(m.history) == 0 {
		m.viewMode = watchtowerViewAggregate
		m.clampSelection()
		return
	}
	entry := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	m.viewMode = entry.viewMode
	m.metricFamily = entry.metricFamily
	m.aggregateFocus = entry.aggregateFocus
	m.detailFocus = entry.detailFocus
	m.selected = entry.selected
	m.matrixPage = entry.matrixPage
	m.clampSelection()
}

func (m WatchtowerModel) metricFamilyIndex() int {
	for i, family := range watchtowerMetricFamilyOrder {
		if family == m.metricFamily {
			return i
		}
	}
	return 0
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

func (m WatchtowerModel) detailFocusFamily() agent.MetricFamily {
	if m.detailFocus >= 0 && m.detailFocus < len(watchtowerMetricFamilyOrder) {
		return watchtowerMetricFamilyOrder[m.detailFocus]
	}
	return agent.MetricFamilyMemory
}

func (m WatchtowerModel) activeMetricFamily() agent.MetricFamily {
	switch m.viewMode {
	case watchtowerViewHostDetail:
		return m.detailFocusFamily()
	case watchtowerViewAggregate:
		if m.aggregateFocus >= 0 && m.aggregateFocus < len(watchtowerMetricFamilyOrder) {
			return watchtowerMetricFamilyOrder[m.aggregateFocus]
		}
	}
	return m.metricFamily
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
	frameWidth := watchtowerAppStyle.GetHorizontalFrameSize()
	frameHeight := watchtowerAppStyle.GetVerticalFrameSize()
	contentWidth := max(width-frameWidth, 1)
	contentHeight := max(height-frameHeight, 1)
	header := ansi.Truncate(m.renderDashboardHeader(contentWidth), contentWidth, "")
	footer := ansi.Truncate(m.renderDashboardFooter(contentWidth), contentWidth, "")
	mainHeight := max(contentHeight-lipgloss.Height(header)-lipgloss.Height(footer), 1)
	body := m.renderWatchtowerBody(contentWidth, mainHeight)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Width(contentWidth).Render(header),
		lipgloss.NewStyle().Width(contentWidth).Height(mainHeight).Render(constrainSurfaceContent(body, contentWidth, mainHeight)),
		lipgloss.NewStyle().Width(contentWidth).Render(footer),
	)
	return watchtowerAppStyle.Width(contentWidth).Height(contentHeight).Render(content)
}

var watchtowerAppStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#383838"))

func (m WatchtowerModel) renderWatchtowerBody(width, height int) string {
	if width < 32 {
		return m.renderStackedWatchtowerBody(width, height)
	}
	return m.renderSplitWatchtowerBody(width, height)
}

func (m WatchtowerModel) renderStackedWatchtowerBody(width, height int) string {
	railHeight := max(min(height/3, 8), 4)
	contentHeight := max(height-railHeight, 1)
	rail := m.renderPanel("FLEET", m.renderFleetRailBody(), width, railHeight, m.viewMode == watchtowerViewMatrix)
	content := lipgloss.NewStyle().Width(width).Height(contentHeight).Render(constrainSurfaceContent(m.renderPrimaryPane(width, contentHeight), width, contentHeight))
	return lipgloss.JoinVertical(lipgloss.Left, rail, content)
}

func (m WatchtowerModel) renderSplitWatchtowerBody(width, height int) string {
	railWidth := max(min(width/5, 22), 16)
	if width-railWidth < 24 {
		railWidth = max(width-24, 14)
	}
	contentWidth := max(width-railWidth, 1)
	rail := m.renderPanel("FLEET", m.renderFleetRailBody(), railWidth, height, m.viewMode == watchtowerViewMatrix)
	content := lipgloss.NewStyle().Width(contentWidth).Height(height).Render(constrainSurfaceContent(m.renderPrimaryPane(contentWidth, height), contentWidth, height))
	return lipgloss.JoinHorizontal(lipgloss.Top, rail, content)
}

func (m WatchtowerModel) renderPrimaryPane(width, height int) string {
	switch m.viewMode {
	case watchtowerViewHostDetail:
		return m.renderDetailPane(width, height)
	case watchtowerViewMatrix:
		return m.renderMatrixPane(width, height)
	default:
		return m.renderAggregatePane(width, height)
	}
}

func (m WatchtowerModel) renderAggregatePane(width, height int) string {
	labels := []string{"MEMORY", "CPU", "STORAGE", "NETWORK"}
	bodies := []string{
		m.renderMemoryAggregateBundle(),
		m.renderCPUAggregateBundle(),
		m.renderStorageAggregateBundle(),
		m.renderNetworkAggregateBundle(),
	}
	activeIndex := m.aggregateFocus
	if width < 48 || height < 12 {
		return m.renderPanelStack(labels, bodies, width, height, activeIndex)
	}
	if width >= 60 && height >= 16 {
		topHeight := max(height/3, 1)
		bottomHeight := max(height-topHeight, 1)
		middleHeight := max((bottomHeight*3)/5, 1)
		networkHeight := max(bottomHeight-middleHeight, 1)
		memoryWidth := max(width/2, 1)
		storageWidth := max(width-memoryWidth, 1)
		cpu := m.renderPanel(labels[1], bodies[1], width, topHeight, activeIndex == 1)
		middle := lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.renderPanel(labels[0], bodies[0], memoryWidth, middleHeight, activeIndex == 0),
			m.renderPanel(labels[2], bodies[2], storageWidth, middleHeight, activeIndex == 2),
		)
		network := m.renderPanel(labels[3], bodies[3], width, networkHeight, activeIndex == 3)
		return lipgloss.JoinVertical(lipgloss.Left, cpu, middle, network)
	}
	leftWidth := max(width/2, 1)
	rightWidth := max(width-leftWidth, 1)
	topHeight := max((height*3)/5, 1)
	bottomHeight := max(height-topHeight, 1)
	top := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderPanel(labels[1], bodies[1], rightWidth, topHeight, activeIndex == 1),
		m.renderPanel(labels[0], bodies[0], leftWidth, topHeight, activeIndex == 0),
	)
	bottom := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderPanel(labels[2], bodies[2], leftWidth, bottomHeight, activeIndex == 2),
		m.renderPanel(labels[3], bodies[3], rightWidth, bottomHeight, activeIndex == 3),
	)
	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

func (m WatchtowerModel) renderMatrixPane(width, height int) string {
	topHeight := max(min(height/4, 7), 4)
	if topHeight >= height {
		topHeight = max(height/2, 1)
	}
	bottomHeight := max(height-topHeight, 1)
	summaryTitle := fmt.Sprintf("%s FLEET", m.metricFamily)
	hostTitle := fmt.Sprintf("%s HOST", m.selectedHostLabel())
	top := m.renderPanel(summaryTitle, m.renderFleetOverviewBody(), width, topHeight, false)
	bottom := m.renderPanel(hostTitle, m.renderSelectedHostBody(false), width, bottomHeight, false)
	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

func (m WatchtowerModel) renderDetailPane(width, height int) string {
	state, ok := m.selectedMemoryHostState()
	if !ok {
		return m.renderPanel("DETAIL", "No host selected.", width, height, true)
	}
	labels := []string{"MEMORY", "CPU", "STORAGE", "NETWORK"}
	bodies := []string{
		m.renderHostDetailModule(agent.MetricFamilyMemory, state),
		m.renderHostDetailModule(agent.MetricFamilyCPU, state),
		m.renderHostDetailModule(agent.MetricFamilyStorage, state),
		m.renderHostDetailModule(agent.MetricFamilyNetwork, state),
	}
	if width < 52 || height < 14 {
		return m.renderPanelStack(labels, bodies, width, height, m.detailFocus)
	}
	cpuWidth := max((width*3)/5, 1)
	memoryWidth := max(width-cpuWidth, 1)
	storageWidth := max((width*11)/20, 1)
	networkWidth := max(width-storageWidth, 1)
	topHeight := max((height*3)/5, 1)
	bottomHeight := max(height-topHeight, 1)
	top := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderPanel(labels[1], bodies[1], cpuWidth, topHeight, m.detailFocus == 1),
		m.renderPanel(labels[0], bodies[0], memoryWidth, topHeight, m.detailFocus == 0),
	)
	bottom := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderPanel(labels[2], bodies[2], storageWidth, bottomHeight, m.detailFocus == 2),
		m.renderPanel(labels[3], bodies[3], networkWidth, bottomHeight, m.detailFocus == 3),
	)
	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

func (m WatchtowerModel) renderPanelStack(labels, bodies []string, width, height, activeIndex int) string {
	if len(labels) == 0 || len(labels) != len(bodies) {
		return ""
	}
	panelHeight := max(height/len(labels), 1)
	remainingHeight := height
	parts := make([]string, 0, len(labels))
	for i := range labels {
		currentHeight := panelHeight
		if i == len(labels)-1 {
			currentHeight = max(remainingHeight, 1)
		}
		parts = append(parts, m.renderPanel(labels[i], bodies[i], width, currentHeight, activeIndex == i))
		remainingHeight -= currentHeight
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m WatchtowerModel) renderDashboardHeader(width int) string {
	state := "AGGREGATE"
	switch m.viewMode {
	case watchtowerViewMatrix:
		state = "MATRIX"
	case watchtowerViewHostDetail:
		state = "DETAIL"
	}
	refresh := ""
	if m.refreshing {
		refresh = " ref"
	}

	family := m.metricFamily
	if m.viewMode == watchtowerViewHostDetail {
		family = m.detailFocusFamily()
	}

	header := fmt.Sprintf(
		"WT %s %s %s %s %s%s",
		state,
		m.selectedHostLabel(),
		m.compactScopeLabel(),
		m.renderMetricTabsForFamily(family),
		m.compactFleetHealthSummaryForFamily(family),
		refresh,
	)

	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Render(ansi.Truncate(header, width, ""))
}

func (m WatchtowerModel) compactScopeLabel() string {
	switch m.scope.Kind {
	case ScopeSelectedHosts:
		return fmt.Sprintf("scope(%d)", len(m.scope.Hosts))
	default:
		return fmt.Sprintf("inv(%d)", len(m.scope.Hosts))
	}
}

func (m WatchtowerModel) compactFleetHealthSummaryForFamily(family agent.MetricFamily) string {
	count, failed := 0, 0
	switch family {
	case agent.MetricFamilyCPU:
		count = len(m.visibleCPUHostStates())
		for _, state := range m.visibleCPUHostStates() {
			if state.freshness == watchtowerFreshnessFailed {
				failed++
			}
		}
	case agent.MetricFamilyStorage:
		count = len(m.visibleStorageHostStates())
		for _, state := range m.visibleStorageHostStates() {
			if state.freshness == watchtowerFreshnessFailed {
				failed++
			}
		}
	case agent.MetricFamilyNetwork:
		count = len(m.visibleNetworkHostStates())
		for _, state := range m.visibleNetworkHostStates() {
			if state.freshness == watchtowerFreshnessFailed {
				failed++
			}
		}
	default:
		count = len(m.visibleMemoryHostStates())
		for _, state := range m.visibleMemoryHostStates() {
			if state.freshness == watchtowerFreshnessFailed {
				failed++
			}
		}
	}
	return fmt.Sprintf("%dh/%df", count, failed)
}

func (m WatchtowerModel) renderMetricTabsForFamily(family agent.MetricFamily) string {
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
		if family == tab.family {
			label = tab.label
		}
		parts = append(parts, fmt.Sprintf("[%s:%s]", tab.key, label))
	}
	return strings.Join(parts, " ")
}

func (m WatchtowerModel) renderFleetHealthSummaryForFamily(family agent.MetricFamily) string {
	return m.currentFamilyHealthSummaryForFamily(family)
}

func (m WatchtowerModel) renderDashboardFooter(width int) string {
	if width <= 36 {
		switch m.viewMode {
		case watchtowerViewHostDetail:
			return "b 1-4 r"
		default:
			return "enter r 1-4"
		}
	}
	if width <= 68 {
		switch m.viewMode {
		case watchtowerViewHostDetail:
			return "[] b g/m/d 1-4 r a/c"
		case watchtowerViewMatrix:
			return "j/k enter g/m/d 1-4 r a/c"
		default:
			return "j/k enter g/m/d 1-4 r a/c"
		}
	}
	switch m.viewMode {
	case watchtowerViewHostDetail:
		return "[] host • b back • g/m/d • 1-4 fam • r • a/c"
	case watchtowerViewMatrix:
		return "j/k move • enter detail • g/m/d • 1-4 fam • r • a/c"
	default:
		return "j/k focus • enter matrix • g/m/d • 1-4 fam • r • a/c"
	}
}

func (m WatchtowerModel) selectedHostLabel() string {
	switch m.activeMetricFamily() {
	case agent.MetricFamilyCPU:
		state, ok := m.selectedCPUHostState()
		if !ok {
			return "no host"
		}
		return state.host.Alias
	case agent.MetricFamilyStorage:
		state, ok := m.selectedStorageHostState()
		if !ok {
			return "no host"
		}
		return state.host.Alias
	case agent.MetricFamilyNetwork:
		state, ok := m.selectedNetworkHostState()
		if !ok {
			return "no host"
		}
		return state.host.Alias
	default:
		state, ok := m.selectedMemoryHostState()
		if !ok {
			return "no host"
		}
		return state.host.Alias
	}
}

func (m WatchtowerModel) renderCompactDashboard(width, height int) string {
	if m.viewMode == watchtowerViewAggregate {
		return m.renderAggregateBody(width, height)
	}

	if m.viewMode == watchtowerViewHostDetail {
		return m.renderHostDetailBody(width, height)
	}

	if height <= 1 {
		return ansi.Truncate(m.renderCompactSelectionBody(), width, "")
	}

	upperHeight := max(height/2, 1)
	lowerHeight := max(height-upperHeight, 1)
	hosts := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render("FLEET MATRIX"),
		constrainSurfaceContent(m.renderHostsBody(), width, max(upperHeight-1, 1)),
	)
	title := "FLEET"
	body := m.renderFleetOverviewBody()
	body = m.renderCompactSelectionBody()
	secondary := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render(title),
		constrainSurfaceContent(body, width, max(lowerHeight-1, 1)),
	)
	return lipgloss.JoinVertical(lipgloss.Left, hosts, secondary)
}

func (m WatchtowerModel) renderSplitDashboard(width, height int) string {
	if m.viewMode == watchtowerViewAggregate {
		return m.renderAggregateBody(width, height)
	}

	if m.viewMode == watchtowerViewHostDetail {
		return m.renderHostDetailBody(width, height)
	}

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
	bottom := m.renderPanel(bottomTitle, bottomBody, rightWidth, bottomHeight, false)

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
	return m.renderFleetRailBody()
}

func (m WatchtowerModel) renderFleetRailBody() string {
	switch m.sidebarFamily() {
	case agent.MetricFamilyCPU:
		visible := m.visibleCPUHostStates()
		if len(visible) == 0 {
			return "No CPU snapshots yet.\nPress r to refresh."
		}
		start, end := m.visiblePageBounds(len(visible))
		page := visible[start:end]
		lines := make([]string, 0, len(page)+2)
		for i, state := range page {
			globalIdx := start + i
			lines = append(lines, m.renderFleetRailRow(globalIdx == m.selected, state.host.Alias, renderCPUMatrixBadges(state)))
		}
		lines = append(lines, "", m.renderFleetRailPageInfo())
		return strings.Join(lines, "\n")
	case agent.MetricFamilyStorage:
		visible := m.visibleStorageHostStates()
		if len(visible) == 0 {
			return "No storage snapshots yet.\nPress r to refresh."
		}
		start, end := m.visiblePageBounds(len(visible))
		page := visible[start:end]
		lines := make([]string, 0, len(page)+2)
		for i, state := range page {
			globalIdx := start + i
			lines = append(lines, m.renderFleetRailRow(globalIdx == m.selected, state.host.Alias, renderStorageMatrixBadges(state)))
		}
		lines = append(lines, "", m.renderFleetRailPageInfo())
		return strings.Join(lines, "\n")
	case agent.MetricFamilyNetwork:
		visible := m.visibleNetworkHostStates()
		if len(visible) == 0 {
			return "No network snapshots yet.\nPress r to refresh."
		}
		start, end := m.visiblePageBounds(len(visible))
		page := visible[start:end]
		lines := make([]string, 0, len(page)+2)
		for i, state := range page {
			globalIdx := start + i
			lines = append(lines, m.renderFleetRailRow(globalIdx == m.selected, state.host.Alias, renderNetworkMatrixBadges(state)))
		}
		lines = append(lines, "", m.renderFleetRailPageInfo())
		return strings.Join(lines, "\n")
	default:
		visible := m.visibleMemoryHostStates()
		if len(visible) == 0 {
			return "No memory snapshots yet.\nPress r to refresh."
		}
		start, end := m.visiblePageBounds(len(visible))
		page := visible[start:end]
		lines := make([]string, 0, len(page)+2)
		for i, state := range page {
			globalIdx := start + i
			lines = append(lines, m.renderFleetRailRow(globalIdx == m.selected, state.host.Alias, renderMemoryMatrixBadges(state)))
		}
		lines = append(lines, "", m.renderFleetRailPageInfo())
		return strings.Join(lines, "\n")
	}
}

func (m WatchtowerModel) renderFleetRailRow(selected bool, alias, badge string) string {
	marker := " "
	if selected {
		marker = ">"
	}
	return fmt.Sprintf("%s %s %s", marker, alias, badge)
}

func (m WatchtowerModel) renderFleetRailPageInfo() string {
	return fmt.Sprintf("%d/%d", m.matrixPage+1, m.matrixPageCount())
}

func (m WatchtowerModel) sidebarFamily() agent.MetricFamily {
	switch m.viewMode {
	case watchtowerViewHostDetail:
		return m.detailFocusFamily()
	case watchtowerViewAggregate:
		if m.aggregateFocus >= 0 && m.aggregateFocus < len(watchtowerMetricFamilyOrder) {
			return watchtowerMetricFamilyOrder[m.aggregateFocus]
		}
	}
	return m.metricFamily
}

func (m WatchtowerModel) visiblePageBounds(total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	pageSize := m.matrixPageSize
	if pageSize <= 0 {
		pageSize = 4
	}
	start := m.matrixPage * pageSize
	if start >= total {
		start = max(total-pageSize, 0)
	}
	end := min(start+pageSize, total)
	return start, end
}

func (m WatchtowerModel) renderAggregateBody(width, height int) string {
	families := watchtowerMetricFamilyOrder
	labels := []string{"MEMORY", "CPU", "STORAGE", "NETWORK"}

	var lines []string
	for i, family := range families {
		marker := "  "
		if i == m.aggregateFocus {
			marker = "> "
		}
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render(marker+labels[i]))
		for line := range strings.SplitSeq(m.renderAggregateBundle(family), "\n") {
			lines = append(lines, "  "+line)
		}
	}

	return constrainSurfaceContent(strings.Join(lines, "\n"), width, height)
}

func (m WatchtowerModel) renderAggregateBundle(family agent.MetricFamily) string {
	switch family {
	case agent.MetricFamilyCPU:
		return m.renderCPUAggregateBundle()
	case agent.MetricFamilyStorage:
		return m.renderStorageAggregateBundle()
	case agent.MetricFamilyNetwork:
		return m.renderNetworkAggregateBundle()
	default:
		return m.renderMemoryAggregateBundle()
	}
}

func (m WatchtowerModel) renderFleetOverviewBody() string {
	return m.renderAggregateBundle(m.metricFamily)
}

func (m WatchtowerModel) renderMemoryAggregateBundle() string {
	states := m.visibleMemoryHostStates()
	if len(states) == 0 {
		return "No memory fleet metrics available."
	}
	var usedTotal uint64
	var capacityTotal uint64
	var percentSum float64
	validCount := 0
	criticalCount := 0
	elevatedCount := 0
	staleCount := 0
	failedCount := 0
	missingCount := 0
	var peak watchtowerMemoryHostState
	havePeak := false
	overallSeverity := watchtowerSeverityNormal
	for _, state := range states {
		switch state.freshness {
		case watchtowerFreshnessMissing:
			missingCount++
		case watchtowerFreshnessFailed:
			failedCount++
		case watchtowerFreshnessStale:
			staleCount++
		}

		if !state.hasSnapshot || state.freshness == watchtowerFreshnessMissing || state.freshness == watchtowerFreshnessFailed {
			continue
		}

		usedTotal += state.snapshot.UsedBytes
		capacityTotal += state.snapshot.TotalBytes
		percentSum += state.snapshot.UsedPercent
		validCount++

		if !havePeak || state.snapshot.UsedPercent > peak.snapshot.UsedPercent {
			peak = state
			havePeak = true
		}
		if state.severity > overallSeverity {
			overallSeverity = state.severity
		}
		if state.freshness == watchtowerFreshnessFresh {
			switch state.severity {
			case watchtowerSeverityCritical:
				criticalCount++
			case watchtowerSeverityElevated:
				elevatedCount++
			}
		}
	}

	usedLine := "Used memory: -- / --"
	averageLine := "Average usage: --.-%"
	peakLine := "Peak host: none"
	severityLine := fmt.Sprintf("Severity: %s", renderAggregateSeverity(overallSeverity))
	if validCount > 0 {
		avg := percentSum / float64(validCount)
		usedLine = fmt.Sprintf("Used memory: %s / %s", formatBytes(usedTotal), formatBytes(capacityTotal))
		averageLine = fmt.Sprintf("Average usage: %.1f%%", avg)
		peakLine = fmt.Sprintf("Peak host: %s (%.1f%%)", peak.host.Alias, peak.snapshot.UsedPercent)
	}

	badges := make([]string, 0, 5)
	if criticalCount > 0 {
		badges = append(badges, fmt.Sprintf("[CRIT %d]", criticalCount))
	}
	if elevatedCount > 0 {
		badges = append(badges, fmt.Sprintf("[ELEV %d]", elevatedCount))
	}
	if staleCount > 0 {
		badges = append(badges, fmt.Sprintf("[STALE %d]", staleCount))
	}
	if failedCount > 0 {
		badges = append(badges, fmt.Sprintf("[FAILED %d]", failedCount))
	}
	if missingCount > 0 {
		badges = append(badges, fmt.Sprintf("[MISSING %d]", missingCount))
	}
	if len(badges) == 0 {
		badges = append(badges, "[OK]")
	}

	fleetTrend := m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyMemory, hostAlias: watchtowerFleetTrendHostAlias}]
	return strings.Join([]string{
		usedLine,
		averageLine,
		peakLine,
		severityLine,
		fmt.Sprintf("Notable: %s", strings.Join(badges, " ")),
		fmt.Sprintf("Trend: %s", renderTrendStrip(fleetTrend, 8)),
	}, "\n")
}

func (m WatchtowerModel) renderCPUAggregateBundle() string {
	states := m.visibleCPUHostStates()
	if len(states) == 0 {
		return "No CPU fleet metrics available."
	}
	var sum float64
	validCount := 0
	criticalCount := 0
	elevatedCount := 0
	staleCount := 0
	failedCount := 0
	missingCount := 0
	var peak watchtowerCPUHostState
	havePeak := false
	overallSeverity := watchtowerSeverityNormal
	for _, state := range states {
		switch state.freshness {
		case watchtowerFreshnessMissing:
			missingCount++
		case watchtowerFreshnessFailed:
			failedCount++
		case watchtowerFreshnessStale:
			staleCount++
		}
		if !state.hasSnapshot || state.freshness == watchtowerFreshnessMissing || state.freshness == watchtowerFreshnessFailed {
			continue
		}
		sum += state.snapshot.UsagePercent
		validCount++
		if !havePeak || state.snapshot.UsagePercent > peak.snapshot.UsagePercent {
			peak = state
			havePeak = true
		}
		if state.severity > overallSeverity {
			overallSeverity = state.severity
		}
		if state.freshness == watchtowerFreshnessFresh {
			switch state.severity {
			case watchtowerSeverityCritical:
				criticalCount++
			case watchtowerSeverityElevated:
				elevatedCount++
			}
		}
	}

	avgLine := "Average CPU: --.-%"
	peakLine := "Peak host: none"
	severityLine := fmt.Sprintf("Severity: %s", renderAggregateSeverity(overallSeverity))
	if validCount > 0 {
		avgLine = fmt.Sprintf("Average CPU: %.1f%%", sum/float64(validCount))
		peakLine = fmt.Sprintf("Peak host: %s (%.1f%%)", peak.host.Alias, peak.snapshot.UsagePercent)
	}

	badges := make([]string, 0, 5)
	if criticalCount > 0 {
		badges = append(badges, fmt.Sprintf("[CRIT %d]", criticalCount))
	}
	if elevatedCount > 0 {
		badges = append(badges, fmt.Sprintf("[ELEV %d]", elevatedCount))
	}
	if staleCount > 0 {
		badges = append(badges, fmt.Sprintf("[STALE %d]", staleCount))
	}
	if failedCount > 0 {
		badges = append(badges, fmt.Sprintf("[FAILED %d]", failedCount))
	}
	if missingCount > 0 {
		badges = append(badges, fmt.Sprintf("[MISSING %d]", missingCount))
	}
	if len(badges) == 0 {
		badges = append(badges, "[OK]")
	}

	fleetTrend := m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyCPU, hostAlias: watchtowerFleetTrendHostAlias}]
	return strings.Join([]string{
		avgLine,
		peakLine,
		severityLine,
		fmt.Sprintf("Notable: %s", strings.Join(badges, " ")),
		fmt.Sprintf("Trend: %s", renderTrendStrip(fleetTrend, 8)),
	}, "\n")
}

func (m WatchtowerModel) renderStorageAggregateBundle() string {
	states := m.visibleStorageHostStates()
	if len(states) == 0 {
		return "No storage fleet metrics available."
	}
	var usedTotal uint64
	var capacityTotal uint64
	var percentSum float64
	validCount := 0
	criticalCount := 0
	elevatedCount := 0
	staleCount := 0
	failedCount := 0
	missingCount := 0
	var peak watchtowerStorageHostState
	havePeak := false
	overallSeverity := watchtowerSeverityNormal
	for _, state := range states {
		switch state.freshness {
		case watchtowerFreshnessMissing:
			missingCount++
		case watchtowerFreshnessFailed:
			failedCount++
		case watchtowerFreshnessStale:
			staleCount++
		}
		if !state.hasSnapshot || state.freshness == watchtowerFreshnessMissing || state.freshness == watchtowerFreshnessFailed {
			continue
		}
		usedTotal += state.snapshot.UsedBytes
		capacityTotal += state.snapshot.TotalBytes
		percentSum += state.snapshot.UsedPercent
		validCount++
		if !havePeak || state.snapshot.UsedPercent > peak.snapshot.UsedPercent {
			peak = state
			havePeak = true
		}
		if state.severity > overallSeverity {
			overallSeverity = state.severity
		}
		if state.freshness == watchtowerFreshnessFresh {
			switch state.severity {
			case watchtowerSeverityCritical:
				criticalCount++
			case watchtowerSeverityElevated:
				elevatedCount++
			}
		}
	}
	avg := 0.0
	if validCount > 0 {
		avg = percentSum / float64(validCount)
	}
	usedLine := "Used storage: -- / --"
	peakLine := "Peak host: none"
	severityLine := fmt.Sprintf("Severity: %s", renderAggregateSeverity(overallSeverity))
	if validCount > 0 {
		usedLine = fmt.Sprintf("Used storage: %s / %s", formatBytes(usedTotal), formatBytes(capacityTotal))
		peakLine = fmt.Sprintf("Peak host: %s (%.1f%%)", peak.host.Alias, peak.snapshot.UsedPercent)
	}
	badges := make([]string, 0, 5)
	if criticalCount > 0 {
		badges = append(badges, fmt.Sprintf("[CRIT %d]", criticalCount))
	}
	if elevatedCount > 0 {
		badges = append(badges, fmt.Sprintf("[ELEV %d]", elevatedCount))
	}
	if staleCount > 0 {
		badges = append(badges, fmt.Sprintf("[STALE %d]", staleCount))
	}
	if failedCount > 0 {
		badges = append(badges, fmt.Sprintf("[FAILED %d]", failedCount))
	}
	if missingCount > 0 {
		badges = append(badges, fmt.Sprintf("[MISSING %d]", missingCount))
	}
	if len(badges) == 0 {
		badges = append(badges, "[OK]")
	}
	fleetTrend := m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyStorage, hostAlias: watchtowerFleetTrendHostAlias}]
	return strings.Join([]string{
		usedLine,
		fmt.Sprintf("Average usage: %.1f%%", avg),
		peakLine,
		severityLine,
		fmt.Sprintf("Notable: %s", strings.Join(badges, " ")),
		fmt.Sprintf("Trend: %s", renderTrendStrip(fleetTrend, 8)),
	}, "\n")
}

func (m WatchtowerModel) renderNetworkAggregateBundle() string {
	states := m.visibleNetworkHostStates()
	if len(states) == 0 {
		return "No network fleet metrics available."
	}
	var rxTotal uint64
	var txTotal uint64
	validCount := 0
	criticalCount := 0
	elevatedCount := 0
	staleCount := 0
	failedCount := 0
	missingCount := 0
	var peak watchtowerNetworkHostState
	havePeak := false
	overallSeverity := watchtowerSeverityNormal
	for _, state := range states {
		switch state.freshness {
		case watchtowerFreshnessMissing:
			missingCount++
		case watchtowerFreshnessFailed:
			failedCount++
		case watchtowerFreshnessStale:
			staleCount++
		}
		if !state.hasSnapshot || state.freshness == watchtowerFreshnessMissing || state.freshness == watchtowerFreshnessFailed {
			continue
		}
		rxTotal += state.snapshot.RxBytesPerSec
		txTotal += state.snapshot.TxBytesPerSec
		validCount++
		if !havePeak || state.snapshot.RxBytesPerSec+state.snapshot.TxBytesPerSec > peak.snapshot.RxBytesPerSec+peak.snapshot.TxBytesPerSec {
			peak = state
			havePeak = true
		}
		if state.severity > overallSeverity {
			overallSeverity = state.severity
		}
		if state.freshness == watchtowerFreshnessFresh {
			switch state.severity {
			case watchtowerSeverityCritical:
				criticalCount++
			case watchtowerSeverityElevated:
				elevatedCount++
			}
		}
	}
	peakLine := "Peak host: none"
	severityLine := fmt.Sprintf("Severity: %s", renderAggregateSeverity(overallSeverity))
	if validCount > 0 {
		peakLine = fmt.Sprintf("Peak host: %s (%s/s)", peak.host.Alias, formatBytes(peak.snapshot.RxBytesPerSec+peak.snapshot.TxBytesPerSec))
	}
	badges := make([]string, 0, 5)
	if criticalCount > 0 {
		badges = append(badges, fmt.Sprintf("[CRIT %d]", criticalCount))
	}
	if elevatedCount > 0 {
		badges = append(badges, fmt.Sprintf("[ELEV %d]", elevatedCount))
	}
	if staleCount > 0 {
		badges = append(badges, fmt.Sprintf("[STALE %d]", staleCount))
	}
	if failedCount > 0 {
		badges = append(badges, fmt.Sprintf("[FAILED %d]", failedCount))
	}
	if missingCount > 0 {
		badges = append(badges, fmt.Sprintf("[MISSING %d]", missingCount))
	}
	if len(badges) == 0 {
		badges = append(badges, "[OK]")
	}
	fleetTrend := m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyNetwork, hostAlias: watchtowerFleetTrendHostAlias}]
	return strings.Join([]string{
		fmt.Sprintf("Ingress: %s/s", formatBytes(rxTotal)),
		fmt.Sprintf("Egress: %s/s", formatBytes(txTotal)),
		peakLine,
		severityLine,
		fmt.Sprintf("Notable: %s", strings.Join(badges, " ")),
		fmt.Sprintf("Trend: %s", renderTrendStrip(fleetTrend, 8)),
	}, "\n")
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
		state, ok := m.selectedNetworkHostState()
		if !ok {
			return "No host selected."
		}
		lines := []string{
			fmt.Sprintf("Host: %s (%s)", state.host.Alias, state.host.IP),
			renderNetworkHostDetailModule(state),
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

func (m WatchtowerModel) renderHostDetailBody(width, height int) string {
	state, ok := m.selectedMemoryHostState()
	if !ok {
		return constrainSurfaceContent("No host selected.", width, height)
	}

	labels := []string{"MEMORY", "CPU", "STORAGE", "NETWORK"}
	var lines []string
	for i, family := range watchtowerMetricFamilyOrder {
		marker := "  "
		if i == m.detailFocus {
			marker = "> "
		}
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render(marker+labels[i]))
		lines = append(lines, "  "+m.renderHostDetailModule(family, state))
	}
	return constrainSurfaceContent(strings.Join(lines, "\n"), width, height)
}

func (m WatchtowerModel) renderHostDetailModule(family agent.MetricFamily, memoryState watchtowerMemoryHostState) string {
	switch family {
	case agent.MetricFamilyCPU:
		state, ok := m.selectedCPUHostState()
		if !ok {
			return "No CPU data."
		}
		return renderCPUHostDetailModule(state)
	case agent.MetricFamilyStorage:
		state, ok := m.selectedStorageHostState()
		if !ok {
			return "No storage data."
		}
		return renderStorageHostDetailModule(state)
	case agent.MetricFamilyNetwork:
		state, ok := m.selectedNetworkHostState()
		if !ok {
			return "No network data."
		}
		return renderNetworkHostDetailModule(state)
	default:
		return renderMemoryHostDetailModule(memoryState)
	}
}

func (m WatchtowerModel) renderCompactSelectionBody() string {
	if m.viewMode == watchtowerViewHostDetail {
		return m.renderSelectedHostBody(true)
	}
	return strings.Join([]string{m.renderFleetOverviewBody(), "", "Press Enter for focused detail."}, "\n")
}

func (m WatchtowerModel) currentFamilyHealthSummaryForFamily(family agent.MetricFamily) string {
	count, failed := 0, 0
	switch family {
	case agent.MetricFamilyCPU:
		count = len(m.visibleCPUHostStates())
		for _, state := range m.visibleCPUHostStates() {
			if state.freshness == watchtowerFreshnessFailed {
				failed++
			}
		}
	case agent.MetricFamilyStorage:
		count = len(m.visibleStorageHostStates())
		for _, state := range m.visibleStorageHostStates() {
			if state.freshness == watchtowerFreshnessFailed {
				failed++
			}
		}
	case agent.MetricFamilyNetwork:
		count = len(m.visibleNetworkHostStates())
		for _, state := range m.visibleNetworkHostStates() {
			if state.freshness == watchtowerFreshnessFailed {
				failed++
			}
		}
	default:
		count = len(m.visibleMemoryHostStates())
		for _, state := range m.visibleMemoryHostStates() {
			if state.freshness == watchtowerFreshnessFailed {
				failed++
			}
		}
	}
	if count == 0 {
		return fmt.Sprintf("NO %s DATA", family)
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

func (m WatchtowerModel) selectedMemoryHostState() (watchtowerMemoryHostState, bool) {
	visible := m.visibleMemoryHostStates()
	if len(visible) == 0 || m.selected >= len(visible) {
		return watchtowerMemoryHostState{}, false
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

func (m WatchtowerModel) selectedCPUHostState() (watchtowerCPUHostState, bool) {
	visible := m.visibleCPUHostStates()
	if len(visible) == 0 || m.selected >= len(visible) {
		return watchtowerCPUHostState{}, false
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

func (m WatchtowerModel) selectedStorageHostState() (watchtowerStorageHostState, bool) {
	visible := m.visibleStorageHostStates()
	if len(visible) == 0 || m.selected >= len(visible) {
		return watchtowerStorageHostState{}, false
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

func (m WatchtowerModel) selectedNetworkHostState() (watchtowerNetworkHostState, bool) {
	visible := m.visibleNetworkHostStates()
	if len(visible) == 0 || m.selected >= len(visible) {
		return watchtowerNetworkHostState{}, false
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
	switch m.activeMetricFamily() {
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

func (m WatchtowerModel) visibleMemoryHosts() []inventory.TargetHost {
	hosts := cloneHosts(m.inventory)
	if m.scope.Kind == ScopeSelectedHosts && len(m.scope.Hosts) > 0 {
		hosts = cloneHosts(m.scope.Hosts)
	}
	known := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		known[host.Alias] = struct{}{}
	}
	for _, snapshot := range m.memorySnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		if _, ok := known[snapshot.HostAlias]; ok {
			continue
		}
		hosts = append(hosts, inventory.TargetHost{Alias: snapshot.HostAlias, IP: snapshot.HostIP})
		known[snapshot.HostAlias] = struct{}{}
	}
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].Alias < hosts[j].Alias })
	return hosts
}

func (m WatchtowerModel) visibleMemoryHostStates() []watchtowerMemoryHostState {
	hosts := m.visibleMemoryHosts()
	if len(hosts) == 0 {
		return nil
	}
	snapshotsByAlias := make(map[string]agent.MemorySnapshot, len(m.memorySnapshots))
	for _, snapshot := range m.memorySnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		snapshotsByAlias[snapshot.HostAlias] = snapshot
	}
	states := make([]watchtowerMemoryHostState, 0, len(hosts))
	for _, host := range hosts {
		state := watchtowerMemoryHostState{host: host}
		snapshot, ok := snapshotsByAlias[host.Alias]
		if ok {
			state.snapshot = snapshot
			state.hasSnapshot = true
			state.freshness = m.memoryFreshnessState(snapshot)
			state.severity = memorySeverity(snapshot)
			state.trend = append([]watchtowerTrendPoint(nil), m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyMemory, hostAlias: host.Alias}]...)
		} else {
			state.freshness = watchtowerFreshnessMissing
			state.trend = append([]watchtowerTrendPoint(nil), m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyMemory, hostAlias: host.Alias}]...)
		}
		states = append(states, state)
	}
	return states
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

func (m WatchtowerModel) visibleCPUHosts() []inventory.TargetHost {
	hosts := cloneHosts(m.inventory)
	if m.scope.Kind == ScopeSelectedHosts && len(m.scope.Hosts) > 0 {
		hosts = cloneHosts(m.scope.Hosts)
	}
	known := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		known[host.Alias] = struct{}{}
	}
	for _, snapshot := range m.cpuSnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		if _, ok := known[snapshot.HostAlias]; ok {
			continue
		}
		hosts = append(hosts, inventory.TargetHost{Alias: snapshot.HostAlias, IP: snapshot.HostIP})
		known[snapshot.HostAlias] = struct{}{}
	}
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].Alias < hosts[j].Alias })
	return hosts
}

func (m WatchtowerModel) visibleCPUHostStates() []watchtowerCPUHostState {
	hosts := m.visibleCPUHosts()
	if len(hosts) == 0 {
		return nil
	}
	snapshotsByAlias := make(map[string]agent.CPUSnapshot, len(m.cpuSnapshots))
	for _, snapshot := range m.cpuSnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		snapshotsByAlias[snapshot.HostAlias] = snapshot
	}
	states := make([]watchtowerCPUHostState, 0, len(hosts))
	for _, host := range hosts {
		state := watchtowerCPUHostState{host: host}
		snapshot, ok := snapshotsByAlias[host.Alias]
		if ok {
			state.snapshot = snapshot
			state.hasSnapshot = true
			state.freshness = m.cpuFreshnessState(snapshot)
			state.severity = cpuSeverity(snapshot)
			state.trend = append([]watchtowerTrendPoint(nil), m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyCPU, hostAlias: host.Alias}]...)
		} else {
			state.freshness = watchtowerFreshnessMissing
			state.trend = append([]watchtowerTrendPoint(nil), m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyCPU, hostAlias: host.Alias}]...)
		}
		states = append(states, state)
	}
	return states
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

func (m WatchtowerModel) visibleStorageHosts() []inventory.TargetHost {
	hosts := cloneHosts(m.inventory)
	if m.scope.Kind == ScopeSelectedHosts && len(m.scope.Hosts) > 0 {
		hosts = cloneHosts(m.scope.Hosts)
	}
	known := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		known[host.Alias] = struct{}{}
	}
	for _, snapshot := range m.storageSnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		if _, ok := known[snapshot.HostAlias]; ok {
			continue
		}
		hosts = append(hosts, inventory.TargetHost{Alias: snapshot.HostAlias, IP: snapshot.HostIP})
		known[snapshot.HostAlias] = struct{}{}
	}
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].Alias < hosts[j].Alias })
	return hosts
}

func (m WatchtowerModel) visibleStorageHostStates() []watchtowerStorageHostState {
	hosts := m.visibleStorageHosts()
	if len(hosts) == 0 {
		return nil
	}
	snapshotsByAlias := make(map[string]agent.StorageSnapshot, len(m.storageSnapshots))
	for _, snapshot := range m.storageSnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		snapshotsByAlias[snapshot.HostAlias] = snapshot
	}
	states := make([]watchtowerStorageHostState, 0, len(hosts))
	for _, host := range hosts {
		state := watchtowerStorageHostState{host: host}
		snapshot, ok := snapshotsByAlias[host.Alias]
		if ok {
			state.snapshot = snapshot
			state.hasSnapshot = true
			state.freshness = m.storageFreshnessState(snapshot)
			state.severity = storageSeverity(snapshot)
			state.trend = append([]watchtowerTrendPoint(nil), m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyStorage, hostAlias: host.Alias}]...)
		} else {
			state.freshness = watchtowerFreshnessMissing
			state.trend = append([]watchtowerTrendPoint(nil), m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyStorage, hostAlias: host.Alias}]...)
		}
		states = append(states, state)
	}
	return states
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

func (m WatchtowerModel) visibleNetworkHosts() []inventory.TargetHost {
	hosts := cloneHosts(m.inventory)
	if m.scope.Kind == ScopeSelectedHosts && len(m.scope.Hosts) > 0 {
		hosts = cloneHosts(m.scope.Hosts)
	}
	known := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		known[host.Alias] = struct{}{}
	}
	for _, snapshot := range m.networkSnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		if _, ok := known[snapshot.HostAlias]; ok {
			continue
		}
		hosts = append(hosts, inventory.TargetHost{Alias: snapshot.HostAlias, IP: snapshot.HostIP})
		known[snapshot.HostAlias] = struct{}{}
	}
	sort.Slice(hosts, func(i, j int) bool { return hosts[i].Alias < hosts[j].Alias })
	return hosts
}

func (m WatchtowerModel) visibleNetworkHostStates() []watchtowerNetworkHostState {
	hosts := m.visibleNetworkHosts()
	if len(hosts) == 0 {
		return nil
	}
	snapshotsByAlias := make(map[string]agent.NetworkSnapshot, len(m.networkSnapshots))
	for _, snapshot := range m.networkSnapshots {
		if m.scope.Kind == ScopeSelectedHosts && !m.scope.Includes(snapshot.HostAlias) {
			continue
		}
		snapshotsByAlias[snapshot.HostAlias] = snapshot
	}
	states := make([]watchtowerNetworkHostState, 0, len(hosts))
	for _, host := range hosts {
		state := watchtowerNetworkHostState{host: host}
		snapshot, ok := snapshotsByAlias[host.Alias]
		if ok {
			state.snapshot = snapshot
			state.hasSnapshot = true
			state.freshness = m.networkFreshnessState(snapshot)
			state.severity = networkSeverity(snapshot)
			state.trend = append([]watchtowerTrendPoint(nil), m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyNetwork, hostAlias: host.Alias}]...)
		} else {
			state.freshness = watchtowerFreshnessMissing
			state.trend = append([]watchtowerTrendPoint(nil), m.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyNetwork, hostAlias: host.Alias}]...)
		}
		states = append(states, state)
	}
	return states
}

func (m WatchtowerModel) visibleRowCount() int {
	switch m.metricFamily {
	case agent.MetricFamilyCPU:
		return len(m.visibleCPUHostStates())
	case agent.MetricFamilyStorage:
		return len(m.visibleStorageHostStates())
	case agent.MetricFamilyNetwork:
		return len(m.visibleNetworkHostStates())
	default:
		return len(m.visibleMemoryHostStates())
	}
}

func (m *WatchtowerModel) recordMemoryTrendWindows(snapshots []agent.MemorySnapshot) {
	for _, snapshot := range snapshots {
		if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.CollectedAt.IsZero() {
			continue
		}
		m.appendTrendPoint(watchtowerTrendKey{family: agent.MetricFamilyMemory, hostAlias: snapshot.HostAlias}, watchtowerTrendPoint{
			Value:       snapshot.UsedPercent,
			CollectedAt: snapshot.CollectedAt,
		})
	}

	var sum float64
	count := 0
	latest := time.Time{}
	for _, snapshot := range snapshots {
		if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.CollectedAt.IsZero() {
			continue
		}
		sum += snapshot.UsedPercent
		count++
		if snapshot.CollectedAt.After(latest) {
			latest = snapshot.CollectedAt
		}
	}
	if count > 0 {
		m.appendTrendPoint(watchtowerTrendKey{family: agent.MetricFamilyMemory, hostAlias: watchtowerFleetTrendHostAlias}, watchtowerTrendPoint{
			Value:       sum / float64(count),
			CollectedAt: latest,
		})
	}
}

func (m *WatchtowerModel) recordCPUTrendWindows(snapshots []agent.CPUSnapshot) {
	for _, snapshot := range snapshots {
		if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.CollectedAt.IsZero() {
			continue
		}
		m.appendTrendPoint(watchtowerTrendKey{family: agent.MetricFamilyCPU, hostAlias: snapshot.HostAlias}, watchtowerTrendPoint{
			Value:       snapshot.UsagePercent,
			CollectedAt: snapshot.CollectedAt,
		})
	}
	var sum float64
	count := 0
	latest := time.Time{}
	for _, snapshot := range snapshots {
		if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.CollectedAt.IsZero() {
			continue
		}
		sum += snapshot.UsagePercent
		count++
		if snapshot.CollectedAt.After(latest) {
			latest = snapshot.CollectedAt
		}
	}
	if count > 0 {
		m.appendTrendPoint(watchtowerTrendKey{family: agent.MetricFamilyCPU, hostAlias: watchtowerFleetTrendHostAlias}, watchtowerTrendPoint{
			Value:       sum / float64(count),
			CollectedAt: latest,
		})
	}
}

func (m *WatchtowerModel) recordStorageTrendWindows(snapshots []agent.StorageSnapshot) {
	for _, snapshot := range snapshots {
		if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.CollectedAt.IsZero() {
			continue
		}
		m.appendTrendPoint(watchtowerTrendKey{family: agent.MetricFamilyStorage, hostAlias: snapshot.HostAlias}, watchtowerTrendPoint{
			Value:       snapshot.UsedPercent,
			CollectedAt: snapshot.CollectedAt,
		})
	}
	var sum float64
	count := 0
	latest := time.Time{}
	for _, snapshot := range snapshots {
		if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.CollectedAt.IsZero() {
			continue
		}
		sum += snapshot.UsedPercent
		count++
		if snapshot.CollectedAt.After(latest) {
			latest = snapshot.CollectedAt
		}
	}
	if count > 0 {
		m.appendTrendPoint(watchtowerTrendKey{family: agent.MetricFamilyStorage, hostAlias: watchtowerFleetTrendHostAlias}, watchtowerTrendPoint{
			Value:       sum / float64(count),
			CollectedAt: latest,
		})
	}
}

func (m *WatchtowerModel) recordNetworkTrendWindows(snapshots []agent.NetworkSnapshot) {
	for _, snapshot := range snapshots {
		if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.CollectedAt.IsZero() {
			continue
		}
		m.appendTrendPoint(watchtowerTrendKey{family: agent.MetricFamilyNetwork, hostAlias: snapshot.HostAlias}, watchtowerTrendPoint{
			Value:       float64(snapshot.RxBytesPerSec + snapshot.TxBytesPerSec),
			CollectedAt: snapshot.CollectedAt,
		})
	}
	var sum float64
	count := 0
	latest := time.Time{}
	for _, snapshot := range snapshots {
		if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.CollectedAt.IsZero() {
			continue
		}
		sum += float64(snapshot.RxBytesPerSec + snapshot.TxBytesPerSec)
		count++
		if snapshot.CollectedAt.After(latest) {
			latest = snapshot.CollectedAt
		}
	}
	if count > 0 {
		m.appendTrendPoint(watchtowerTrendKey{family: agent.MetricFamilyNetwork, hostAlias: watchtowerFleetTrendHostAlias}, watchtowerTrendPoint{
			Value:       sum / float64(count),
			CollectedAt: latest,
		})
	}
}

func (m *WatchtowerModel) appendTrendPoint(key watchtowerTrendKey, point watchtowerTrendPoint) {
	window := append(m.trendWindows[key], point)
	if len(window) > watchtowerTrendWindowLimit {
		window = append([]watchtowerTrendPoint(nil), window[len(window)-watchtowerTrendWindowLimit:]...)
		m.trendWindows[key] = window
		return
	}
	m.trendWindows[key] = window
}

func (m WatchtowerModel) memoryFreshnessState(snapshot agent.MemorySnapshot) watchtowerFreshnessState {
	if snapshot.Status == agent.SnapshotStatusFailed {
		return watchtowerFreshnessFailed
	}
	if snapshot.CollectedAt.IsZero() {
		return watchtowerFreshnessStale
	}
	now := time.Now()
	if m.now != nil {
		now = m.now()
	}
	if now.Sub(snapshot.CollectedAt) > watchtowerMemoryStaleThreshold {
		return watchtowerFreshnessStale
	}
	return watchtowerFreshnessFresh
}

func (m WatchtowerModel) cpuFreshnessState(snapshot agent.CPUSnapshot) watchtowerFreshnessState {
	if snapshot.Status == agent.SnapshotStatusFailed {
		return watchtowerFreshnessFailed
	}
	if snapshot.CollectedAt.IsZero() {
		return watchtowerFreshnessStale
	}
	now := time.Now()
	if m.now != nil {
		now = m.now()
	}
	if now.Sub(snapshot.CollectedAt) > watchtowerMemoryStaleThreshold {
		return watchtowerFreshnessStale
	}
	return watchtowerFreshnessFresh
}

func (m WatchtowerModel) storageFreshnessState(snapshot agent.StorageSnapshot) watchtowerFreshnessState {
	if snapshot.Status == agent.SnapshotStatusFailed {
		return watchtowerFreshnessFailed
	}
	if snapshot.CollectedAt.IsZero() {
		return watchtowerFreshnessStale
	}
	now := time.Now()
	if m.now != nil {
		now = m.now()
	}
	if now.Sub(snapshot.CollectedAt) > watchtowerMemoryStaleThreshold {
		return watchtowerFreshnessStale
	}
	return watchtowerFreshnessFresh
}

func (m WatchtowerModel) networkFreshnessState(snapshot agent.NetworkSnapshot) watchtowerFreshnessState {
	if snapshot.Status == agent.SnapshotStatusFailed {
		return watchtowerFreshnessFailed
	}
	if snapshot.CollectedAt.IsZero() {
		return watchtowerFreshnessStale
	}
	now := time.Now()
	if m.now != nil {
		now = m.now()
	}
	if now.Sub(snapshot.CollectedAt) > watchtowerMemoryStaleThreshold {
		return watchtowerFreshnessStale
	}
	return watchtowerFreshnessFresh
}

func memorySeverity(snapshot agent.MemorySnapshot) watchtowerSeverity {
	switch {
	case snapshot.UsedPercent >= watchtowerMemoryCriticalThreshold:
		return watchtowerSeverityCritical
	case snapshot.UsedPercent >= watchtowerMemoryElevatedThreshold:
		return watchtowerSeverityElevated
	default:
		return watchtowerSeverityNormal
	}
}

func cpuSeverity(snapshot agent.CPUSnapshot) watchtowerSeverity {
	switch {
	case snapshot.UsagePercent >= watchtowerCPUCriticalThreshold:
		return watchtowerSeverityCritical
	case snapshot.UsagePercent >= watchtowerCPUElevatedThreshold:
		return watchtowerSeverityElevated
	default:
		return watchtowerSeverityNormal
	}
}

func storageSeverity(snapshot agent.StorageSnapshot) watchtowerSeverity {
	switch {
	case snapshot.UsedPercent >= watchtowerStorageCriticalThreshold:
		return watchtowerSeverityCritical
	case snapshot.UsedPercent >= watchtowerStorageElevatedThreshold:
		return watchtowerSeverityElevated
	default:
		return watchtowerSeverityNormal
	}
}

func networkSeverity(snapshot agent.NetworkSnapshot) watchtowerSeverity {
	throughput := snapshot.RxBytesPerSec + snapshot.TxBytesPerSec
	switch {
	case throughput >= watchtowerNetworkCriticalThreshold:
		return watchtowerSeverityCritical
	case throughput >= watchtowerNetworkElevatedThreshold:
		return watchtowerSeverityElevated
	default:
		return watchtowerSeverityNormal
	}
}

func memoryMatrixValue(state watchtowerMemoryHostState) string {
	if !state.hasSnapshot || state.freshness == watchtowerFreshnessMissing || state.freshness == watchtowerFreshnessFailed {
		return "--.-%"
	}
	return fmt.Sprintf("%5.1f%%", state.snapshot.UsedPercent)
}

func renderMemoryMatrixBadges(state watchtowerMemoryHostState) string {
	switch state.freshness {
	case watchtowerFreshnessMissing:
		return "[MISSING]"
	case watchtowerFreshnessFailed:
		return "[FAILED]"
	case watchtowerFreshnessStale:
		return "[STALE]"
	}
	switch state.severity {
	case watchtowerSeverityCritical:
		return "[CRIT]"
	case watchtowerSeverityElevated:
		return "[ELEV]"
	default:
		return "[OK]"
	}
}

func cpuMatrixValue(state watchtowerCPUHostState) string {
	if !state.hasSnapshot || state.freshness == watchtowerFreshnessMissing || state.freshness == watchtowerFreshnessFailed {
		return "--.-%"
	}
	return fmt.Sprintf("%5.1f%%", state.snapshot.UsagePercent)
}

func renderCPUMatrixBadges(state watchtowerCPUHostState) string {
	switch state.freshness {
	case watchtowerFreshnessMissing:
		return "[MISSING]"
	case watchtowerFreshnessFailed:
		return "[FAILED]"
	case watchtowerFreshnessStale:
		return "[STALE]"
	}
	switch state.severity {
	case watchtowerSeverityCritical:
		return "[CRIT]"
	case watchtowerSeverityElevated:
		return "[ELEV]"
	default:
		return "[OK]"
	}
}

func storageMatrixValue(state watchtowerStorageHostState) string {
	if !state.hasSnapshot || state.freshness == watchtowerFreshnessMissing || state.freshness == watchtowerFreshnessFailed {
		return "--.-%"
	}
	return fmt.Sprintf("%5.1f%%", state.snapshot.UsedPercent)
}

func renderStorageMatrixBadges(state watchtowerStorageHostState) string {
	switch state.freshness {
	case watchtowerFreshnessMissing:
		return "[MISSING]"
	case watchtowerFreshnessFailed:
		return "[FAILED]"
	case watchtowerFreshnessStale:
		return "[STALE]"
	}
	switch state.severity {
	case watchtowerSeverityCritical:
		return "[CRIT]"
	case watchtowerSeverityElevated:
		return "[ELEV]"
	default:
		return "[OK]"
	}
}

func networkMatrixValue(state watchtowerNetworkHostState) string {
	if !state.hasSnapshot || state.freshness == watchtowerFreshnessMissing || state.freshness == watchtowerFreshnessFailed {
		return "--.-/s"
	}
	return fmt.Sprintf("%s/s", formatBytes(state.snapshot.RxBytesPerSec+state.snapshot.TxBytesPerSec))
}

func renderNetworkMatrixBadges(state watchtowerNetworkHostState) string {
	switch state.freshness {
	case watchtowerFreshnessMissing:
		return "[MISSING]"
	case watchtowerFreshnessFailed:
		return "[FAILED]"
	case watchtowerFreshnessStale:
		return "[STALE]"
	}
	switch state.severity {
	case watchtowerSeverityCritical:
		return "[CRIT]"
	case watchtowerSeverityElevated:
		return "[ELEV]"
	default:
		return "[OK]"
	}
}

func renderTrendStrip(points []watchtowerTrendPoint, width int) string {
	if width <= 0 {
		return ""
	}
	if len(points) == 0 {
		return strings.Repeat("·", width)
	}
	if len(points) > width {
		points = points[len(points)-width:]
	}
	minValue := points[0].Value
	maxValue := points[0].Value
	for _, point := range points[1:] {
		if point.Value < minValue {
			minValue = point.Value
		}
		if point.Value > maxValue {
			maxValue = point.Value
		}
	}
	levels := []rune("▁▂▃▄▅▆▇█")
	var builder strings.Builder
	padding := width - len(points)
	for i := 0; i < padding; i++ {
		builder.WriteRune('·')
	}
	for _, point := range points {
		if maxValue == minValue {
			builder.WriteRune(levels[len(levels)-1])
			continue
		}
		idx := int(((point.Value - minValue) / (maxValue - minValue)) * float64(len(levels)-1))
		idx = max(0, min(idx, len(levels)-1))
		builder.WriteRune(levels[idx])
	}
	return builder.String()
}

func renderAggregateSeverity(severity watchtowerSeverity) string {
	switch severity {
	case watchtowerSeverityCritical:
		return "CRITICAL"
	case watchtowerSeverityElevated:
		return "ELEVATED"
	default:
		return "HEALTHY"
	}
}

func renderMemoryHostDetailModule(state watchtowerMemoryHostState) string {
	lines := []string{
		fmt.Sprintf("Status: %s", renderMemoryMatrixBadges(state)),
		fmt.Sprintf("Trend: %s", renderTrendStrip(state.trend, 8)),
	}
	switch state.freshness {
	case watchtowerFreshnessMissing:
		lines = append(lines, "No memory snapshot collected yet")
		return strings.Join(lines, "\n")
	case watchtowerFreshnessFailed:
		lines = append(lines, "Memory snapshot collection failed")
		if state.snapshot.Error != "" {
			lines = append(lines, fmt.Sprintf("Error: %s", state.snapshot.Error))
		}
		return strings.Join(lines, "\n")
	}
	lines = append(lines,
		fmt.Sprintf("Used: %s / %s", formatBytes(state.snapshot.UsedBytes), formatBytes(state.snapshot.TotalBytes)),
		fmt.Sprintf("Usage: %.1f%% %s", state.snapshot.UsedPercent, renderUsageBar(state.snapshot.UsedPercent, 12)),
		fmt.Sprintf("Freshness: %s", memoryFreshnessLabel(state.freshness)),
	)
	return strings.Join(lines, "\n")
}

func renderCPUHostDetailModule(state watchtowerCPUHostState) string {
	lines := []string{
		fmt.Sprintf("Status: %s", renderCPUMatrixBadges(state)),
		fmt.Sprintf("Trend: %s", renderTrendStrip(state.trend, 8)),
	}
	switch state.freshness {
	case watchtowerFreshnessMissing:
		lines = append(lines, "No CPU snapshot collected yet")
		return strings.Join(lines, "\n")
	case watchtowerFreshnessFailed:
		lines = append(lines, "CPU snapshot collection failed")
		if state.snapshot.Error != "" {
			lines = append(lines, fmt.Sprintf("Error: %s", state.snapshot.Error))
		}
		return strings.Join(lines, "\n")
	}
	lines = append(lines,
		fmt.Sprintf("CPU Usage: %.1f%% %s", state.snapshot.UsagePercent, renderUsageBar(state.snapshot.UsagePercent, 12)),
		fmt.Sprintf("Freshness: %s", memoryFreshnessLabel(state.freshness)),
	)
	return strings.Join(lines, "\n")
}

func renderStorageHostDetailModule(state watchtowerStorageHostState) string {
	lines := []string{
		fmt.Sprintf("Status: %s", renderStorageMatrixBadges(state)),
		fmt.Sprintf("Trend: %s", renderTrendStrip(state.trend, 8)),
	}
	switch state.freshness {
	case watchtowerFreshnessMissing:
		lines = append(lines, "No storage snapshot collected yet")
		return strings.Join(lines, "\n")
	case watchtowerFreshnessFailed:
		lines = append(lines, "Storage snapshot collection failed")
		if state.snapshot.Error != "" {
			lines = append(lines, fmt.Sprintf("Error: %s", state.snapshot.Error))
		}
		return strings.Join(lines, "\n")
	}
	lines = append(lines,
		fmt.Sprintf("Used: %s / %s", formatBytes(state.snapshot.UsedBytes), formatBytes(state.snapshot.TotalBytes)),
		fmt.Sprintf("Usage: %.1f%% %s", state.snapshot.UsedPercent, renderUsageBar(state.snapshot.UsedPercent, 12)),
		fmt.Sprintf("Freshness: %s", memoryFreshnessLabel(state.freshness)),
	)
	return strings.Join(lines, "\n")
}

func renderNetworkHostDetailModule(state watchtowerNetworkHostState) string {
	lines := []string{
		fmt.Sprintf("Status: %s", renderNetworkMatrixBadges(state)),
		fmt.Sprintf("Trend: %s", renderTrendStrip(state.trend, 8)),
	}
	switch state.freshness {
	case watchtowerFreshnessMissing:
		lines = append(lines, "No network snapshot collected yet")
		return strings.Join(lines, "\n")
	case watchtowerFreshnessFailed:
		lines = append(lines, "Network snapshot collection failed")
		if state.snapshot.Error != "" {
			lines = append(lines, fmt.Sprintf("Error: %s", state.snapshot.Error))
		}
		return strings.Join(lines, "\n")
	}
	lines = append(lines,
		fmt.Sprintf("Ingress: %s/s", formatBytes(state.snapshot.RxBytesPerSec)),
		fmt.Sprintf("Egress: %s/s", formatBytes(state.snapshot.TxBytesPerSec)),
		fmt.Sprintf("Throughput: %s/s", formatBytes(state.snapshot.RxBytesPerSec+state.snapshot.TxBytesPerSec)),
		fmt.Sprintf("Freshness: %s", memoryFreshnessLabel(state.freshness)),
	)
	return strings.Join(lines, "\n")
}

func memoryFreshnessLabel(state watchtowerFreshnessState) string {
	switch state {
	case watchtowerFreshnessMissing:
		return "MISSING"
	case watchtowerFreshnessFailed:
		return "FAILED"
	case watchtowerFreshnessStale:
		return "STALE"
	default:
		return "FRESH"
	}
}

func (m *WatchtowerModel) clampSelection() {
	visible := m.visibleRowCount()
	if visible == 0 {
		m.selected = 0
		m.matrixPage = 0
		m.viewMode = watchtowerViewAggregate
		return
	}
	if m.selected >= visible || m.selected < 0 {
		m.selected = 0
	}
	m.ensureSelectedVisible()
}

func (m *WatchtowerModel) ensureSelectedVisible() {
	visible := m.visibleRowCount()
	if visible == 0 {
		m.matrixPage = 0
		return
	}
	if m.matrixPageSize <= 0 {
		m.matrixPageSize = 4
	}
	pageCount := m.matrixPageCount()
	if m.matrixPage < 0 {
		m.matrixPage = 0
	}
	if m.matrixPage >= pageCount {
		m.matrixPage = pageCount - 1
	}
	start := m.matrixPage * m.matrixPageSize
	end := min(start+m.matrixPageSize, visible)
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= end {
		m.matrixPage = min(m.selected/m.matrixPageSize, pageCount-1)
		start = m.matrixPage * m.matrixPageSize
		end = min(start+m.matrixPageSize, visible)
	}
	if m.selected < start {
		m.selected = start
	}
	if m.selected >= end {
		m.selected = end - 1
	}
}

func (m *WatchtowerModel) matrixPageCount() int {
	visible := m.visibleRowCount()
	if visible == 0 {
		return 1
	}
	if m.matrixPageSize <= 0 {
		m.matrixPageSize = 4
	}
	return max((visible+m.matrixPageSize-1)/m.matrixPageSize, 1)
}

func (m WatchtowerModel) currentCollectorMissing() bool {
	switch m.activeMetricFamily() {
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
	if m.activeMetricFamily() == agent.MetricFamilyCPU {
		states := m.visibleCPUHostStates()
		if len(states) == 0 {
			return WatchtowerEscalationPayload{}, false
		}
		state := states[m.selected]
		observation := fmt.Sprintf("CPU %.1f%% used • %s", state.snapshot.UsagePercent, renderFreshnessLabel(state.freshness))
		if state.freshness == watchtowerFreshnessFailed {
			observation = fmt.Sprintf("Collection failed: %s", state.snapshot.Error)
		}
		return WatchtowerEscalationPayload{
			Target:       target,
			MetricFamily: agent.MetricFamilyCPU,
			Scope:        m.scope.Clone(),
			SelectedHost: state.host.Alias,
			Observation:  observation,
			ViewMode:     m.viewMode,
		}, true
	}
	if m.activeMetricFamily() == agent.MetricFamilyStorage {
		states := m.visibleStorageHostStates()
		if len(states) == 0 {
			return WatchtowerEscalationPayload{}, false
		}
		state := states[m.selected]
		observation := fmt.Sprintf("Storage %.1f%% used (%s / %s) • %s", state.snapshot.UsedPercent, formatBytes(state.snapshot.UsedBytes), formatBytes(state.snapshot.TotalBytes), renderFreshnessLabel(state.freshness))
		if state.freshness == watchtowerFreshnessFailed {
			observation = fmt.Sprintf("Collection failed: %s", state.snapshot.Error)
		}
		return WatchtowerEscalationPayload{
			Target:       target,
			MetricFamily: agent.MetricFamilyStorage,
			Scope:        m.scope.Clone(),
			SelectedHost: state.host.Alias,
			Observation:  observation,
			ViewMode:     m.viewMode,
		}, true
	}
	if m.activeMetricFamily() == agent.MetricFamilyNetwork {
		states := m.visibleNetworkHostStates()
		if len(states) == 0 {
			return WatchtowerEscalationPayload{}, false
		}
		state := states[m.selected]
		observation := fmt.Sprintf("Network %s/s in • %s/s out • %s", formatBytes(state.snapshot.RxBytesPerSec), formatBytes(state.snapshot.TxBytesPerSec), renderFreshnessLabel(state.freshness))
		if state.freshness == watchtowerFreshnessFailed {
			observation = fmt.Sprintf("Collection failed: %s", state.snapshot.Error)
		}
		return WatchtowerEscalationPayload{
			Target:       target,
			MetricFamily: agent.MetricFamilyNetwork,
			Scope:        m.scope.Clone(),
			SelectedHost: state.host.Alias,
			Observation:  observation,
			ViewMode:     m.viewMode,
		}, true
	}
	states := m.visibleMemoryHostStates()
	if len(states) == 0 {
		return WatchtowerEscalationPayload{}, false
	}
	state := states[m.selected]
	observation := fmt.Sprintf("Memory %.1f%% used (%s / %s) • %s", state.snapshot.UsedPercent, formatBytes(state.snapshot.UsedBytes), formatBytes(state.snapshot.TotalBytes), renderFreshnessLabel(state.freshness))
	if state.freshness == watchtowerFreshnessFailed {
		observation = fmt.Sprintf("Collection failed: %s", state.snapshot.Error)
	}
	return WatchtowerEscalationPayload{
		Target:       target,
		MetricFamily: agent.MetricFamilyMemory,
		Scope:        m.scope.Clone(),
		SelectedHost: state.host.Alias,
		Observation:  observation,
		ViewMode:     m.viewMode,
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

func renderFreshnessLabel(freshness watchtowerFreshnessState) string {
	switch freshness {
	case watchtowerFreshnessMissing:
		return "MISSING"
	case watchtowerFreshnessFailed:
		return "FAILED"
	case watchtowerFreshnessStale:
		return "STALE"
	default:
		return "FRESH"
	}
}
