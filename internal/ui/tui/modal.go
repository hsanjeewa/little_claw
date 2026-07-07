package tui

import (
	"fmt"
	"strings"
)

func centeredModal(content string, width, height int) string {
	lines := strings.Split(content, "\n")
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	maxWidth = min(maxWidth, width-10)

	paddedLines := make([]string, len(lines))
	for i, line := range lines {
		paddedLines[i] = fmt.Sprintf("%*s", maxWidth, line)
	}

	return strings.Join(paddedLines, "\n")
}