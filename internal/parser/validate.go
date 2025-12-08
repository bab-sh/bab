package parser

import "github.com/bab-sh/bab/internal/babfile"

func validateDependencies(tasks babfile.TaskMap) error {
	for name, task := range tasks {
		for _, dep := range task.Deps {
			if !tasks.Has(dep) {
				return &NotFoundError{
					TaskName:     dep,
					ReferencedBy: name,
					Available:    tasks.Names(),
				}
			}
		}
	}
	return nil
}

func validateRunTaskRefs(tasks babfile.TaskMap) error {
	for name, task := range tasks {
		for _, item := range task.Run {
			if tr, ok := item.(babfile.TaskRun); ok {
				if !tasks.Has(tr.Task) {
					return &NotFoundError{
						TaskName:     tr.Task,
						ReferencedBy: name,
						Available:    tasks.Names(),
					}
				}
			}
		}
	}
	return nil
}

func validateRunCycles(tasks babfile.TaskMap) error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(name string, chain []string) error
	dfs = func(name string, chain []string) error {
		visited[name] = true
		recStack[name] = true
		chain = append(chain, name)

		task := tasks[name]
		for _, item := range task.Run {
			if tr, ok := item.(babfile.TaskRun); ok {
				if recStack[tr.Task] {
					chain = append(chain, tr.Task)
					return &CircularError{
						Type:  "task run",
						Chain: chain,
					}
				}
				if !visited[tr.Task] {
					if err := dfs(tr.Task, chain); err != nil {
						return err
					}
				}
			}
		}

		recStack[name] = false
		return nil
	}

	for name := range tasks {
		if !visited[name] {
			if err := dfs(name, nil); err != nil {
				return err
			}
		}
	}
	return nil
}
