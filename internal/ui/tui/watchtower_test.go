package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

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

func newWatchtowerWithAllSnapshotsForTest(inv []inventory.TargetHost) WatchtowerModel {
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, inv, TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(inv)}, nil, nil, nil, nil, 0)
	updated, _ := model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: testCPUSnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerStorageSnapshotsMsg{snapshots: testStorageSnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerNetworkSnapshotsMsg{snapshots: testNetworkSnapshots()})
	return updated.(WatchtowerModel)
	}

func trustSignalInventory() []inventory.TargetHost {
	return []inventory.TargetHost{
		{Alias: "api-critical", IP: "10.0.0.30", Port: 22, User: "root"},
		{Alias: "cache-failed", IP: "10.0.0.40", Port: 22, User: "root"},
		{Alias: "queue-missing", IP: "10.0.0.50", Port: 22, User: "root"},
		{Alias: "web-stale", IP: "10.0.0.60", Port: 22, User: "root"},
	}
}

func trustSignalSnapshots(now time.Time) []agent.MemorySnapshot {
	return []agent.MemorySnapshot{
		{
			HostAlias:   "api-critical",
			HostIP:      "10.0.0.30",
			TotalBytes:  16 * 1024 * 1024 * 1024,
			UsedBytes:   15 * 1024 * 1024 * 1024,
			UsedPercent: 93.0,
			CollectedAt: now.Add(-2 * time.Minute),
			Status:      agent.SnapshotStatusSuccess,
		},
		{
			HostAlias:   "cache-failed",
			HostIP:      "10.0.0.40",
			CollectedAt: now.Add(-1 * time.Minute),
			Status:      agent.SnapshotStatusFailed,
			Error:       "ssh timeout",
		},
		{
			HostAlias:   "web-stale",
			HostIP:      "10.0.0.60",
			TotalBytes:  8 * 1024 * 1024 * 1024,
			UsedBytes:   6 * 1024 * 1024 * 1024,
			UsedPercent: 75.0,
			CollectedAt: now.Add(-11 * time.Minute),
			Status:      agent.SnapshotStatusSuccess,
		},
	}
}

func cpuTrustSignalSnapshots(now time.Time) []agent.CPUSnapshot {
	return []agent.CPUSnapshot{
		{
			HostAlias:    "api-critical",
			HostIP:       "10.0.0.30",
			UsagePercent: 92.0,
			CollectedAt:  now.Add(-2 * time.Minute),
			Status:       agent.SnapshotStatusSuccess,
		},
		{
			HostAlias:   "cache-failed",
			HostIP:      "10.0.0.40",
			CollectedAt: now.Add(-1 * time.Minute),
			Status:      agent.SnapshotStatusFailed,
			Error:       "ssh timeout",
		},
		{
			HostAlias:    "web-stale",
			HostIP:       "10.0.0.60",
			UsagePercent: 72.0,
			CollectedAt:  now.Add(-11 * time.Minute),
			Status:       agent.SnapshotStatusSuccess,
		},
	}
}

func storageTrustSignalSnapshots(now time.Time) []agent.StorageSnapshot {
	return []agent.StorageSnapshot{
		{
			HostAlias:   "api-critical",
			HostIP:      "10.0.0.30",
			TotalBytes:  512 * 1024 * 1024 * 1024,
			UsedBytes:   480 * 1024 * 1024 * 1024,
			UsedPercent: 93.0,
			CollectedAt: now.Add(-2 * time.Minute),
			Status:      agent.SnapshotStatusSuccess,
		},
		{
			HostAlias:   "cache-failed",
			HostIP:      "10.0.0.40",
			CollectedAt: now.Add(-1 * time.Minute),
			Status:      agent.SnapshotStatusFailed,
			Error:       "ssh timeout",
		},
		{
			HostAlias:   "web-stale",
			HostIP:      "10.0.0.60",
			TotalBytes:  256 * 1024 * 1024 * 1024,
			UsedBytes:   192 * 1024 * 1024 * 1024,
			UsedPercent: 75.0,
			CollectedAt: now.Add(-11 * time.Minute),
			Status:      agent.SnapshotStatusSuccess,
		},
	}
}

func networkTrustSignalSnapshots(now time.Time) []agent.NetworkSnapshot {
	return []agent.NetworkSnapshot{
		{
			HostAlias:     "api-critical",
			HostIP:        "10.0.0.30",
			RxBytesPerSec: 90 * 1024 * 1024,
			TxBytesPerSec: 30 * 1024 * 1024,
			CollectedAt:   now.Add(-2 * time.Minute),
			Status:        agent.SnapshotStatusSuccess,
		},
		{
			HostAlias:   "cache-failed",
			HostIP:      "10.0.0.40",
			CollectedAt: now.Add(-1 * time.Minute),
			Status:      agent.SnapshotStatusFailed,
			Error:       "ssh timeout",
		},
		{
			HostAlias:     "web-stale",
			HostIP:        "10.0.0.60",
			RxBytesPerSec: 48 * 1024 * 1024,
			TxBytesPerSec: 16 * 1024 * 1024,
			CollectedAt:   now.Add(-11 * time.Minute),
			Status:        agent.SnapshotStatusSuccess,
		},
	}
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

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	if !strings.Contains(view, "MEMORY") {
		t.Fatalf("expected metric family in view, got:\n%s", view)
	}
	assertContainsOneOf(t, view, []string{"HOSTS", "FLEET", "SELECTION"}, "expected host-list pane label")
	rail := model.renderFleetRailBody()
	if !strings.Contains(rail, "db-master") || !strings.Contains(rail, "web-prod-01") {
		t.Fatalf("expected both hosts in fleet matrix rail, got:\n%s", rail)
	}
	if !strings.Contains(rail, "FAILED") {
		t.Fatalf("expected failed host status to be visible, got:\n%s", rail)
	}
	selected := model.renderSelectedHostBody(false)
	if !strings.Contains(selected, "UPDATED 10:30:00") {
		t.Fatalf("expected visible freshness indicator, got:\n%s", selected)
	}
}

func TestWatchtowerModel_MemoryAggregateShowsTrustBadges(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	model := newWatchtowerForTest(trustSignalInventory(), nil)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 36})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: trustSignalSnapshots(now)})
	model = updated.(WatchtowerModel)

	bundle := model.renderMemoryAggregateBundle()
	for _, want := range []string{"[CRIT 1]", "[STALE 1]", "[FAILED 1]", "[MISSING 1]"} {
		if !strings.Contains(bundle, want) {
			t.Fatalf("expected aggregate trust badge %q, got:\n%s", want, bundle)
		}
	}
	if !strings.Contains(bundle, "Severity: CRITICAL") {
		t.Fatalf("expected aggregate severity label, got:\n%s", bundle)
	}
	if !strings.Contains(bundle, "Trend:") {
		t.Fatalf("expected aggregate trend strip, got:\n%s", bundle)
	}
}

