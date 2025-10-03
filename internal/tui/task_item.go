package tui

import (
	"strings"

	"github.com/bab-sh/bab/internal/registry"
)

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

// IsGrouped returns true if the task is part of a group (contains a colon).
func (t TaskItem) IsGrouped() bool {
	return strings.Contains(t.task.Name, ":")
}

// GroupName returns the group portion of a grouped task name.
func (t TaskItem) GroupName() string {
	if !t.IsGrouped() {
		return ""
	}
	parts := strings.SplitN(t.task.Name, ":", 2)
	return parts[0]
}

// ShortName returns the task name without the group prefix.
func (t TaskItem) ShortName() string {
	if !t.IsGrouped() {
		return t.task.Name
	}
	parts := strings.SplitN(t.task.Name, ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return t.task.Name
}
