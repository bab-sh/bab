package cmd

import (
	"sort"
	"strings"

	"github.com/bab-sh/bab/internal/runner"
	"github.com/spf13/cobra"
)

func completeTaskNames(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	tasks, err := runner.LoadTasks()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for taskName, task := range tasks {
		if task == nil || !strings.HasPrefix(taskName, toComplete) {
			continue
		}

		if task.Description != "" {
			completions = append(completions, taskName+"\t"+task.Description)
		} else {
			completions = append(completions, taskName)
		}
	}

	sort.Strings(completions)
	return completions, cobra.ShellCompDirectiveNoFileComp
}
