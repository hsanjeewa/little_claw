package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/inventory"
)

func testMemorySnapshots() []agent.MemorySnapshot {
	now := time.Date(2026, 7, 3, 10, 30, 0, 0, time.UTC)
	return []agent.MemorySnapshot{
		{
			HostAlias:   "db-master",
			HostIP:      "10.0.0.20",
			TotalBytes:  16 * 1024 * 1024 * 1024,
			UsedBytes:   8 * 1024 * 1024 * 1024,
			UsedPercent: 50.0,
			CollectedAt: now,
			Status:      agent.SnapshotStatusSuccess,
		},
		{
			HostAlias:   "web-prod-01",
			HostIP:      "10.0.0.10",
			CollectedAt: now,
			Status:      agent.SnapshotStatusFailed,
			Error:       "ssh timeout",
		},
	}
}

func testCPUSnapshots() []agent.CPUSnapshot {
	now := time.Date(2026, 7, 3, 10, 35, 0, 0, time.UTC)
	return []agent.CPUSnapshot{
		{
			HostAlias:    "db-master",
			HostIP:       "10.0.0.20",
			UsagePercent: 37.5,
			CollectedAt:  now,
			Status:       agent.SnapshotStatusSuccess,
		},
		{
			HostAlias:    "web-prod-01",
			HostIP:       "10.0.0.10",
			UsagePercent: 82.2,
			CollectedAt:  now,
			Status:       agent.SnapshotStatusSuccess,
		},
	}
}

func newWatchtowerForTest(inv []inventory.TargetHost, collector MemorySnapshotCollector) WatchtowerModel {
	taskChan, logChan, hitlChan := testChannels()
	return NewWatchtowerModel(taskChan, logChan, hitlChan, nil, inv, TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(inv)}, collector)
}

func TestWatchtowerModel_FleetMatrixRendersMemorySnapshotsAndFailures(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	if !strings.Contains(view, "MEMORY") {
		t.Fatalf("expected metric family in view, got:\n%s", view)
	}
	if !strings.Contains(view, "FLEET MATRIX") {
		t.Fatalf("expected fleet matrix in view, got:\n%s", view)
	}
	if !strings.Contains(view, "db-master") || !strings.Contains(view, "web-prod-01") {
		t.Fatalf("expected both hosts in fleet matrix, got:\n%s", view)
	}
	if !strings.Contains(view, "FAILED") {
		t.Fatalf("expected failed host status to be visible, got:\n%s", view)
	}
	if !strings.Contains(view, "UPDATED 10:30:00") {
		t.Fatalf("expected visible freshness indicator, got:\n%s", view)
	}
}

func TestWatchtowerModel_HostDetailDrillDownAndBack(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	view := model.View()
	if !strings.Contains(view, "HOST DETAIL") {
		t.Fatalf("expected host detail view, got:\n%s", view)
	}
	if !strings.Contains(view, "db-master") {
		t.Fatalf("expected selected host detail, got:\n%s", view)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)
	if !strings.Contains(model.View(), "FLEET MATRIX") {
		t.Fatalf("expected return to fleet matrix, got:\n%s", model.View())
	}
}

func TestWatchtowerModel_ScopeFiltersFleetMatrix(t *testing.T) {
	inv := testInventory()
	model := newWatchtowerForTest(inv, nil)
	model = model.SetScope(TargetScope{Kind: ScopeSelectedHosts, Hosts: []inventory.TargetHost{inv[0]}})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	if strings.Contains(view, "db-master") {
		t.Fatalf("expected filtered fleet matrix to hide db-master, got:\n%s", view)
	}
	if !strings.Contains(view, "web-prod-01") {
		t.Fatalf("expected selected host to remain visible, got:\n%s", view)
	}
}

func TestWatchtowerModel_RefreshUsesCollectorAndUpdatesSnapshots(t *testing.T) {
	called := false
	collector := func(_ context.Context, hosts []inventory.TargetHost) ([]agent.MemorySnapshot, error) {
		called = true
		if len(hosts) != 2 {
			t.Fatalf("expected refresh to target 2 hosts, got %d", len(hosts))
		}
		return testMemorySnapshots(), nil
	}
	model := newWatchtowerForTest(testInventory(), collector)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = updated.(WatchtowerModel)
	if cmd == nil {
		t.Fatal("expected refresh command")
	}

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(WatchtowerModel)

	if !called {
		t.Fatal("expected collector to be called on refresh")
	}
	if !strings.Contains(model.View(), "db-master") {
		t.Fatalf("expected snapshots from collector to render, got:\n%s", model.View())
	}
}

func TestWatchtowerModel_SwitchesToCPUFleetMatrix(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, nil, nil)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: testCPUSnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	if !strings.Contains(view, string(agent.MetricFamilyCPU)) {
		t.Fatalf("expected CPU family in view, got:\n%s", view)
	}
	if !strings.Contains(view, "37.5%") {
		t.Fatalf("expected CPU usage in fleet matrix, got:\n%s", view)
	}
}

func TestWatchtowerModel_CPUHostDetail(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, nil, nil)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: testCPUSnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	view := model.View()
	if !strings.Contains(view, "CPU Usage: 37.5%") {
		t.Fatalf("expected CPU host detail, got:\n%s", view)
	}
}

func TestWatchtowerModel_CPURefreshUsesCPUCollector(t *testing.T) {
	called := false
	cpuCollector := func(_ context.Context, hosts []inventory.TargetHost) ([]agent.CPUSnapshot, error) {
		called = true
		if len(hosts) != 2 {
			t.Fatalf("expected cpu refresh to target 2 hosts, got %d", len(hosts))
		}
		return testCPUSnapshots(), nil
	}
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, nil, cpuCollector)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = updated.(WatchtowerModel)
	if cmd == nil {
		t.Fatal("expected cpu refresh command")
	}
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(WatchtowerModel)
	if !called {
		t.Fatal("expected cpu collector to be called")
	}
	if !strings.Contains(model.View(), "37.5%") {
		t.Fatalf("expected cpu collector data to render, got:\n%s", model.View())
	}
}
