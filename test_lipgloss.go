package main
import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)
func main() {
	style := lipgloss.NewStyle().Width(10)
	rendered := style.Render("This is a very long string that might wrap or overflow")
	fmt.Printf("Rendered:\n%s\n", rendered)
	fmt.Printf("Height: %d\n", lipgloss.Height(rendered))
}
