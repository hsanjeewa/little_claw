package ssh

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/inventory"
)

type WatchtowerMemoryCollector struct {
	client *SSHClient
}

func NewWatchtowerMemoryCollector(client *SSHClient) *WatchtowerMemoryCollector {
	return &WatchtowerMemoryCollector{client: client}
}

func (c *WatchtowerMemoryCollector) CollectMemory(ctx context.Context, hosts []inventory.TargetHost) ([]agent.MemorySnapshot, error) {
	snapshots := make([]agent.MemorySnapshot, 0, len(hosts))
	for _, host := range hosts {
		snapshot := agent.MemorySnapshot{
			HostAlias:   host.Alias,
			HostIP:      host.IP,
			CollectedAt: time.Now(),
		}

		task, err := agent.NewTask(
			fmt.Sprintf("watchtower-%s", host.Alias),
			host.Alias,
			host.IP,
			host.Port,
			host.User,
			"cat /proc/meminfo",
			false,
		)
		if err != nil {
			snapshot.Status = agent.SnapshotStatusFailed
			snapshot.Error = err.Error()
			snapshots = append(snapshots, snapshot)
			continue
		}

		output, err := c.client.Execute(ctx, task)
		if err != nil {
			snapshot.Status = agent.SnapshotStatusFailed
			snapshot.Error = err.Error()
			snapshots = append(snapshots, snapshot)
			continue
		}

		total, available, err := parseMemInfo(output)
		if err != nil {
			snapshot.Status = agent.SnapshotStatusFailed
			snapshot.Error = err.Error()
			snapshots = append(snapshots, snapshot)
			continue
		}

		used := total - available
		usedPercent := 0.0
		if total > 0 {
			usedPercent = (float64(used) / float64(total)) * 100
		}

		snapshot.TotalBytes = total
		snapshot.UsedBytes = used
		snapshot.UsedPercent = usedPercent
		snapshot.Status = agent.SnapshotStatusSuccess
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

func parseMemInfo(output string) (uint64, uint64, error) {
	var totalKB uint64
	var availableKB uint64

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf("parse MemTotal: %w", err)
			}
			totalKB = value
		case "MemAvailable:":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf("parse MemAvailable: %w", err)
			}
			availableKB = value
		}
	}

	if totalKB == 0 {
		return 0, 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
	}
	if availableKB == 0 {
		return 0, 0, fmt.Errorf("MemAvailable not found in /proc/meminfo")
	}
	if availableKB > totalKB {
		availableKB = totalKB
	}

	return totalKB * 1024, availableKB * 1024, nil
}
