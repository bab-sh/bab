package cmd

import (
	"github.com/bab-sh/bab/internal/runner"
	"github.com/bab-sh/bab/internal/tui"
	"github.com/charmbracelet/log"
)

func (c *CLI) runInteractive() error {
	tasks, err := runner.LoadTasks()
	if err != nil {
		return err
	}

	selected, err := tui.PickTask(tasks)
	if err != nil {
		return err
	}

	if selected == nil {
		log.Debug("No task selected")
		return nil
	}

	log.Debug("Running task", "name", selected.Name)
	return c.runTask(selected.Name)
}
