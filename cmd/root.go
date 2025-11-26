package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/bab-sh/bab/internal/executor"
	"github.com/bab-sh/bab/internal/finder"
	"github.com/bab-sh/bab/internal/parser"
	"github.com/bab-sh/bab/internal/tui"
	"github.com/bab-sh/bab/internal/version"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	dryRun  bool
	rootCtx context.Context

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
			if len(args) > 0 {
				return executeTask(rootCtx, args[0])
			}
			return runInteractive(rootCtx)
		},
		Args: cobra.ArbitraryArgs,
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Show commands without executing")
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
