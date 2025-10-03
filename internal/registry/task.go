package registry

import (
	"fmt"
	"strings"
)

// Task represents a runnable task with commands.
type Task struct {
	Name        string
	Description string
	Commands    []string
}

// NewTask creates a new task with the given name.
func NewTask(name string) *Task {
	return &Task{
		Name:     name,
		Commands: []string{},
	}
}

// IsGrouped returns true if the task name contains a colon separator.
func (t *Task) IsGrouped() bool {
	return strings.Contains(t.Name, ":")
}

// Validate checks if the task is valid.
func (t *Task) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	if len(t.Commands) == 0 {
		return fmt.Errorf("task %s has no commands", t.Name)
	}
	return nil
}
