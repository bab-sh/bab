package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bab-sh/bab/internal/finder"
	"github.com/bab-sh/bab/internal/parser"
	"github.com/charmbracelet/log"
)

type status int

const (
	_ status = iota
	running
	done
)

type Runner struct {
	DryRun bool
}

func New(dryRun bool) *Runner {
	return &Runner{DryRun: dryRun}
}

func LoadTasks() (parser.TaskMap, error) {
	path, err := finder.FindBabfile()
	if err != nil {
		return nil, err
	}
	return parser.Parse(path)
}

func (r *Runner) Run(ctx context.Context, taskName string) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}
	return r.RunWithTasks(ctx, taskName, tasks)
}

func (r *Runner) RunWithTasks(ctx context.Context, taskName string, tasks parser.TaskMap) error {
	state := make(map[string]status)
	return r.runTask(ctx, taskName, tasks, state)
}

func (r *Runner) runTask(ctx context.Context, name string, tasks parser.TaskMap, state map[string]status) error {
	switch state[name] {
	case done:
		return nil
	case running:
		return fmt.Errorf("circular dependency detected: %s", buildChain(name, tasks, state))
	}

	task, ok := tasks[name]
	if !ok {
		return fmt.Errorf("task %q not found", name)
	}

	state[name] = running

	for _, dep := range task.Dependencies {
		log.Debug("Running dependency", "task", name, "dep", dep)
		if err := r.runTask(ctx, dep, tasks, state); err != nil {
			return fmt.Errorf("dependency %q failed: %w", dep, err)
		}
	}

	if err := r.executeTask(ctx, task); err != nil {
		return err
	}

	state[name] = done
	return nil
}

func (r *Runner) executeTask(ctx context.Context, task *parser.Task) error {
	if len(task.Commands) == 0 {
		return fmt.Errorf("task %q has no commands", task.Name)
	}

	shell, shellArg := shellCommand()
	platform := runtime.GOOS
	executed := 0

	log.Debug("Executing task", "name", task.Name, "commands", len(task.Commands), "dryRun", r.DryRun)

	for i, cmd := range task.Commands {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancelled: %w", ctx.Err())
		default:
		}

		if !cmd.ShouldRunOnPlatform(platform) {
			log.Debug("Skipping command", "task", task.Name, "index", i+1, "reason", "platform")
			continue
		}

		if strings.TrimSpace(cmd.Cmd) == "" {
			return fmt.Errorf("task %q command %d is empty", task.Name, i+1)
		}

		if r.DryRun {
			log.Info("Would run", "cmd", cmd.Cmd)
		} else {
			log.Debug("Running", "cmd", cmd.Cmd)
			if err := runCommand(ctx, shell, shellArg, cmd.Cmd); err != nil {
				return fmt.Errorf("command %d failed: %w", i+1, err)
			}
		}
		executed++
	}

	if executed == 0 {
		return fmt.Errorf("task %q has no commands for platform %q", task.Name, platform)
	}

	return nil
}

func shellCommand() (string, string) {
	if runtime.GOOS == "windows" {
		return "cmd", "/C"
	}
	return "sh", "-c"
}

func runCommand(ctx context.Context, shell, shellArg, command string) error {
	cmd := exec.CommandContext(ctx, shell, shellArg, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func buildChain(current string, tasks parser.TaskMap, state map[string]status) string {
	chain := []string{current}
	seen := make(map[string]bool)

	for {
		last := chain[len(chain)-1]
		if seen[last] {
			break
		}
		seen[last] = true

		task, ok := tasks[last]
		if !ok || len(task.Dependencies) == 0 {
			break
		}

		for _, dep := range task.Dependencies {
			if state[dep] == running {
				chain = append(chain, dep)
				if dep == current {
					return strings.Join(chain, " → ")
				}
				break
			}
		}
	}
	return strings.Join(chain, " → ")
}
