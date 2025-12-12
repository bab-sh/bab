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
	"github.com/bab-sh/bab/internal/output"
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
	DryRun    bool
	Babfile   string
	GlobalEnv map[string]string
}

func New(dryRun bool, babfile string) *Runner {
	return &Runner{DryRun: dryRun, Babfile: babfile}
}

func LoadTasks(customPath string) (*parser.ParseResult, error) {
	var path string
	if customPath != "" {
		path = customPath
	} else {
		p, err := finder.FindBabfile()
		if err != nil {
			return nil, err
		}
		path = p
	}
	return parser.Parse(path)
}

func (r *Runner) Run(ctx context.Context, taskName string) error {
	result, err := LoadTasks(r.Babfile)
	if err != nil {
		return err
	}
	r.GlobalEnv = result.GlobalEnv
	return r.RunWithTasks(ctx, taskName, result.Tasks)
}

func (r *Runner) RunWithTasks(ctx context.Context, taskName string, tasks babfile.TaskMap) error {
	state := make(map[string]status)
	return r.runTask(ctx, taskName, tasks, state, true)
}

func (r *Runner) runTask(ctx context.Context, name string, tasks babfile.TaskMap, state map[string]status, isMain bool) error {
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

	if !r.DryRun {
		if isMain {
			output.Task(name)
		} else {
			output.Dep(name)
		}
	}

	for _, dep := range task.Deps {
		log.Debug("Running dependency", "task", name, "dep", dep)
		if err := r.runTask(ctx, dep, tasks, state, false); err != nil {
			return fmt.Errorf("dependency %q failed: %w", dep, err)
		}
	}

	if len(task.Run) > 0 {
		if err := r.executeTask(ctx, task, tasks, state); err != nil {
			return err
		}
	}

	state[name] = done
	return nil
}

func (r *Runner) executeTask(ctx context.Context, task *babfile.Task, tasks babfile.TaskMap, state map[string]status) error {
	shell, shellArg := shellCommand()
	platform := runtime.GOOS
	executed := 0

	taskEnv := babfile.MergeEnvMaps(r.GlobalEnv, task.Env)

	log.Debug("Executing task", "name", task.Name, "runItems", len(task.Run), "dryRun", r.DryRun, "envVars", len(taskEnv))

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

			cmdEnv := babfile.MergeEnvMaps(taskEnv, v.Env)

			if r.DryRun {
				log.Info("Would run", "cmd", v.Cmd, "env", len(cmdEnv))
			} else {
				output.Cmd(v.Cmd)
				if err := runCommand(ctx, shell, shellArg, v.Cmd, cmdEnv); err != nil {
					return fmt.Errorf("task %q: command %d failed: %w", task.Name, i+1, err)
				}
			}

		case babfile.TaskRun:
			if r.DryRun {
				log.Info("Would run task", "task", v.Task)
			} else {
				log.Debug("Running task", "task", v.Task)
				if err := r.runTask(ctx, v.Task, tasks, state, false); err != nil {
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

func runCommand(ctx context.Context, shell, shellArg, command string, env map[string]string) error {
	cmd := exec.CommandContext(ctx, shell, shellArg, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if len(env) > 0 {
		cmd.Env = append(os.Environ(), babfile.MergeEnv(env)...)
	}

	err := cmd.Run()
	if err != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	return err
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
