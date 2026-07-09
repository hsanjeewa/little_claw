package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/devops/agent/internal/infrastructure/inventory"
	"github.com/devops/agent/internal/infrastructure/llm"
)

func TestPrivilegeInstruction_ContextAware(t *testing.T) {
	cases := []struct {
		name    string
		user    string
		mode    SudoMode
		want    string
		forbid  bool // true => must forbid "sudo"
	}{
		{
			name:   "root user needs no sudo",
			user:   "root",
			mode:   SudoNone,
			want:   "do NOT use sudo",
			forbid: true,
		},
		{
			name: "non-root passwordless sudo",
			user: "deployer",
			mode: SudoPasswordless,
			want: "prefix commands with sudo",
		},
		{
			name: "non-root sudo with password",
			user: "deployer",
			mode: SudoPassword,
			want: "prefix commands with sudo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := privilegeInstruction(tc.user, tc.mode)
			if !strings.Contains(got, tc.want) {
				t.Fatalf("expected privilege instruction to contain %q, got %q", tc.want, got)
			}
			if tc.forbid && strings.Contains(strings.ToLower(got), "prefix commands with sudo") {
				t.Fatalf("root host must NOT be told to prefix with sudo, got %q", got)
			}
		})
	}
}

func TestAutopilot_GeneratePlan_CommandTriggersPlanGeneration(t *testing.T) {
	m := NewAutopilotModel()
	m.width = 80
	m.height = 24

	goal := "Check nginx status"
	for _, r := range goal {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(AutopilotModel)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(AutopilotModel)

	if cmd == nil {
		t.Fatal("expected Enter to return command")
	}

	msg := cmd()
	if _, ok := msg.(GeneratePlanMsg); !ok {
		t.Fatal("expected GeneratePlanMsg")
	}
}

func TestAutopilot_PlanGeneration_AddsReasoningToTranscript(t *testing.T) {
	m := NewAutopilotModel()

	goal := "Check nginx"
	cmd := m.handlePlanGeneration(goal)
	result := cmd()

	planMsg, ok := result.(PlanGeneratedMsg)
	if !ok {
		t.Fatal("expected PlanGeneratedMsg")
	}

	m.transcript = append(m.transcript, TranscriptEntry{
		Kind: TranscriptAgent,
		Text: planMsg.Reasoning,
	}.String())

	transcript := strings.Join(m.transcript, "\n")
	if !strings.Contains(transcript, planMsg.Reasoning) {
		t.Fatalf("expected transcript to contain reasoning, got: %s", transcript)
	}

	if !strings.Contains(transcript, "🤖") {
		t.Fatal("expected agent emoji in transcript")
	}
}

func TestAutopilot_PlanePane_DisplaysPlanWithMutativeFlags(t *testing.T) {
	m := NewAutopilotModel()
	m.width = 80
	m.height = 24

	m.plan = []string{
		"1. Check nginx status \n   systemctl status nginx",
		"2. Restart nginx [MUTATIVE]\n   systemctl restart nginx",
	}

	view := m.View()
	if !strings.Contains(view, "1. Check nginx status") {
		t.Fatal("expected plan pane to show step 1")
	}

	if !strings.Contains(view, "[MUTATIVE]") {
		t.Fatal("expected plan pane to show mutative flag")
	}
}

func TestAutopilot_InvalidJSONResponse_HandledGracefully(t *testing.T) {
	m := NewAutopilotModel()

	invalidJSON := `not valid json`
	_, err := llm.ParsePlan(invalidJSON)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	errorMsg := TranscriptEntry{
		Kind: TranscriptSystem,
		Text: fmt.Sprintf("Failed to parse plan: %v", err),
	}
	m.transcript = append(m.transcript, errorMsg.String())

	transcript := strings.Join(m.transcript, "\n")
	if !strings.Contains(transcript, "Failed to parse") {
		t.Fatal("expected transcript to show error message")
	}

	if !strings.Contains(transcript, "ℹ️") {
		t.Fatal("expected system emoji in transcript")
	}
}

func TestGetCapabilities_ContextAwareSudo(t *testing.T) {
	policy := func(alias string) SudoMode {
		switch alias {
		case "web-prod-01":
			return SudoPasswordless // non-root, NOPASSWD sudo
		case "db-master":
			return SudoPassword // non-root, password configured
		case "root-host":
			return SudoNone // runs as root
		}
		return SudoPasswordless
	}
	client := (&LLMClient{}).WithSudoPolicy(policy)

	hosts := []inventory.TargetHost{
		{Alias: "web-prod-01", IP: "127.0.0.1", Port: 2222, User: "deployer"},
		{Alias: "db-master", IP: "127.0.0.1", Port: 2223, User: "postgres"},
		{Alias: "root-host", IP: "127.0.0.1", Port: 22, User: "root"},
	}

	caps := client.getCapabilities(hosts, WatchtowerStateSnapshot{}, map[string]HostEnvironment{})

	// Non-root passwordless host: told to prefix with sudo.
	if !strings.Contains(caps, "web-prod-01: sudo available (passwordless); prefix commands with sudo") {
		t.Fatalf("expected web-prod-01 to require sudo, got:\n%s", caps)
	}
	// Non-root password host: told to prefix with sudo (auto-supplied).
	if !strings.Contains(caps, "db-master: sudo required; prefix commands with sudo (password is supplied automatically)") {
		t.Fatalf("expected db-master to require sudo with password, got:\n%s", caps)
	}
	// Root host: explicitly told NOT to use sudo.
	if !strings.Contains(caps, "root-host: runs as root; do NOT use sudo") {
		t.Fatalf("expected root-host to forbid sudo, got:\n%s", caps)
	}

	// Constraint must reinforce following the per-host capability.
	constraints := client.getConstraints()
	if !strings.Contains(constraints, "follow the per-host capability") {
		t.Fatalf("expected constraint to reference per-host capability, got:\n%s", constraints)
	}
}