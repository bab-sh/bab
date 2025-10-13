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
		Short:         "A modern task runner from simple to scaled",
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

func Execute() {
	log.Debug("Starting bab execution")

	if rootCmd.Execute() == nil {
		log.Debug("Command executed successfully")
		return
	}

	if len(os.Args) < 2 {
		log.Error("No command or task specified")
		os.Exit(1)
	}

	checkVerboseFlag()

	taskName := os.Args[1]
	log.Debug("No command matched, attempting to execute as task", "arg", taskName)
	if err := executeTask(taskName); err != nil {
		os.Exit(1)
	}
	log.Debug("Task executed successfully")
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

func executeTask(taskName string) error {
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
	if err := executor.Execute(context.Background(), task); err != nil {
		log.Error("Task failed", "name", taskName, "error", err)
		return err
	}

	log.Info("Task completed successfully", "name", taskName)
	return nil
}
