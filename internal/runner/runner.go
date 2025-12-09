package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bab-sh/bab/internal/babfile"
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

func LoadTasks() (babfile.TaskMap, error) {
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

func (r *Runner) RunWithTasks(ctx context.Context, taskName string, tasks babfile.TaskMap) error {
	state := make(map[string]status)
	return r.runTask(ctx, taskName, tasks, state)
}

func (r *Runner) runTask(ctx context.Context, name string, tasks babfile.TaskMap, state map[string]status) error {
	switch state[name] {
	case done:
		return nil
	case running:
		return &parser.CircularError{
			Type:  "dependency",
			Chain: buildChainSlice(name, tasks, state),
		}
	}

	task, ok := tasks[name]
	if !ok {
		return &parser.NotFoundError{
			TaskName:  name,
			Available: tasks.Names(),
		}
	}

	state[name] = running

	for _, dep := range task.Deps {
		log.Debug("Running dependency", "task", name, "dep", dep)
		if err := r.runTask(ctx, dep, tasks, state); err != nil {
			return fmt.Errorf("dependency %q failed: %w", dep, err)
		}
	}

	if err := r.executeTask(ctx, task, tasks, state); err != nil {
		return err
	}

	state[name] = done
	return nil
}

func (r *Runner) executeTask(ctx context.Context, task *babfile.Task, tasks babfile.TaskMap, state map[string]status) error {
	if len(task.Run) == 0 {
		return fmt.Errorf("task %q has no run items", task.Name)
	}

	shell, shellArg := shellCommand()
	platform := runtime.GOOS
	executed := 0

	log.Debug("Executing task", "name", task.Name, "runItems", len(task.Run), "dryRun", r.DryRun)

	for i, item := range task.Run {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancelled: %w", ctx.Err())
		default:
		}

		if !item.ShouldRunOnPlatform(platform) {
			log.Debug("Skipping run item", "task", task.Name, "index", i+1, "reason", "platform")
			continue
		}

		switch v := item.(type) {
		case babfile.CommandRun:
			if strings.TrimSpace(v.Cmd) == "" {
				return fmt.Errorf("task %q command %d is empty", task.Name, i+1)
			}

			if r.DryRun {
				log.Info("Would run", "cmd", v.Cmd)
			} else {
				log.Debug("Running", "cmd", v.Cmd)
				if err := runCommand(ctx, shell, shellArg, v.Cmd); err != nil {
					return fmt.Errorf("task %q: command %d failed: %w", task.Name, i+1, err)
				}
			}

		case babfile.TaskRun:
			if r.DryRun {
				log.Info("Would run task", "task", v.Task)
			} else {
				log.Debug("Running task", "task", v.Task)
				if err := r.runTask(ctx, v.Task, tasks, state); err != nil {
					return fmt.Errorf("task %q failed: %w", v.Task, err)
				}
			}
		}

		executed++
	}

	if executed == 0 {
		return fmt.Errorf("task %q has no run items for platform %q", task.Name, platform)
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

func buildChainSlice(current string, tasks babfile.TaskMap, state map[string]status) []string {
	chain := []string{current}
	seen := make(map[string]bool)

	for {
		last := chain[len(chain)-1]
		if seen[last] {
			break
		}
		seen[last] = true

		task, ok := tasks[last]
		if !ok || len(task.Deps) == 0 {
			break
		}

		for _, dep := range task.Deps {
			if state[dep] == running {
				chain = append(chain, dep)
				if dep == current {
					return chain
				}
				break
			}
		}
	}
	return chain
}
