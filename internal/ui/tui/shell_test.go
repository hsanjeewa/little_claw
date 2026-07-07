package tui

import (
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/devops/agent/internal/domain/agent"
)

func assertRenderedWithinBounds(t *testing.T, view string, width, height int) {
	t.Helper()

	if got := lipgloss.Width(view); got > width {
		t.Fatalf("expected rendered width <= %d, got %d\nview:\n%s", width, got, view)
	}
	if got := lipgloss.Height(view); got > height {
		t.Fatalf("expected rendered height <= %d, got %d\nview:\n%s", height, got, view)
	}
}

func testChannels() (chan agent.Task, chan agent.ExecutionLog, chan agent.HitlRequest) {
	return make(chan agent.Task, 1), make(chan agent.ExecutionLog, 1), make(chan agent.HitlRequest, 1)
}

func isQuitCmd(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	return ok
}

func leaderSwitch(t *testing.T, shell Shell, key rune) Shell {
	t.Helper()

	updated, _ := shell.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	shell = updated.(Shell)

	updated, _ = shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
	return updated.(Shell)
}

func TestNewShell_DefaultsToWatchtower(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	if shell.ActiveMode() != ModeWatchtower {
		t.Fatalf("expected default mode Watchtower, got %v", shell.ActiveMode())
	}
}

func TestShell_StatusBadge(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	badge := shell.StatusBadge()
	if !strings.Contains(badge, "WATCHTOWER") {
		t.Fatalf("expected status badge to contain WATCHTOWER, got %q", badge)
	}
}

func TestShell_Hotkeys_Watchtower(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	hotkeys := shell.Hotkeys()
	if len(hotkeys) == 0 {
		t.Fatal("expected non-empty hotkeys for Watchtower")
	}
	if !slices.Contains(hotkeys, "q quit") {
		t.Fatalf("expected Watchtower hotkeys to contain 'q quit', got %v", hotkeys)
	}
	if !slices.Contains(hotkeys, "Ctrl+a w Watchtower") {
		t.Fatalf("expected Watchtower hotkeys to contain leader navigation, got %v", hotkeys)
	}
}

func TestShell_SwitchModeWithLeaderKeys(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	shell = leaderSwitch(t, shell, 'a')
	if shell.ActiveMode() != ModeAutopilot {
		t.Fatalf("expected Autopilot after Ctrl+a,a, got %v", shell.ActiveMode())
	}

	shell = leaderSwitch(t, shell, 'c')
	if shell.ActiveMode() != ModeCopilot {
		t.Fatalf("expected Copilot after Ctrl+a,c, got %v", shell.ActiveMode())
	}

	shell = leaderSwitch(t, shell, 'w')
	if shell.ActiveMode() != ModeWatchtower {
		t.Fatalf("expected Watchtower after Ctrl+a,w, got %v", shell.ActiveMode())
	}
}

func TestShell_AttachesSimulatorWithLeaderKey(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := leaderSwitch(t, NewShell(taskChan, logChan, hitlChan, nil), 'a')

	updated, _ := shell.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	shell = updated.(Shell)
	updated, cmd := shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	shell = updated.(Shell)

	if shell.ActiveMode() != ModeWatchtower {
		t.Fatalf("expected Ctrl+a,z to switch to Watchtower, got %v", shell.ActiveMode())
	}
	if shell.Scope().Kind != ScopeEntireInventory {
		t.Fatalf("expected simulator attach to reset scope to EntireInventory, got %v", shell.Scope().Kind)
	}
	if len(shell.Inventory()) < 4 {
		t.Fatalf("expected simulator fleet with at least 4 hosts, got %d", len(shell.Inventory()))
	}
	if cmd == nil {
		t.Fatal("expected simulator attach to trigger refresh command")
	}

	msg := cmd()
	updated, _ = shell.Update(msg)
	shell = updated.(Shell)

	view := shell.View()
	if !strings.Contains(view, "api-east-01") {
		t.Fatalf("expected simulator host in Watchtower view, got:\n%s", view)
	}
	if !strings.Contains(view, "MEMORY") {
		t.Fatalf("expected Watchtower to render simulator metric family, got:\n%s", view)
	}
	assertRenderedWithinBounds(t, view, 80, 24)
	if strings.Contains(view, "LEADER: w/a/c/z") {
		t.Fatalf("expected leader mode to clear after simulator attach, got:\n%s", view)
	}
}

