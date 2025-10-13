package registry

import (
	"strings"

	baberrors "github.com/bab-sh/bab/internal/errors"
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

// GroupPath returns the group path segments for this task.
func (t *Task) GroupPath() []string {
	if !t.IsGrouped() {
		return []string{}
	}
	parts := strings.Split(t.Name, ":")
	return parts[:len(parts)-1]
}

// LeafName returns the task name without group prefixes.
func (t *Task) LeafName() string {
	if !t.IsGrouped() {
		return t.Name
	}
	parts := strings.Split(t.Name, ":")
	return parts[len(parts)-1]
}

// Validate checks if the task is valid.
func (t *Task) Validate() error {
	if t.Name == "" {
		return baberrors.ErrEmptyTaskName
	}
	if len(t.Commands) == 0 {
		return baberrors.NewTaskValidationError(t.Name, "commands", "task has no commands")
	}
	return nil
}
