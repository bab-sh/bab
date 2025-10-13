// Package registry manages task registration and retrieval.
package registry

import (
	"sort"

	baberrors "github.com/bab-sh/bab/internal/errors"
)

// TreeNode represents a node in the hierarchical task tree.
type TreeNode struct {
	Name     string               // Name of this node (group or task name)
	Task     *Task                // The task (nil if this is a group node)
	Children map[string]*TreeNode // Child nodes (groups and tasks)
}

// NewTreeNode creates a new tree node.
func NewTreeNode(name string) *TreeNode {
	return &TreeNode{
		Name:     name,
		Children: make(map[string]*TreeNode),
	}
}

// IsTask returns true if this node represents a task (leaf node).
func (n *TreeNode) IsTask() bool {
	return n.Task != nil
}

// Registry manages task registration and retrieval.
type Registry interface {
	Register(task *Task) error
	Get(name string) (*Task, error)
	List() []*Task
	Tree() *TreeNode
}

type registry struct {
	tasks map[string]*Task
}

// New creates a new task registry.
func New() Registry {
	return &registry{
		tasks: make(map[string]*Task),
	}
}

// Register adds a task to the registry.
func (r *registry) Register(task *Task) error {
	if task == nil {
		return baberrors.NewTaskValidationError("", "task", "cannot register nil task")
	}

	if err := task.Validate(); err != nil {
		return err
	}

	if _, exists := r.tasks[task.Name]; exists {
		return baberrors.ErrTaskAlreadyExists
	}

	r.tasks[task.Name] = task
	return nil
}

// Get retrieves a task by name from the registry.
func (r *registry) Get(name string) (*Task, error) {
	task, exists := r.tasks[name]
	if !exists {
		return nil, baberrors.NewTaskNotFoundError(name)
	}
	return task, nil
}

// List returns all tasks sorted by name.
func (r *registry) List() []*Task {
	tasks := make([]*Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Name < tasks[j].Name
	})

	return tasks
}

// Tree organizes tasks into a hierarchical tree structure.
func (r *registry) Tree() *TreeNode {
	root := NewTreeNode("")

	// Sort tasks by name for consistent ordering
	tasks := r.List()

	for _, task := range tasks {
		insertTaskIntoTree(root, task)
	}

	return root
}

// insertTaskIntoTree inserts a task into the tree at the appropriate position.
func insertTaskIntoTree(root *TreeNode, task *Task) {
	// get the group path and leaf name
	groupPath := task.GroupPath()
	leafName := task.LeafName()

	// navigate/create the path to the task
	current := root
	for _, segment := range groupPath {
		if _, exists := current.Children[segment]; !exists {
			current.Children[segment] = NewTreeNode(segment)
		}
		current = current.Children[segment]
	}

	taskNode := NewTreeNode(leafName)
	taskNode.Task = task
	current.Children[leafName] = taskNode
}
