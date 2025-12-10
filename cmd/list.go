package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bab-sh/bab/internal/runner"
	"github.com/bab-sh/bab/internal/theme"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/charmbracelet/log"
)

type node struct {
	desc     string
	children map[string]*node
}

func (c *CLI) runList() error {
	tasks, err := runner.LoadTasks(c.babfile)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		log.Warn("No tasks found")
		return nil
	}

	root := &node{children: make(map[string]*node)}
	for name, task := range tasks {
		parts := strings.Split(name, ":")
		current := root
		for i, part := range parts {
			if current.children[part] == nil {
				current.children[part] = &node{children: make(map[string]*node)}
			}
			current = current.children[part]
			if i == len(parts)-1 {
				current.desc = task.Desc
			}
		}
	}

	enumStyle := lipgloss.NewStyle().Foreground(theme.Gray).PaddingRight(1)
	itemStyle := lipgloss.NewStyle().Foreground(theme.Purple).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)

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