func TestWatchtowerModel_MemoryMatrixAndDetailShowExplicitTrustStates(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	model := newWatchtowerForTest(trustSignalInventory(), nil)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 36})
	model = updated.(WatchtowerModel)

	for i, used := range []float64{55, 72, 93} {
		snapshots := trustSignalSnapshots(now.Add(-time.Duration(2-i) * time.Minute))
		snapshots[0].UsedPercent = used
		snapshots[0].UsedBytes = uint64(float64(snapshots[0].TotalBytes) * (used / 100.0))
		updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: snapshots})
		model = updated.(WatchtowerModel)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	rail := model.renderFleetRailBody()
	for _, want := range []string{"[CRIT]", "[FAILED]", "[MISSING]", "[STALE]"} {
		if !strings.Contains(rail, want) {
			t.Fatalf("expected matrix trust badge %q, got:\n%s", want, rail)
		}
	}

	for i := 0; i < 2; i++ {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updated.(WatchtowerModel)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 120, 36)
	if !strings.Contains(view, "queue-missing") {
		t.Fatalf("expected missing host detail, got:\n%s", view)
	}
	state, ok := model.selectedMemoryHostState()
	if !ok {
		t.Fatal("expected selected memory host state")
	}
	detail := model.renderHostDetailModule(agent.MetricFamilyMemory, state)
	for _, want := range []string{"[MISSING]", "Status:", "Trend:"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("expected host detail trust state %q, got:\n%s", want, detail)
		}
	}
	if !strings.Contains(detail, "No memory snapshot collected yet") {
		t.Fatalf("expected missing-data detail guidance, got:\n%s", detail)
	}
}

func TestWatchtowerModel_MemoryTrendWindowIsBounded(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	model := newWatchtowerForTest(testInventory(), nil)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)

	for i := 0; i < watchtowerTrendWindowLimit+5; i++ {
		collectedAt := now.Add(time.Duration(i) * time.Minute)
		updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: []agent.MemorySnapshot{{
			HostAlias:   "db-master",
			HostIP:      "10.0.0.20",
			TotalBytes:  16 * 1024 * 1024 * 1024,
			UsedBytes:   uint64((6 + i) * 1024 * 1024 * 1024 / 2),
			UsedPercent: float64(40 + i),
			CollectedAt: collectedAt,
			Status:      agent.SnapshotStatusSuccess,
		}}})
		model = updated.(WatchtowerModel)
	}

	window := model.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyMemory, hostAlias: "db-master"}]
	if len(window) != watchtowerTrendWindowLimit {
		t.Fatalf("expected bounded memory trend window of %d, got %d", watchtowerTrendWindowLimit, len(window))
	}
	if window[0].Value != 45 {
		t.Fatalf("expected oldest retained value to be 45, got %.1f", window[0].Value)
	}
	if window[len(window)-1].Value != 68 {
		t.Fatalf("expected newest retained value to be 68, got %.1f", window[len(window)-1].Value)
	}
}

func TestWatchtowerModel_CPUAggregateShowsTrustBadges(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithCollectors(taskChan, logChan, hitlChan, nil, trustSignalInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(trustSignalInventory())}, nil, nil)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 28})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: cpuTrustSignalSnapshots(now)})
	model = updated.(WatchtowerModel)

	bundle := model.renderCPUAggregateBundle()
	for _, want := range []string{"[CRIT 1]", "[STALE 1]", "[FAILED 1]", "[MISSING 1]"} {
		if !strings.Contains(bundle, want) {
			t.Fatalf("expected CPU aggregate trust badge %q, got:\n%s", want, bundle)
		}
	}
	if !strings.Contains(bundle, "Severity: CRITICAL") {
		t.Fatalf("expected CPU aggregate severity label, got:\n%s", bundle)
	}
	if !strings.Contains(bundle, "Trend:") {
		t.Fatalf("expected CPU aggregate trend strip, got:\n%s", bundle)
	}
}

func TestWatchtowerModel_CPUMatrixAndDetailShowExplicitTrustStates(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithCollectors(taskChan, logChan, hitlChan, nil, trustSignalInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(trustSignalInventory())}, nil, nil)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 28})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)

	for i, usage := range []float64{55, 72, 92} {
		snapshots := cpuTrustSignalSnapshots(now.Add(-time.Duration(2-i) * time.Minute))
		snapshots[0].UsagePercent = usage
		updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: snapshots})
		model = updated.(WatchtowerModel)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	rail := model.renderFleetRailBody()
	for _, want := range []string{"[CRIT]", "[FAILED]", "[MISSING]", "[STALE]"} {
		if !strings.Contains(rail, want) {
			t.Fatalf("expected CPU matrix trust badge %q, got:\n%s", want, rail)
		}
	}

	for range 2 {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updated.(WatchtowerModel)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 120, 28)
	if !strings.Contains(view, "queue-missing") {
		t.Fatalf("expected missing CPU host detail, got:\n%s", view)
	}
	state, ok := model.selectedMemoryHostState()
	if !ok {
		t.Fatal("expected selected memory host state")
	}
	detail := model.renderHostDetailModule(agent.MetricFamilyCPU, state)
	for _, want := range []string{"[MISSING]", "Status:", "Trend:"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("expected CPU host detail trust state %q, got:\n%s", want, detail)
		}
	}
	if !strings.Contains(detail, "No CPU snapshot collected yet") {
		t.Fatalf("expected missing CPU detail guidance, got:\n%s", detail)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)
	if model.viewMode != watchtowerViewMatrix {
		t.Fatalf("expected CPU host detail back to return to matrix, got %v", model.viewMode)
	}
}

func TestWatchtowerModel_CPUTrendWindowIsBounded(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, nil, nil)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)

	for i := 0; i < watchtowerTrendWindowLimit+5; i++ {
		collectedAt := now.Add(time.Duration(i) * time.Minute)
		updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: []agent.CPUSnapshot{{
			HostAlias:    "db-master",
			HostIP:       "10.0.0.20",
			UsagePercent: float64(40 + i),
			CollectedAt:  collectedAt,
			Status:       agent.SnapshotStatusSuccess,
		}}})
		model = updated.(WatchtowerModel)
	}

	window := model.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyCPU, hostAlias: "db-master"}]
	if len(window) != watchtowerTrendWindowLimit {
		t.Fatalf("expected bounded CPU trend window of %d, got %d", watchtowerTrendWindowLimit, len(window))
	}
	if window[0].Value != 45 {
		t.Fatalf("expected oldest retained CPU value to be 45, got %.1f", window[0].Value)
	}
	if window[len(window)-1].Value != 68 {
		t.Fatalf("expected newest retained CPU value to be 68, got %.1f", window[len(window)-1].Value)
	}
}

