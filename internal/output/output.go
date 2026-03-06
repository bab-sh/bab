package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bab-sh/bab/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

var (
	taskIndicator   = lipgloss.NewStyle().Foreground(theme.Pink)
	taskAction      = lipgloss.NewStyle().Foreground(theme.Purple)
	taskName        = lipgloss.NewStyle().Foreground(theme.White).Bold(true)
	secondary       = lipgloss.NewStyle().Foreground(theme.Gray)
	parallelSuccess = lipgloss.NewStyle().Foreground(theme.Cyan)
	parallelFailure = lipgloss.NewStyle().Foreground(theme.Pink)
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

func RenderTask(name string) string {
	return fmt.Sprintf("%s %s %s",
		taskIndicator.Render("●"),
		taskAction.Render("Running"),
		taskName.Render(name),
	)
}

func RenderCmd(cmd string) string {
	return fmt.Sprintf("%s %s",
		secondary.Render("$"),
		secondary.Render(cmd),
	)
}

func RenderDep(name string) string {
	return fmt.Sprintf("%s %s",
		secondary.Render("▶"),
		secondary.Render(name),
	)
}

func ParallelDone(labels []string, errs []error) {
	if len(errs) < len(labels) {
		return
	}
	var parts []string
	for i, label := range labels {
		if errs[i] != nil {
			parts = append(parts, parallelFailure.Render(label+" ✗"))
		} else {
			parts = append(parts, parallelSuccess.Render(label+" ✓"))
		}
	}

	_, _ = fmt.Fprintf(Writer, "%s %s %s\n",
		taskIndicator.Render("●"),
		taskAction.Render("Parallel"),
		secondary.Render(fmt.Sprintf("[%s]", strings.Join(parts, "  "))),
	)
}
