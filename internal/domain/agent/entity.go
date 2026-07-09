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
	ID               string
	HostAlias        string
	HostIP           string
	Port             int
	User             string
	Command          string
	IsMutative       bool
	Status           TaskStatus
	VerificationTask *Task
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
		ID:               id,
		HostAlias:        hostAlias,
		HostIP:           hostIP,
		Port:             port,
		User:             user,
		Command:          command,
		IsMutative:       isMutative,
		Status:           StatusPending,
		VerificationTask: nil,
	}, nil
}

type ExecutionLog struct {
	ID        string
	Timestamp time.Time
	Host      string
	Command   string
	Status    TaskStatus
	Output    string
	Error     string
}

type SecretVault interface {
	GetPrivateKey(hostAlias string) (string, error)
	GetSudoPassword(hostAlias string) (string, error)
}

type NotificationSink interface {
	Emit(ctx context.Context, topic string, payload map[string]interface{}) error
}

type IdempotencyHelper interface {
	IsSatisfied(ctx context.Context, task Task) (bool, string)
}

type AIAnalyzer interface {
	AnalyzeOutput(ctx context.Context, command string, output string) (string, error)
	PlanTasks(ctx context.Context, goal string) ([]Task, error)
}

type HitlRequest struct {
	Task         Task
	ResponseChan chan bool
}

type AuditRepository interface {
	SaveLog(ctx context.Context, log ExecutionLog) error
	GetLogs(ctx context.Context, host string) ([]ExecutionLog, error)
	GetFailedLogs(ctx context.Context) ([]ExecutionLog, error)
}
