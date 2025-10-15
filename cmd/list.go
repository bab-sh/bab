package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bab-sh/bab/internal/finder"
	"github.com/bab-sh/bab/internal/parser"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all available tasks",
	Aliases: []string{"ls", "tasks"},
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

type node struct {
	desc     string
	children map[string]*node
}

func runList(_ *cobra.Command, _ []string) error {
	log.Debug("Starting list command")

	babfilePath, err := finder.FindBabfile()
	if err != nil {
		log.Error("Failed to locate Babfile", "error", err)
		return fmt.Errorf("failed to locate Babfile: %w", err)
	}
	log.Debug("Found Babfile", "path", babfilePath)

	tasks, err := parser.Parse(babfilePath)
	if err != nil {
		log.Error("Failed to parse Babfile", "error", err)
		return fmt.Errorf("failed to parse Babfile: %w", err)
	}
	log.Debug("Parsed tasks successfully", "count", len(tasks))

	if len(tasks) == 0 {
		log.Warn("No tasks found in Babfile")
		return nil
	}

	log.Debug("Building task tree for display")

	root := &node{children: make(map[string]*node)}
	for name, task := range tasks {
		log.Debug("Adding task to tree", "name", name)
		parts := strings.Split(name, ":")
		current := root
		for i, part := range parts {
			if current.children[part] == nil {
				current.children[part] = &node{children: make(map[string]*node)}
			}
			current = current.children[part]
			if i == len(parts)-1 {
				current.desc = task.Description
			}
		}
	}
	log.Debug("Task tree built successfully")

	enumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingRight(1)
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)

	var buildTree func(*node, string) *tree.Tree
	buildTree = func(n *node, name string) *tree.Tree {
		label := name
		if n.desc != "" {
			label += " " + descStyle.Render(n.desc)
		}
		t := tree.Root(label)
		for _, childName := range sortedKeys(n.children) {
			t.Child(buildTree(n.children[childName], childName))
		}
		return t
	}

	for _, name := range sortedKeys(root.children) {
		fmt.Println(buildTree(root.children[name], name).
			EnumeratorStyle(enumStyle).
			ItemStyle(itemStyle))
	}

	return nil
}

func sortedKeys(m map[string]*node) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
