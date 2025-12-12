package cmd

import (
	"github.com/bab-sh/bab/internal/runner"
	"github.com/charmbracelet/log"
)

func (c *CLI) runValidate() error {
	result, err := runner.LoadTasks(c.babfile)
	if err != nil {
		return err
	}
	log.Info("Babfile is valid", "tasks", len(result.Tasks))
	return nil
}
