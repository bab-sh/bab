package registry

import (
	"testing"
)

func TestNew(t *testing.T) {
	reg := New()
	if reg == nil {
		t.Fatal("New() returned nil")
	}

	tasks := reg.List()
	if len(tasks) != 0 {
		t.Errorf("expected empty registry, got %d tasks", len(tasks))
	}
}

func TestRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		task    *Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "register valid task",
			task: &Task{
				Name:     "build",
				Commands: []string{"go build"},
			},
			wantErr: false,
		},
		{
			name:    "register nil task",
			task:    nil,
			wantErr: true,
			errMsg:  "cannot register nil task",
		},
		{
			name: "register invalid task with empty name",
			task: &Task{
				Name:     "",
				Commands: []string{"go build"},
			},
			wantErr: true,
			errMsg:  "invalid task: task name cannot be empty",
		},
		{
			name: "register invalid task with no commands",
			task: &Task{
				Name:     "build",
				Commands: []string{},
			},
			wantErr: true,
			errMsg:  "invalid task: task build has no commands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := New()
			err := reg.Register(tt.task)

			if tt.wantErr {
				if err == nil {
					t.Error("Register() expected error, got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Register() error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("Register() unexpected error: %v", err)
			}
		})
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	reg := New()
	task := &Task{
		Name:     "build",
		Commands: []string{"go build"},
	}

	// First registration should succeed
	if err := reg.Register(task); err != nil {
		t.Fatalf("first Register() failed: %v", err)
	}

	// Second registration should fail
	err := reg.Register(task)
	if err == nil {
		t.Fatal("Register() expected error for duplicate task, got nil")
	}

	expected := "task build already registered"
	if err.Error() != expected {
		t.Errorf("Register() error = %q, want %q", err.Error(), expected)
	}
}

func TestRegistry_Get(t *testing.T) {
	reg := New()
	task := &Task{
		Name:        "build",
		Description: "Build the project",
		Commands:    []string{"go build"},
	}

	if err := reg.Register(task); err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	// Test successful retrieval
	t.Run("get existing task", func(t *testing.T) {
		retrieved, err := reg.Get("build")
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if retrieved == nil {
			t.Fatal("Get() returned nil task")
		}
		if retrieved.Name != task.Name {
			t.Errorf("Get() name = %q, want %q", retrieved.Name, task.Name)
		}
		if retrieved.Description != task.Description {
			t.Errorf("Get() description = %q, want %q", retrieved.Description, task.Description)
		}
	})

	// Test non-existent task
	t.Run("get non-existent task", func(t *testing.T) {
		retrieved, err := reg.Get("nonexistent")
		if err == nil {
			t.Fatal("Get() expected error for non-existent task, got nil")
		}
		if retrieved != nil {
			t.Error("Get() expected nil task for non-existent task")
		}

		expected := "task nonexistent not found"
		if err.Error() != expected {
			t.Errorf("Get() error = %q, want %q", err.Error(), expected)
		}
	})
}

func TestRegistry_List(t *testing.T) {
	reg := New()

	// Test empty registry
	t.Run("empty registry", func(t *testing.T) {
		tasks := reg.List()
		if len(tasks) != 0 {
			t.Errorf("List() expected 0 tasks, got %d", len(tasks))
		}
	})

	// Add tasks
	tasks := []*Task{
		{Name: "test", Commands: []string{"go test"}},
		{Name: "build", Commands: []string{"go build"}},
		{Name: "lint", Commands: []string{"golangci-lint run"}},
	}

	for _, task := range tasks {
		if err := reg.Register(task); err != nil {
			t.Fatalf("Register() failed: %v", err)
		}
	}

	// Test list returns all tasks sorted
	t.Run("list with tasks", func(t *testing.T) {
		list := reg.List()
		if len(list) != len(tasks) {
			t.Errorf("List() expected %d tasks, got %d", len(tasks), len(list))
		}

		// Verify sorted order (alphabetically)
		expected := []string{"build", "lint", "test"}
		for i, task := range list {
			if task.Name != expected[i] {
				t.Errorf("List()[%d] name = %q, want %q", i, task.Name, expected[i])
			}
		}
	})
}

func TestRegistry_Tree(t *testing.T) {
	reg := New()

	tasks := []*Task{
		{Name: "build", Commands: []string{"go build"}},
		{Name: "test", Commands: []string{"go test"}},
		{Name: "dev:start", Commands: []string{"npm run dev"}},
		{Name: "dev:watch", Commands: []string{"npm run watch"}},
		{Name: "prod:deploy", Commands: []string{"./deploy.sh"}},
	}

	for _, task := range tasks {
		if err := reg.Register(task); err != nil {
			t.Fatalf("Register() failed: %v", err)
		}
	}

	tree := reg.Tree()

	// Test root tasks
	t.Run("root tasks", func(t *testing.T) {
		rootTasks, exists := tree[""]
		if !exists {
			t.Fatal("Tree() missing root tasks")
		}
		if len(rootTasks) != 2 {
			t.Errorf("Tree() expected 2 root tasks, got %d", len(rootTasks))
		}

		// Verify sorted order
		expected := []string{"build", "test"}
		for i, task := range rootTasks {
			if task.Name != expected[i] {
				t.Errorf("Tree()[%q][%d] name = %q, want %q", "", i, task.Name, expected[i])
			}
		}
	})

	// Test grouped tasks
	t.Run("dev group", func(t *testing.T) {
		devTasks, exists := tree["dev"]
		if !exists {
			t.Fatal("Tree() missing 'dev' group")
		}
		if len(devTasks) != 2 {
			t.Errorf("Tree() expected 2 dev tasks, got %d", len(devTasks))
		}

		expected := []string{"dev:start", "dev:watch"}
		for i, task := range devTasks {
			if task.Name != expected[i] {
				t.Errorf("Tree()[%q][%d] name = %q, want %q", "dev", i, task.Name, expected[i])
			}
		}
	})

	t.Run("prod group", func(t *testing.T) {
		prodTasks, exists := tree["prod"]
		if !exists {
			t.Fatal("Tree() missing 'prod' group")
		}
		if len(prodTasks) != 1 {
			t.Errorf("Tree() expected 1 prod task, got %d", len(prodTasks))
		}
		if prodTasks[0].Name != "prod:deploy" {
			t.Errorf("Tree()[%q][0] name = %q, want %q", "prod", prodTasks[0].Name, "prod:deploy")
		}
	})
}
