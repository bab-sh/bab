package output

import (
	"fmt"
	"io"
	"os"

	"github.com/bab-sh/bab/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

var (
	taskIndicator = lipgloss.NewStyle().Foreground(theme.Pink)
	taskAction    = lipgloss.NewStyle().Foreground(theme.Purple)
	taskName      = lipgloss.NewStyle().Foreground(theme.White).Bold(true)
	secondary     = lipgloss.NewStyle().Foreground(theme.Gray)
)

var Writer io.Writer = os.Stderr

func Task(name string) {
	_, _ = fmt.Fprintf(Writer, "%s %s %s\n",
		taskIndicator.Render("●"),
		taskAction.Render("Running"),
		taskName.Render(name),
	)
}

func Cmd(cmd string) {
	_, _ = fmt.Fprintf(Writer, "%s %s\n",
		secondary.Render("$"),
		secondary.Render(cmd),
	)
}

func Dep(name string) {
	_, _ = fmt.Fprintf(Writer, "%s %s\n",
		secondary.Render("▶"),
		secondary.Render(name),
	)
}
