package tui

import (
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/infrastructure/inventory"
)

func testInventory() []inventory.TargetHost {
	return []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "10.0.0.10", Port: 22, User: "root"},
		{Alias: "db-master", IP: "10.0.0.20", Port: 2222, User: "admin"},
	}
}

func TestNewShell_DefaultScopeIsEntireInventory(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	shell := NewShell(taskChan, logChan, hitlChan, nil)

	scope := shell.Scope()
	if scope.Kind != ScopeEntireInventory {
		t.Fatalf("expected default scope EntireInventory, got %v", scope.Kind)
	}
}

func TestNewShellWithInventory_ScopeIncludesAllHosts(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := testInventory()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	scope := shell.Scope()
	if scope.Kind != ScopeEntireInventory {
		t.Fatalf("expected EntireInventory, got %v", scope.Kind)
	}
	if len(scope.Hosts) != len(inv) {
		t.Fatalf("expected scope to include all %d hosts, got %d", len(inv), len(scope.Hosts))
	}
}

func TestShell_SelectHosts_SetsSelectedScope(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := testInventory()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	shell, err := shell.SelectHosts([]string{"db-master"})
	if err != nil {
		t.Fatalf("unexpected error selecting host: %v", err)
	}

	scope := shell.Scope()
	if scope.Kind != ScopeSelectedHosts {
		t.Fatalf("expected SelectedHosts, got %v", scope.Kind)
	}
	if len(scope.Hosts) != 1 || scope.Hosts[0].Alias != "db-master" {
		t.Fatalf("expected single selected host db-master, got %+v", scope.Hosts)
	}
}

func TestShell_SelectHosts_PreservesHostData(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := testInventory()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	shell, err := shell.SelectHosts([]string{"web-prod-01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	host := shell.Scope().Hosts[0]
	if host.IP != "10.0.0.10" || host.Port != 22 || host.User != "root" {
		t.Fatalf("selected host data did not match inventory: %+v", host)
	}
}

func TestShell_SelectHosts_UnknownHostReturnsError(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := testInventory()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	before := shell.Scope()
	shell, err := shell.SelectHosts([]string{"missing-host"})
	if err == nil {
		t.Fatal("expected error for unknown host alias")
	}

	if shell.Scope().Kind != before.Kind || len(shell.Scope().Hosts) != len(before.Hosts) {
		t.Fatal("scope must remain unchanged after a failed selection")
	}
}

func TestShell_SelectHosts_EmptySelectsEntireInventory(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := testInventory()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	shell, err := shell.SelectHosts([]string{"db-master"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	shell, err = shell.SelectHosts(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if shell.Scope().Kind != ScopeEntireInventory {
		t.Fatalf("expected empty selection to reset to EntireInventory, got %v", shell.Scope().Kind)
	}
}

func TestShell_SetEntireInventory_RevertsSelection(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := testInventory()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	shell, _ = shell.SelectHosts([]string{"db-master"})
	shell = shell.SetEntireInventory()

	scope := shell.Scope()
	if scope.Kind != ScopeEntireInventory {
		t.Fatalf("expected EntireInventory after reset, got %v", scope.Kind)
	}
	if len(scope.Hosts) != len(inv) {
		t.Fatalf("expected reset to include all hosts, got %d", len(scope.Hosts))
	}
}

func TestShell_ScopeBadge_ReflectsEntireInventory(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := testInventory()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	badge := shell.ScopeBadge()
	if !strings.Contains(badge, "Entire Inventory") {
		t.Fatalf("expected badge to contain 'Entire Inventory', got %q", badge)
	}
}

func TestShell_ScopeBadge_ReflectsSelectedHosts(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := testInventory()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	shell, _ = shell.SelectHosts([]string{"db-master", "web-prod-01"})
	badge := shell.ScopeBadge()
	if !strings.Contains(badge, "Selected Hosts") {
		t.Fatalf("expected badge to contain 'Selected Hosts', got %q", badge)
	}
	if !strings.Contains(badge, "db-master") || !strings.Contains(badge, "web-prod-01") {
		t.Fatalf("expected badge to list selected aliases, got %q", badge)
	}
}

func TestShell_View_SurfacesCurrentScope(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := testInventory()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	updated, _ := shell.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	shell = updated.(Shell)

	view := shell.View()
	if !strings.Contains(view, "Entire Inventory") {
		t.Fatalf("expected shell view to surface Entire Inventory scope, got:\n%s", view)
	}
}

func TestShell_ModeSwitchPreservesScope(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := testInventory()
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	shell, _ = shell.SelectHosts([]string{"web-prod-01"})

	shell = leaderSwitch(t, shell, 'a')

	if shell.ActiveMode() != ModeAutopilot {
		t.Fatalf("expected mode switch to Autopilot")
	}
	if shell.Scope().Kind != ScopeSelectedHosts {
		t.Fatalf("expected scope to persist across mode switch, got %v", shell.Scope().Kind)
	}
	if !slices.ContainsFunc(shell.Scope().Hosts, func(h inventory.TargetHost) bool { return h.Alias == "web-prod-01" }) {
		t.Fatalf("expected selected host to persist across mode switch, got %+v", shell.Scope().Hosts)
	}
}
