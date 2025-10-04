package tui

import "github.com/bab-sh/bab/internal/registry"

// TaskItem wraps a registry.Task for display in the TUI.
type TaskItem struct {
	task *registry.Task
}

// NewTaskItem creates a new TaskItem from a registry.Task.
func NewTaskItem(task *registry.Task) TaskItem {
	return TaskItem{task: task}
}

// FilterValue returns the string value to use for fuzzy searching.
func (t TaskItem) FilterValue() string {
	return t.task.Name
}

// Title returns the full task name.
func (t TaskItem) Title() string {
	return t.task.Name
}

// Description returns the task description.
func (t TaskItem) Description() string {
	return t.task.Description
}

// Task returns the underlying registry.Task.
func (t TaskItem) Task() *registry.Task {
	return t.task
}
