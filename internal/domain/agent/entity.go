package agent

import (
	"context"
	"fmt"
	"time"
)

type TaskStatus string

const (
	StatusPending    TaskStatus = "PENDING"
	StatusRunning    TaskStatus = "RUNNING"
	StatusSuccess    TaskStatus = "SUCCESS"
	StatusChanged    TaskStatus = "CHANGED"
	StatusFailed     TaskStatus = "FAILED"
	StatusIdempotent TaskStatus = "IDEMPOTENT"
	StatusSkipped    TaskStatus = "SKIPPED"
)

type Task struct {
	ID         string
	HostAlias  string
	HostIP     string
	Port       int
	User       string
	Command    string
	IsMutative bool
	Status     TaskStatus
}

func NewTask(id, hostAlias, hostIP string, port int, user, command string, isMutative bool) (Task, error) {
	if hostIP == "" {
		return Task{}, fmt.Errorf("context: %w", fmt.Errorf("HostIP cannot be empty"))
	}
	if port <= 0 || port >= 65536 {
		return Task{}, fmt.Errorf("context: %w", fmt.Errorf("Port must be between 1 and 65535"))
	}
	if command == "" {
		return Task{}, fmt.Errorf("context: %w", fmt.Errorf("Command cannot be empty"))
	}
	if user == "" {
		return Task{}, fmt.Errorf("context: %w", fmt.Errorf("User cannot be empty"))
	}

	return Task{
		ID:         id,
		HostAlias:  hostAlias,
		HostIP:     hostIP,
		Port:       port,
		User:       user,
		Command:    command,
		IsMutative: isMutative,
		Status:     StatusPending,
	}, nil
}

type ExecutionLog struct {
	ID        string
	Timestamp time.Time
	Host      string
	Command   string
	Status    TaskStatus
	Output    string
}

type SecretVault interface {
	GetPassword(host string) (string, error)
}

type NotificationSink interface {
	Notify(ctx context.Context, msg string) error
}

type AuditRepository interface {
	SaveLog(ctx context.Context, log ExecutionLog) error
	GetLogs(ctx context.Context, host string) ([]ExecutionLog, error)
}
