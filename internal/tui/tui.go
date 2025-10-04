package tui

import (
	"fmt"
	"os"

	"github.com/bab-sh/bab/internal/executor"
	"github.com/bab-sh/bab/internal/history"
	"github.com/bab-sh/bab/internal/registry"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

// Run starts the interactive TUI for task selection and executes the selected task.
func Run(reg registry.Registry, projectRoot string, dryRun bool, verbose bool) error {
	tasks := reg.List()
	if len(tasks) == 0 {
		log.Info("No tasks available")
		return nil
	}

	if !isInteractive() {
		log.Warn("Non-interactive terminal detected, use 'bab list' to see tasks")
		return fmt.Errorf("interactive mode requires a TTY")
	}

	var historyManager *history.Manager
	if projectRoot != "" {
		hm, err := history.NewManager(projectRoot)
		if err != nil {
			log.Debug("Failed to initialize history manager for TUI", "error", err)
		} else {
			historyManager = hm
		}
	}

	model := NewModel(reg, historyManager)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	m, ok := finalModel.(Model)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	selectedTask := m.SelectedTask()
	if selectedTask == nil {
		return nil
	}

	exec := executor.New(
		executor.WithDryRun(dryRun),
		executor.WithVerbose(verbose),
		executor.WithProjectRoot(projectRoot),
	)

	return exec.Execute(selectedTask)
}

func isInteractive() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
