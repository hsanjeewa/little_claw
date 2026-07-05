package tui

import (
	"fmt"

	"github.com/devops/agent/internal/domain/agent"
)

type WatchtowerEscalationPayload struct {
	Target       Mode
	MetricFamily agent.MetricFamily
	Scope        TargetScope
	SelectedHost string
	Observation  string
	ViewMode     watchtowerViewMode
}

func (p WatchtowerEscalationPayload) Summary() string {
	if p.SelectedHost != "" {
		return fmt.Sprintf("%s • %s • %s", p.SelectedHost, p.MetricFamily, p.Scope.String())
	}
	return fmt.Sprintf("%s • %s", p.MetricFamily, p.Scope.String())
}

type watchtowerEscalationMsg struct {
	Payload WatchtowerEscalationPayload
}
