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

type WatchtowerCPUCollector struct {
	client *SSHClient
}

func NewWatchtowerCPUCollector(client *SSHClient) *WatchtowerCPUCollector {
	return &WatchtowerCPUCollector{client: client}
}

func (c *WatchtowerCPUCollector) CollectCPU(ctx context.Context, hosts []inventory.TargetHost) ([]agent.CPUSnapshot, error) {
	snapshots := make([]agent.CPUSnapshot, 0, len(hosts))
	for _, host := range hosts {
		snapshot := agent.CPUSnapshot{
			HostAlias:   host.Alias,
			HostIP:      host.IP,
			CollectedAt: time.Now(),
		}

		task, err := agent.NewTask(
			fmt.Sprintf("watchtower-cpu-%s", host.Alias),
			host.Alias,
			host.IP,
			host.Port,
			host.User,
			`sh -lc 'read cpu u n s i w irq sirq st g gn < /proc/stat; total1=$((u+n+s+i+w+irq+sirq+st)); idle1=$((i+w)); sleep 0.2; read cpu u n s i w irq sirq st g gn < /proc/stat; total2=$((u+n+s+i+w+irq+sirq+st)); idle2=$((i+w)); diff_total=$((total2-total1)); diff_idle=$((idle2-idle1)); if [ "$diff_total" -eq 0 ]; then echo 0; else awk "BEGIN { printf \"%.1f\", (100*($diff_total-$diff_idle)/$diff_total) }"; fi'`,
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

		usage, err := parseCPUPercent(output)
		if err != nil {
			snapshot.Status = agent.SnapshotStatusFailed
			snapshot.Error = err.Error()
			snapshots = append(snapshots, snapshot)
			continue
		}

		snapshot.UsagePercent = usage
		snapshot.Status = agent.SnapshotStatusSuccess
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

func parseCPUPercent(output string) (float64, error) {
	value := strings.TrimSpace(output)
	if value == "" {
		return 0, fmt.Errorf("empty CPU output")
	}
	usage, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse cpu percent: %w", err)
	}
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}
	return usage, nil
}
