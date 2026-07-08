package agent

import "context"

type AutopilotPlan struct {
	ID            string
	Root          PlanNode
	Status        TaskStatus
	BlockedSubset []string
}

func NewAutopilotPlan(id string, root PlanNode) AutopilotPlan {
	return AutopilotPlan{
		ID:     id,
		Root:   root,
		Status: StatusPending,
	}
}

type TaskRunner interface {
	RunTask(ctx context.Context, task Task, hitlChan chan<- HitlRequest)
}

type RecoveryManager interface {
	Retry(ctx context.Context, hostAlias string, node PlanNode) error
	Skip(ctx context.Context, hostAlias string, node PlanNode) error
}

type AutopilotRunner struct {
	runner TaskRunner
}

func NewAutopilotRunner(runner TaskRunner) *AutopilotRunner {
	return &AutopilotRunner{
		runner: runner,
	}
}

func (r *AutopilotRunner) ExecutePlan(ctx context.Context, plan AutopilotPlan) error {
	return r.executeNode(ctx, plan.Root)
}

func (r *AutopilotRunner) executeNode(ctx context.Context, node PlanNode) error {
	if !node.Enabled {
		return nil
	}
	if node.Type == NodeTask && node.Task != nil {
		r.runner.RunTask(ctx, *node.Task, nil)
		if node.Task.VerificationTask != nil {
			r.runner.RunTask(ctx, *node.Task.VerificationTask, nil)
		}
	}
	for _, child := range node.Children {
		if err := r.executeNode(ctx, child); err != nil {
			return err
		}
	}
	return nil
}
