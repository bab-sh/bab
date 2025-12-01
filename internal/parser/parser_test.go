package parser

import (
	"path/filepath"
	"testing"
)

const testBuildTask = "build"

func TestParseSimpleTask(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "simple.yml"))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
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
}

func TestParseMultiCommand(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "multi_command.yml"))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	task, exists := tasks[testBuildTask]
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
}

func TestParseTasksWithDependencies(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "dependencies.yml"))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}

	clean := tasks["clean"]
	if clean == nil {
		t.Fatal("task 'clean' not found")
		return
	}
	if len(clean.Dependencies) != 0 {
		t.Errorf("clean should have no dependencies, got %v", clean.Dependencies)
	}

	build := tasks[testBuildTask]
	if build == nil {
		t.Fatal("task 'build' not found")
		return
	}
	if len(build.Dependencies) != 1 || build.Dependencies[0] != "clean" {
		t.Errorf("build should depend on 'clean', got %v", build.Dependencies)
	}

	test := tasks["test"]
	if test == nil {
		t.Fatal("task 'test' not found")
		return
	}
	if len(test.Dependencies) != 1 || test.Dependencies[0] != testBuildTask {
		t.Errorf("test should depend on 'build', got %v", test.Dependencies)
	}
}

func TestParseNestedTasks(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "nested.yml"))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
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
		return
	}
	if len(full.Dependencies) != 2 {
		t.Errorf("ci:full should have 2 dependencies, got %d", len(full.Dependencies))
	}
}

func TestParseDescriptions(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "with_description.yml"))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	test := tasks["test"]
	if test == nil {
		t.Fatal("task 'test' not found")
		return
	}
	if test.Description != "Run all unit tests" {
		t.Errorf("expected description 'Run all unit tests', got %q", test.Description)
	}

	coverage := tasks["coverage"]
	if coverage == nil {
		t.Fatal("task 'coverage' not found")
		return
	}
	if coverage.Description != "Generate coverage report" {
		t.Errorf("expected description 'Generate coverage report', got %q", coverage.Description)
	}
}

func TestParseEmptyTasks(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "empty.yml"))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestParseInvalidFiles(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		errMsg string
	}{
		{"invalid YAML syntax", filepath.Join("testdata", "invalid_yaml.yml"), "failed to parse YAML"},
		{"root is not a map", filepath.Join("testdata", "root_not_map.yml"), "root of Babfile must be a map"},
		{"missing tasks key", filepath.Join("testdata", "missing_tasks_key.yml"), "babfile must contain a 'tasks' key"},
		{"tasks not a map", filepath.Join("testdata", "tasks_not_map.yml"), "'tasks' must be a map"},
		{"invalid dependency reference", filepath.Join("testdata", "invalid_deps.yml"), "dependency validation failed"},
		{"empty command", filepath.Join("testdata", "empty_command.yml"), "command cannot be"},
		{"nonexistent file", "nonexistent.yml", "failed to read Babfile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.path)
			if err == nil {
				t.Errorf("Parse() expected error containing %q, got nil", tt.errMsg)
				return
			}
			if !contains(err.Error(), tt.errMsg) {
				t.Errorf("Parse() error = %q, want error containing %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestParseInvalidPaths(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		errMsg string
	}{
		{"empty path", "", "path cannot be empty"},
		{"whitespace path", "   ", "path cannot be"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.path)
			if err == nil {
				t.Errorf("Parse() expected error containing %q, got nil", tt.errMsg)
				return
			}
			if !contains(err.Error(), tt.errMsg) {
				t.Errorf("Parse() error = %q, want error containing %q", err.Error(), tt.errMsg)
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
