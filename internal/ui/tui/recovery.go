package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/infrastructure/llm"
)

type RecoveryPlanGeneratedMsg struct {
	Plan []llm.PlanStep
}

type RecoveryApprovedMsg struct{}

type RecoveryRejectedMsg struct{}

type RecoveryFailedMsg struct {
	Error string
}

const (
	defaultMaxRetries = 3
)

func (m AutopilotModel) triggerRecovery(errorMsg string) (AutopilotModel, tea.Cmd) {
	if m.run.RetryCount >= m.run.MaxRetries {
		m.run.State = RunStateFailed
		m.transcript = append(m.transcript, TranscriptEntry{
			Kind: TranscriptSystem,
			Text: fmt.Sprintf("❌ Recovery failed after %d attempts. Max retries reached.", m.run.RetryCount),
		}.String())
		return m, nil
	}

	retryCount := m.run.RetryCount + 1
	m.run.RetryCount = retryCount
	m.run.OriginalError = errorMsg
	m.run.State = RunStateDrafting

	m.transcript = append(m.transcript, TranscriptEntry{
		Kind: TranscriptSystem,
		Text: fmt.Sprintf("🔄 Generating recovery plan (attempt %d/%d)...", retryCount, m.run.MaxRetries),
	}.String())

	return m, func() tea.Msg {
		recoveryPlan := m.generateRecoveryPlan(errorMsg)
		return RecoveryPlanGeneratedMsg{Plan: recoveryPlan}
	}
}

func (m AutopilotModel) generateRecoveryPlan(errorMsg string) []llm.PlanStep {
	if len(m.run.Plan) == 0 || m.run.LastCompletedStep >= len(m.run.Plan) {
		return []llm.PlanStep{
			{
				Description: "Investigate and resolve underlying issue",
				Command:     "echo 'Manual investigation required'",
				IsMutative:  true,
			},
		}
	}

	var recoverySteps []llm.PlanStep

	if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "refused") {
		recoverySteps = []llm.PlanStep{
			{
				Description: "Check network connectivity",
				Command:     "ping -c 3 8.8.8.8",
				IsMutative:  false,
			},
			{
				Description: "Verify SSH service is running",
				Command:     "ps aux | grep -E '(sshd|ssh)' | grep -v grep",
				IsMutative:  false,
			},
			{
				Description: "Check for listening ports",
				Command:     "ss -tlnp | grep -E '(22|ssh)' || netstat -tlnp | grep -E '(22|ssh)'",
				IsMutative:  false,
			},
		}
	} else if strings.Contains(errorMsg, "permission") || strings.Contains(errorMsg, "access") {
		recoverySteps = []llm.PlanStep{
			{
				Description: "Check current user permissions",
				Command:     "id",
				IsMutative:  false,
			},
			{
				Description: "Verify sudo access",
				Command:     "sudo -v",
				IsMutative:  false,
			},
			{
				Description: "Check file permissions",
				Command:     "ls -la /var/log/nginx 2>/dev/null || ls -la /var/log 2>/dev/null",
				IsMutative:  false,
			},
		}
	} else {
		recoverySteps = []llm.PlanStep{
			{
				Description: "Check system logs for errors",
				Command:     "dmesg | tail -20 || tail -20 /var/log/syslog 2>/dev/null || tail -20 /var/log/messages 2>/dev/null",
				IsMutative:  false,
			},
			{
				Description: "Verify service status",
				Command:     "ps aux | head -20",
				IsMutative:  false,
			},
			{
				Description: "Check disk space",
				Command:     "df -h",
				IsMutative:  false,
			},
		}
	}

	return recoverySteps
}