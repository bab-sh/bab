package display

import (
	"fmt"
	"strings"

	"github.com/bab/bab/internal/registry"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/charmbracelet/log"
)

func ListTasks(reg registry.Registry) error {
	tasks := reg.List()
	if len(tasks) == 0 {
		log.Info("No tasks available")
		return nil
	}

	taskNameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	descriptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	groupNameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	enumeratorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	fmt.Println("\nAvailable tasks:")

	taskTree := reg.Tree()

	if rootTasks, exists := taskTree[""]; exists && len(rootTasks) > 0 {
		maxLen := 0
		for _, task := range rootTasks {
			if len(task.Name) > maxLen {
				maxLen = len(task.Name)
			}
		}

		for _, task := range rootTasks {
			t := tree.New().
				Root(formatTaskWithPadding(task.Name, task.Description, maxLen, taskNameStyle, descriptionStyle)).
				Enumerator(tree.RoundedEnumerator).
				EnumeratorStyle(enumeratorStyle)
			fmt.Println(t)
		}
	}

	for group, groupTasks := range taskTree {
		if group == "" {
			continue
		}

		maxLen := 0
		for _, task := range groupTasks {
			shortName := strings.TrimPrefix(task.Name, group+":")
			if len(shortName) > maxLen {
				maxLen = len(shortName)
			}
		}

		t := tree.New().
			Root(groupNameStyle.Render(group)).
			Enumerator(tree.RoundedEnumerator).
			EnumeratorStyle(enumeratorStyle)

		for _, task := range groupTasks {
			shortName := strings.TrimPrefix(task.Name, group+":")
			taskDisplay := formatTaskWithPadding(shortName, task.Description, maxLen, taskNameStyle, descriptionStyle)
			t.Child(taskDisplay)
		}

		fmt.Println(t)
	}

	fmt.Println("\nRun 'bab <task>' to execute a task")
	fmt.Println("Run 'bab <task> --dry-run' to see what would be executed")

	return nil
}

func formatTaskWithPadding(name, description string, maxLen int, nameStyle, descStyle lipgloss.Style) string {
	if description != "" {
		padding := strings.Repeat(" ", maxLen-len(name))
		return nameStyle.Render(name) + padding + " - " + descStyle.Render(description)
	}
	return nameStyle.Render(name)
}
