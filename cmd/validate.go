package cmd

import (
	"github.com/bab-sh/bab/internal/runner"
	"github.com/charmbracelet/log"
)

func (c *CLI) runValidate() error {
	tasks, err := runner.LoadTasks()
	if err != nil {
		return err
	}
	log.Info("Babfile is valid", "tasks", len(tasks))
	return nil
}
