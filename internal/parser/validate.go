package parser

func validateDependencies(tasks TaskMap) error {
	for name, task := range tasks {
		for _, dep := range task.Dependencies {
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

func validateRunTaskRefs(tasks TaskMap) error {
	for name, task := range tasks {
		for _, item := range task.RunItems {
			if tr, ok := item.(TaskRun); ok {
				if !tasks.Has(tr.TaskRef) {
					return &NotFoundError{
						TaskName:     tr.TaskRef,
						ReferencedBy: name,
						Available:    tasks.Names(),
					}
				}
			}
		}
	}
	return nil
}

func validateRunCycles(tasks TaskMap) error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(name string, chain []string) error
	dfs = func(name string, chain []string) error {
		visited[name] = true
		recStack[name] = true
		chain = append(chain, name)

		task := tasks[name]
		for _, item := range task.RunItems {
			if tr, ok := item.(TaskRun); ok {
				if recStack[tr.TaskRef] {
					chain = append(chain, tr.TaskRef)
					return &CircularError{
						Type:  "task run",
						Chain: chain,
					}
				}
				if !visited[tr.TaskRef] {
					if err := dfs(tr.TaskRef, chain); err != nil {
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
