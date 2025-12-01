package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bab-sh/bab/internal/executor"
	"github.com/bab-sh/bab/internal/finder"
	"github.com/bab-sh/bab/internal/parser"
	"github.com/bab-sh/bab/internal/tui"
	"github.com/bab-sh/bab/internal/version"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	verbose    bool
	dryRun     bool
	listTasks  bool
	completion string
	rootCtx    context.Context

	rootCmd = &cobra.Command{
		Use:           "bab [task]",
		Short:         "Custom commands for every project",
		Version:       version.Version,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if completion != "" {
				return runCompletion(cmd, completion)
			}
			if listTasks {
				return runList()
			}
			if len(args) > 0 {
				return executeTask(rootCtx, args[0])
			}
			return runInteractive(rootCtx)
		},
		Args: cobra.ArbitraryArgs,
	}
)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Show commands without executing")
	rootCmd.Flags().BoolVarP(&listTasks, "list", "l", false, "List all available tasks")
	rootCmd.Flags().StringVarP(&completion, "completion", "c", "", "Generate completion script (bash|zsh|fish|powershell)")
}

func ExecuteContext(ctx context.Context) error {
	log.Debug("Starting bab execution")
	rootCtx = ctx

	return rootCmd.Execute()
}

func executeTask(ctx context.Context, taskName string) error {
	log.Debug("Executing task", "name", taskName, "dry-run", dryRun)

	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	executed := make(map[string]bool)
	executing := make(map[string]bool)

	return executeTaskWithDeps(ctx, taskName, tasks, executed, executing)
}

func loadTasks() (parser.TaskMap, error) {
	path, err := finder.FindBabfile()
	if err != nil {
		log.Error("Failed to locate Babfile", "error", err)
		return nil, err
	}
	log.Debug("Found Babfile", "path", path)

	tasks, err := parser.Parse(path)
	if err != nil {
		log.Error("Failed to parse Babfile", "error", err)
		return nil, err
	}
	log.Debug("Parsed Babfile", "task-count", len(tasks))

	return tasks, nil
}

func executeTaskWithDeps(ctx context.Context, taskName string, tasks parser.TaskMap, executed, executing map[string]bool) error {
	if executed[taskName] {
		log.Debug("Task already executed, skipping", "name", taskName)
		return nil
	}

	if executing[taskName] {
		chain := buildDependencyChain(taskName, executing, tasks)
		return fmt.Errorf("circular dependency detected: %s", chain)
	}

	task, exists := tasks[taskName]
	if !exists {
		return fmt.Errorf("task %q not found", taskName)
	}

	executing[taskName] = true
	defer delete(executing, taskName)

	for _, dep := range task.Dependencies {
		log.Debug("Executing dependency", "task", taskName, "dependency", dep)
		if err := executeTaskWithDeps(ctx, dep, tasks, executed, executing); err != nil {
			return fmt.Errorf("dependency %q of task %q failed: %w", dep, taskName, err)
		}
	}

	log.Debug("Executing task", "name", taskName, "commands", len(task.Commands))

	var err error
	if dryRun {
		err = executor.DryRun(ctx, task)
	} else {
		err = executor.Execute(ctx, task)
	}

	if err != nil {
		log.Error("Task failed", "name", taskName, "error", err)
		return err
	}

	log.Info("Task completed", "name", taskName)
	executed[taskName] = true
	return nil
}

func buildDependencyChain(currentTask string, executing map[string]bool, tasks parser.TaskMap) string {
	chain := []string{currentTask}
	visited := make(map[string]bool)

	for len(chain) < len(tasks) {
		lastTask := chain[len(chain)-1]
		if visited[lastTask] {
			break
		}
		visited[lastTask] = true

		task, exists := tasks[lastTask]
		if !exists || len(task.Dependencies) == 0 {
			break
		}

		for _, dep := range task.Dependencies {
			if executing[dep] {
				chain = append(chain, dep)
				if dep == currentTask {
					return strings.Join(chain, " → ")
				}
				break
			}
		}
	}

	return strings.Join(chain, " → ")
}

func runInteractive(ctx context.Context) error {
	log.Debug("Starting interactive task picker")

	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	selected, err := tui.PickTask(tasks)
	if err != nil {
		log.Error("Task picker failed", "error", err)
		return err
	}

	if selected == nil {
		log.Debug("No task selected")
		return nil
	}

	log.Debug("Task selected", "name", selected.Name)
	return executeTask(ctx, selected.Name)
}

type node struct {
	desc     string
	children map[string]*node
}

func runList() error {
	log.Debug("Starting list command")

	tasks, err := loadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
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

func runCompletion(cmd *cobra.Command, shell string) error {
	switch shell {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	default:
		return fmt.Errorf("unknown shell %q (valid: bash, zsh, fish, powershell)", shell)
	}
}
