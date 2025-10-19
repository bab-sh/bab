package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bab-sh/bab/internal/parser"
)

func TestContainsString(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{"found in slice", []string{"a", "b", "c"}, "b", true},
		{"not found in slice", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
		{"empty string in slice", []string{"", "b"}, "", true},
		{"empty string not in slice", []string{"a", "b"}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsString(tt.slice, tt.item)
			if got != tt.want {
				t.Errorf("containsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecuteTask(t *testing.T) {
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
			babfileYAML: `hello:
  run: echo "Hello World"`,
			wantErr: false,
		},
		{
			name:     "execute task with dependencies",
			taskName: "test",
			babfileYAML: `build:
  run: echo "Building"
test:
  deps: build
  run: echo "Testing"`,
			wantErr: false,
		},
		{
			name:     "task not found",
			taskName: "nonexistent",
			babfileYAML: `hello:
  run: echo "Hello"`,
			wantErr: true,
			errMsg:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			babfilePath := filepath.Join(tmpDir, "Babfile")

			if err := os.WriteFile(babfilePath, []byte(tt.babfileYAML), 0644); err != nil {
				t.Fatalf("failed to create test Babfile: %v", err)
			}

			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}

			oldDryRun := dryRun
			dryRun = true
			defer func() { dryRun = oldDryRun }()

			ctx := context.Background()
			err := executeTask(ctx, tt.taskName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("executeTask() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("executeTask() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("executeTask() unexpected error: %v", err)
			}
		})
	}
}

func TestExecuteTaskWithDeps(t *testing.T) {
	tests := []struct {
		name      string
		taskName  string
		tasks     parser.TaskMap
		wantErr   bool
		errMsg    string
		executed  map[string]bool
		executing map[string]bool
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
			wantErr:   false,
			executed:  make(map[string]bool),
			executing: make(map[string]bool),
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
			wantErr:   false,
			executed:  make(map[string]bool),
			executing: make(map[string]bool),
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
			wantErr:   false,
			executed:  make(map[string]bool),
			executing: make(map[string]bool),
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
			wantErr:   true,
			errMsg:    "not found",
			executed:  make(map[string]bool),
			executing: make(map[string]bool),
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
			wantErr:   true,
			errMsg:    "circular dependency",
			executed:  make(map[string]bool),
			executing: make(map[string]bool),
		},
		{
			name:     "skip already executed task",
			taskName: "hello",
			tasks: parser.TaskMap{
				"hello": &parser.Task{
					Name:     "hello",
					Commands: []string{"echo hello"},
				},
			},
			wantErr: false,
			executed: map[string]bool{
				"hello": true,
			},
			executing: make(map[string]bool),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldDryRun := dryRun
			dryRun = true
			defer func() { dryRun = oldDryRun }()

			ctx := context.Background()
			err := executeTaskWithDeps(ctx, tt.taskName, tt.tasks, tt.executed, tt.executing)

			if tt.wantErr {
				if err == nil {
					t.Errorf("executeTaskWithDeps() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("executeTaskWithDeps() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("executeTaskWithDeps() unexpected error: %v", err)
			}
		})
	}
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
			got := buildDependencyChain(tt.currentTask, tt.executing, tt.tasks)
			if got != tt.wantChain {
				t.Errorf("buildDependencyChain() = %q, want %q", got, tt.wantChain)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
