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
	reg, err := loadRegistry()
	if err != nil {
		return err
	}

	return display.ListTasks(reg)
}
