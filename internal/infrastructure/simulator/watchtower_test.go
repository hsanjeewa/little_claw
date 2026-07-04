package simulator

import (
	"context"
	"testing"
	"time"

	"github.com/devops/agent/internal/domain/agent"
)

func TestWatchtowerBackend_FleetHasAtLeastFourHosts(t *testing.T) {
	backend := NewWatchtowerBackend()
	fleet := backend.Fleet()

	if len(fleet) < 4 {
		t.Fatalf("expected at least 4 hosts in simulator fleet, got %d", len(fleet))
	}
}

func TestWatchtowerBackend_CollectsAllMetricFamilies(t *testing.T) {
	backend := NewWatchtowerBackendWithClock(func() time.Time {
		return time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	})
	fleet := backend.Fleet()

	memory, err := backend.CollectMemory(context.Background(), fleet)
	if err != nil {
		t.Fatalf("unexpected memory collection error: %v", err)
	}
	cpu, err := backend.CollectCPU(context.Background(), fleet)
	if err != nil {
		t.Fatalf("unexpected cpu collection error: %v", err)
	}
	storage, err := backend.CollectStorage(context.Background(), fleet)
	if err != nil {
		t.Fatalf("unexpected storage collection error: %v", err)
	}
	network, err := backend.CollectNetwork(context.Background(), fleet)
	if err != nil {
		t.Fatalf("unexpected network collection error: %v", err)
	}

	if len(memory) != len(fleet) || len(cpu) != len(fleet) || len(storage) != len(fleet) || len(network) != len(fleet) {
		t.Fatalf("expected all metric families to cover the whole fleet")
	}

	assertSuccessfulMemorySnapshot(t, memory[0])
	assertSuccessfulCPUSnapshot(t, cpu[0])
	assertSuccessfulStorageSnapshot(t, storage[0])
	assertSuccessfulNetworkSnapshot(t, network[0])
}

func TestWatchtowerBackend_ValuesChangeOverTime(t *testing.T) {
	times := []time.Time{
		time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 4, 12, 1, 10, 0, time.UTC),
	}
	idx := 0
	backend := NewWatchtowerBackendWithClock(func() time.Time {
		current := times[idx]
		if idx < len(times)-1 {
			idx++
		}
		return current
	})
	fleet := backend.Fleet()

	memory1, _ := backend.CollectMemory(context.Background(), fleet)
	memory2, _ := backend.CollectMemory(context.Background(), fleet)
	if memory1[0].UsedPercent == memory2[0].UsedPercent {
		t.Fatalf("expected memory percent to change over time, got %.1f and %.1f", memory1[0].UsedPercent, memory2[0].UsedPercent)
	}
	if !memory2[0].CollectedAt.After(memory1[0].CollectedAt) {
		t.Fatalf("expected newer collection timestamp, got %v then %v", memory1[0].CollectedAt, memory2[0].CollectedAt)
	}

	idx = 0
	cpu1, _ := backend.CollectCPU(context.Background(), fleet)
	cpu2, _ := backend.CollectCPU(context.Background(), fleet)
	if cpu1[0].UsagePercent == cpu2[0].UsagePercent {
		t.Fatalf("expected cpu percent to change over time, got %.1f and %.1f", cpu1[0].UsagePercent, cpu2[0].UsagePercent)
	}

	idx = 0
	storage1, _ := backend.CollectStorage(context.Background(), fleet)
	storage2, _ := backend.CollectStorage(context.Background(), fleet)
	if storage1[0].UsedPercent == storage2[0].UsedPercent {
		t.Fatalf("expected storage percent to change over time, got %.1f and %.1f", storage1[0].UsedPercent, storage2[0].UsedPercent)
	}

	idx = 0
	network1, _ := backend.CollectNetwork(context.Background(), fleet)
	network2, _ := backend.CollectNetwork(context.Background(), fleet)
	if network1[0].RxBytesPerSec == network2[0].RxBytesPerSec && network1[0].TxBytesPerSec == network2[0].TxBytesPerSec {
		t.Fatalf("expected network throughput to change over time, got rx=%d/%d tx=%d/%d", network1[0].RxBytesPerSec, network2[0].RxBytesPerSec, network1[0].TxBytesPerSec, network2[0].TxBytesPerSec)
	}
}

func assertSuccessfulMemorySnapshot(t *testing.T, snapshot agent.MemorySnapshot) {
	t.Helper()
	if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.TotalBytes == 0 || snapshot.UsedBytes == 0 {
		t.Fatalf("unexpected memory snapshot: %+v", snapshot)
	}
}

func assertSuccessfulCPUSnapshot(t *testing.T, snapshot agent.CPUSnapshot) {
	t.Helper()
	if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.UsagePercent <= 0 {
		t.Fatalf("unexpected cpu snapshot: %+v", snapshot)
	}
}

func assertSuccessfulStorageSnapshot(t *testing.T, snapshot agent.StorageSnapshot) {
	t.Helper()
	if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.TotalBytes == 0 || snapshot.UsedBytes == 0 {
		t.Fatalf("unexpected storage snapshot: %+v", snapshot)
	}
}

func assertSuccessfulNetworkSnapshot(t *testing.T, snapshot agent.NetworkSnapshot) {
	t.Helper()
	if snapshot.Status != agent.SnapshotStatusSuccess || snapshot.RxBytesPerSec == 0 || snapshot.TxBytesPerSec == 0 {
		t.Fatalf("unexpected network snapshot: %+v", snapshot)
	}
}
