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

	if tree == nil {
		t.Fatal("Tree() returned nil")
	}

	if len(tree.Children) != 4 {
		t.Errorf("Tree() root expected 4 children, got %d", len(tree.Children))
	}

	// Test root tasks
	t.Run("root tasks", func(t *testing.T) {
		buildNode, exists := tree.Children["build"]
		if !exists {
			t.Fatal("Tree() missing 'build' task")
		}
		if !buildNode.IsTask() {
			t.Error("'build' should be a task node")
		}
		if buildNode.Task == nil || buildNode.Task.Name != "build" {
			t.Error("'build' node has incorrect task")
		}

		testNode, exists := tree.Children["test"]
		if !exists {
			t.Fatal("Tree() missing 'test' task")
		}
		if !testNode.IsTask() {
			t.Error("'test' should be a task node")
		}
	})

	// Test grouped tasks
	t.Run("dev group", func(t *testing.T) {
		devNode, exists := tree.Children["dev"]
		if !exists {
			t.Fatal("Tree() missing 'dev' group")
		}
		if !devNode.IsGroup() {
			t.Fatal("'dev' should be a group node")
		}

		if len(devNode.Children) != 2 {
			t.Errorf("'dev' group expected 2 children, got %d", len(devNode.Children))
		}

		startNode, exists := devNode.Children["start"]
		if !exists {
			t.Fatal("'dev' group missing 'start' task")
		}
		if !startNode.IsTask() {
			t.Error("'dev:start' should be a task node")
		}

		watchNode, exists := devNode.Children["watch"]
		if !exists {
			t.Fatal("'dev' group missing 'watch' task")
		}
		if !watchNode.IsTask() {
			t.Error("'dev:watch' should be a task node")
		}
	})

	t.Run("prod group", func(t *testing.T) {
		prodNode, exists := tree.Children["prod"]
		if !exists {
			t.Fatal("Tree() missing 'prod' group")
		}
		if !prodNode.IsGroup() {
			t.Fatal("'prod' should be a group node")
		}

		if len(prodNode.Children) != 1 {
			t.Errorf("'prod' group expected 1 child, got %d", len(prodNode.Children))
		}

		deployNode, exists := prodNode.Children["deploy"]
		if !exists {
			t.Fatal("'prod' group missing 'deploy' task")
		}
		if !deployNode.IsTask() {
			t.Error("'prod:deploy' should be a task node")
		}
	})
}

func TestRegistry_TreeDeepNesting(t *testing.T) {
	reg := New()

	tasks := []*Task{
		{Name: "build:platforms:linux:amd64", Commands: []string{"build linux amd64"}},
		{Name: "build:platforms:linux:arm64", Commands: []string{"build linux arm64"}},
		{Name: "build:platforms:windows:amd64", Commands: []string{"build windows amd64"}},
		{Name: "build:platforms:macos:arm64", Commands: []string{"build macos arm64"}},
		{Name: "test:unit:fast", Commands: []string{"test unit fast"}},
		{Name: "test:unit:slow", Commands: []string{"test unit slow"}},
		{Name: "test:integration:api", Commands: []string{"test integration api"}},
	}

	for _, task := range tasks {
		if err := reg.Register(task); err != nil {
			t.Fatalf("Register() failed: %v", err)
		}
	}

	tree := reg.Tree()

	// Test build hierarchy: build -> platforms -> (linux/windows/macos) -> (amd64/arm64)
	t.Run("deep nesting - build platforms", func(t *testing.T) {
		buildNode, exists := tree.Children["build"]
		if !exists {
			t.Fatal("Tree() missing 'build' group")
		}
		if !buildNode.IsGroup() {
			t.Fatal("'build' should be a group node")
		}

		platformsNode, exists := buildNode.Children["platforms"]
		if !exists {
			t.Fatal("'build' missing 'platforms' group")
		}
		if !platformsNode.IsGroup() {
			t.Fatal("'platforms' should be a group node")
		}

		// Check linux group
		linuxNode, exists := platformsNode.Children["linux"]
		if !exists {
			t.Fatal("'platforms' missing 'linux' group")
		}
		if !linuxNode.IsGroup() {
			t.Fatal("'linux' should be a group node")
		}
		if len(linuxNode.Children) != 2 {
			t.Errorf("'linux' expected 2 children, got %d", len(linuxNode.Children))
		}

		// Check linux:amd64 task
		amd64Node, exists := linuxNode.Children["amd64"]
		if !exists {
			t.Fatal("'linux' missing 'amd64' task")
		}
		if !amd64Node.IsTask() {
			t.Error("'linux:amd64' should be a task node")
		}

		// Check linux:arm64 task
		arm64Node, exists := linuxNode.Children["arm64"]
		if !exists {
			t.Fatal("'linux' missing 'arm64' task")
		}
		if !arm64Node.IsTask() {
			t.Error("'linux:arm64' should be a task node")
		}
	})

	// Test hierarchy: test -> (unit/integration) -> tasks
	t.Run("deep nesting - test groups", func(t *testing.T) {
		testNode, exists := tree.Children["test"]
		if !exists {
			t.Fatal("Tree() missing 'test' group")
		}

		unitNode, exists := testNode.Children["unit"]
		if !exists {
			t.Fatal("'test' missing 'unit' group")
		}
		if len(unitNode.Children) != 2 {
			t.Errorf("'unit' expected 2 children, got %d", len(unitNode.Children))
		}

		integrationNode, exists := testNode.Children["integration"]
		if !exists {
			t.Fatal("'test' missing 'integration' group")
		}
		if len(integrationNode.Children) != 1 {
			t.Errorf("'integration' expected 1 child, got %d", len(integrationNode.Children))
		}
	})
}