func TestWatchtowerModel_StorageAggregateShowsTrustBadges(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, trustSignalInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(trustSignalInventory())}, nil, nil, nil, nil, 0)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 28})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerStorageSnapshotsMsg{snapshots: storageTrustSignalSnapshots(now)})
	model = updated.(WatchtowerModel)

	bundle := model.renderStorageAggregateBundle()
	for _, want := range []string{"[CRIT 1]", "[STALE 1]", "[FAILED 1]", "[MISSING 1]"} {
		if !strings.Contains(bundle, want) {
			t.Fatalf("expected storage aggregate trust badge %q, got:\n%s", want, bundle)
		}
	}
	if !strings.Contains(bundle, "Severity: CRITICAL") {
		t.Fatalf("expected storage aggregate severity label, got:\n%s", bundle)
	}
	if !strings.Contains(bundle, "Trend:") {
		t.Fatalf("expected storage aggregate trend strip, got:\n%s", bundle)
	}
}

func TestWatchtowerModel_StorageMatrixAndDetailShowExplicitTrustStates(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, trustSignalInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(trustSignalInventory())}, nil, nil, nil, nil, 0)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 28})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model = updated.(WatchtowerModel)

	for i, usage := range []float64{55, 72, 93} {
		snapshots := storageTrustSignalSnapshots(now.Add(-time.Duration(2-i) * time.Minute))
		snapshots[0].UsedPercent = usage
		snapshots[0].UsedBytes = uint64(float64(snapshots[0].TotalBytes) * (usage / 100.0))
		updated, _ = model.Update(watchtowerStorageSnapshotsMsg{snapshots: snapshots})
		model = updated.(WatchtowerModel)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	rail := model.renderFleetRailBody()
	for _, want := range []string{"[CRIT]", "[FAILED]", "[MISSING]", "[STALE]"} {
		if !strings.Contains(rail, want) {
			t.Fatalf("expected storage matrix trust badge %q, got:\n%s", want, rail)
		}
	}

	for range 2 {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updated.(WatchtowerModel)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 120, 28)
	if !strings.Contains(view, "queue-missing") {
		t.Fatalf("expected missing storage host detail, got:\n%s", view)
	}
	state, ok := model.selectedMemoryHostState()
	if !ok {
		t.Fatal("expected selected memory host state")
	}
	detail := model.renderHostDetailModule(agent.MetricFamilyStorage, state)
	for _, want := range []string{"[MISSING]", "Status:", "Trend:"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("expected storage host detail trust state %q, got:\n%s", want, detail)
		}
	}
	if !strings.Contains(detail, "No storage snapshot collected yet") {
		t.Fatalf("expected missing storage detail guidance, got:\n%s", detail)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)
	if model.viewMode != watchtowerViewMatrix {
		t.Fatalf("expected storage host detail back to return to matrix, got %v", model.viewMode)
	}
}

func TestWatchtowerModel_StorageTrendWindowIsBounded(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, nil, nil, nil, nil, 0)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model = updated.(WatchtowerModel)

	for i := 0; i < watchtowerTrendWindowLimit+5; i++ {
		collectedAt := now.Add(time.Duration(i) * time.Minute)
		updated, _ = model.Update(watchtowerStorageSnapshotsMsg{snapshots: []agent.StorageSnapshot{{
			HostAlias:   "db-master",
			HostIP:      "10.0.0.20",
			TotalBytes:  512 * 1024 * 1024 * 1024,
			UsedBytes:   uint64(512*(40+i)/100) * 1024 * 1024 * 1024,
			UsedPercent: float64(40 + i),
			CollectedAt: collectedAt,
			Status:      agent.SnapshotStatusSuccess,
		}}})
		model = updated.(WatchtowerModel)
	}

	window := model.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyStorage, hostAlias: "db-master"}]
	if len(window) != watchtowerTrendWindowLimit {
		t.Fatalf("expected bounded storage trend window of %d, got %d", watchtowerTrendWindowLimit, len(window))
	}
	if window[0].Value != 45 {
		t.Fatalf("expected oldest retained storage value to be 45, got %.1f", window[0].Value)
	}
	if window[len(window)-1].Value != 68 {
		t.Fatalf("expected newest retained storage value to be 68, got %.1f", window[len(window)-1].Value)
	}
}

func TestWatchtowerModel_NetworkAggregateShowsTrustBadges(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, trustSignalInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(trustSignalInventory())}, nil, nil, nil, nil, 0)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 28})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerNetworkSnapshotsMsg{snapshots: networkTrustSignalSnapshots(now)})
	model = updated.(WatchtowerModel)

	// Verify the network bundle content directly (aggregate view may truncate at small heights).
	bundle := model.renderNetworkAggregateBundle()
	for _, want := range []string{"[CRIT 1]", "[STALE 1]", "[FAILED 1]", "[MISSING 1]"} {
		if !strings.Contains(bundle, want) {
			t.Fatalf("expected network aggregate trust badge %q, got:\n%s", want, bundle)
		}
	}
	if !strings.Contains(bundle, "Severity: CRITICAL") {
		t.Fatalf("expected network aggregate severity label, got:\n%s", bundle)
	}
	if !strings.Contains(bundle, "Trend:") {
		t.Fatalf("expected network aggregate trend strip, got:\n%s", bundle)
	}

	view := model.View()
	assertRenderedWithinBounds(t, view, 120, 28)
}

