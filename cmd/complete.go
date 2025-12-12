package cmd

import (
	"sort"
	"strings"

	"github.com/bab-sh/bab/internal/runner"
	"github.com/spf13/cobra"
)

func completeTaskNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	babfile, _ := cmd.Flags().GetString("babfile")
	result, err := runner.LoadTasks(babfile)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for taskName, task := range result.Tasks {
		if task == nil || !strings.HasPrefix(taskName, toComplete) {
			continue
		}

		if task.Desc != "" {
			completions = append(completions, taskName+"\t"+task.Desc)
		} else {
			completions = append(completions, taskName)
		}
	}

	sort.Strings(completions)
	return completions, cobra.ShellCompDirectiveNoFileComp
}
