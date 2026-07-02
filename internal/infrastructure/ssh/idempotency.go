package ssh

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/devops/agent/internal/domain/agent"
)

type LinuxIdempotencyAnalyzer struct {
	client *SSHClient
}

func NewLinuxIdempotencyAnalyzer(client *SSHClient) *LinuxIdempotencyAnalyzer {
	return &LinuxIdempotencyAnalyzer{client: client}
}

func (h *LinuxIdempotencyAnalyzer) IsSatisfied(ctx context.Context, task agent.Task) (bool, string) {
	lower := strings.ToLower(task.Command)
	
	if strings.Contains(lower, "apt-get install") || strings.Contains(lower, "apt install") {
		fields := strings.Fields(task.Command)
		pkg := fields[len(fields)-1]
		
		checkTask := task
		checkTask.Command = fmt.Sprintf("dpkg-query -W -f='${Status}' %s 2>/dev/null", pkg)
		checkTask.IsMutative = false
		
		output, err := h.client.Execute(ctx, checkTask)
		if err == nil && strings.Contains(output, "ok installed") {
			return true, fmt.Sprintf("Skipped: Package [%s] is already installed on the system state.", pkg)
		}
	}
	
	return false, ""
}

var _ agent.IdempotencyHelper = (*LinuxIdempotencyAnalyzer)(nil)