func TestWatchtowerModel_NetworkMatrixAndDetailShowExplicitTrustStates(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, trustSignalInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(trustSignalInventory())}, nil, nil, nil, nil, 0)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 28})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model = updated.(WatchtowerModel)

	throughputs := [][2]uint64{{55 * 1024 * 1024, 15 * 1024 * 1024}, {60 * 1024 * 1024, 12 * 1024 * 1024}, {90 * 1024 * 1024, 30 * 1024 * 1024}}
	for i, pair := range throughputs {
		snapshots := networkTrustSignalSnapshots(now.Add(-time.Duration(2-i) * time.Minute))
		snapshots[0].RxBytesPerSec = pair[0]
		snapshots[0].TxBytesPerSec = pair[1]
		updated, _ = model.Update(watchtowerNetworkSnapshotsMsg{snapshots: snapshots})
		model = updated.(WatchtowerModel)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	rail := model.renderFleetRailBody()
	for _, want := range []string{"[CRIT]", "[FAILED]", "[MISSING]", "[STALE]"} {
		if !strings.Contains(rail, want) {
			t.Fatalf("expected network matrix trust badge %q, got:\n%s", want, rail)
		}
	}

	for range 2 {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updated.(WatchtowerModel)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 120, 28)
	if !strings.Contains(view, "queue-missing") {
		t.Fatalf("expected missing network host detail, got:\n%s", view)
	}
	state, ok := model.selectedMemoryHostState()
	if !ok {
		t.Fatal("expected selected memory host state")
	}
	detail := model.renderHostDetailModule(agent.MetricFamilyNetwork, state)
	for _, want := range []string{"[MISSING]", "Status:", "Trend:"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("expected network host detail trust state %q, got:\n%s", want, detail)
		}
	}
	if !strings.Contains(detail, "No network snapshot collected yet") {
		t.Fatalf("expected missing network detail guidance, got:\n%s", detail)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)
	if model.viewMode != watchtowerViewMatrix {
		t.Fatalf("expected network host detail back to return to matrix, got %v", model.viewMode)
	}
}

func TestWatchtowerModel_NetworkTrendWindowIsBounded(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, nil, nil, nil, nil, 0)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model = updated.(WatchtowerModel)

	for i := 0; i < watchtowerTrendWindowLimit+5; i++ {
		collectedAt := now.Add(time.Duration(i) * time.Minute)
		updated, _ = model.Update(watchtowerNetworkSnapshotsMsg{snapshots: []agent.NetworkSnapshot{{
			HostAlias:     "db-master",
			HostIP:        "10.0.0.20",
			RxBytesPerSec: uint64(20+i) * 1024 * 1024,
			TxBytesPerSec: uint64(20+i) * 1024 * 1024,
			CollectedAt:   collectedAt,
			Status:        agent.SnapshotStatusSuccess,
		}}})
		model = updated.(WatchtowerModel)
	}

	window := model.trendWindows[watchtowerTrendKey{family: agent.MetricFamilyNetwork, hostAlias: "db-master"}]
	if len(window) != watchtowerTrendWindowLimit {
		t.Fatalf("expected bounded network trend window of %d, got %d", watchtowerTrendWindowLimit, len(window))
	}

	oldestExpected := float64((20 + 5) * 2 * 1024 * 1024)
	newestExpected := float64((20 + 28) * 2 * 1024 * 1024)
	if window[0].Value != oldestExpected {
		t.Fatalf("expected oldest retained network value to be %.0f, got %.1f", oldestExpected, window[0].Value)
	}
	if window[len(window)-1].Value != newestExpected {
		t.Fatalf("expected newest retained network value to be %.0f, got %.1f", newestExpected, window[len(window)-1].Value)
	}
}

func TestWatchtowerModel_HostDetailDrillDownAndBack(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	assertContainsOneOf(t, view, []string{"DETAIL"}, "expected host detail pane label")
	if !strings.Contains(view, "db-master") {
		t.Fatalf("expected selected host detail, got:\n%s", view)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)
	resizedBack, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	model = resizedBack.(WatchtowerModel)
	backView := model.View()
	assertRenderedWithinBounds(t, backView, 80, 20)
	if model.viewMode != watchtowerViewMatrix {
		t.Fatalf("expected return to fleet matrix view, got %v", model.viewMode)
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

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
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

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)

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
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
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
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: testCPUSnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	state, ok := model.selectedMemoryHostState()
	if !ok {
		t.Fatal("expected selected memory host state")
	}
	detail := model.renderHostDetailModule(agent.MetricFamilyCPU, state)
	if !strings.Contains(detail, "CPU Usage: 37.5%") && !strings.Contains(detail, "37.5%") {
		t.Fatalf("expected CPU host detail, got:\n%s", detail)
	}
	assertContainsOneOf(t, view, []string{"DETAIL"}, "expected host detail pane label")
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

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	if !strings.Contains(view, "37.5%") {
		t.Fatalf("expected cpu collector data to render, got:\n%s", view)
	}
}

func TestWatchtowerModel_SwitchesToStorageFleet(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, nil, nil, nil, nil, 0)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerStorageSnapshotsMsg{snapshots: testStorageSnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	if !strings.Contains(view, string(agent.MetricFamilyStorage)) {
		t.Fatalf("expected storage family in view, got:\n%s", view)
	}
	selected := model.renderSelectedHostBody(false)
	if !strings.Contains(selected, "50.0%") {
		t.Fatalf("expected storage usage in selected host pane, got:\n%s", selected)
	}
}

func TestWatchtowerModel_SwitchesToNetworkFleet(t *testing.T) {
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, nil, nil, nil, nil, 0)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerNetworkSnapshotsMsg{snapshots: testNetworkSnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	if model.metricFamily != agent.MetricFamilyNetwork {
		t.Fatalf("expected network family to be active, got %v", model.metricFamily)
	}
	if !strings.Contains(view, "/s") {
		t.Fatalf("expected network throughput in view, got:\n%s", view)
	}
}

func TestWatchtowerModel_DefaultViewIsFleetAggregate(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	if !strings.Contains(view, "AGGREGATE") {
		t.Fatalf("expected default view to be Fleet Aggregate, got:\n%s", view)
	}
	bundle := model.renderMemoryAggregateBundle()
	if !strings.Contains(bundle, "Used memory:") {
		t.Fatalf("expected memory aggregate data, got:\n%s", bundle)
	}
	if !strings.Contains(bundle, "Peak host:") {
		t.Fatalf("expected aggregate outlier visibility, got:\n%s", bundle)
	}
}

func TestWatchtowerModel_ChromeShowsSelectedHostAndView(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	if !strings.Contains(view, "db-master") {
		t.Fatalf("expected chrome to show selected host, got:\n%s", view)
	}
	if !strings.Contains(view, "AGGREGATE") {
		t.Fatalf("expected chrome to show active view, got:\n%s", view)
	}
}

func TestWatchtowerModel_HeaderUsesCompactScopeSummary(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	header := model.renderDashboardHeader(72)
	if !strings.Contains(header, "AGGREGATE") {
		t.Fatalf("expected compact header to preserve active view, got:\n%s", header)
	}
	if !strings.Contains(header, "db-master") {
		t.Fatalf("expected compact header to preserve selected host, got:\n%s", header)
	}
	if strings.Contains(header, "Entire Inventory") {
		t.Fatalf("expected compact header to replace verbose scope label, got:\n%s", header)
	}
}

