package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/inventory"
)

// Mode identifies one of the three top-level shell surfaces.
type Mode int

const (
	ModeWatchtower Mode = iota
	ModeAutopilot
	ModeCopilot
)

// String returns a human-readable badge label for the mode.
func (m Mode) String() string {
	switch m {
	case ModeWatchtower:
		return "WATCHTOWER"
	case ModeAutopilot:
		return "AUTOPILOT"
	case ModeCopilot:
		return "COPILOT"
	default:
		return "UNKNOWN"
	}
}

// Shell is the root Bubble Tea model. It wraps the existing Watchtower TUI
// and provides interactive layout skeletons for Autopilot and Copilot.
type Shell struct {
	mode       Mode
	width      int
	height     int
	leaderMode bool
	inventory  []inventory.TargetHost
	scope      TargetScope
	watchtower tea.Model
	autopilot  tea.Model
	copilot    tea.Model
}

// NewShell creates the root shell model, defaulting to Watchtower mode and an
// empty Entire Inventory scope.
func NewShell(
	taskChan chan agent.Task,
	logChan chan agent.ExecutionLog,
	hitlChan chan agent.HitlRequest,
	initialTasks []agent.Task,
) Shell {
	initialScope := TargetScope{Kind: ScopeEntireInventory}
	return Shell{
		mode:       ModeWatchtower,
		scope:      initialScope,
		watchtower: NewWatchtowerModel(taskChan, logChan, hitlChan, initialTasks, nil, initialScope, nil),
		autopilot:  NewAutopilotModel(),
		copilot:    NewCopilotModel(),
	}
}

// NewShellWithInventory creates the root shell model bound to an inventory.
// The initial scope is Entire Inventory.
func NewShellWithInventory(
	taskChan chan agent.Task,
	logChan chan agent.ExecutionLog,
	hitlChan chan agent.HitlRequest,
	initialTasks []agent.Task,
	inv []inventory.TargetHost,
) Shell {
	return NewShellWithInventoryAndCollector(taskChan, logChan, hitlChan, initialTasks, inv, nil)
}

// NewShellWithInventoryAndCollector creates the root shell model with
// inventory-backed scope and a Watchtower memory collector.
func NewShellWithInventoryAndCollector(
	taskChan chan agent.Task,
	logChan chan agent.ExecutionLog,
	hitlChan chan agent.HitlRequest,
	initialTasks []agent.Task,
	inv []inventory.TargetHost,
	collector MemorySnapshotCollector,
) Shell {
	return NewShellWithInventoryAndCollectors(taskChan, logChan, hitlChan, initialTasks, inv, collector, nil)
}

// NewShellWithInventoryAndCollectors creates the root shell model with
// inventory-backed scope and Watchtower collectors for multiple metric families.
func NewShellWithInventoryAndCollectors(
	taskChan chan agent.Task,
	logChan chan agent.ExecutionLog,
	hitlChan chan agent.HitlRequest,
	initialTasks []agent.Task,
	inv []inventory.TargetHost,
	memoryCollector MemorySnapshotCollector,
	cpuCollector CPUSnapshotCollector,
) Shell {
	s := NewShell(taskChan, logChan, hitlChan, initialTasks)
	s.inventory = cloneHosts(inv)
	s.scope = TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(inv)}
	s.watchtower = NewWatchtowerModelWithCollectors(taskChan, logChan, hitlChan, initialTasks, inv, s.scope, memoryCollector, cpuCollector)
	return s
}

// ActiveMode returns the currently selected shell mode.
func (s Shell) ActiveMode() Mode {
	return s.mode
}

// StatusBadge returns the mode label displayed as a status surface.
func (s Shell) StatusBadge() string {
	return s.mode.String()
}

// Scope returns the current shell-level Target Scope.
func (s Shell) Scope() TargetScope {
	return s.scope.Clone()
}

// Inventory returns the full inventory backing the shell scope.
func (s Shell) Inventory() []inventory.TargetHost {
	return cloneHosts(s.inventory)
}

// ScopeBadge returns the human-readable scope description for the shell chrome.
func (s Shell) ScopeBadge() string {
	return s.scope.String()
}

// SetEntireInventory resets the scope to the full inventory.
func (s Shell) SetEntireInventory() Shell {
	s.scope = TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(s.inventory)}
	s.syncWatchtowerScope()
	return s
}

// SelectHosts narrows the scope to the given aliases. Each alias must exist in
// the inventory. An empty alias list resets the scope to Entire Inventory.
func (s Shell) SelectHosts(aliases []string) (Shell, error) {
	if len(aliases) == 0 {
		return s.SetEntireInventory(), nil
	}

	selected := make([]inventory.TargetHost, 0, len(aliases))
	for _, alias := range aliases {
		found := false
		for _, h := range s.inventory {
			if h.Alias == alias {
				selected = append(selected, h)
				found = true
				break
			}
		}
		if !found {
			return s, fmt.Errorf("host %q not found in inventory", alias)
		}
	}

	s.scope = TargetScope{Kind: ScopeSelectedHosts, Hosts: selected}
	s.syncWatchtowerScope()
	return s, nil
}

