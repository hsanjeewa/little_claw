package main
import (
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/lipgloss"
)
func main() {
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("FAILED")
	line := fmt.Sprintf("> [%s] very long string", styled)
	trunc := ansi.Truncate(line, 15, "...")
	fmt.Printf("Truncated:\n%s\n", trunc)
}
