package agent

type ApprovalMode string

const (
	ModePreApproved  ApprovalMode = "PRE_APPROVED"
	ModeApprovalGated ApprovalMode = "APPROVAL_GATED"
)

type Runbook struct {
	ID           string
	Name         string
	Schedule     string
	ApprovalMode ApprovalMode
	Plan         AutopilotPlan
}

func NewRunbook(id, name, schedule string, mode ApprovalMode, plan AutopilotPlan) Runbook {
	return Runbook{
		ID:           id,
		Name:         name,
		Schedule:     schedule,
		ApprovalMode: mode,
		Plan:         plan,
	}
}
