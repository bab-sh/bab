package parser

import (
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr bool
		errMsg  string
		validate func(t *testing.T, tasks TaskMap)
	}{
		{
			name:    "simple task with single command",
			file:    "simple.yml",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 1 {
					t.Errorf("expected 1 task, got %d", len(tasks))
				}
				task, exists := tasks["hello"]
				if !exists {
					t.Fatal("task 'hello' not found")
				}
				if task.Name != "hello" {
					t.Errorf("expected name 'hello', got %q", task.Name)
				}
				if len(task.Commands) != 1 {
					t.Errorf("expected 1 command, got %d", len(task.Commands))
				}
				if task.Commands[0] != `echo "Hello World"` {
					t.Errorf("unexpected command: %q", task.Commands[0])
				}
			},
		},
		{
			name:    "task with multiple commands",
			file:    "multi_command.yml",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				task, exists := tasks["build"]
				if !exists {
					t.Fatal("task 'build' not found")
				}
				if task.Description != "Build the project" {
					t.Errorf("expected description 'Build the project', got %q", task.Description)
				}
				if len(task.Commands) != 3 {
					t.Fatalf("expected 3 commands, got %d", len(task.Commands))
				}
				expectedCmds := []string{
					`echo "Compiling..."`,
					`echo "Linking..."`,
					`echo "Done!"`,
				}
				for i, expected := range expectedCmds {
					if task.Commands[i] != expected {
						t.Errorf("command[%d]: expected %q, got %q", i, expected, task.Commands[i])
					}
				}
			},
		},
		{
			name:    "tasks with dependencies",
			file:    "dependencies.yml",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 3 {
					t.Errorf("expected 3 tasks, got %d", len(tasks))
				}

				clean := tasks["clean"]
				if clean == nil {
					t.Fatal("task 'clean' not found")
				}
				if len(clean.Dependencies) != 0 {
					t.Errorf("clean should have no dependencies, got %v", clean.Dependencies)
				}

				build := tasks["build"]
				if build == nil {
					t.Fatal("task 'build' not found")
				}
				if len(build.Dependencies) != 1 || build.Dependencies[0] != "clean" {
					t.Errorf("build should depend on 'clean', got %v", build.Dependencies)
				}

				test := tasks["test"]
				if test == nil {
					t.Fatal("task 'test' not found")
				}
				if len(test.Dependencies) != 1 || test.Dependencies[0] != "build" {
					t.Errorf("test should depend on 'build', got %v", test.Dependencies)
				}
			},
		},
		{
			name:    "nested tasks",
			file:    "nested.yml",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				if len(tasks) != 4 {
					t.Errorf("expected 4 tasks, got %d", len(tasks))
				}

				expectedTasks := []string{"ci:test", "ci:lint", "ci:full", "dev:start"}
				for _, name := range expectedTasks {
					if tasks[name] == nil {
						t.Errorf("expected task %q not found", name)
					}
				}

				full := tasks["ci:full"]
				if full == nil {
					t.Fatal("task 'ci:full' not found")
				}
				if len(full.Dependencies) != 2 {
					t.Errorf("ci:full should have 2 dependencies, got %d", len(full.Dependencies))
				}
			},
		},
		{
			name:    "tasks with descriptions",
			file:    "with_description.yml",
			wantErr: false,
			validate: func(t *testing.T, tasks TaskMap) {
				test := tasks["test"]
				if test == nil {
					t.Fatal("task 'test' not found")
				}
				if test.Description != "Run all unit tests" {
					t.Errorf("expected description 'Run all unit tests', got %q", test.Description)
				}

				coverage := tasks["coverage"]
				if coverage == nil {
					t.Fatal("task 'coverage' not found")
				}
				if coverage.Description != "Generate coverage report" {
					t.Errorf("expected description 'Generate coverage report', got %q", coverage.Description)
				}
			},
		},
		{
			name:    "invalid YAML syntax",
			file:    "invalid_yaml.yml",
			wantErr: true,
			errMsg:  "failed to parse YAML",
		},
		{
			name:    "empty file",
			file:    "empty.yml",
			wantErr: true,
			errMsg:  "root of Babfile must be a map",
		},
		{
			name:    "root is not a map",
			file:    "root_not_map.yml",
			wantErr: true,
			errMsg:  "root of Babfile must be a map",
		},
		{
			name:    "invalid dependency reference",
			file:    "invalid_deps.yml",
			wantErr: true,
			errMsg:  "dependency validation failed",
		},
		{
			name:    "empty command",
			file:    "empty_command.yml",
			wantErr: true,
			errMsg:  "command cannot be",
		},
		{
			name:    "empty path",
			file:    "",
			wantErr: true,
			errMsg:  "path cannot be empty",
		},
		{
			name:    "whitespace path",
			file:    "   ",
			wantErr: true,
			errMsg:  "path cannot be",
		},
		{
			name:    "nonexistent file",
			file:    "nonexistent.yml",
			wantErr: true,
			errMsg:  "failed to read Babfile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.file != "" && tt.file != "   " && tt.file != "nonexistent.yml" {
				path = filepath.Join("testdata", tt.file)
			} else {
				path = tt.file
			}

			tasks, err := Parse(path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Parse() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, tasks)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path", "/some/path/file.yml", false},
		{"valid relative path", "file.yml", false},
		{"empty path", "", true},
		{"whitespace only", "   ", true},
		{"tab only", "\t", true},
		{"newline only", "\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
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
