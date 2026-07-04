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

func testStorageSnapshots() []agent.StorageSnapshot {
	now := time.Date(2026, 7, 3, 10, 40, 0, 0, time.UTC)
	return []agent.StorageSnapshot{
		{HostAlias: "db-master", HostIP: "10.0.0.20", TotalBytes: 512 * 1024 * 1024 * 1024, UsedBytes: 256 * 1024 * 1024 * 1024, UsedPercent: 50.0, CollectedAt: now, Status: agent.SnapshotStatusSuccess},
		{HostAlias: "web-prod-01", HostIP: "10.0.0.10", TotalBytes: 128 * 1024 * 1024 * 1024, UsedBytes: 96 * 1024 * 1024 * 1024, UsedPercent: 75.0, CollectedAt: now, Status: agent.SnapshotStatusSuccess},
	}
}

func testNetworkSnapshots() []agent.NetworkSnapshot {
	now := time.Date(2026, 7, 3, 10, 45, 0, 0, time.UTC)
	return []agent.NetworkSnapshot{
		{HostAlias: "db-master", HostIP: "10.0.0.20", RxBytesPerSec: 64 * 1024 * 1024, TxBytesPerSec: 22 * 1024 * 1024, CollectedAt: now, Status: agent.SnapshotStatusSuccess},
		{HostAlias: "web-prod-01", HostIP: "10.0.0.10", RxBytesPerSec: 48 * 1024 * 1024, TxBytesPerSec: 31 * 1024 * 1024, CollectedAt: now, Status: agent.SnapshotStatusSuccess},
	}
}

func newWatchtowerForTest(inv []inventory.TargetHost, collector MemorySnapshotCollector) WatchtowerModel {
	taskChan, logChan, hitlChan := testChannels()
	return NewWatchtowerModel(taskChan, logChan, hitlChan, nil, inv, TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(inv)}, collector)
}

func assertContainsOneOf(t *testing.T, view string, expected []string, msg string) {
	t.Helper()
	for _, label := range expected {
		if strings.Contains(view, label) {
			return
		}
	}
	t.Fatalf("%s (expected one of %v), got:\n%s", msg, expected, view)
}

func assertWatchtowerKeySections(t *testing.T, view string, metricFamily string) {
	t.Helper()

	if !strings.Contains(view, metricFamily) {
		t.Fatalf("expected metric family %q in view, got:\n%s", metricFamily, view)
	}
	if !strings.Contains(view, "db-master") || !strings.Contains(view, "web-prod-01") {
		t.Fatalf("expected both hosts to remain visible, got:\n%s", view)
	}
	if !strings.Contains(view, "FAILED") && !strings.Contains(view, "UPDATED") {
		t.Fatalf("expected freshness/status indicators, got:\n%s", view)
	}

	lower := strings.ToLower(view)
	hintCount := 0
	for _, hint := range []string{"enter", "refresh", "memory", "cpu", "storage", "network", "j/k", "back", "esc"} {
		if strings.Contains(lower, hint) {
			hintCount++
		}
	}
	if hintCount < 2 {
		t.Fatalf("expected navigation/footer hints in view, got:\n%s", view)
	}

	assertContainsOneOf(t, view, []string{"HOSTS", "FLEET", "DETAIL", "SELECTION"}, "expected a host or metric pane label")
}

func TestWatchtowerModel_FleetMatrixRendersMemorySnapshotsAndFailures(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	if !strings.Contains(view, "MEMORY") {
		t.Fatalf("expected metric family in view, got:\n%s", view)
	}
	assertContainsOneOf(t, view, []string{"FLEET MATRIX", "HOSTS", "METRICS"}, "expected host-list pane label")
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
	assertRenderedWithinBounds(t, view, 100, 24)
	assertContainsOneOf(t, view, []string{"HOST DETAIL", "DETAIL", "INFO"}, "expected host detail pane label")
	if !strings.Contains(view, "db-master") {
		t.Fatalf("expected selected host detail, got:\n%s", view)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)
	resizedBack, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	model = resizedBack.(WatchtowerModel)
	backView := model.View()
	assertRenderedWithinBounds(t, backView, 80, 20)
	assertContainsOneOf(t, backView, []string{"FLEET MATRIX", "HOSTS", "METRICS"}, "expected return to fleet/dashboard view")
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
	assertRenderedWithinBounds(t, view, 100, 24)
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
	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	if !strings.Contains(view, "db-master") || !strings.Contains(view, "web-prod-01") {
		t.Fatalf("expected snapshots from collector to render, got:\n%s", view)
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
	assertRenderedWithinBounds(t, view, 100, 24)
	if !strings.Contains(view, string(agent.MetricFamilyCPU)) {
		t.Fatalf("expected CPU family in view, got:\n%s", view)
	}
	if !strings.Contains(view, "37.5%") {
		t.Fatalf("expected CPU usage in fleet matrix, got:\n%s", view)
	}
	assertWatchtowerKeySections(t, view, string(agent.MetricFamilyCPU))
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
	assertRenderedWithinBounds(t, view, 100, 24)
	if !strings.Contains(view, "CPU Usage: 37.5%") && !strings.Contains(view, "37.5%") {
		t.Fatalf("expected CPU host detail, got:\n%s", view)
	}
	assertContainsOneOf(t, view, []string{"HOST DETAIL", "DETAIL", "INFO"}, "expected host detail pane label")
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
	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	if !strings.Contains(view, "37.5%") {
		t.Fatalf("expected cpu collector data to render, got:\n%s", view)
	}
}

func TestWatchtowerModel_SwitchesToStorageFleet(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, nil, nil, nil, nil)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerStorageSnapshotsMsg{snapshots: testStorageSnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	if !strings.Contains(view, string(agent.MetricFamilyStorage)) {
		t.Fatalf("expected storage family in view, got:\n%s", view)
	}
	if !strings.Contains(view, "75%") {
		t.Fatalf("expected storage usage in view, got:\n%s", view)
	}
}

func TestWatchtowerModel_SwitchesToNetworkFleet(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, nil, nil, nil, nil)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerNetworkSnapshotsMsg{snapshots: testNetworkSnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	if !strings.Contains(view, string(agent.MetricFamilyNetwork)) {
		t.Fatalf("expected network family in view, got:\n%s", view)
	}
	if !strings.Contains(view, "/s") {
		t.Fatalf("expected network throughput in view, got:\n%s", view)
	}
}