func TestShell_AttachesSimulatorBootstrapsAllWatchtowerFamilies(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := leaderSwitch(t, NewShell(taskChan, logChan, hitlChan, nil), 'a')

	updated, _ := shell.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	shell = updated.(Shell)
	updated, cmd := shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})
	shell = updated.(Shell)
	if cmd == nil {
		t.Fatal("expected simulator attach bootstrap command")
	}

	updated, _ = shell.Update(cmd())
	shell = updated.(Shell)

	watchtower, ok := shell.watchtower.(WatchtowerModel)
	if !ok {
		t.Fatal("expected watchtower child model")
	}

	if len(watchtower.memorySnapshots) == 0 {
		t.Fatal("expected simulator attach to bootstrap memory snapshots")
	}
	if len(watchtower.cpuSnapshots) == 0 {
		t.Fatal("expected simulator attach to bootstrap cpu snapshots")
	}
	if len(watchtower.storageSnapshots) == 0 {
		t.Fatal("expected simulator attach to bootstrap storage snapshots")
	}
	if len(watchtower.networkSnapshots) == 0 {
		t.Fatal("expected simulator attach to bootstrap network snapshots")
	}

	updated, _ = shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	shell = updated.(Shell)
	watchtower, ok = shell.watchtower.(WatchtowerModel)
	if !ok {
		t.Fatal("expected watchtower child model after cpu switch")
	}
	if !strings.Contains(watchtower.renderCPUAggregateBundle(), "%") {
		t.Fatalf("expected bootstrapped cpu simulator data in cpu aggregate bundle, got:\n%s", watchtower.renderCPUAggregateBundle())
	}

	updated, _ = shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	shell = updated.(Shell)
	watchtower, ok = shell.watchtower.(WatchtowerModel)
	if !ok {
		t.Fatal("expected watchtower child model after storage switch")
	}
	if !strings.Contains(watchtower.renderStorageAggregateBundle(), "GiB") {
		t.Fatalf("expected bootstrapped storage simulator data in storage aggregate bundle, got:\n%s", watchtower.renderStorageAggregateBundle())
	}

	updated, _ = shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	shell = updated.(Shell)
	watchtower, ok = shell.watchtower.(WatchtowerModel)
	if !ok {
		t.Fatal("expected watchtower child model after network switch")
	}
	if !strings.Contains(watchtower.renderNetworkAggregateBundle(), "/s") {
		t.Fatalf("expected bootstrapped network simulator data in network aggregate bundle, got:\n%s", watchtower.renderNetworkAggregateBundle())
	}
}

func TestShell_LeaderKeyShowsPendingStateUntilModeKey(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	updated, _ := shell.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	shell = updated.(Shell)

	if !strings.Contains(shell.View(), "LEADER: w/a/c") {
		t.Fatalf("expected shell chrome to show pending leader state, got:\n%s", shell.View())
	}
}

func TestShell_ModePersistsAcrossWindowResize(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	shell = leaderSwitch(t, shell, 'a')

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	shell = updated.(Shell)

	if shell.ActiveMode() != ModeAutopilot {
		t.Fatalf("expected mode to persist as Autopilot after resize, got %v", shell.ActiveMode())
	}
}

func TestShell_WatchtowerViewRendersFleetMatrix(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, testInventory(), nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	shell = updated.(Shell)

	updated, _ = shell.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	shell = updated.(Shell)

	view := shell.View()
	if (!strings.Contains(view, "HOSTS") && !strings.Contains(view, "FLEET") && !strings.Contains(view, "DETAIL") && !strings.Contains(view, "AGGREGATE")) || !strings.Contains(view, "db-master") {
		t.Fatalf("expected Watchtower shell view to render dashboard panes, got:\n%s", view)
	}
}

func TestShell_InitialWatchtowerViewFitsFallbackViewport(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, testInventory(), nil)

	updated, _ := shell.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	shell = updated.(Shell)

	view := shell.View()
	assertRenderedWithinBounds(t, view, 80, 24)

	if !strings.Contains(view, "WATCHTOWER") {
		t.Fatalf("expected shell chrome to remain visible on initial render, got:\n%s", view)
	}
	if !strings.Contains(view, "MEMORY") {
		t.Fatalf("expected watchtower title to remain visible on initial render, got:\n%s", view)
	}
}

