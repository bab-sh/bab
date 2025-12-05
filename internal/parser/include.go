package parser

import (
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

func resolveInclude(namespace, babfilePath, baseDir string, tasks TaskMap, visited map[string]bool) error {
	incPath := babfilePath
	if !filepath.IsAbs(incPath) {
		incPath = filepath.Join(baseDir, incPath)
	}
	incPath = filepath.Clean(incPath)

	log.Debug("Resolving include", "namespace", namespace, "path", incPath)

	includedTasks, err := parseFile(incPath, visited)
	if err != nil {
		return err
	}

	for name, task := range includedTasks {
		prefixedName := namespace + ":" + name
		if tasks.Has(prefixedName) {
			return &ParseError{Path: incPath, Message: "task name collision: " + prefixedName}
		}
		tasks[prefixedName] = &Task{
			Name:         prefixedName,
			Description:  task.Description,
			RunItems:     prefixTaskRuns(task.RunItems, namespace),
			Dependencies: prefixDeps(task.Dependencies, namespace),
		}
	}

	return nil
}

func prefixDeps(deps []string, namespace string) []string {
	if len(deps) == 0 {
		return deps
	}
	prefixed := make([]string, len(deps))
	for i, dep := range deps {
		if !strings.Contains(dep, ":") {
			prefixed[i] = namespace + ":" + dep
		} else {
			prefixed[i] = dep
		}
	}
	return prefixed
}

func prefixTaskRuns(items []RunItem, namespace string) []RunItem {
	if len(items) == 0 {
		return items
	}
	prefixed := make([]RunItem, len(items))
	for i, item := range items {
		switch v := item.(type) {
		case CommandRun:
			prefixed[i] = v
		case TaskRun:
			if !strings.Contains(v.TaskRef, ":") {
				prefixed[i] = TaskRun{
					TaskRef:   namespace + ":" + v.TaskRef,
					Platforms: v.Platforms,
				}
			} else {
				prefixed[i] = v
			}
		}
	}
	return prefixed
}
