package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/bab-sh/bab/internal/parser"
	"github.com/charmbracelet/log"
)

const (
	windowsShell = "cmd"
	windowsArg   = "/C"
	unixShell    = "sh"
	unixArg      = "-c"
)

func getShellCommand() (string, string) {
	if runtime.GOOS == "windows" {
		return windowsShell, windowsArg
	}
	return unixShell, unixArg
}

func Execute(ctx context.Context, task *parser.Task) error {
	if task == nil {
		log.Debug("Execute called with nil task")
		return fmt.Errorf("task cannot be nil")
	}

	log.Debug("Starting task execution", "task", task.Name, "command-count", len(task.Commands))

	if len(task.Commands) == 0 {
		log.Debug("Task has no commands", "task", task.Name)
		return fmt.Errorf("task %q has no commands to execute", task.Name)
	}

	shell, shellArg := getShellCommand()
	log.Debug("Using shell", "shell", shell, "arg", shellArg, "platform", runtime.GOOS)

	for i, command := range task.Commands {
		select {
		case <-ctx.Done():
			return fmt.Errorf("task execution cancelled: %w", ctx.Err())
		default:
		}

		log.Debug("Executing command", "task", task.Name, "index", i+1, "total", len(task.Commands), "command", command)

		if err := validateCommand(command); err != nil {
			log.Debug("Invalid command detected", "task", task.Name, "index", i+1, "error", err)
			return fmt.Errorf("task %q has invalid command at index %d: %w", task.Name, i+1, err)
		}

		if err := executeCommand(ctx, shell, shellArg, command); err != nil {
			log.Debug("Command failed", "task", task.Name, "index", i+1, "error", err)
			return fmt.Errorf("command %d failed: %w", i+1, err)
		}

		log.Debug("Command completed successfully", "task", task.Name, "index", i+1)
	}

	log.Debug("Task execution completed", "task", task.Name)
	return nil
}

func executeCommand(ctx context.Context, shell, shellArg, command string) error {
	cmd := exec.CommandContext(ctx, shell, shellArg, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func DryRun(ctx context.Context, task *parser.Task) error {
	if task == nil {
		log.Debug("DryRun called with nil task")
		return fmt.Errorf("task cannot be nil")
	}

	log.Debug("Starting dry-run for task", "task", task.Name, "command-count", len(task.Commands))

	if len(task.Commands) == 0 {
		log.Debug("Task has no commands", "task", task.Name)
		return fmt.Errorf("task %q has no commands to execute", task.Name)
	}

	if task.Description != "" {
		log.Debug("Task description", "desc", task.Description)
	}

	if len(task.Dependencies) > 0 {
		log.Debug("Dependencies", "deps", task.Dependencies)
	}

	for i, command := range task.Commands {
		select {
		case <-ctx.Done():
			return fmt.Errorf("dry-run cancelled: %w", ctx.Err())
		default:
		}

		log.Debug("Command", "step", fmt.Sprintf("[%d/%d]", i+1, len(task.Commands)), "cmd", command)
	}

	log.Debug("Dry-run completed", "task", task.Name)
	return nil
}
