package output

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

var (
	taskIndicator     = lipgloss.NewStyle().Foreground(theme.Pink)
	taskAction        = lipgloss.NewStyle().Foreground(theme.Purple)
	taskName          = lipgloss.NewStyle().Foreground(theme.White).Bold(true)
	secondary         = lipgloss.NewStyle().Foreground(theme.Gray)
	parallelSuccess   = lipgloss.NewStyle().Foreground(theme.Cyan)
	parallelFailure   = lipgloss.NewStyle().Foreground(theme.Pink)
	parallelCancelled = lipgloss.NewStyle().Foreground(theme.Dim)
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

var logLevelStyles = map[babfile.LogLevel]lipgloss.Style{
	babfile.LogLevelDebug: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")),
	babfile.LogLevelInfo:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")),
	babfile.LogLevelWarn:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("192")),
	babfile.LogLevelError: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("204")),
}

func RenderLog(msg string, level babfile.LogLevel) string {
	style, ok := logLevelStyles[level]
	if !ok {
		style = logLevelStyles[babfile.LogLevelInfo]
	}
	return fmt.Sprintf("%s %s", style.Render(strings.ToUpper(string(level))), msg)
}

func ParallelDone(labels []string, errs []error) {
	if len(errs) < len(labels) {
		return
	}
	var parts []string
	for i, label := range labels {
		if errs[i] != nil {
			if errors.Is(errs[i], context.Canceled) {
				parts = append(parts, parallelCancelled.Render(label+" ⊘"))
			} else {
				parts = append(parts, parallelFailure.Render(label+" ✗"))
			}
		} else {
			parts = append(parts, parallelSuccess.Render(label+" ✓"))
		}
	}

	_, _ = fmt.Fprintf(Writer, "%s %s %s%s%s\n",
		taskIndicator.Render("●"),
		taskAction.Render("Parallel"),
		secondary.Render("["),
		strings.Join(parts, "  "),
		secondary.Render("]"),
	)
}
