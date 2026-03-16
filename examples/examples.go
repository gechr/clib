package examples

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	demoBulletStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#4ADE80"})
	demoFlagStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#1D4ED8", Dark: "#93C5FD"})
)

func DemoMessage() string {
	lines := []string{
		"This is a demo of clib's help rendering and CLI framework integration:",
		"",
		fmt.Sprintf(
			"  %s Specify %s for short help",
			demoBulletStyle.Render("•"),
			demoFlagStyle.Render("-h"),
		),
		fmt.Sprintf(
			"  %s Specify %s for long help with examples",
			demoBulletStyle.Render("•"),
			demoFlagStyle.Render("--help"),
		),
	}
	return strings.Join(lines, "\n")
}
