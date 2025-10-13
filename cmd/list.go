// Package cmd provides command-line interface commands for the bab task runner.
package cmd

import (
	"github.com/bab-sh/bab/internal/display"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available tasks (non-interactive)",
		RunE:  runList,
	}

	return cmd
}

func runList(_ *cobra.Command, _ []string) error {
	reg, _, err := loadRegistry()
	if err != nil {
		return err
	}

	return display.ListTasks(reg)
}
