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

type WatchtowerStorageCollector struct {
	client *SSHClient
}

func NewWatchtowerStorageCollector(client *SSHClient) *WatchtowerStorageCollector {
	return &WatchtowerStorageCollector{client: client}
}

func (c *WatchtowerStorageCollector) CollectStorage(ctx context.Context, hosts []inventory.TargetHost) ([]agent.StorageSnapshot, error) {
	snapshots := make([]agent.StorageSnapshot, 0, len(hosts))
	for _, host := range hosts {
		snapshot := agent.StorageSnapshot{
			HostAlias:   host.Alias,
			HostIP:      host.IP,
			CollectedAt: time.Now(),
		}

		task, err := agent.NewTask(
			fmt.Sprintf("watchtower-storage-%s", host.Alias),
			host.Alias,
			host.IP,
			host.Port,
			host.User,
			"df / | tail -1 | awk '{print $2, $3}'",
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

		totalBytes, usedBytes, err := parseDFOutput(output)
		if err != nil {
			snapshot.Status = agent.SnapshotStatusFailed
			snapshot.Error = err.Error()
			snapshots = append(snapshots, snapshot)
			continue
		}

		usedPercent := 0.0
		if totalBytes > 0 {
			usedPercent = (float64(usedBytes) / float64(totalBytes)) * 100
		}

		snapshot.TotalBytes = totalBytes
		snapshot.UsedBytes = usedBytes
		snapshot.UsedPercent = usedPercent
		snapshot.Status = agent.SnapshotStatusSuccess
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// parseDFOutput parses "total_kb used_kb" from df output.
func parseDFOutput(output string) (uint64, uint64, error) {
	line := strings.TrimSpace(output)
	if line == "" {
		return 0, 0, fmt.Errorf("empty df output")
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("unexpected df output: %q", output)
	}

	totalKB, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse total blocks: %w", err)
	}

	usedKB, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse used blocks: %w", err)
	}

	if usedKB > totalKB {
		usedKB = totalKB
	}

	return totalKB * 1024, usedKB * 1024, nil
}