func TestShell_WatchtowerViewRespectsWindowResizeBounds(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, testInventory(), nil)

	updated, _ := shell.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	shell = updated.(Shell)

	updated, _ = shell.Update(tea.WindowSizeMsg{Width: 48, Height: 6})
	shell = updated.(Shell)

	compact := shell.View()
	assertRenderedWithinBounds(t, compact, 48, 6)

	updated, _ = shell.Update(tea.WindowSizeMsg{Width: 72, Height: 10})
	shell = updated.(Shell)

	resized := shell.View()
	assertRenderedWithinBounds(t, resized, 72, 10)

	if !strings.Contains(resized, "db-master") {
		t.Fatalf("expected resized watchtower view to keep rendering host data, got:\n%s", resized)
	}
}

func TestShell_AutopilotViewShowsModeLabel(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	shell = updated.(Shell)

	shell = leaderSwitch(t, shell, 'a')

	view := shell.View()
	if !strings.Contains(view, "AUTOPILOT") {
		t.Fatalf("expected Autopilot placeholder view to render label, got:\n%s", view)
	}
}

func TestShell_InitialAutopilotViewFitsFallbackViewport(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := leaderSwitch(t, NewShell(taskChan, logChan, hitlChan, nil), 'a')

	view := shell.View()
	assertRenderedWithinBounds(t, view, 80, 24)

	if !strings.Contains(view, "WATCHTOWER | Entire Inventory") && !strings.Contains(view, "AUTOPILOT | Entire Inventory") {
		t.Fatalf("expected shell chrome to remain visible on initial Autopilot render, got:\n%s", view)
	}
	if !strings.Contains(view, commandBarLabel) {
		t.Fatalf("expected Autopilot command bar to remain visible on initial render, got:\n%s", view)
	}
}

func TestShell_AutopilotViewRespectsWindowResizeBounds(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := leaderSwitch(t, NewShell(taskChan, logChan, hitlChan, nil), 'a')

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 48, Height: 8})
	shell = updated.(Shell)

	compact := shell.View()
	assertRenderedWithinBounds(t, compact, 48, 8)

	updated, _ = shell.Update(tea.WindowSizeMsg{Width: 72, Height: 10})
	shell = updated.(Shell)

	resized := shell.View()
	assertRenderedWithinBounds(t, resized, 72, 10)
	if !strings.Contains(resized, "PLAN") {
		t.Fatalf("expected resized Autopilot view to keep rendering pane labels, got:\n%s", resized)
	}
}

func TestShell_InitialCopilotViewFitsFallbackViewport(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := leaderSwitch(t, NewShell(taskChan, logChan, hitlChan, nil), 'c')

	view := shell.View()
	assertRenderedWithinBounds(t, view, 80, 24)

	if !strings.Contains(view, "COPILOT") {
		t.Fatalf("expected shell chrome to remain visible on initial Copilot render, got:\n%s", view)
	}
	if !strings.Contains(view, commandBarLabel) {
		t.Fatalf("expected Copilot command bar to remain visible on initial render, got:\n%s", view)
	}
}

func TestShell_CopilotViewRespectsWindowResizeBounds(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := leaderSwitch(t, NewShell(taskChan, logChan, hitlChan, nil), 'c')

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 48, Height: 8})
	shell = updated.(Shell)

	compact := shell.View()
	assertRenderedWithinBounds(t, compact, 48, 8)

	updated, _ = shell.Update(tea.WindowSizeMsg{Width: 72, Height: 10})
	shell = updated.(Shell)

	resized := shell.View()
	assertRenderedWithinBounds(t, resized, 72, 10)
	if !strings.Contains(resized, "TERMINAL") {
		t.Fatalf("expected resized Copilot view to keep rendering pane labels, got:\n%s", resized)
	}
}

func TestShell_CopilotDoesNotQuitOnQ(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	shell = leaderSwitch(t, shell, 'c')

	_, cmd := shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if isQuitCmd(cmd) {
		t.Fatal("expected 'q' not to quit in Copilot mode")
	}
}

func TestShell_CopilotQuitsOnSlashExit(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	shell = leaderSwitch(t, shell, 'c')

	for _, r := range "/exit" {
		updated, _ := shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		shell = updated.(Shell)
	}

	_, cmd := shell.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !isQuitCmd(cmd) {
		t.Fatal("expected '/exit' to quit in Copilot mode")
	}
}

func TestShell_AutopilotDoesNotQuitOnQ(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	shell = leaderSwitch(t, shell, 'a')

	_, cmd := shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if isQuitCmd(cmd) {
		t.Fatal("expected 'q' not to quit in Autopilot mode")
	}
}