// Hotkeys returns the contextual hotkey surface for the active mode.
func (s Shell) Hotkeys() []string {
	switch s.mode {
	case ModeWatchtower:
		return []string{
			"j/k navigate hosts",
			"Enter host detail",
			"b back",
			"1 memory",
			"2 cpu",
			"r refresh",
			"a escalate autopilot",
			"c escalate copilot",
			"q quit",
			"Ctrl+a w Watchtower",
			"Ctrl+a a Autopilot",
			"Ctrl+a c Copilot",
		}
	case ModeAutopilot:
		return []string{
			"runs autonomously",
			"Ctrl+a w Watchtower",
			"Ctrl+a a Autopilot",
			"Ctrl+a c Copilot",
		}
	case ModeCopilot:
		return []string{
			"collaborative mode",
			"Ctrl+a w Watchtower",
			"Ctrl+a a Autopilot",
			"Ctrl+a c Copilot",
		}
	default:
		return []string{}
	}
}

func (s Shell) Init() tea.Cmd {
	return s.activeChild().Init()
}

func (s Shell) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := strings.ToLower(msg.String())

		if s.leaderMode {
			s.leaderMode = false
			switch key {
			case "w":
				return s.switchMode(ModeWatchtower)
			case "a":
				return s.switchMode(ModeAutopilot)
			case "c":
				return s.switchMode(ModeCopilot)
			}
			return s, nil
		}

		switch key {
		case "ctrl+a":
			s.leaderMode = true
			return s, nil
		case "ctrl+c":
			return s, tea.Quit
		case "q":
			if s.mode != ModeWatchtower {
				return s, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	case watchtowerEscalationMsg:
		return s.applyWatchtowerEscalation(msg.Payload)
	}

	child := s.activeChild()
	updated, cmd := child.Update(msg)
	s.updateChild(updated)

	return s, cmd
}

var shellChromeStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("#383838")).
	Foreground(lipgloss.Color("#FAFAFA")).
	Padding(0, 1)

// View renders the active child surface with a shell chrome line that surfaces
// the current mode and target scope.
func (s Shell) View() string {
	w := s.width
	if w == 0 {
		w = 80
	}

	chromeText := fmt.Sprintf("%s | %s", s.StatusBadge(), s.ScopeBadge())
	if s.leaderMode {
		chromeText += " | LEADER: w/a/c"
	}
	chrome := shellChromeStyle.Width(w - shellChromeStyle.GetHorizontalFrameSize()).Render(chromeText)

	return lipgloss.JoinVertical(lipgloss.Left, chrome, s.activeChild().View())
}

func (s Shell) activeChild() tea.Model {
	switch s.mode {
	case ModeAutopilot:
		return s.autopilot
	case ModeCopilot:
		return s.copilot
	default:
		return s.watchtower
	}
}

func (s *Shell) syncWatchtowerScope() {
	watchtower, ok := s.watchtower.(WatchtowerModel)
	if !ok {
		return
	}
	s.watchtower = watchtower.SetScope(s.scope)
}

func (s *Shell) updateChild(updated tea.Model) {
	switch s.mode {
	case ModeAutopilot:
		s.autopilot = updated
	case ModeCopilot:
		s.copilot = updated
	default:
		s.watchtower = updated
	}
}

func (s Shell) applyWatchtowerEscalation(payload WatchtowerEscalationPayload) (tea.Model, tea.Cmd) {
	s.leaderMode = false
	s.mode = payload.Target

	switch payload.Target {
	case ModeAutopilot:
		autopilot, ok := s.autopilot.(AutopilotModel)
		if ok {
			s.autopilot = autopilot.ApplyWatchtowerEscalation(payload)
		}
	case ModeCopilot:
		copilot, ok := s.copilot.(CopilotModel)
		if ok {
			s.copilot = copilot.ApplyWatchtowerEscalation(payload)
		}
	}

	var cmds []tea.Cmd
	cmds = append(cmds, s.activeChild().Init())
	if s.width > 0 || s.height > 0 {
		cmds = append(cmds, func() tea.Msg {
			return tea.WindowSizeMsg{Width: s.width, Height: s.height}
		})
	}
	return s, tea.Batch(cmds...)
}

func (s Shell) switchMode(m Mode) (tea.Model, tea.Cmd) {
	s.mode = m

	var cmds []tea.Cmd
	cmds = append(cmds, s.activeChild().Init())

	if s.width > 0 || s.height > 0 {
		cmds = append(cmds, func() tea.Msg {
			return tea.WindowSizeMsg{Width: s.width, Height: s.height}
		})
	}

	return s, tea.Batch(cmds...)
}