func TestWatchtowerModel_FooterDegradesByWidth(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)

	wide := model.renderDashboardFooter(100)
	medium := model.renderDashboardFooter(60)
	narrow := model.renderDashboardFooter(34)

	if !strings.Contains(wide, "j/k") || !strings.Contains(wide, "a/c") {
		t.Fatalf("expected wide footer to keep compact navigation hints, got:\n%s", wide)
	}
	if strings.Contains(wide, "autopilot") || strings.Contains(wide, "copilot") {
		t.Fatalf("expected wide footer to avoid verbose action labels, got:\n%s", wide)
	}
	if !strings.Contains(medium, "g/m/d") || !strings.Contains(medium, "1-4") {
		t.Fatalf("expected medium footer to keep compressed mode/family hints, got:\n%s", medium)
	}
	if !strings.Contains(narrow, "enter") || !strings.Contains(narrow, "r") {
		t.Fatalf("expected narrow footer to preserve critical drill/refresh hints, got:\n%s", narrow)
	}
	if strings.Contains(narrow, "j/k") || strings.Contains(narrow, "g/m/d") {
		t.Fatalf("expected narrow footer to drop lower-priority hints, got:\n%s", narrow)
	}
}

func TestWatchtowerModel_DetailFooterHintsActualHostCycleKeys(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	model.viewMode = watchtowerViewHostDetail

	footer := model.renderDashboardFooter(100)
	if strings.Contains(footer, "[/]") {
		t.Fatalf("expected detail footer to avoid advertising '/' as a host cycle key, got:\n%s", footer)
	}
	if !strings.Contains(footer, "[]") {
		t.Fatalf("expected detail footer to mention bracket host cycling, got:\n%s", footer)
	}
}

