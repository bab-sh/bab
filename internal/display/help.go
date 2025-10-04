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

var (
	taskNameStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	descriptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	groupNameStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	enumeratorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
)

// ListTasks displays all available tasks from the registry.
func ListTasks(reg registry.Registry) error {
	tasks := reg.List()
	if len(tasks) == 0 {
		log.Info("No tasks available")
		return nil
	}

	fmt.Println("\nAvailable tasks:")

	root := reg.Tree()

	// Get sorted children names for consistent ordering
	childNames := getSortedChildNames(root)

	for _, childName := range childNames {
		child := root.Children[childName]
		displayNode(child)
	}

	fmt.Println("\nRun 'bab <task>' to execute a task")
	fmt.Println("Run 'bab <task> --dry-run' to see what would be executed")

	return nil
}

// displayNode recursively displays a tree node and its children.
func displayNode(node *registry.TreeNode) {
	if node.IsTask() {
		// Display as a standalone task
		t := tree.New().
			Root(formatTaskWithPadding(node.Name, node.Task.Description, 0)).
			Enumerator(tree.RoundedEnumerator).
			EnumeratorStyle(enumeratorStyle)
		fmt.Println(t)
	} else {
		// Display as a group with children
		t := tree.New().
			Root(groupNameStyle.Render(node.Name)).
			Enumerator(tree.RoundedEnumerator).
			EnumeratorStyle(enumeratorStyle)

		// Calculate max length for padding
		maxLen := calculateMaxLength(node)

		// Add children recursively
		addChildrenToTree(t, node, maxLen)

		fmt.Println(t)
	}
}

// addChildrenToTree recursively adds children to a tree.
func addChildrenToTree(t *tree.Tree, node *registry.TreeNode, maxLen int) {
	childNames := getSortedChildNames(node)

	for _, childName := range childNames {
		child := node.Children[childName]

		if child.IsTask() {
			// Add task as leaf
			taskDisplay := formatTaskWithPadding(child.Name, child.Task.Description, maxLen)
			t.Child(taskDisplay)
		} else {
			// Add group as subtree
			subTree := tree.New().
				Root(groupNameStyle.Render(child.Name)).
				Enumerator(tree.RoundedEnumerator)

			// Recursively add children of this group
			subMaxLen := calculateMaxLength(child)
			addChildrenToTree(subTree, child, subMaxLen)

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

func formatTaskWithPadding(name, description string, maxLen int) string {
	if description != "" {
		paddingLen := maxLen - len(name)
		if paddingLen < 0 {
			paddingLen = 0
		}
		padding := strings.Repeat(" ", paddingLen)
		return taskNameStyle.Render(name) + padding + " - " + descriptionStyle.Render(description)
	}
	return taskNameStyle.Render(name)
}
