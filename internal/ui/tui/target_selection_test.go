package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/infrastructure/inventory"
)

func TestTargetSelectionModel_RendersAllHostsWithCheckboxes(t *testing.T) {
	hosts := []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "10.0.0.10", Port: 22, User: "root"},
		{Alias: "db-master", IP: "10.0.0.20", Port: 2222, User: "admin"},
		{Alias: "cache-01", IP: "10.0.0.30", Port: 22, User: "redis"},
	}

	model := NewTargetSelectionModel(hosts)
	view := model.View()

	for _, host := range hosts {
		if !strings.Contains(view, host.Alias) {
			t.Fatalf("expected view to contain host alias %q, got:\n%s", host.Alias, view)
		}
	}

	if !strings.Contains(view, "[ ]") && !strings.Contains(view, "[x]") {
		t.Fatalf("expected view to contain checkbox indicators, got:\n%s", view)
	}
}

func TestTargetSelectionModel_SpaceTogglesHostSelection(t *testing.T) {
	hosts := []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "10.0.0.10", Port: 22, User: "root"},
		{Alias: "db-master", IP: "10.0.0.20", Port: 2222, User: "admin"},
	}

	model := NewTargetSelectionModel(hosts)

	initialSelected := model.SelectedHosts()
	if len(initialSelected) != 0 {
		t.Fatalf("expected no hosts selected initially, got %d", len(initialSelected))
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(TargetSelectionModel)

	selected := model.SelectedHosts()
	if len(selected) != 1 {
		t.Fatalf("expected 1 host selected after Space on first item, got %d", len(selected))
	}
	if selected[0].Alias != "web-prod-01" {
		t.Fatalf("expected web-prod-01 to be selected, got %s", selected[0].Alias)
	}
}

func TestTargetSelectionModel_CtrlASelectsAll(t *testing.T) {
	hosts := []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "10.0.0.10", Port: 22, User: "root"},
		{Alias: "db-master", IP: "10.0.0.20", Port: 2222, User: "admin"},
		{Alias: "cache-01", IP: "10.0.0.30", Port: 22, User: "redis"},
	}

	model := NewTargetSelectionModel(hosts)
	initialSelected := model.SelectedHosts()

	if len(initialSelected) != 0 {
		t.Fatalf("expected no hosts selected initially, got %d", len(initialSelected))
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	model = updated.(TargetSelectionModel)

	selected := model.SelectedHosts()
	if len(selected) != len(hosts) {
		t.Fatalf("expected all %d hosts selected after Ctrl+a, got %d", len(hosts), len(selected))
	}
}

func TestTargetSelectionModel_EnterReturnsConfirmedMessage(t *testing.T) {
	hosts := []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "10.0.0.10", Port: 22, User: "root"},
		{Alias: "db-master", IP: "10.0.0.20", Port: 2222, User: "admin"},
	}

	model := NewTargetSelectionModel(hosts)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	_, ok := cmd().(TargetSelectionConfirmedMsg)
	if !ok {
		t.Fatal("expected Enter to return TargetSelectionConfirmedMsg command")
	}

	updatedModel, ok := updated.(TargetSelectionModel)
	if !ok {
		t.Fatal("expected updated model to be TargetSelectionModel")
	}

	selected := updatedModel.SelectedHosts()
	if len(selected) != 0 {
		t.Fatalf("expected no hosts selected on Enter with no selection, got %d", len(selected))
	}
}

func TestTargetSelectionModel_EscReturnsCancelledMessage(t *testing.T) {
	hosts := []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "10.0.0.10", Port: 22, User: "root"},
	}

	model := NewTargetSelectionModel(hosts)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	_, ok := cmd().(TargetSelectionCancelledMsg)
	if !ok {
		t.Fatal("expected Esc to return TargetSelectionCancelledMsg command")
	}

	_, ok = updated.(TargetSelectionModel)
	if !ok {
		t.Fatal("expected updated model to be TargetSelectionModel")
	}
}

func TestTargetSelectionModel_HandlesWindowSize(t *testing.T) {
	hosts := []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "10.0.0.10", Port: 22, User: "root"},
	}

	model := NewTargetSelectionModel(hosts)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	model = updated.(TargetSelectionModel)

	if model.width != 100 {
		t.Fatalf("expected width 100, got %d", model.width)
	}
	if model.height != 40 {
		t.Fatalf("expected height 40, got %d", model.height)
	}
}