func TestWatchtowerModel_DirectViewSwitching_gmd(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	if model.viewMode != watchtowerViewMatrix {
		t.Fatalf("expected m to switch to Fleet Matrix, got %v", model.viewMode)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	model = updated.(WatchtowerModel)
	if model.viewMode != watchtowerViewAggregate {
		t.Fatalf("expected g to switch to Fleet Aggregate, got %v", model.viewMode)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updated.(WatchtowerModel)
	if model.viewMode != watchtowerViewHostDetail {
		t.Fatalf("expected d to switch to Host Detail, got %v", model.viewMode)
	}
}

func TestWatchtowerModel_AggregateFocusMovesWithJK(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)

	if model.metricFamily != agent.MetricFamilyMemory {
		t.Fatalf("expected default aggregate focus on memory, got %v", model.metricFamily)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(WatchtowerModel)
	if model.metricFamily != agent.MetricFamilyCPU {
		t.Fatalf("expected j to move aggregate focus to CPU, got %v", model.metricFamily)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(WatchtowerModel)
	if model.metricFamily != agent.MetricFamilyStorage {
		t.Fatalf("expected j to move aggregate focus to storage, got %v", model.metricFamily)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(WatchtowerModel)
	if model.metricFamily != agent.MetricFamilyCPU {
		t.Fatalf("expected k to move aggregate focus back to CPU, got %v", model.metricFamily)
	}
}

func TestWatchtowerModel_AggregateFocusClamps(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(WatchtowerModel)
	if model.metricFamily != agent.MetricFamilyMemory {
		t.Fatalf("expected k at top to clamp to memory, got %v", model.metricFamily)
	}

	for range 5 {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updated.(WatchtowerModel)
	}
	if model.metricFamily != agent.MetricFamilyNetwork {
		t.Fatalf("expected j at bottom to clamp to network, got %v", model.metricFamily)
	}
}

func TestWatchtowerModel_AggregateDrillIntoMatrix(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: testCPUSnapshots()})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	if model.viewMode != watchtowerViewMatrix {
		t.Fatalf("expected Enter from aggregate to drill into Fleet Matrix, got %v", model.viewMode)
	}
	selected := model.renderSelectedHostBody(false)
	if !strings.Contains(selected, "37.5%") {
		t.Fatalf("expected matrix to show CPU data, got:\n%s", selected)
	}
}

func TestWatchtowerModel_AggregateHonorsScope(t *testing.T) {
	inv := testInventory()
	model := newWatchtowerForTest(inv, nil).SetScope(TargetScope{Kind: ScopeSelectedHosts, Hosts: []inventory.TargetHost{inv[0]}})
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	if !strings.Contains(view, "AGGREGATE") {
		t.Fatalf("expected aggregate view, got:\n%s", view)
	}
	if strings.Contains(view, "db-master") {
		t.Fatalf("expected scoped aggregate to hide db-master, got:\n%s", view)
	}
	if !strings.Contains(view, "web-prod-01") {
		t.Fatalf("expected scoped aggregate to show web-prod-01, got:\n%s", view)
	}
}

func TestWatchtowerModel_SelectedHostRehomesOnScopeChange(t *testing.T) {
	inv := testInventory()
	model := newWatchtowerForTest(inv, nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(WatchtowerModel)

	model = model.SetScope(TargetScope{Kind: ScopeSelectedHosts, Hosts: []inventory.TargetHost{inv[0]}})

	view := model.View()
	if strings.Contains(view, "db-master") {
		t.Fatalf("expected selected host to re-home after scope change, got:\n%s", view)
	}
	if !strings.Contains(view, "web-prod-01") {
		t.Fatalf("expected re-homed selected host to be first visible host, got:\n%s", view)
	}
}

func TestWatchtowerModel_MemoryMatrixToHostDetail(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	view := model.View()
	if model.viewMode != watchtowerViewHostDetail {
		t.Fatalf("expected Enter from matrix to open Host Detail, got %v", model.viewMode)
	}
	if !strings.Contains(view, "db-master") {
		t.Fatalf("expected Host Detail to show selected host, got:\n%s", view)
	}
	state, ok := model.selectedMemoryHostState()
	if !ok {
		t.Fatal("expected selected memory host state")
	}
	detail := model.renderHostDetailModule(agent.MetricFamilyMemory, state)
	if !strings.Contains(detail, "Used: 8.0GiB / 16.0GiB") {
		t.Fatalf("expected Host Detail to show memory usage, got:\n%s", detail)
	}
}

func TestWatchtowerModel_DetailFamilyHotkeysUpdateFocusedModule(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: testCPUSnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)

	if model.detailFocusFamily() != agent.MetricFamilyCPU {
		t.Fatalf("expected detail hotkey to focus CPU module, got %v", model.detailFocusFamily())
	}
	if model.metricFamily != agent.MetricFamilyCPU {
		t.Fatalf("expected detail hotkey to keep active family aligned with focused module, got %v", model.metricFamily)
	}
}

func TestWatchtowerModel_DetailRefreshUsesFocusedFamily(t *testing.T) {
	memoryCalled := false
	cpuCalled := false
	memoryCollector := func(_ context.Context, hosts []inventory.TargetHost) ([]agent.MemorySnapshot, error) {
		memoryCalled = true
		return testMemorySnapshots(), nil
	}
	cpuCollector := func(_ context.Context, hosts []inventory.TargetHost) ([]agent.CPUSnapshot, error) {
		cpuCalled = true
		return testCPUSnapshots(), nil
	}
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithCollectors(taskChan, logChan, hitlChan, nil, testInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(testInventory())}, memoryCollector, cpuCollector)

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: testCPUSnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = updated.(WatchtowerModel)
	if cmd == nil {
		t.Fatal("expected refresh command in detail mode")
	}
	msg := cmd()
	if _, ok := msg.(watchtowerCPUSnapshotsMsg); !ok {
		t.Fatalf("expected detail refresh to target focused CPU family, got %T", msg)
	}
	if memoryCalled {
		t.Fatal("expected detail refresh not to call memory collector when CPU is focused")
	}
	if !cpuCalled {
		t.Fatal("expected detail refresh to call CPU collector when CPU is focused")
	}
}

func TestWatchtowerModel_DetailEscalationUsesFocusedFamily(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerCPUSnapshotsMsg{snapshots: testCPUSnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)

	payload, ok := model.escalationPayload(ModeCopilot)
	if !ok {
		t.Fatal("expected escalation payload from detail mode")
	}
	if payload.MetricFamily != agent.MetricFamilyCPU {
		t.Fatalf("expected detail escalation to use focused CPU family, got %v", payload.MetricFamily)
	}
}

func TestWatchtowerModel_MemoryHostDetailCyclesScopedHosts(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	if model.selectedHostLabel() != "db-master" {
		t.Fatalf("expected initial detail host db-master, got %q", model.selectedHostLabel())
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model = updated.(WatchtowerModel)
	if model.selectedHostLabel() != "web-prod-01" {
		t.Fatalf("expected ] to move to next scoped host, got %q", model.selectedHostLabel())
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model = updated.(WatchtowerModel)
	if model.selectedHostLabel() != "db-master" {
		t.Fatalf("expected ] to wrap to first scoped host, got %q", model.selectedHostLabel())
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	model = updated.(WatchtowerModel)
	if model.selectedHostLabel() != "web-prod-01" {
		t.Fatalf("expected [ to wrap to previous scoped host, got %q", model.selectedHostLabel())
	}
}

func TestWatchtowerModel_MemoryHostDetailDrillsBackToMatrix(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	if model.viewMode != watchtowerViewHostDetail {
		t.Fatalf("expected Host Detail after drill from storage matrix, got %v", model.viewMode)
	}
	if model.detailFocusFamily() != agent.MetricFamilyStorage {
		t.Fatalf("expected detail to open on storage family, got %v", model.detailFocusFamily())
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)

	if model.viewMode != watchtowerViewMatrix {
		t.Fatalf("expected b from detail to restore matrix, got %v", model.viewMode)
	}
	if model.metricFamily != agent.MetricFamilyStorage {
		t.Fatalf("expected matrix to keep storage family, got %v", model.metricFamily)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)
	if model.viewMode != watchtowerViewAggregate {
		t.Fatalf("expected second b from matrix to restore aggregate, got %v", model.viewMode)
	}
	if model.aggregateFocus != 2 {
		t.Fatalf("expected aggregate focus to restore storage, got %d", model.aggregateFocus)
	}
}

func TestWatchtowerModel_BackReclampsSelectionAfterScopeChange(t *testing.T) {
	inv := []inventory.TargetHost{
		{Alias: "alpha", IP: "10.0.0.1", Port: 22, User: "root"},
		{Alias: "bravo", IP: "10.0.0.2", Port: 22, User: "root"},
		{Alias: "charlie", IP: "10.0.0.3", Port: 22, User: "root"},
		{Alias: "delta", IP: "10.0.0.4", Port: 22, User: "root"},
		{Alias: "echo", IP: "10.0.0.5", Port: 22, User: "root"},
	}
	model := newWatchtowerForTest(inv, nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	snapshots := []agent.MemorySnapshot{
		{HostAlias: "alpha", HostIP: "10.0.0.1", TotalBytes: 16 * 1024 * 1024 * 1024, UsedBytes: 8 * 1024 * 1024 * 1024, UsedPercent: 50, CollectedAt: time.Date(2026, 7, 3, 10, 30, 0, 0, time.UTC), Status: agent.SnapshotStatusSuccess},
		{HostAlias: "bravo", HostIP: "10.0.0.2", TotalBytes: 16 * 1024 * 1024 * 1024, UsedBytes: 7 * 1024 * 1024 * 1024, UsedPercent: 43.75, CollectedAt: time.Date(2026, 7, 3, 10, 30, 0, 0, time.UTC), Status: agent.SnapshotStatusSuccess},
		{HostAlias: "charlie", HostIP: "10.0.0.3", TotalBytes: 16 * 1024 * 1024 * 1024, UsedBytes: 9 * 1024 * 1024 * 1024, UsedPercent: 56.25, CollectedAt: time.Date(2026, 7, 3, 10, 30, 0, 0, time.UTC), Status: agent.SnapshotStatusSuccess},
		{HostAlias: "delta", HostIP: "10.0.0.4", TotalBytes: 16 * 1024 * 1024 * 1024, UsedBytes: 10 * 1024 * 1024 * 1024, UsedPercent: 62.5, CollectedAt: time.Date(2026, 7, 3, 10, 30, 0, 0, time.UTC), Status: agent.SnapshotStatusSuccess},
		{HostAlias: "echo", HostIP: "10.0.0.5", TotalBytes: 16 * 1024 * 1024 * 1024, UsedBytes: 11 * 1024 * 1024 * 1024, UsedPercent: 68.75, CollectedAt: time.Date(2026, 7, 3, 10, 30, 0, 0, time.UTC), Status: agent.SnapshotStatusSuccess},
	}
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: snapshots})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	model = model.SetScope(TargetScope{Kind: ScopeSelectedHosts, Hosts: []inventory.TargetHost{inv[0]}})
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)

	if model.viewMode != watchtowerViewMatrix {
		t.Fatalf("expected back to restore matrix view, got %v", model.viewMode)
	}
	if model.selected != 0 {
		t.Fatalf("expected back to reclamp selected host after scope change, got %d", model.selected)
	}
	if model.matrixPage != 0 {
		t.Fatalf("expected back to reclamp matrix page after scope change, got %d", model.matrixPage)
	}
}

func TestWatchtowerModel_MemoryHostDetailFocusedModuleDrillsToMatrix(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)

	if model.viewMode != watchtowerViewMatrix {
		t.Fatalf("expected focused module drill-back to land in Fleet Matrix, got %v", model.viewMode)
	}
	if model.metricFamily != agent.MetricFamilyMemory {
		t.Fatalf("expected matrix family to match focused memory module, got %v", model.metricFamily)
	}
}

func TestWatchtowerModel_MemoryHostDetailFitsCompactViewport(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.WindowSizeMsg{Width: 36, Height: 8})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 36, 8)
	if !strings.Contains(view, "db-master") {
		t.Fatalf("expected compact Host Detail to retain selected host, got:\n%s", view)
	}
	if !strings.Contains(view, "DETAIL") {
		t.Fatalf("expected compact Host Detail to retain view label, got:\n%s", view)
	}
}

