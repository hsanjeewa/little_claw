package tui

import (
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const commandBarLabel = "🦀"

var (
	syntaxTheme     = styles.Get("monokai")
	syntaxFormatter = formatters.Get("terminal16m")
	codeFenceRegex  = regexp.MustCompile("(?s)^\\s*```([a-zA-Z0-9_+-]+)\\s*\\n")
)

func init() {
	if syntaxTheme == nil {
		syntaxTheme = styles.Fallback
	}
	if syntaxFormatter == nil {
		syntaxFormatter = formatters.Fallback
	}
}

// renderPaneBody highlights the supplied lines when a recognizable language is
// detected, then truncates and limits the result to the requested dimensions.
// The returned string is safe to place inside a Lipgloss panel of the same
// inner width/height.
func renderPaneBody(lines []string, width, maxLines int) string {
	if width < 1 {
		width = 1
	}
	if maxLines < 0 {
		maxLines = 0
	}

	highlighted := highlightCode(lines)
	return constrainHighlighted(highlighted, width, maxLines)
}

// highlightCode attempts to detect the language of the joined lines and apply
// Chroma syntax highlighting. If detection fails or highlighting errors, the
// original text is returned unchanged so rendering never breaks.
func highlightCode(lines []string) string {
	raw := strings.Join(lines, "\n")
	lang := detectLanguage(lines)
	if lang == "" {
		return raw
	}

	lexer := lexers.Get(lang)
	if lexer == nil {
		return raw
	}

	iterator, err := lexer.Tokenise(nil, raw)
	if err != nil {
		return raw
	}

	var buf strings.Builder
	if err := syntaxFormatter.Format(&buf, syntaxTheme, iterator); err != nil {
		return raw
	}
	return buf.String()
}

// detectLanguage returns a Chroma lexer alias for the content, or an empty
// string when no confident match is found.
func detectLanguage(lines []string) string {
	raw := strings.Join(lines, "\n")

	if match := codeFenceRegex.FindStringSubmatch(raw); len(match) > 1 {
		return normalizeLanguage(match[1])
	}

	if lexer := lexers.Analyse(raw); lexer != nil {
		return normalizeLanguage(lexer.Config().Name)
	}

	firstLine := strings.TrimSpace(raw)
	if idx := strings.IndexByte(firstLine, '\n'); idx != -1 {
		firstLine = firstLine[:idx]
	}

	switch {
	case strings.HasPrefix(firstLine, "$ "),
		strings.HasPrefix(firstLine, "# "),
		strings.HasPrefix(firstLine, "> "):
		return "bash"
	case strings.HasPrefix(firstLine, "{"),
		strings.HasPrefix(firstLine, "["):
		return "json"
	case looksLikeYAML(raw):
		return "yaml"
	}

	return ""
}

// normalizeLanguage maps common aliases to Chroma lexer names.
func normalizeLanguage(lang string) string {
	lower := strings.ToLower(strings.TrimSpace(lang))
	switch lower {
	case "sh", "shell", "zsh", "console":
		return "bash"
	case "yml":
		return "yaml"
	case "md", "mkd":
		return "markdown"
	}
	return lower
}

// looksLikeYAML applies a fast heuristic for YAML content when Chroma analysis
// is inconclusive.
func looksLikeYAML(s string) bool {
	if s == "" {
		return false
	}

	lines := strings.Split(s, "\n")
	colons, dashes := 0, 0
	hasKeyValue := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") {
			dashes++
		}
		if before, _, ok := strings.Cut(trimmed, ":"); ok && strings.TrimSpace(before) != "" {
			colons++
			if strings.Contains(trimmed, ": ") || strings.Contains(trimmed, ":\"") {
				hasKeyValue = true
			}
		}
	}

	return colons > 0 && (dashes > 0 || hasKeyValue)
}

// constrainHighlighted truncates each visible line to width and limits the
// total number of lines to maxLines. ANSI escape sequences are preserved.
func constrainHighlighted(highlighted string, width, maxLines int) string {
	lines := strings.Split(highlighted, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}

	return strings.Join(lines, "\n")
}

// renderPane is the shared implementation for drawing titled Lipgloss panels
// used by Autopilot and Copilot. It applies syntax highlighting to the pane
// body and respects the supplied outer dimensions.
func renderPane(title string, lines []string, width, height int, active bool) string {
	style := panelStyle
	if active {
		style = activePanelStyle
	}

	frameWidth := style.GetHorizontalFrameSize()
	frameHeight := style.GetVerticalFrameSize()
	styleWidth := max(width-frameWidth, 0)
	styleHeight := max(height-frameHeight, 0)
	innerWidth := max(styleWidth-style.GetHorizontalPadding(), 1)
	innerHeight := max(styleHeight-style.GetVerticalPadding()-1, 0)

	header := lipgloss.NewStyle().Bold(true).Render(title)
	body := renderPaneBody(lines, innerWidth, innerHeight)
	if body == "" {
		body = " "
	}

	bodyStyle := lipgloss.NewStyle().Width(innerWidth).Height(innerHeight)
	content := lipgloss.JoinVertical(lipgloss.Left, header, bodyStyle.Render(body))
	return style.Width(styleWidth).Height(styleHeight).Render(content)
}
