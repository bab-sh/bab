package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bab-sh/bab/internal/executor"
	"github.com/bab-sh/bab/internal/finder"
	"github.com/bab-sh/bab/internal/parser"
	"github.com/bab-sh/bab/internal/version"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	dryRun  bool

	rootCmd = &cobra.Command{
		Use:           "bab",
		Short:         "Custom commands for every project",
		Version:       version.Version,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "Show commands without executing")
}

func ExecuteContext(ctx context.Context) error {
	log.Debug("Starting bab execution")

	if err := rootCmd.Execute(); err == nil {
		log.Debug("Command executed successfully")
		return nil
	}

	if len(os.Args) < 2 {
		log.Error("No command or task specified")
		return fmt.Errorf("no command or task specified")
	}

	if err := rootCmd.ParseFlags(os.Args[1:]); err != nil {
		log.Error("Failed to parse flags", "error", err)
		return err
	}

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	var taskName string
	for _, arg := range os.Args[1:] {
		if !strings.HasPrefix(arg, "-") {
			taskName = arg
			break
		}
	}

	if taskName == "" {
		log.Error("No task specified")
		return fmt.Errorf("no task specified")
	}

	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == taskName || containsString(cmd.Aliases, taskName) {
			log.Debug("Command error occurred", "command", taskName)
			return fmt.Errorf("command %q failed", taskName)
		}
	}

	log.Debug("No command matched, attempting to execute as task", "arg", taskName)
	if err := executeTask(ctx, taskName); err != nil {
		return err
	}
	log.Debug("Task executed successfully")
	return nil
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func executeTask(ctx context.Context, taskName string) error {
	log.Debug("Executing task", "name", taskName, "dry-run", dryRun)

	babfilePath, err := finder.FindBabfile()
	if err != nil {
		log.Error("Failed to locate Babfile", "error", err)
		return err
	}
	log.Debug("Found Babfile", "path", babfilePath)

	tasks, err := parser.Parse(babfilePath)
	if err != nil {
		log.Error("Failed to parse Babfile", "error", err)
		return err
	}
	log.Debug("Parsed Babfile", "task-count", len(tasks))

	executed := make(map[string]bool)
	executing := make(map[string]bool)

	return executeTaskWithDeps(ctx, taskName, tasks, executed, executing)
}

func executeTaskWithDeps(ctx context.Context, taskName string, tasks parser.TaskMap, executed map[string]bool, executing map[string]bool) error {
	if executed[taskName] {
		log.Debug("Task already executed, skipping", "name", taskName)
		return nil
	}

	if executing[taskName] {
		chain := buildDependencyChain(taskName, executing, tasks)
		log.Error("Circular dependency detected", "chain", chain)
		return fmt.Errorf("circular dependency detected: %s", chain)
	}

	task, exists := tasks[taskName]
	if !exists {
		log.Error("Task not found", "name", taskName)
		return fmt.Errorf("task %q not found", taskName)
	}

	executing[taskName] = true
	defer delete(executing, taskName)

	if len(task.Dependencies) > 0 {
		log.Debug("Executing dependencies first", "task", taskName, "deps", task.Dependencies)
		for _, dep := range task.Dependencies {
			log.Debug("Executing dependency", "task", taskName, "dependency", dep)
			if err := executeTaskWithDeps(ctx, dep, tasks, executed, executing); err != nil {
				return fmt.Errorf("dependency %q of task %q failed: %w", dep, taskName, err)
			}
		}
	}

	log.Debug("Found task", "name", taskName, "commands", len(task.Commands))

	if dryRun {
		log.Info("Running task", "name", taskName, "dry-run", true)
		if err := executor.DryRun(ctx, task); err != nil {
			log.Error("Task dry-run failed", "name", taskName, "error", err)
			return err
		}
		log.Info("Task dry-run completed", "name", taskName)
	} else {
		log.Info("Executing task", "name", taskName)
		if err := executor.Execute(ctx, task); err != nil {
			log.Error("Task failed", "name", taskName, "error", err)
			return err
		}
		log.Info("Task completed successfully", "name", taskName)
	}

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
