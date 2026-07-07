package agent

type NodeType string

const (
	NodeGoal   NodeType = "GOAL"
	NodeBranch NodeType = "BRANCH"
	NodeTask   NodeType = "TASK"
)

type PlanNode struct {
	ID       string
	Type     NodeType
	Children []PlanNode
	Task     *Task
	Enabled  bool
}

func NewPlanNode(id string, nodeType NodeType, task *Task) PlanNode {
	return PlanNode{
		ID:      id,
		Type:    nodeType,
		Task:    task,
		Enabled: true,
	}
}

func (n *PlanNode) Disable() {
	n.Enabled = false
}
