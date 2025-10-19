package parser

import (
	"testing"
)

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

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseCommands() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("parseCommands() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseCommands() unexpected error: %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("parseCommands() got %d commands, want %d", len(got), len(tt.want))
				return
			}

			for i, cmd := range got {
				if cmd != tt.want[i] {
					t.Errorf("parseCommands()[%d] = %q, want %q", i, cmd, tt.want[i])
				}
			}
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

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDependencies() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("parseDependencies() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseDependencies() unexpected error: %v", err)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("parseDependencies() got %d deps, want %d", len(got), len(tt.want))
				return
			}

			for i, dep := range got {
				if dep != tt.want[i] {
					t.Errorf("parseDependencies()[%d] = %q, want %q", i, dep, tt.want[i])
				}
			}
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

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"valid command", "echo hello", false},
		{"command with flags", "ls -la", false},
		{"complex command", "docker run -it --rm ubuntu", false},
		{"empty command", "", true},
		{"whitespace only", "   ", true},
		{"tab only", "\t", true},
		{"newline only", "\n", true},
		{"mixed whitespace", " \t\n ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
