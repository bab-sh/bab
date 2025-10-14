package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/bab-sh/bab/internal/executor"
	"github.com/bab-sh/bab/internal/finder"
	"github.com/bab-sh/bab/internal/parser"
	"github.com/bab-sh/bab/internal/version"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	verbose bool

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
}

func Execute() error {
	return ExecuteContext(context.Background())
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

	commandName := os.Args[1]
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == commandName || containsString(cmd.Aliases, commandName) {
			log.Debug("Command error occurred", "command", commandName)
			return fmt.Errorf("command %q failed", commandName)
		}
	}

	checkVerboseFlag()

	log.Debug("No command matched, attempting to execute as task", "arg", commandName)
	if err := executeTask(ctx, commandName); err != nil {
		return err
	}
	log.Debug("Task executed successfully")
	return nil
}

func checkVerboseFlag() {
	for _, arg := range os.Args {
		if arg == "-v" || arg == "--verbose" {
			log.SetLevel(log.DebugLevel)
			verbose = true
			break
		}
	}
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
	log.Debug("Executing task", "name", taskName)

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

	task, exists := tasks[taskName]
	if !exists {
		log.Error("Task not found", "name", taskName)
		return fmt.Errorf("task %q not found", taskName)
	}
	log.Debug("Found task", "name", taskName, "commands", len(task.Commands))

	log.Info("Executing task", "name", taskName)
	if err := executor.Execute(ctx, task); err != nil {
		log.Error("Task failed", "name", taskName, "error", err)
		return err
	}

	log.Info("Task completed successfully", "name", taskName)
	return nil
}
