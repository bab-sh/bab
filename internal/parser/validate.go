package parser

import "github.com/bab-sh/bab/internal/babfile"

func validateAll(path string, tasks babfile.TaskMap) error {
	errs := &ValidationErrors{}
	validateDependencies(path, tasks, errs)
	validateRunTaskRefs(path, tasks, errs)
	validateRunCycles(path, tasks, errs)
	return errs.OrNil()
}

func validateDependencies(path string, tasks babfile.TaskMap, errs *ValidationErrors) {
	for name, task := range tasks {
		for _, dep := range task.Deps {
			if !tasks.Has(dep) {
				line := task.DepsLine
				if line == 0 {
					line = task.Line
				}
				errs.Add(&NotFoundError{
					Path:         path,
					Line:         line,
					TaskName:     dep,
					ReferencedBy: name,
					Available:    tasks.Names(),
				})
			}
		}
	}
}

func validateRunTaskRefs(path string, tasks babfile.TaskMap, errs *ValidationErrors) {
	for name, task := range tasks {
		for _, item := range task.Run {
			if tr, ok := item.(babfile.TaskRun); ok {
				if !tasks.Has(tr.Task) {
					errs.Add(&NotFoundError{
						Path:         path,
						Line:         tr.Line,
						TaskName:     tr.Task,
						ReferencedBy: name,
						Available:    tasks.Names(),
					})
				}
			}
		}
	}
}

func validateRunCycles(path string, tasks babfile.TaskMap, errs *ValidationErrors) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(name string, chain []string)
	dfs = func(name string, chain []string) {
		visited[name] = true
		recStack[name] = true
		chain = append(chain, name)

		task := tasks[name]
		if task == nil {
			recStack[name] = false
			return
		}
		for _, item := range task.Run {
			if tr, ok := item.(babfile.TaskRun); ok {
				if recStack[tr.Task] {
					chain = append(chain, tr.Task)
					errs.Add(&CircularError{
						Path:  path,
						Type:  "dependency",
						Chain: chain,
					})
					return
				}
				if !visited[tr.Task] && tasks.Has(tr.Task) {
					dfs(tr.Task, chain)
				}
			}
		}

		recStack[name] = false
	}

	for name := range tasks {
		if !visited[name] {
			dfs(name, nil)
		}
	}
}
