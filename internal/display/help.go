package display

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/bab/bab/internal/registry"
	"github.com/fatih/color"
)

func ListTasks(reg registry.Registry) error {
	tasks := reg.List()
	if len(tasks) == 0 {
		fmt.Println("No tasks available")
		return nil
	}

	header := color.New(color.FgCyan, color.Bold)
	header.Println("\nAvailable tasks:")

	tree := reg.Tree()

	if rootTasks, exists := tree[""]; exists && len(rootTasks) > 0 {
		fmt.Println("\n Root tasks:")
		printTaskList(rootTasks)
	}

	for group, groupTasks := range tree {
		if group == "" {
			continue
		}

		groupHeader := color.New(color.FgYellow, color.Bold)
		groupHeader.Printf("\n %s:\n", group)
		printTaskList(groupTasks)
	}

	fmt.Println("\nRun 'bab <task>' to execute a task")
	fmt.Println("Run 'bab <task> --dry-run' to see what would be executed")

	return nil
}

func printTaskList(tasks []*registry.Task) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	for _, task := range tasks {
		taskColor := color.New(color.FgGreen)
		name := taskColor.Sprint(task.Name)

		if task.Description != "" {
			fmt.Fprintf(w, "  %s\t%s\n", name, task.Description)
		} else {
			fmt.Fprintf(w, "  %s\t\n", name)
		}
	}

	w.Flush()
}