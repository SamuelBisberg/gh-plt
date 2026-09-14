package pkg

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Theme is a small, consistent stylesheet used for every status message
// gh-plt prints: green for success, red for errors, yellow for warnings, and
// cyan for informational output.
type Theme struct {
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style
}

// NewTheme builds a Theme using lipgloss's default terminal color
// detection.
func NewTheme() *Theme {
	r := lipgloss.NewRenderer(os.Stdout)
	return &Theme{
		Success: r.NewStyle().Foreground(lipgloss.Color("42")).Bold(true),
		Error:   r.NewStyle().Foreground(lipgloss.Color("204")).Bold(true),
		Warning: r.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		Info:    r.NewStyle().Foreground(lipgloss.Color("39")),
	}
}

// Successf prints a "✓"-prefixed success message to stdout.
func (t *Theme) Successf(format string, a ...any) {
	fmt.Println(t.Success.Render("✓") + " " + fmt.Sprintf(format, a...))
}

// Errorf prints a "✗"-prefixed error message to stderr.
func (t *Theme) Errorf(format string, a ...any) {
	fmt.Fprintln(os.Stderr, t.Error.Render("✗")+" "+fmt.Sprintf(format, a...))
}

// Warningf prints a "⚠"-prefixed warning message to stdout.
func (t *Theme) Warningf(format string, a ...any) {
	fmt.Println(t.Warning.Render("⚠") + " " + fmt.Sprintf(format, a...))
}

// Infof prints an "ℹ"-prefixed informational message to stdout.
func (t *Theme) Infof(format string, a ...any) {
	fmt.Println(t.Info.Render("ℹ") + " " + fmt.Sprintf(format, a...))
}
