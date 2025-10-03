package registry

import (
	"fmt"
	"strings"
)

type Task struct {
	Name        string
	Description string
	Commands    []string
}

func NewTask(name string) *Task {
	return &Task{
		Name:     name,
		Commands: []string{},
	}
}

func (t *Task) IsGrouped() bool {
	return strings.Contains(t.Name, ":")
}

func (t *Task) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	if len(t.Commands) == 0 {
		return fmt.Errorf("task %s has no commands", t.Name)
	}
	return nil
}
