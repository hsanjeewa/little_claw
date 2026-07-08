package llm

type GeneratedPlan struct {
	Reasoning string     `json:"reasoning"`
	Steps     []PlanStep `json:"steps"`
}

type PlanStep struct {
	Description string `json:"description"`
	Command     string `json:"command"`
	IsMutative  bool   `json:"is_mutative"`
	VerifyCmd   string `json:"verify_command,omitempty"`
}