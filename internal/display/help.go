// Package display provides functionality for displaying task lists and help information.
package display

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bab-sh/bab/internal/registry"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/charmbracelet/log"
)

// ListTasks displays all available tasks from the registry.
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

	root := reg.Tree()

	// Get sorted children names for consistent ordering
	childNames := getSortedChildNames(root)

	for _, childName := range childNames {
		child := root.Children[childName]
		displayNode(child, taskNameStyle, descriptionStyle, groupNameStyle, enumeratorStyle)
	}

	fmt.Println("\nRun 'bab <task>' to execute a task")
	fmt.Println("Run 'bab <task> --dry-run' to see what would be executed")

	return nil
}

// displayNode recursively displays a tree node and its children.
func displayNode(node *registry.TreeNode, taskStyle, descStyle, groupStyle, enumStyle lipgloss.Style) {
	if node.IsTask() {
		// Display as a standalone task
		t := tree.New().
			Root(formatTaskWithPadding(node.Name, node.Task.Description, 0, taskStyle, descStyle)).
			Enumerator(tree.RoundedEnumerator).
			EnumeratorStyle(enumStyle)
		fmt.Println(t)
	} else {
		// Display as a group with children
		t := tree.New().
			Root(groupStyle.Render(node.Name)).
			Enumerator(tree.RoundedEnumerator).
			EnumeratorStyle(enumStyle)

		// Calculate max length for padding
		maxLen := calculateMaxLength(node)

		// Add children recursively
		addChildrenToTree(t, node, maxLen, taskStyle, descStyle, groupStyle)

		fmt.Println(t)
	}
}

// addChildrenToTree recursively adds children to a tree.
func addChildrenToTree(t *tree.Tree, node *registry.TreeNode, maxLen int, taskStyle, descStyle, groupStyle lipgloss.Style) {
	childNames := getSortedChildNames(node)

	for _, childName := range childNames {
		child := node.Children[childName]

		if child.IsTask() {
			// Add task as leaf
			taskDisplay := formatTaskWithPadding(child.Name, child.Task.Description, maxLen, taskStyle, descStyle)
			t.Child(taskDisplay)
		} else {
			// Add group as subtree
			subTree := tree.New().
				Root(groupStyle.Render(child.Name)).
				Enumerator(tree.RoundedEnumerator)

			// Recursively add children of this group
			subMaxLen := calculateMaxLength(child)
			addChildrenToTree(subTree, child, subMaxLen, taskStyle, descStyle, groupStyle)

			t.Child(subTree)
		}
	}
}

// calculateMaxLength calculates the maximum name length among direct task children.
func calculateMaxLength(node *registry.TreeNode) int {
	maxLen := 0
	for _, child := range node.Children {
		if child.IsTask() && len(child.Name) > maxLen {
			maxLen = len(child.Name)
		}
	}
	return maxLen
}

// getSortedChildNames returns sorted child names for consistent ordering.
func getSortedChildNames(node *registry.TreeNode) []string {
	names := make([]string, 0, len(node.Children))
	for name := range node.Children {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func formatTaskWithPadding(name, description string, maxLen int, nameStyle, descStyle lipgloss.Style) string {
	if description != "" {
		paddingLen := maxLen - len(name)
		if paddingLen < 0 {
			paddingLen = 0
		}
		padding := strings.Repeat(" ", paddingLen)
		return nameStyle.Render(name) + padding + " - " + descStyle.Render(description)
	}
	return nameStyle.Render(name)
}
