package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/bab-sh/bab/internal/parser"
	"github.com/charmbracelet/log"
)

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

	for i, command := range task.Commands {
		log.Debug("Executing command", "task", task.Name, "index", i+1, "total", len(task.Commands), "command", command)

		cmd := exec.CommandContext(ctx, "sh", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			log.Debug("Command failed", "task", task.Name, "index", i+1, "error", err)
			return fmt.Errorf("command %d failed: %w", i+1, err)
		}

		log.Debug("Command completed successfully", "task", task.Name, "index", i+1)
	}

	log.Debug("Task execution completed", "task", task.Name)
	return nil
}
