package agent

import "time"

// MetricFamily identifies the Watchtower metric family currently being shown.
type MetricFamily string

const (
	MetricFamilyMemory  MetricFamily = "MEMORY"
	MetricFamilyCPU     MetricFamily = "CPU"
	MetricFamilyStorage MetricFamily = "STORAGE"
	MetricFamilyNetwork MetricFamily = "NETWORK"
)

// SnapshotStatus describes whether a host snapshot was collected successfully.
type SnapshotStatus string

const (
	SnapshotStatusSuccess SnapshotStatus = "SUCCESS"
	SnapshotStatusFailed  SnapshotStatus = "FAILED"
)

// MemorySnapshot is the normalized memory view of one host at one point in time.
type MemorySnapshot struct {
	HostAlias   string
	HostIP      string
	TotalBytes  uint64
	UsedBytes   uint64
	UsedPercent float64
	CollectedAt time.Time
	Status      SnapshotStatus
	Error       string
}

// CPUSnapshot is the normalized CPU view of one host at one point in time.
type CPUSnapshot struct {
	HostAlias    string
	HostIP       string
	UsagePercent float64
	CollectedAt  time.Time
	Status       SnapshotStatus
	Error        string
}

// StorageSnapshot is the normalized storage view of one host at one point in time.
type StorageSnapshot struct {
	HostAlias   string
	HostIP      string
	TotalBytes  uint64
	UsedBytes   uint64
	UsedPercent float64
	CollectedAt time.Time
	Status      SnapshotStatus
	Error       string
}

// NetworkSnapshot is the normalized network view of one host at one point in time.
type NetworkSnapshot struct {
	HostAlias     string
	HostIP        string
	RxBytesPerSec uint64
	TxBytesPerSec uint64
	CollectedAt   time.Time
	Status        SnapshotStatus
	Error         string
}
