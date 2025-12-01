package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/bab-sh/bab/internal/executor"
	"github.com/bab-sh/bab/internal/finder"
	"github.com/bab-sh/bab/internal/parser"
	"github.com/charmbracelet/log"
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

func (r *Runner) Run(ctx context.Context, taskName string) error {
	log.Debug("Executing task", "name", taskName, "dry-run", r.DryRun)

	tasks, err := LoadTasks()
	if err != nil {
		return err
	}

	executed := make(map[string]bool)
	executing := make(map[string]bool)

	return r.runWithDeps(ctx, taskName, tasks, executed, executing)
}

func (r *Runner) RunWithTasks(ctx context.Context, taskName string, tasks parser.TaskMap) error {
	executed := make(map[string]bool)
	executing := make(map[string]bool)

	return r.runWithDeps(ctx, taskName, tasks, executed, executing)
}

func (r *Runner) runWithDeps(ctx context.Context, taskName string, tasks parser.TaskMap, executed, executing map[string]bool) error {
	if executed[taskName] {
		log.Debug("Task already executed, skipping", "name", taskName)
		return nil
	}

	if executing[taskName] {
		chain := BuildDependencyChain(taskName, executing, tasks)
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
		if err := r.runWithDeps(ctx, dep, tasks, executed, executing); err != nil {
			return fmt.Errorf("dependency %q of task %q failed: %w", dep, taskName, err)
		}
	}

	log.Debug("Executing task", "name", taskName, "commands", len(task.Commands))

	var err error
	if r.DryRun {
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

func BuildDependencyChain(currentTask string, executing map[string]bool, tasks parser.TaskMap) string {
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
