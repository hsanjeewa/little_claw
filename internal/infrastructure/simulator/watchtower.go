package simulator

import (
	"context"
	"math"
	"time"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/inventory"
)

type WatchtowerBackend struct {
	fleet []inventory.TargetHost
	now   func() time.Time
}

func NewWatchtowerBackend() *WatchtowerBackend {
	return NewWatchtowerBackendWithClock(time.Now)
}

func NewWatchtowerBackendWithClock(now func() time.Time) *WatchtowerBackend {
	return &WatchtowerBackend{
		fleet: []inventory.TargetHost{
			{Alias: "api-east-01", IP: "10.10.0.11", Port: 22, User: "root"},
			{Alias: "db-core-02", IP: "10.10.0.12", Port: 22, User: "root"},
			{Alias: "worker-west-03", IP: "10.10.0.13", Port: 22, User: "root"},
			{Alias: "cache-edge-04", IP: "10.10.0.14", Port: 22, User: "root"},
		},
		now: now,
	}
}

func (b *WatchtowerBackend) Fleet() []inventory.TargetHost {
	cloned := make([]inventory.TargetHost, len(b.fleet))
	copy(cloned, b.fleet)
	return cloned
}

func (b *WatchtowerBackend) CollectMemory(_ context.Context, hosts []inventory.TargetHost) ([]agent.MemorySnapshot, error) {
	resolved := b.resolveHosts(hosts)
	total := map[string]uint64{
		"api-east-01":    16 * gib,
		"db-core-02":     32 * gib,
		"worker-west-03": 24 * gib,
		"cache-edge-04":  12 * gib,
	}
	basePercent := map[string]float64{
		"api-east-01":    37.5,
		"db-core-02":     66.0,
		"worker-west-03": 46.0,
		"cache-edge-04":  75.0,
	}
	collectedAt := b.now().UTC().Truncate(time.Second)
	phase := int(collectedAt.Unix() / 5)

	snapshots := make([]agent.MemorySnapshot, 0, len(resolved))
	for idx, host := range resolved {
		totalBytes := total[host.Alias]
		usedPercent := oscillatePercent(basePercent[host.Alias], 18, 92, phase, idx, 18)
		usedBytes := bytesFromPercent(totalBytes, usedPercent)
		snapshots = append(snapshots, agent.MemorySnapshot{
			HostAlias:   host.Alias,
			HostIP:      host.IP,
			TotalBytes:  totalBytes,
			UsedBytes:   usedBytes,
			UsedPercent: percent(usedBytes, totalBytes),
			CollectedAt: collectedAt,
			Status:      agent.SnapshotStatusSuccess,
		})
	}
	return snapshots, nil
}

func (b *WatchtowerBackend) CollectCPU(_ context.Context, hosts []inventory.TargetHost) ([]agent.CPUSnapshot, error) {
	resolved := b.resolveHosts(hosts)
	baseUsage := map[string]float64{
		"api-east-01":    23.5,
		"db-core-02":     67.2,
		"worker-west-03": 48.9,
		"cache-edge-04":  12.4,
	}
	collectedAt := b.now().UTC().Truncate(time.Second)
	phase := int(collectedAt.Unix() / 3)

	snapshots := make([]agent.CPUSnapshot, 0, len(resolved))
	for idx, host := range resolved {
		snapshots = append(snapshots, agent.CPUSnapshot{
			HostAlias:    host.Alias,
			HostIP:       host.IP,
			UsagePercent: oscillatePercent(baseUsage[host.Alias], 4, 96, phase, idx, 20),
			CollectedAt:  collectedAt,
			Status:       agent.SnapshotStatusSuccess,
		})
	}
	return snapshots, nil
}

