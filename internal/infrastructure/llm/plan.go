package llm

import (
	"encoding/json"
	"errors"
	"fmt"
)

func (p PlanStep) Validate() error {
	if p.Description == "" {
		return errors.New("description is required")
	}
	if p.Command == "" {
		return errors.New("command is required")
	}
	return nil
}

func ValidatePlan(plan []PlanStep) error {
	for i, step := range plan {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("step %d validation failed: %w", i, err)
		}
	}
	return nil
}

func ParsePlan(jsonPlan string) ([]PlanStep, error) {
	var steps []PlanStep
	err := json.Unmarshal([]byte(jsonPlan), &steps)
	if err != nil {
		return nil, fmt.Errorf("context: %w", fmt.Errorf("invalid JSON: %w", err))
	}

	for i, step := range steps {
		if err := step.Validate(); err != nil {
			return nil, fmt.Errorf("context: %w", fmt.Errorf("step %d validation failed: %w", i, err))
		}
	}

	return steps, nil
}