package parser

import (
	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/errs"
)

func validateAll(path string, tasks babfile.TaskMap) error {
	verrs := &errs.ValidationErrors{}
	validateDependencies(path, tasks, verrs)
	validateRunTaskRefs(path, tasks, verrs)
	validateRunCycles(path, tasks, verrs)
	validateAliases(path, tasks, verrs)
	return verrs.OrNil()
}

func validateAliases(path string, tasks babfile.TaskMap, verrs *errs.ValidationErrors) {
	type aliasInfo struct {
		taskName string
		line     int
	}
	aliasToTask := make(map[string]aliasInfo)

	for name, task := range tasks {
		for _, alias := range task.GetAllAliases() {
			if alias == "" {
				continue
			}

			if tasks.Has(alias) {
				verrs.Add(&errs.AliasConflictError{
					Path:     path,
					Line:     task.Line,
					Alias:    alias,
					TaskName: name,
				})
				continue
			}

			if existing, exists := aliasToTask[alias]; exists {
				verrs.Add(&errs.DuplicateAliasError{
					Path:         path,
					Line:         task.Line,
					Alias:        alias,
					TaskName:     name,
					OriginalTask: existing.taskName,
					OriginalLine: existing.line,
				})
				continue
			}

			aliasToTask[alias] = aliasInfo{taskName: name, line: task.Line}
		}
	}
}

func validateDependencies(path string, tasks babfile.TaskMap, verrs *errs.ValidationErrors) {
	for name, task := range tasks {
		for _, dep := range task.Deps {
			if !tasks.Has(dep) {
				line := task.DepsLine
				if line == 0 {
					line = task.Line
				}
				verrs.Add(&errs.TaskNotFoundError{
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

func validateRunTaskRefs(path string, tasks babfile.TaskMap, verrs *errs.ValidationErrors) {
	for name, task := range tasks {
		for _, item := range task.Run {
			if tr, ok := item.(babfile.TaskRun); ok {
				if !tasks.Has(tr.Task) {
					verrs.Add(&errs.TaskNotFoundError{
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

func validateRunCycles(path string, tasks babfile.TaskMap, verrs *errs.ValidationErrors) {
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
					verrs.Add(&errs.CircularDepError{
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
