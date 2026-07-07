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

const networkSampleSleep = "0.5"

type WatchtowerNetworkCollector struct {
	client *SSHClient
}

func NewWatchtowerNetworkCollector(client *SSHClient) *WatchtowerNetworkCollector {
	return &WatchtowerNetworkCollector{client: client}
}

func (c *WatchtowerNetworkCollector) CollectNetwork(ctx context.Context, hosts []inventory.TargetHost) ([]agent.NetworkSnapshot, error) {
	snapshots := make([]agent.NetworkSnapshot, 0, len(hosts))
	for _, host := range hosts {
		snapshot := agent.NetworkSnapshot{
			HostAlias:   host.Alias,
			HostIP:      host.IP,
			CollectedAt: time.Now(),
		}

		command := fmt.Sprintf(
			`read rx1 tx1 _ < <(awk 'NR>2&&$1!="lo:"{print $2,$10;exit}' /proc/net/dev); sleep %s; read rx2 tx2 _ < <(awk 'NR>2&&$1!="lo:"{print $2,$10;exit}' /proc/net/dev); echo $((rx2-rx1)) $((tx2-tx1))`,
			networkSampleSleep,
		)

		task, err := agent.NewTask(
			fmt.Sprintf("watchtower-network-%s", host.Alias),
			host.Alias,
			host.IP,
			host.Port,
			host.User,
			command,
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

		rxDelta, txDelta, err := parseNetworkDelta(output)
		if err != nil {
			snapshot.Status = agent.SnapshotStatusFailed
			snapshot.Error = err.Error()
			snapshots = append(snapshots, snapshot)
			continue
		}

		snapshot.RxBytesPerSec = uint64(float64(rxDelta) / 0.5)
		snapshot.TxBytesPerSec = uint64(float64(txDelta) / 0.5)
		snapshot.Status = agent.SnapshotStatusSuccess
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// parseNetworkDevOutput extracts rx and tx bytes from the first non-loopback
// interface in /proc/net/dev output.
func parseNetworkDevOutput(output string) (uint64, uint64, error) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Inter-") || strings.HasPrefix(line, " face") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 10 {
			continue
		}

		rx, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse rx bytes for %s: %w", iface, err)
		}

		tx, err := strconv.ParseUint(fields[8], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse tx bytes for %s: %w", iface, err)
		}

		return rx, tx, nil
	}

	return 0, 0, fmt.Errorf("no non-loopback interface found in /proc/net/dev")
}

// parseNetworkDelta parses "rx_delta tx_delta" output from the SSH command.
func parseNetworkDelta(output string) (uint64, uint64, error) {
	line := strings.TrimSpace(output)
	if line == "" {
		return 0, 0, fmt.Errorf("empty network delta output")
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("unexpected network delta output: %q", output)
	}

	rx, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse rx delta: %w", err)
	}

	tx, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse tx delta: %w", err)
	}

	return rx, tx, nil
}