func TestShell_AutopilotQuitsOnSlashExit(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	shell = leaderSwitch(t, shell, 'a')

	for _, r := range "/exit" {
		updated, _ := shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		shell = updated.(Shell)
	}

	_, cmd := shell.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !isQuitCmd(cmd) {
		t.Fatal("expected '/exit' to quit in Autopilot mode")
	}
}

func TestShell_WatchtowerEscalatesToAutopilotWithContext(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, testInventory(), nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	shell = updated.(Shell)
	updated, _ = shell.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	shell = updated.(Shell)

	_, cmd := shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd == nil {
		t.Fatal("expected escalation command to Autopilot")
	}

	msg := cmd()
	updated, _ = shell.Update(msg)
	shell = updated.(Shell)

	if shell.ActiveMode() != ModeAutopilot {
		t.Fatalf("expected shell to switch to Autopilot, got %v", shell.ActiveMode())
	}
	view := shell.View()
	if !strings.Contains(view, "WATCHTOWER HANDOFF") || !strings.Contains(view, "db-master") || !strings.Contains(view, "MEMORY") {
		t.Fatalf("expected Autopilot to render Watchtower handoff context, got:\n%s", view)
	}
}

func TestShell_WatchtowerEscalatesToCopilotWithContext(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, testInventory(), nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	shell = updated.(Shell)
	updated, _ = shell.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	shell = updated.(Shell)

	_, cmd := shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if cmd == nil {
		t.Fatal("expected escalation command to Copilot")
	}

	msg := cmd()
	updated, _ = shell.Update(msg)
	shell = updated.(Shell)

	if shell.ActiveMode() != ModeCopilot {
		t.Fatalf("expected shell to switch to Copilot, got %v", shell.ActiveMode())
	}
	view := shell.View()
	if !strings.Contains(view, "WATCHTOWER HANDOFF") || !strings.Contains(view, "db-master") || !strings.Contains(view, "MEMORY") {
		t.Fatalf("expected Copilot to render Watchtower handoff context, got:\n%s", view)
	}
}

func TestShell_WatchtowerEscalatesToAutopilotCompactHeight(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, testInventory(), nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 15})
	shell = updated.(Shell)
	updated, _ = shell.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	shell = updated.(Shell)

	_, cmd := shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd == nil {
		t.Fatal("expected escalation command to Autopilot")
	}

	msg := cmd()
	updated, _ = shell.Update(msg)
	shell = updated.(Shell)

	updated, _ = shell.Update(tea.WindowSizeMsg{Width: 80, Height: 15})
	shell = updated.(Shell)

	if shell.ActiveMode() != ModeAutopilot {
		t.Fatalf("expected shell to switch to Autopilot, got %v", shell.ActiveMode())
	}

	view := shell.View()
	assertRenderedWithinBounds(t, view, 80, 15)
	for _, label := range []string{"WATCHTOWER HANDOFF", "db-master", "MEMORY", "PLAN", "TRANSCRIPT", commandBarLabel} {
		if !strings.Contains(view, label) {
			t.Fatalf("expected compact escalated Autopilot shell view to contain %q, got:\n%s", label, view)
		}
	}
}

func TestShell_WatchtowerEscalatesToCopilotCompactHeight(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, testInventory(), nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 21})
	shell = updated.(Shell)
	updated, _ = shell.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	shell = updated.(Shell)

	_, cmd := shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if cmd == nil {
		t.Fatal("expected escalation command to Copilot")
	}

	msg := cmd()
	updated, _ = shell.Update(msg)
	shell = updated.(Shell)

	updated, _ = shell.Update(tea.WindowSizeMsg{Width: 80, Height: 21})
	shell = updated.(Shell)

	if shell.ActiveMode() != ModeCopilot {
		t.Fatalf("expected shell to switch to Copilot, got %v", shell.ActiveMode())
	}

	view := shell.View()
	assertRenderedWithinBounds(t, view, 80, 21)
	for _, label := range []string{"WATCHTOWER HANDOFF", "db-master", "MEMORY", "TERMINAL", "ADVISORY", "GUIDANCE", commandBarLabel} {
		if !strings.Contains(view, label) {
			t.Fatalf("expected compact escalated Copilot shell view to contain %q, got:\n%s", label, view)
		}
	}
}

