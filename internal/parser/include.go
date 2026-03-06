package parser

import (
	"path/filepath"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/errs"
	"github.com/charmbracelet/log"
)

func resolveInclude(namespace, babfilePath, baseDir string, tasks babfile.TaskMap, visited map[string]bool) error {
	incPath := babfilePath
	if !filepath.IsAbs(incPath) {
		incPath = filepath.Join(baseDir, incPath)
	}
	incPath = filepath.Clean(incPath)

	log.Debug("Resolving include", "namespace", namespace, "path", incPath)

	result, err := parseFile(incPath, visited)
	if err != nil {
		return err
	}

	for name, task := range result.Tasks {
		prefixedName := namespace + ":" + name
		if tasks.Has(prefixedName) {
			return &errs.ParseError{Path: incPath, Message: "task name collision: " + prefixedName}
		}

		taskVars := babfile.MergeVarMaps(result.GlobalVars, task.Vars)
		taskEnv := babfile.MergeEnvMaps(result.GlobalEnv, task.Env)

		taskDir := task.Dir
		if taskDir == "" {
			taskDir = result.GlobalDir
		}

		taskSilent := task.Silent
		if taskSilent == nil {
			taskSilent = result.GlobalSilent
		}

		taskOutput := task.Output
		if taskOutput == nil {
			taskOutput = result.GlobalOutput
		}

		tasks[prefixedName] = &babfile.Task{
			Name:       prefixedName,
			Line:       task.Line,
			DepsLine:   task.DepsLine,
			SourcePath: task.SourcePath,
			Desc:       task.Desc,
			Alias:      prefixAlias(task.Alias, namespace),
			Aliases:    prefixAliases(task.Aliases, namespace),
			Vars:       taskVars,
			Env:        taskEnv,
			Silent:     taskSilent,
			Output:     taskOutput,
			Dir:        taskDir,
			When:       task.When,
			Deps:       prefixDeps(task.Deps, namespace),
			Run:        prefixTaskRuns(task.Run, namespace),
		}
	}

	return nil
}

func prefixAlias(alias, namespace string) string {
	if alias == "" {
		return ""
	}
	return namespace + ":" + alias
}

func prefixAliases(aliases []string, namespace string) []string {
	if len(aliases) == 0 {
		return nil
	}
	prefixed := make([]string, len(aliases))
	for i, alias := range aliases {
		prefixed[i] = namespace + ":" + alias
	}
	return prefixed
}

func prefixDeps(deps []string, namespace string) []string {
	if len(deps) == 0 {
		return nil
	}
	prefixed := make([]string, len(deps))
	for i, dep := range deps {
		prefixed[i] = namespace + ":" + dep
	}
	return prefixed
}

func prefixTaskRuns(items []babfile.RunItem, namespace string) []babfile.RunItem {
	prefixed := make([]babfile.RunItem, len(items))
	for i, item := range items {
		switch v := item.(type) {
		case babfile.CommandRun:
			prefixed[i] = v
		case babfile.TaskRun:
			prefixed[i] = babfile.TaskRun{
				Line:      v.Line,
				Task:      namespace + ":" + v.Task,
				Silent:    v.Silent,
				Output:    v.Output,
				Platforms: v.Platforms,
				When:      v.When,
			}
		case babfile.LogRun:
			prefixed[i] = v
		case babfile.PromptRun:
			prefixed[i] = v
		}
	}
	return prefixed
}
