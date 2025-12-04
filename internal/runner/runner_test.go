package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bab-sh/bab/internal/parser"
)

func TestRunSimpleTask(t *testing.T) {
	tasks := parser.TaskMap{
		"hello": &parser.Task{
			Name:     "hello",
			Commands: []parser.Command{{Cmd: "echo hello"}},
		},
	}

	r := New(true)
	err := r.RunWithTasks(context.Background(), "hello", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunWithDependencies(t *testing.T) {
	tasks := parser.TaskMap{
		"build": &parser.Task{
			Name:     "build",
			Commands: []parser.Command{{Cmd: "echo building"}},
		},
		"test": &parser.Task{
			Name:         "test",
			Commands:     []parser.Command{{Cmd: "echo testing"}},
			Dependencies: []string{"build"},
		},
	}

	r := New(true)
	err := r.RunWithTasks(context.Background(), "test", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunTaskNotFound(t *testing.T) {
	tasks := parser.TaskMap{
		"hello": &parser.Task{
			Name:     "hello",
			Commands: []parser.Command{{Cmd: "echo hello"}},
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
	tasks := parser.TaskMap{
		"a": &parser.Task{
			Name:         "a",
			Commands:     []parser.Command{{Cmd: "echo a"}},
			Dependencies: []string{"b"},
		},
		"b": &parser.Task{
			Name:         "b",
			Commands:     []parser.Command{{Cmd: "echo b"}},
			Dependencies: []string{"a"},
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
	babfile := filepath.Join(tmpDir, "Babfile.yml")

	yaml := `tasks:
  hello:
    run:
      - cmd: echo "Hello"
  world:
    run:
      - cmd: echo "World"`

	if err := os.WriteFile(babfile, []byte(yaml), 0600); err != nil {
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
