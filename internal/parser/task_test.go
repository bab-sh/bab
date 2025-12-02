package parser

import (
	"testing"
)

func assertSliceResults(t *testing.T, funcName string, got []string, want []string, err error, wantErr bool, errMsg string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Errorf("%s() expected error containing %q, got nil", funcName, errMsg)
			return
		}
		if errMsg != "" && !contains(err.Error(), errMsg) {
			t.Errorf("%s() error = %q, want error containing %q", funcName, err.Error(), errMsg)
		}
		return
	}

	if err != nil {
		t.Errorf("%s() unexpected error: %v", funcName, err)
		return
	}

	if len(got) != len(want) {
		t.Errorf("%s() got %d items, want %d", funcName, len(got), len(want))
		return
	}

	for i, item := range got {
		if item != want[i] {
			t.Errorf("%s()[%d] = %q, want %q", funcName, i, item, want[i])
		}
	}
}

func TestParseCommands(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		runCmd   interface{}
		want     []string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "single string command",
			taskName: "test",
			runCmd:   "echo hello",
			want:     []string{"echo hello"},
			wantErr:  false,
		},
		{
			name:     "multiple commands as slice",
			taskName: "build",
			runCmd:   []interface{}{"npm install", "npm build", "npm test"},
			want:     []string{"npm install", "npm build", "npm test"},
			wantErr:  false,
		},
		{
			name:     "single command in slice",
			taskName: "lint",
			runCmd:   []interface{}{"golangci-lint run"},
			want:     []string{"golangci-lint run"},
			wantErr:  false,
		},
		{
			name:     "nil command",
			taskName: "test",
			runCmd:   nil,
			wantErr:  true,
			errMsg:   "nil 'run' command",
		},
		{
			name:     "empty string command",
			taskName: "test",
			runCmd:   "",
			wantErr:  true,
			errMsg:   "command cannot be empty",
		},
		{
			name:     "whitespace only command",
			taskName: "test",
			runCmd:   "   ",
			wantErr:  true,
			errMsg:   "command cannot be",
		},
		{
			name:     "empty slice",
			taskName: "test",
			runCmd:   []interface{}{},
			wantErr:  true,
			errMsg:   "empty 'run' command list",
		},
		{
			name:     "slice with empty string",
			taskName: "test",
			runCmd:   []interface{}{"echo hello", ""},
			wantErr:  true,
			errMsg:   "command cannot be empty",
		},
		{
			name:     "slice with non-string element",
			taskName: "test",
			runCmd:   []interface{}{"echo hello", 123},
			want:     []string{"echo hello", "123"},
			wantErr:  false,
		},
		{
			name:     "integer command converted to string",
			taskName: "test",
			runCmd:   123,
			want:     []string{"123"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommands(tt.taskName, tt.runCmd)
			assertSliceResults(t, "parseCommands", got, tt.want, err, tt.wantErr, tt.errMsg)
		})
	}
}

func TestParseDependencies(t *testing.T) {
	tests := []struct {
		name      string
		taskName  string
		depsValue interface{}
		want      []string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "single dependency as string",
			taskName:  "build",
			depsValue: "clean",
			want:      []string{"clean"},
			wantErr:   false,
		},
		{
			name:      "multiple dependencies as slice",
			taskName:  "test",
			depsValue: []interface{}{"build", "lint"},
			want:      []string{"build", "lint"},
			wantErr:   false,
		},
		{
			name:      "single dependency in slice",
			taskName:  "deploy",
			depsValue: []interface{}{"test"},
			want:      []string{"test"},
			wantErr:   false,
		},
		{
			name:      "nil dependency",
			taskName:  "test",
			depsValue: nil,
			wantErr:   true,
			errMsg:    "nil 'deps' value",
		},
		{
			name:      "empty string dependency",
			taskName:  "test",
			depsValue: "",
			wantErr:   true,
			errMsg:    "empty 'deps' string",
		},
		{
			name:      "empty slice",
			taskName:  "test",
			depsValue: []interface{}{},
			wantErr:   true,
			errMsg:    "empty 'deps' list",
		},
		{
			name:      "slice with empty string",
			taskName:  "test",
			depsValue: []interface{}{"build", ""},
			wantErr:   true,
			errMsg:    "empty dependency",
		},
		{
			name:      "slice with non-string element",
			taskName:  "test",
			depsValue: []interface{}{"build", 123},
			want:      []string{"build", "123"},
			wantErr:   false,
		},
		{
			name:      "integer dependency converted to string",
			taskName:  "test",
			depsValue: 456,
			want:      []string{"456"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDependencies(tt.taskName, tt.depsValue)
			assertSliceResults(t, "parseDependencies", got, tt.want, err, tt.wantErr, tt.errMsg)
		})
	}
}

