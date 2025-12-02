package parser

import (
	"fmt"
	"strings"
)

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path cannot be only whitespace")
	}
	return nil
}

func ValidateDependencies(tasks TaskMap) error {
	for taskName, task := range tasks {
		if len(task.Dependencies) == 0 {
			continue
		}

		for _, dep := range task.Dependencies {
			if _, exists := tasks[dep]; !exists {
				availableTasks := make([]string, 0, len(tasks))
				for name := range tasks {
					availableTasks = append(availableTasks, name)
				}
				return fmt.Errorf("task %q has invalid dependency %q (available tasks: %s)",
					taskName, dep, strings.Join(availableTasks, ", "))
			}
		}
	}
	return nil
}
