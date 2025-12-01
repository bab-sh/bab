package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bab-sh/bab/internal/parser"
)

func TestRunner_Run(t *testing.T) {
	tests := []struct {
		name        string
		taskName    string
		babfileYAML string
		wantErr     bool
		errMsg      string
	}{
		{
			name:     "execute simple task",
			taskName: "hello",
			babfileYAML: `tasks:
  hello:
    run: echo "Hello World"`,
			wantErr: false,
		},
		{
			name:     "execute task with dependencies",
			taskName: "test",
			babfileYAML: `tasks:
  build:
    run: echo "Building"
  test:
    deps: build
    run: echo "Testing"`,
			wantErr: false,
		},
		{
			name:     "task not found",
			taskName: "nonexistent",
			babfileYAML: `tasks:
  hello:
    run: echo "Hello"`,
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			babfilePath := filepath.Join(tmpDir, "Babfile")

			if err := os.WriteFile(babfilePath, []byte(tt.babfileYAML), 0600); err != nil {
				t.Fatalf("failed to create test Babfile: %v", err)
			}

			oldDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(oldDir) }()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}

			r := New(true)
			ctx := context.Background()
			err := r.Run(ctx, tt.taskName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Run() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Run() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Run() unexpected error: %v", err)
			}
		})
	}
}

func TestRunner_RunWithTasks(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		tasks    parser.TaskMap
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "execute task with no dependencies",
			taskName: "hello",
			tasks: parser.TaskMap{
				"hello": &parser.Task{
					Name:     "hello",
					Commands: []string{"echo hello"},
				},
			},
			wantErr: false,
		},
		{
			name:     "execute task with one dependency",
			taskName: "test",
			tasks: parser.TaskMap{
				"build": &parser.Task{
					Name:     "build",
					Commands: []string{"echo building"},
				},
				"test": &parser.Task{
					Name:         "test",
					Commands:     []string{"echo testing"},
					Dependencies: []string{"build"},
				},
			},
			wantErr: false,
		},
		{
			name:     "execute task with multiple dependencies",
			taskName: "deploy",
			tasks: parser.TaskMap{
				"build": &parser.Task{
					Name:     "build",
					Commands: []string{"echo building"},
				},
				"test": &parser.Task{
					Name:     "test",
					Commands: []string{"echo testing"},
				},
				"deploy": &parser.Task{
					Name:         "deploy",
					Commands:     []string{"echo deploying"},
					Dependencies: []string{"build", "test"},
				},
			},
			wantErr: false,
		},
		{
			name:     "task not found",
			taskName: "nonexistent",
			tasks: parser.TaskMap{
				"hello": &parser.Task{
					Name:     "hello",
					Commands: []string{"echo hello"},
				},
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name:     "circular dependency detected",
			taskName: "task_a",
			tasks: parser.TaskMap{
				"task_a": &parser.Task{
					Name:         "task_a",
					Commands:     []string{"echo a"},
					Dependencies: []string{"task_b"},
				},
				"task_b": &parser.Task{
					Name:         "task_b",
					Commands:     []string{"echo b"},
					Dependencies: []string{"task_a"},
				},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(true)
			ctx := context.Background()
			err := r.RunWithTasks(ctx, tt.taskName, tt.tasks)

			if tt.wantErr {
				if err == nil {
					t.Errorf("RunWithTasks() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("RunWithTasks() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("RunWithTasks() unexpected error: %v", err)
			}
		})
	}
}

func TestLoadTasks(t *testing.T) {
	tests := []struct {
		name        string
		babfileYAML string
		createFile  bool
		wantErr     bool
		wantCount   int
	}{
		{
			name: "load simple tasks",
			babfileYAML: `tasks:
  hello:
    run: echo "Hello"
  world:
    run: echo "World"`,
			createFile: true,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name: "load nested tasks",
			babfileYAML: `tasks:
  ci:
    test:
      run: echo "Testing"
    lint:
      run: echo "Linting"`,
			createFile: true,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:        "no babfile",
			babfileYAML: "",
			createFile:  false,
			wantErr:     true,
			wantCount:   0,
		},
		{
			name:        "empty babfile",
			babfileYAML: "",
			createFile:  true,
			wantErr:     true,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.createFile {
				babfilePath := filepath.Join(tmpDir, "Babfile")
				if err := os.WriteFile(babfilePath, []byte(tt.babfileYAML), 0600); err != nil {
					t.Fatalf("failed to create test Babfile: %v", err)
				}
			}

			oldDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(oldDir) }()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}

			tasks, err := LoadTasks()

			if tt.wantErr {
				if err == nil {
					t.Error("LoadTasks() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("LoadTasks() unexpected error: %v", err)
				return
			}

			if len(tasks) != tt.wantCount {
				t.Errorf("LoadTasks() returned %d tasks, want %d", len(tasks), tt.wantCount)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("dry run false", func(t *testing.T) {
		r := New(false)
		if r.DryRun {
			t.Error("DryRun should be false")
		}
	})

	t.Run("dry run true", func(t *testing.T) {
		r := New(true)
		if !r.DryRun {
			t.Error("DryRun should be true")
		}
	})
}

func TestBuildDependencyChain(t *testing.T) {
	tests := []struct {
		name        string
		currentTask string
		executing   map[string]bool
		tasks       parser.TaskMap
		wantChain   string
	}{
		{
			name:        "simple circular dependency",
			currentTask: "task_a",
			executing: map[string]bool{
				"task_a": true,
				"task_b": true,
			},
			tasks: parser.TaskMap{
				"task_a": &parser.Task{
					Name:         "task_a",
					Commands:     []string{"echo a"},
					Dependencies: []string{"task_b"},
				},
				"task_b": &parser.Task{
					Name:         "task_b",
					Commands:     []string{"echo b"},
					Dependencies: []string{"task_a"},
				},
			},
			wantChain: "task_a → task_b",
		},
		{
			name:        "no circular dependency",
			currentTask: "task_a",
			executing: map[string]bool{
				"task_a": true,
			},
			tasks: parser.TaskMap{
				"task_a": &parser.Task{
					Name:     "task_a",
					Commands: []string{"echo a"},
				},
			},
			wantChain: "task_a",
		},
		{
			name:        "longer chain",
			currentTask: "task_a",
			executing: map[string]bool{
				"task_a": true,
				"task_b": true,
				"task_c": true,
			},
			tasks: parser.TaskMap{
				"task_a": &parser.Task{
					Name:         "task_a",
					Commands:     []string{"echo a"},
					Dependencies: []string{"task_b"},
				},
				"task_b": &parser.Task{
					Name:         "task_b",
					Commands:     []string{"echo b"},
					Dependencies: []string{"task_c"},
				},
				"task_c": &parser.Task{
					Name:         "task_c",
					Commands:     []string{"echo c"},
					Dependencies: []string{"task_a"},
				},
			},
			wantChain: "task_a → task_b → task_c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildDependencyChain(tt.currentTask, tt.executing, tt.tasks)
			if got != tt.wantChain {
				t.Errorf("BuildDependencyChain() = %q, want %q", got, tt.wantChain)
			}
		})
	}
}
