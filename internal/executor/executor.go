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

func getShellCommand() (string, string) {
	if runtime.GOOS == "windows" {
		return "cmd", "/C"
	}
	return "sh", "-c"
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
		log.Debug("Executing command", "task", task.Name, "index", i+1, "total", len(task.Commands), "command", command)

		if command == "" {
			log.Debug("Empty command detected", "task", task.Name, "index", i+1)
			return fmt.Errorf("task %q has empty command at index %d", task.Name, i+1)
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