func TestWatchtowerModel_MemoryAggregateDrillIntoMatrix(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	if model.viewMode != watchtowerViewMatrix {
		t.Fatalf("expected Enter from aggregate to drill into Fleet Matrix, got %v", model.viewMode)
	}
	if model.metricFamily != agent.MetricFamilyMemory {
		t.Fatalf("expected matrix to keep memory family, got %v", model.metricFamily)
	}
	rail := model.renderFleetRailBody()
	if !strings.Contains(rail, "db-master") || !strings.Contains(rail, "web-prod-01") {
		t.Fatalf("expected matrix to show per-host memory cards, got:\n%s", rail)
	}
}

func TestWatchtowerModel_MemoryMatrixFocusedCardBecomesSelectedHost(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(WatchtowerModel)

	if model.selectedHostLabel() != "web-prod-01" {
		t.Fatalf("expected focused card to become selected host, got %q", model.selectedHostLabel())
	}
}

func TestWatchtowerModel_MemoryMatrixPagesAcrossVisibleCards(t *testing.T) {
	snapshots := make([]agent.MemorySnapshot, 0, 10)
	now := time.Date(2026, 7, 3, 10, 30, 0, 0, time.UTC)
	for i := range 10 {
		snapshots = append(snapshots, agent.MemorySnapshot{
			HostAlias:   fmt.Sprintf("host-%02d", i),
			HostIP:      fmt.Sprintf("10.0.0.%d", i),
			TotalBytes:  16 * 1024 * 1024 * 1024,
			UsedBytes:   uint64(i) * 1024 * 1024 * 1024,
			UsedPercent: float64(i) * 10.0,
			CollectedAt: now,
			Status:      agent.SnapshotStatusSuccess,
		})
	}

	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: snapshots})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)

	if model.matrixPage != 0 {
		t.Fatalf("expected matrix to start on page 0, got %d", model.matrixPage)
	}

	for range 3 {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		model = updated.(WatchtowerModel)
	}
	if model.selected != 3 {
		t.Fatalf("expected selection to advance to 3, got %d", model.selected)
	}
	if model.matrixPage != 0 {
		t.Fatalf("expected page to remain 0 when still visible, got %d", model.matrixPage)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(WatchtowerModel)
	if model.matrixPage != 1 {
		t.Fatalf("expected page to advance when crossing visible grid edge, got %d", model.matrixPage)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	model = updated.(WatchtowerModel)
	if model.matrixPage != 0 || model.selected != 0 {
		t.Fatalf("expected [ to move to previous page and select first card, got page %d selected %d", model.matrixPage, model.selected)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model = updated.(WatchtowerModel)
	if model.matrixPage != 1 {
		t.Fatalf("expected ] to move to next page, got %d", model.matrixPage)
	}
}

func TestWatchtowerModel_MemoryMatrixBackRestoresAggregateState(t *testing.T) {
	model := newWatchtowerForTest(testInventory(), nil)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(WatchtowerModel)
	if model.aggregateFocus != 2 {
		t.Fatalf("expected aggregate focus on storage before drill, got %d", model.aggregateFocus)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(WatchtowerModel)

	if model.viewMode != watchtowerViewAggregate {
		t.Fatalf("expected b to restore aggregate view, got %v", model.viewMode)
	}
	if model.aggregateFocus != 2 {
		t.Fatalf("expected aggregate focus to be restored, got %d", model.aggregateFocus)
	}
	if model.metricFamily != agent.MetricFamilyStorage {
		t.Fatalf("expected aggregate family to be restored, got %v", model.metricFamily)
	}
}

func TestWatchtowerEscalationPayload_IncludesWatchtowerViewContext(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, trustSignalInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(trustSignalInventory())}, nil, nil, nil, nil, 0)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	payload, ok := model.escalationPayload(ModeCopilot)
	if !ok {
		t.Fatal("expected escalation payload to be available")
	}
	if payload.ViewMode != watchtowerViewAggregate {
		t.Fatalf("expected view mode to be aggregate, got %v", payload.ViewMode)
	}
	if payload.Scope.Kind != ScopeEntireInventory {
		t.Fatalf("expected scope to be entire inventory, got %v", payload.Scope.Kind)
	}
}

func TestWatchtowerEscalationPayload_IncludesHostStateFreshness(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, trustSignalInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(trustSignalInventory())}, nil, nil, nil, nil, 0)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	payload, ok := model.escalationPayload(ModeCopilot)
	if !ok {
		t.Fatal("expected escalation payload to be available")
	}
	if payload.Observation == "" {
		t.Fatal("expected observation to be non-empty")
	}
	if !strings.Contains(payload.Observation, "FRESH") && !strings.Contains(payload.Observation, "FAILED") && !strings.Contains(payload.Observation, "STALE") && !strings.Contains(payload.Observation, "MISSING") {
		t.Fatalf("expected observation to contain freshness label, got: %s", payload.Observation)
	}
}

