package tui

import (
	"strings"
	"testing"

	"github.com/devops/agent/internal/infrastructure/inventory"
)

func TestShell_Update_OpenTargetSelectionModal(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "10.0.0.10", Port: 22, User: "root"},
		{Alias: "db-master", IP: "10.0.0.20", Port: 2222, User: "admin"},
	}
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	updated, _ := shell.Update(OpenTargetSelectionMsg{})
	shell = updated.(Shell)

	view := shell.View()
	if !strings.Contains(view, "Select Targets") {
		t.Fatalf("expected shell view to show target selection modal, got:\n%s", view)
	}
}

func TestShell_Update_AppliesConfirmedSelectionToTargetScope(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	inv := []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "10.0.0.10", Port: 22, User: "root"},
		{Alias: "db-master", IP: "10.0.0.20", Port: 2222, User: "admin"},
	}
	shell := NewShellWithInventory(taskChan, logChan, hitlChan, nil, inv, nil)

	updated, _ := shell.Update(OpenTargetSelectionMsg{})
	shell = updated.(Shell)

	hosts := []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "10.0.0.10", Port: 22, User: "root"},
	}
	updated, _ = shell.Update(TargetSelectionConfirmedMsg{hosts: hosts})
	shell = updated.(Shell)

	scope := shell.Scope()
	if scope.Kind != ScopeSelectedHosts {
		t.Fatalf("expected scope kind SelectedHosts, got %v", scope.Kind)
	}
	if len(scope.Hosts) != 1 {
		t.Fatalf("expected 1 host in scope, got %d", len(scope.Hosts))
	}
	if scope.Hosts[0].Alias != "web-prod-01" {
		t.Fatalf("expected web-prod-01 in scope, got %s", scope.Hosts[0].Alias)
	}
}