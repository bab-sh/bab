package registry

import (
	"fmt"
	"sort"
	"strings"
)

type Registry interface {
	Register(task *Task) error
	Get(name string) (*Task, error)
	List() []*Task
	Tree() map[string][]*Task
}

type registry struct {
	tasks map[string]*Task
}

func New() Registry {
	return &registry{
		tasks: make(map[string]*Task),
	}
}

func (r *registry) Register(task *Task) error {
	if task == nil {
		return fmt.Errorf("cannot register nil task")
	}

	if err := task.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	if _, exists := r.tasks[task.Name]; exists {
		return fmt.Errorf("task %s already registered", task.Name)
	}

	r.tasks[task.Name] = task
	return nil
}

func (r *registry) Get(name string) (*Task, error) {
	task, exists := r.tasks[name]
	if !exists {
		return nil, fmt.Errorf("task %s not found", name)
	}
	return task, nil
}

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

func (r *registry) Tree() map[string][]*Task {
	tree := make(map[string][]*Task)

	for _, task := range r.tasks {
		if !task.IsGrouped() {
			tree[""] = append(tree[""], task)
		} else {
			group := strings.Split(task.Name, ":")[0]
			tree[group] = append(tree[group], task)
		}
	}

	for group := range tree {
		sort.Slice(tree[group], func(i, j int) bool {
			return tree[group][i].Name < tree[group][j].Name
		})
	}

	return tree
}