func TestWatchtowerEscalationPayload_IncludesRecentObservations(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	taskChan, logChan, hitlChan := testChannels()
	model := NewWatchtowerModelWithAllCollectors(taskChan, logChan, hitlChan, nil, trustSignalInventory(), TargetScope{Kind: ScopeEntireInventory, Hosts: cloneHosts(trustSignalInventory())}, nil, nil, nil, nil, 0)
	model.now = func() time.Time { return now }

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(watchtowerSnapshotsMsg{snapshots: testMemorySnapshots()})
	model = updated.(WatchtowerModel)

	payload, ok := model.escalationPayload(ModeCopilot)
	if !ok {
		t.Fatal("expected escalation payload to be available")
	}
	if payload.SelectedHost == "" {
		t.Fatal("expected selected host to be non-empty")
	}
	if payload.MetricFamily != agent.MetricFamilyMemory {
		t.Fatalf("expected metric family to be memory, got %v", payload.MetricFamily)
	}
}

func TestWatchtowerModel_PersistentFleetRailAcrossViews(t *testing.T) {
	model := newWatchtowerWithAllSnapshotsForTest(testInventory())

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = updated.(WatchtowerModel)

	for _, key := range []rune{'g', 'm', 'd'} {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}})
		model = updated.(WatchtowerModel)

		view := model.View()
		assertRenderedWithinBounds(t, view, 120, 30)
		if !strings.Contains(view, "FLEET") {
			t.Fatalf("expected persistent fleet rail in view %q, got:\n%s", string(key), view)
		}
		for _, host := range []string{"db-master", "web-prod-01"} {
			if !strings.Contains(view, host) {
				t.Fatalf("expected host %q to remain visible in fleet rail for view %q, got:\n%s", host, string(key), view)
			}
		}
	}
}

func TestWatchtowerModel_DetailLayoutRespondsToResizeWithVisibleAppBorder(t *testing.T) {
	model := newWatchtowerWithAllSnapshotsForTest(testInventory())

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 52, Height: 16})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updated.(WatchtowerModel)

	compact := model.View()
	assertRenderedWithinBounds(t, compact, 52, 16)
	compactPlain := ansi.Strip(compact)
	if !strings.Contains(compactPlain, "┌") || !strings.Contains(compactPlain, "┘") {
		t.Fatalf("expected compact detail view to keep app border visible, got:\n%s", compactPlain)
	}

	updated, _ = model.Update(tea.WindowSizeMsg{Width: 96, Height: 28})
	model = updated.(WatchtowerModel)
	resized := model.View()
	assertRenderedWithinBounds(t, resized, 96, 28)
	for _, label := range []string{"MEMORY", "CPU", "STORAGE", "NETWORK"} {
		if !strings.Contains(resized, label) {
			t.Fatalf("expected resized detail layout to keep module %q visible, got:\n%s", label, resized)
		}
	}
}

func TestWatchtowerModel_FleetRailUsesCompactPagingAndRows(t *testing.T) {
	model := newWatchtowerWithAllSnapshotsForTest(testInventory())

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)

	rail := model.renderFleetRailBody()
	if !strings.Contains(rail, "1/1") {
		t.Fatalf("expected compact rail paging, got:\n%s", rail)
	}
	if strings.Contains(rail, "page 1/1") || strings.Contains(rail, "hosts)") {
		t.Fatalf("expected compact paging instead of verbose page copy, got:\n%s", rail)
	}
	if strings.Contains(rail, "50.0%") || strings.Contains(rail, "UPDATED") {
		t.Fatalf("expected dense rail rows without full value/freshness text, got:\n%s", rail)
	}
	for _, want := range []string{"db-master", "web-prod-01", "[STALE]"} {
		if !strings.Contains(rail, want) {
			t.Fatalf("expected compact rail to preserve %q, got:\n%s", want, rail)
		}
	}
}

func TestWatchtowerModel_SplitRailLeavesMoreRoomForPrimaryPane(t *testing.T) {
	model := newWatchtowerWithAllSnapshotsForTest(testInventory())

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 24})
	model = updated.(WatchtowerModel)
	view := model.View()
	assertRenderedWithinBounds(t, view, 100, 24)
	for _, want := range []string{"CPU", "MEMORY", "NETWORK"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected primary pane to keep aggregate module %q after tighter rail, got:\n%s", want, view)
		}
	}
}

func TestWatchtowerModel_DetailLayoutPromotesCPUInWideView(t *testing.T) {
	model := newWatchtowerWithAllSnapshotsForTest(testInventory())

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 108, Height: 28})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 108, 28)
	for _, label := range []string{"MEMORY", "CPU", "STORAGE", "NETWORK"} {
		if !strings.Contains(view, label) {
			t.Fatalf("expected wide detail layout to keep module %q visible, got:\n%s", label, view)
		}
	}
	body := strings.Join(strings.Split(view, "\n")[2:], "\n")
	if strings.Index(body, "CPU") > strings.Index(body, "MEMORY") {
		t.Fatalf("expected CPU module to lead the wide detail layout ahead of memory, got:\n%s", view)
	}
}

func TestWatchtowerModel_AggregateLayoutPromotesCPUInWideView(t *testing.T) {
	model := newWatchtowerWithAllSnapshotsForTest(testInventory())

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 108, Height: 28})
	model = updated.(WatchtowerModel)
	view := model.View()
	assertRenderedWithinBounds(t, view, 108, 28)
	for _, label := range []string{"MEMORY", "CPU", "STORAGE", "NETWORK"} {
		if !strings.Contains(view, label) {
			t.Fatalf("expected wide aggregate layout to keep module %q visible, got:\n%s", label, view)
		}
	}
	body := strings.Join(strings.Split(view, "\n")[2:], "\n")
	if strings.Index(body, "CPU") > strings.Index(body, "MEMORY") {
		t.Fatalf("expected CPU module to lead the wide aggregate layout ahead of memory, got:\n%s", view)
	}
}

func TestWatchtowerModel_MatrixLayoutKeepsCompactSummaryAndFocusedHostInspection(t *testing.T) {
	model := newWatchtowerWithAllSnapshotsForTest(testInventory())

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 96, Height: 24})
	model = updated.(WatchtowerModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(WatchtowerModel)

	view := model.View()
	assertRenderedWithinBounds(t, view, 96, 24)
	if !strings.Contains(view, "MEMORY FLEET") {
		t.Fatalf("expected matrix summary strip to preserve active family summary, got:\n%s", view)
	}
	if !strings.Contains(view, "db-master HOST") {
		t.Fatalf("expected matrix to keep focused host inspection pane, got:\n%s", view)
	}
	if strings.Contains(view, "SELECTION") {
		t.Fatalf("expected matrix pane to use focused host inspection instead of legacy selection pane, got:\n%s", view)
	}
}