func TestBuildTask(t *testing.T) {
	tests := []struct {
		name     string
		taskName string
		taskMap  map[string]interface{}
		runCmd   interface{}
		wantErr  bool
		validate func(t *testing.T, task *Task)
	}{
		{
			name:     "simple task with command only",
			taskName: "build",
			taskMap: map[string]interface{}{
				"run": "go build",
			},
			runCmd:  "go build",
			wantErr: false,
			validate: func(t *testing.T, task *Task) {
				if task.Name != "build" {
					t.Errorf("Name = %q, want %q", task.Name, "build")
				}
				if len(task.Commands) != 1 {
					t.Errorf("got %d commands, want 1", len(task.Commands))
				}
				if task.Commands[0] != "go build" {
					t.Errorf("Commands[0] = %q, want %q", task.Commands[0], "go build")
				}
				if task.Description != "" {
					t.Errorf("Description = %q, want empty", task.Description)
				}
				if len(task.Dependencies) != 0 {
					t.Errorf("got %d dependencies, want 0", len(task.Dependencies))
				}
			},
		},
		{
			name:     "task with description",
			taskName: "test",
			taskMap: map[string]interface{}{
				"desc": "Run tests",
				"run":  "go test ./...",
			},
			runCmd:  "go test ./...",
			wantErr: false,
			validate: func(t *testing.T, task *Task) {
				if task.Description != "Run tests" {
					t.Errorf("Description = %q, want %q", task.Description, "Run tests")
				}
			},
		},
		{
			name:     "task with dependencies",
			taskName: "deploy",
			taskMap: map[string]interface{}{
				"deps": []interface{}{"build", "test"},
				"run":  "kubectl apply",
			},
			runCmd:  "kubectl apply",
			wantErr: false,
			validate: func(t *testing.T, task *Task) {
				if len(task.Dependencies) != 2 {
					t.Fatalf("got %d dependencies, want 2", len(task.Dependencies))
				}
				if task.Dependencies[0] != "build" {
					t.Errorf("Dependencies[0] = %q, want %q", task.Dependencies[0], "build")
				}
				if task.Dependencies[1] != "test" {
					t.Errorf("Dependencies[1] = %q, want %q", task.Dependencies[1], "test")
				}
			},
		},
		{
			name:     "complete task with all fields",
			taskName: "release",
			taskMap: map[string]interface{}{
				"desc": "Release new version",
				"deps": "test",
				"run":  []interface{}{"git tag v1.0.0", "git push --tags"},
			},
			runCmd:  []interface{}{"git tag v1.0.0", "git push --tags"},
			wantErr: false,
			validate: func(t *testing.T, task *Task) {
				if task.Name != "release" {
					t.Errorf("Name = %q, want %q", task.Name, "release")
				}
				if task.Description != "Release new version" {
					t.Errorf("Description = %q, want %q", task.Description, "Release new version")
				}
				if len(task.Dependencies) != 1 || task.Dependencies[0] != "test" {
					t.Errorf("Dependencies = %v, want [test]", task.Dependencies)
				}
				if len(task.Commands) != 2 {
					t.Errorf("got %d commands, want 2", len(task.Commands))
				}
			},
		},
		{
			name:     "task with invalid description type",
			taskName: "test",
			taskMap: map[string]interface{}{
				"desc": 123,
				"run":  "echo test",
			},
			runCmd:  "echo test",
			wantErr: false,
			validate: func(t *testing.T, task *Task) {
				if task.Description != "123" {
					t.Errorf("Description = %q, want %q", task.Description, "123")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := buildTask(tt.taskName, tt.taskMap, tt.runCmd)

			if tt.wantErr {
				if err == nil {
					t.Error("buildTask() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("buildTask() unexpected error: %v", err)
				return
			}

			if task == nil {
				t.Fatal("buildTask() returned nil task")
			}

			if tt.validate != nil {
				tt.validate(t, task)
			}
		})
	}
}