func (b *WatchtowerBackend) CollectStorage(_ context.Context, hosts []inventory.TargetHost) ([]agent.StorageSnapshot, error) {
	resolved := b.resolveHosts(hosts)
	total := map[string]uint64{
		"api-east-01":    256 * gib,
		"db-core-02":     1024 * gib,
		"worker-west-03": 512 * gib,
		"cache-edge-04":  128 * gib,
	}
	basePercent := map[string]float64{
		"api-east-01":    46.9,
		"db-core-02":     66.4,
		"worker-west-03": 60.5,
		"cache-edge-04":  70.3,
	}
	collectedAt := b.now().UTC().Truncate(time.Second)
	phase := int(collectedAt.Unix() / 30)

	snapshots := make([]agent.StorageSnapshot, 0, len(resolved))
	for idx, host := range resolved {
		totalBytes := total[host.Alias]
		usedPercent := oscillatePercent(basePercent[host.Alias], 20, 94, phase, idx, 24)
		usedBytes := bytesFromPercent(totalBytes, usedPercent)
		snapshots = append(snapshots, agent.StorageSnapshot{
			HostAlias:   host.Alias,
			HostIP:      host.IP,
			TotalBytes:  totalBytes,
			UsedBytes:   usedBytes,
			UsedPercent: percent(usedBytes, totalBytes),
			CollectedAt: collectedAt,
			Status:      agent.SnapshotStatusSuccess,
		})
	}
	return snapshots, nil
}

func (b *WatchtowerBackend) CollectNetwork(_ context.Context, hosts []inventory.TargetHost) ([]agent.NetworkSnapshot, error) {
	resolved := b.resolveHosts(hosts)
	baseRX := map[string]uint64{
		"api-east-01":    84 * mib,
		"db-core-02":     146 * mib,
		"worker-west-03": 58 * mib,
		"cache-edge-04":  33 * mib,
	}
	baseTX := map[string]uint64{
		"api-east-01":    41 * mib,
		"db-core-02":     77 * mib,
		"worker-west-03": 92 * mib,
		"cache-edge-04":  24 * mib,
	}
	collectedAt := b.now().UTC().Truncate(time.Second)
	phase := int(collectedAt.Unix() / 2)

	snapshots := make([]agent.NetworkSnapshot, 0, len(resolved))
	for idx, host := range resolved {
		snapshots = append(snapshots, agent.NetworkSnapshot{
			HostAlias:     host.Alias,
			HostIP:        host.IP,
			RxBytesPerSec: oscillateBytes(baseRX[host.Alias], phase, idx, 16),
			TxBytesPerSec: oscillateBytes(baseTX[host.Alias], phase, idx+3, 18),
			CollectedAt:   collectedAt,
			Status:        agent.SnapshotStatusSuccess,
		})
	}
	return snapshots, nil
}

func (b *WatchtowerBackend) resolveHosts(hosts []inventory.TargetHost) []inventory.TargetHost {
	if len(hosts) == 0 {
		return b.Fleet()
	}
	allowed := map[string]inventory.TargetHost{}
	for _, host := range b.fleet {
		allowed[host.Alias] = host
	}
	resolved := make([]inventory.TargetHost, 0, len(hosts))
	for _, host := range hosts {
		if fleetHost, ok := allowed[host.Alias]; ok {
			resolved = append(resolved, fleetHost)
		}
	}
	if len(resolved) == 0 {
		return b.Fleet()
	}
	return resolved
}

func percent(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return (float64(used) / float64(total)) * 100
}

func bytesFromPercent(total uint64, usedPercent float64) uint64 {
	return uint64((usedPercent / 100) * float64(total))
}

func oscillatePercent(base, minValue, maxValue float64, phase, offset, period int) float64 {
	if period < 2 {
		return clampFloat(base, minValue, maxValue)
	}
	swing := math.Min(base-minValue, maxValue-base)
	if swing <= 0 {
		return clampFloat(base, minValue, maxValue)
	}
	step := (phase + offset) % period
	progress := float64(step) / float64(period-1)
	wave := 1 - math.Abs((2*progress)-1)
	value := base + ((wave - 0.5) * 2 * swing)
	return clampFloat(value, minValue, maxValue)
}

func oscillateBytes(base uint64, phase, offset, period int) uint64 {
	factor := oscillatePercent(100, 55, 145, phase, offset, period) / 100
	return uint64(float64(base) * factor)
}

func clampFloat(value, minValue, maxValue float64) float64 {
	return math.Max(minValue, math.Min(value, maxValue))
}

const (
	gib = 1024 * 1024 * 1024
	mib = 1024 * 1024
)
