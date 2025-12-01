package cmd

import (
	"github.com/bab-sh/bab/internal/runner"
	"github.com/bab-sh/bab/internal/tui"
	"github.com/charmbracelet/log"
)

func (c *CLI) runInteractive() error {
	log.Debug("Starting interactive task picker")

	tasks, err := runner.LoadTasks()
	if err != nil {
		return err
	}

	selected, err := tui.PickTask(tasks)
	if err != nil {
		log.Error("Task picker failed", "error", err)
		return err
	}

	if selected == nil {
		log.Debug("No task selected")
		return nil
	}

	log.Debug("Task selected", "name", selected.Name)
	return c.runTask(selected.Name)
}
