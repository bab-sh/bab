package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bab-sh/bab/internal/babfile"
)

func TestRunSimpleTask(t *testing.T) {
	tasks := babfile.TaskMap{
		"hello": &babfile.Task{
			Name: "hello",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo hello"}},
		},
	}

	r := New(true)
	err := r.RunWithTasks(context.Background(), "hello", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunWithDependencies(t *testing.T) {
	tasks := babfile.TaskMap{
		"build": &babfile.Task{
			Name: "build",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo building"}},
		},
		"test": &babfile.Task{
			Name: "test",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo testing"}},
			Deps: []string{"build"},
		},
	}

	r := New(true)
	err := r.RunWithTasks(context.Background(), "test", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunTaskNotFound(t *testing.T) {
	tasks := babfile.TaskMap{
		"hello": &babfile.Task{
			Name: "hello",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo hello"}},
		},
	}

	r := New(true)
	err := r.RunWithTasks(context.Background(), "nonexistent", tasks)
	if err == nil {
		t.Fatal("expected error for nonexistent task")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestRunCircularDependency(t *testing.T) {
	tasks := babfile.TaskMap{
		"a": &babfile.Task{
			Name: "a",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo a"}},
			Deps: []string{"b"},
		},
		"b": &babfile.Task{
			Name: "b",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo b"}},
			Deps: []string{"a"},
		},
	}

	r := New(true)
	err := r.RunWithTasks(context.Background(), "a", tasks)
	if err == nil {
		t.Fatal("expected error for circular dependency")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected 'circular' error, got: %v", err)
	}
}

func TestLoadTasks(t *testing.T) {
	tmpDir := t.TempDir()
	babfilePath := filepath.Join(tmpDir, "Babfile.yml")

	yaml := `tasks:
  hello:
    run:
      - cmd: echo "Hello"
  world:
    run:
      - cmd: echo "World"`

	if err := os.WriteFile(babfilePath, []byte(yaml), 0600); err != nil {
		t.Fatalf("failed to create Babfile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	tasks, err := LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestNew(t *testing.T) {
	r := New(false)
	if r.DryRun {
		t.Error("DryRun should be false")
	}

	r = New(true)
	if !r.DryRun {
		t.Error("DryRun should be true")
	}
}

func TestRunTaskWithTaskRef(t *testing.T) {
	tasks := babfile.TaskMap{
		"main": &babfile.Task{
			Name: "main",
			Run: []babfile.RunItem{
				babfile.CommandRun{Cmd: "echo main"},
				babfile.TaskRun{Task: "helper"},
			},
		},
		"helper": &babfile.Task{
			Name: "helper",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo helper"}},
		},
	}

	r := New(true)
	err := r.RunWithTasks(context.Background(), "main", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunTaskRefCircular(t *testing.T) {
	tasks := babfile.TaskMap{
		"a": &babfile.Task{
			Name: "a",
			Run:  []babfile.RunItem{babfile.TaskRun{Task: "b"}},
		},
		"b": &babfile.Task{
			Name: "b",
			Run:  []babfile.RunItem{babfile.TaskRun{Task: "a"}},
		},
	}

	r := New(false)
	err := r.RunWithTasks(context.Background(), "a", tasks)
	if err == nil {
		t.Fatal("expected error for circular task run")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected 'circular' error, got: %v", err)
	}
}
