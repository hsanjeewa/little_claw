package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func containsANSI(s string) bool {
	return strings.ContainsRune(s, '\x1b')
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected string
	}{
		{
			name:     "bash via prompt",
			lines:    []string{"$ echo hello"},
			expected: "bash",
		},
		{
			name:     "json via brace",
			lines:    []string{`{"key": "value"}`},
			expected: "json",
		},
		{
			name:     "yaml via colon",
			lines:    []string{"name: test", "items:", "  - one"},
			expected: "yaml",
		},
		{
			name:     "code fence alias",
			lines:    []string{"```sh", "echo hi", "```"},
			expected: "bash",
		},
		{
			name:     "plain text returns empty",
			lines:    []string{"Discover inventory", "Generate plan"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectLanguage(tt.lines)
			if got != tt.expected {
				t.Fatalf("expected language %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestHighlightCode_HighlightsRecognizedLanguages(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
	}{
		{
			name:  "bash",
			lines: []string{"$ echo hello", "world"},
		},
		{
			name:  "json",
			lines: []string{`{"key": "value", "flag": true}`},
		},
		{
			name:  "yaml",
			lines: []string{"name: test", "items:", "  - a", "  - b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := highlightCode(tt.lines)
			if !containsANSI(out) {
				t.Fatalf("expected highlighted %s output to contain ANSI sequences, got:\n%s", tt.name, out)
			}
		})
	}
}

func TestHighlightCode_FallsBackToPlainText(t *testing.T) {
	lines := []string{"No code here", "Just a sentence"}
	out := highlightCode(lines)
	if containsANSI(out) {
		t.Fatalf("expected plain text to remain unhighlighted, got:\n%s", out)
	}
	if out != strings.Join(lines, "\n") {
		t.Fatalf("expected plain text unchanged, got:\n%s", out)
	}
}

func TestRenderPaneBody_RespectsDimensions(t *testing.T) {
	lines := []string{
		"$ echo start",
		strings.Repeat("a", 200),
		"middle",
		"end",
	}

	out := renderPaneBody(lines, 20, 2)
	for _, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > 20 {
			t.Fatalf("line width %d exceeds limit 20: %q", w, line)
		}
	}
	if got := strings.Count(out, "\n") + 1; got > 2 {
		t.Fatalf("expected at most 2 lines, got %d", got)
	}
}

func TestRenderPane_HighlightedContentFitsBounds(t *testing.T) {
	lines := []string{
		"$ systemctl status nginx",
		"```json",
		`{"status": "ok"}`,
		"```",
	}

	pane := renderPane("TEST", lines, 40, 10, false)
	assertRenderedWithinBounds(t, pane, 40, 10)
	if !strings.Contains(pane, "TEST") {
		t.Fatalf("expected pane title to be visible, got:\n%s", pane)
	}
}

func TestAutopilotModel_PaneHighlightsCode(t *testing.T) {
	m := NewAutopilotModel()
	m.transcript = []string{"$ echo hello"}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AutopilotModel)

	view := m.View()
	assertRenderedWithinBounds(t, view, 80, 24)
	if !strings.Contains(view, "echo") {
		t.Fatalf("expected transcript content to be visible, got:\n%s", view)
	}
}

func TestCopilotModel_PaneHighlightsCode(t *testing.T) {
	m := NewCopilotModel()
	m.terminal = []string{"$ ls -la"}
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(CopilotModel)

	view := m.View()
	assertRenderedWithinBounds(t, view, 80, 24)
	if !strings.Contains(view, "ls") {
		t.Fatalf("expected terminal content to be visible, got:\n%s", view)
	}
}