func TestShell_AutopilotViewRendersCommandBarAndPanes(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	shell = updated.(Shell)

	shell = leaderSwitch(t, shell, 'a')

	view := shell.View()
	if !strings.Contains(view, "AUTOPILOT") {
		t.Fatalf("expected Autopilot view to render mode label, got:\n%s", view)
	}
	if !strings.Contains(view, commandBarLabel) {
		t.Fatalf("expected Autopilot view to render command bar label, got:\n%s", view)
	}
	if !strings.Contains(view, "PLAN") {
		t.Fatalf("expected Autopilot view to render plan pane, got:\n%s", view)
	}
	if !strings.Contains(view, "TRANSCRIPT") {
		t.Fatalf("expected Autopilot view to render transcript pane, got:\n%s", view)
	}
}

func TestShell_CopilotViewRendersCommandBarAndPanes(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	shell = updated.(Shell)

	shell = leaderSwitch(t, shell, 'c')

	view := shell.View()
	if !strings.Contains(view, "COPILOT") {
		t.Fatalf("expected Copilot view to render mode label, got:\n%s", view)
	}
	if !strings.Contains(view, commandBarLabel) {
		t.Fatalf("expected Copilot view to render command bar label, got:\n%s", view)
	}
	if !strings.Contains(view, "TERMINAL") {
		t.Fatalf("expected Copilot view to render terminal pane, got:\n%s", view)
	}
	if !strings.Contains(view, "ADVISORY") {
		t.Fatalf("expected Copilot view to render advisory pane, got:\n%s", view)
	}
	if !strings.Contains(view, "GUIDANCE") {
		t.Fatalf("expected Copilot view to render guidance control surface, got:\n%s", view)
	}
}

func TestShell_AutopilotCommandInputPersistsAcrossTabSwitches(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	shell = updated.(Shell)

	shell = leaderSwitch(t, shell, 'a')

	for _, r := range "deploy web" {
		updated, _ = shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		shell = updated.(Shell)
	}

	shell = leaderSwitch(t, shell, 'w')

	shell = leaderSwitch(t, shell, 'a')

	view := shell.View()
	if !strings.Contains(view, "deploy web") {
		t.Fatalf("expected Autopilot command input to persist after tab switch, got:\n%s", view)
	}
}

func TestShell_CopilotCommandInputPersistsAcrossTabSwitches(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	shell = updated.(Shell)

	shell = leaderSwitch(t, shell, 'c')

	for _, r := range "check disk" {
		updated, _ = shell.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		shell = updated.(Shell)
	}

	shell = leaderSwitch(t, shell, 'w')

	shell = leaderSwitch(t, shell, 'c')

	view := shell.View()
	if !strings.Contains(view, "check disk") {
		t.Fatalf("expected Copilot command input to persist after tab switch, got:\n%s", view)
	}
}

func TestShell_AutopilotFocusSurvivesResizeAndTabSwitch(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	shell = updated.(Shell)

	shell = leaderSwitch(t, shell, 'a')

	updated, _ = shell.Update(tea.KeyMsg{Type: tea.KeyTab})
	shell = updated.(Shell)
	beforeSwitch := shell.autopilot.(AutopilotModel).focusedPane

	updated, _ = shell.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	shell = updated.(Shell)

	shell = leaderSwitch(t, shell, 'w')

	shell = leaderSwitch(t, shell, 'a')

	if shell.autopilot.(AutopilotModel).focusedPane != beforeSwitch {
		t.Fatalf("expected Autopilot focus to survive resize and tab switch, got %d want %d",
			shell.autopilot.(AutopilotModel).focusedPane, beforeSwitch)
	}
}

func TestShell_CopilotFocusSurvivesResizeAndTabSwitch(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	shell = updated.(Shell)

	shell = leaderSwitch(t, shell, 'c')

	updated, _ = shell.Update(tea.KeyMsg{Type: tea.KeyTab})
	shell = updated.(Shell)
	beforeSwitch := shell.copilot.(CopilotModel).focusedPane

	updated, _ = shell.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	shell = updated.(Shell)

	shell = leaderSwitch(t, shell, 'w')

	shell = leaderSwitch(t, shell, 'c')

	if shell.copilot.(CopilotModel).focusedPane != beforeSwitch {
		t.Fatalf("expected Copilot focus to survive resize and tab switch, got %d want %d",
			shell.copilot.(CopilotModel).focusedPane, beforeSwitch)
	}
}
