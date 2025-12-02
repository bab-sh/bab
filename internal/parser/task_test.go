package parser

import (
	"strings"
	"testing"
)

func assertCommandResults(t *testing.T, funcName string, got []Command, want []Command, err error, wantErr bool, errMsg string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Errorf("%s() expected error containing %q, got nil", funcName, errMsg)
			return
		}
		if errMsg != "" && !strings.Contains(err.Error(), errMsg) {
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
		if item.Cmd != want[i].Cmd {
			t.Errorf("%s()[%d].Cmd = %q, want %q", funcName, i, item.Cmd, want[i].Cmd)
		}
		if len(item.Platforms) != len(want[i].Platforms) {
			t.Errorf("%s()[%d].Platforms length = %d, want %d", funcName, i, len(item.Platforms), len(want[i].Platforms))
			continue
		}
		for j, p := range item.Platforms {
			if p != want[i].Platforms[j] {
				t.Errorf("%s()[%d].Platforms[%d] = %q, want %q", funcName, i, j, p, want[i].Platforms[j])
			}
		}
	}
}

func assertSliceResults(t *testing.T, funcName string, got []string, want []string, err error, wantErr bool, errMsg string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Errorf("%s() expected error containing %q, got nil", funcName, errMsg)
			return
		}
		if errMsg != "" && !strings.Contains(err.Error(), errMsg) {
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
		want     []Command
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "single command object",
			taskName: "test",
			runCmd: []interface{}{
				map[string]interface{}{"cmd": "echo hello"},
			},
			want:    []Command{{Cmd: "echo hello"}},
			wantErr: false,
		},
		{
			name:     "multiple command objects",
			taskName: "build",
			runCmd: []interface{}{
				map[string]interface{}{"cmd": "npm install"},
				map[string]interface{}{"cmd": "npm build"},
				map[string]interface{}{"cmd": "npm test"},
			},
			want: []Command{
				{Cmd: "npm install"},
				{Cmd: "npm build"},
				{Cmd: "npm test"},
			},
			wantErr: false,
		},
		{
			name:     "command with platforms",
			taskName: "deploy",
			runCmd: []interface{}{
				map[string]interface{}{
					"cmd":       "bash scripts/deploy.sh",
					"platforms": []interface{}{"linux", "darwin"},
				},
			},
			want: []Command{
				{Cmd: "bash scripts/deploy.sh", Platforms: []string{"linux", "darwin"}},
			},
			wantErr: false,
		},
		{
			name:     "mixed commands with and without platforms",
			taskName: "setup",
			runCmd: []interface{}{
				map[string]interface{}{"cmd": "echo 'Starting...'"},
				map[string]interface{}{
					"cmd":       "./unix-setup.sh",
					"platforms": []interface{}{"linux", "darwin"},
				},
				map[string]interface{}{
					"cmd":       "powershell setup.ps1",
					"platforms": []interface{}{"windows"},
				},
			},
			want: []Command{
				{Cmd: "echo 'Starting...'"},
				{Cmd: "./unix-setup.sh", Platforms: []string{"linux", "darwin"}},
				{Cmd: "powershell setup.ps1", Platforms: []string{"windows"}},
			},
			wantErr: false,
		},
		{
			name:     "nil command",
			taskName: "test",
			runCmd:   nil,
			wantErr:  true,
			errMsg:   "nil 'run' field",
		},
		{
			name:     "string command (not allowed)",
			taskName: "test",
			runCmd:   "echo hello",
			wantErr:  true,
			errMsg:   "'run' must be a list of commands",
		},
		{
			name:     "empty slice",
			taskName: "test",
			runCmd:   []interface{}{},
			wantErr:  true,
			errMsg:   "'run' field cannot be empty",
		},
		{
			name:     "slice with string (not object)",
			taskName: "test",
			runCmd:   []interface{}{"echo hello"},
			wantErr:  true,
			errMsg:   "must be an object with 'cmd' field",
		},
		{
			name:     "object missing cmd field",
			taskName: "test",
			runCmd: []interface{}{
				map[string]interface{}{"platforms": []interface{}{"linux"}},
			},
			wantErr: true,
			errMsg:  "missing required 'cmd' field",
		},
		{
			name:     "empty cmd field",
			taskName: "test",
			runCmd: []interface{}{
				map[string]interface{}{"cmd": ""},
			},
			wantErr: true,
			errMsg:  "command cannot be empty",
		},
		{
			name:     "whitespace only cmd",
			taskName: "test",
			runCmd: []interface{}{
				map[string]interface{}{"cmd": "   "},
			},
			wantErr: true,
			errMsg:  "command cannot be",
		},
		{
			name:     "invalid platform",
			taskName: "test",
			runCmd: []interface{}{
				map[string]interface{}{
					"cmd":       "echo hello",
					"platforms": []interface{}{"invalid_os"},
				},
			},
			wantErr: true,
			errMsg:  "invalid platform",
		},
		{
			name:     "platforms not a list",
			taskName: "test",
			runCmd: []interface{}{
				map[string]interface{}{
					"cmd":       "echo hello",
					"platforms": "linux",
				},
			},
			wantErr: true,
			errMsg:  "'platforms' must be a list of strings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommands(tt.taskName, tt.runCmd)
			assertCommandResults(t, "parseCommands", got, tt.want, err, tt.wantErr, tt.errMsg)
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
			errMsg:    "empty dependency at index 0",
		},
		{
			name:      "empty slice",
			taskName:  "test",
			depsValue: []interface{}{},
			wantErr:   true,
			errMsg:    "'deps' list cannot be empty",
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
				"run": []interface{}{
					map[string]interface{}{"cmd": "go build"},
				},
			},
			runCmd: []interface{}{
				map[string]interface{}{"cmd": "go build"},
			},
			wantErr: false,
			validate: func(t *testing.T, task *Task) {
				if task.Name != "build" {
					t.Errorf("Name = %q, want %q", task.Name, "build")
				}
				if len(task.Commands) != 1 {
					t.Errorf("got %d commands, want 1", len(task.Commands))
				}
				if task.Commands[0].Cmd != "go build" {
					t.Errorf("Commands[0].Cmd = %q, want %q", task.Commands[0].Cmd, "go build")
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
				"run": []interface{}{
					map[string]interface{}{"cmd": "go test ./..."},
				},
			},
			runCmd: []interface{}{
				map[string]interface{}{"cmd": "go test ./..."},
			},
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
				"run": []interface{}{
					map[string]interface{}{"cmd": "kubectl apply"},
				},
			},
			runCmd: []interface{}{
				map[string]interface{}{"cmd": "kubectl apply"},
			},
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
				"run": []interface{}{
					map[string]interface{}{"cmd": "git tag v1.0.0"},
					map[string]interface{}{"cmd": "git push --tags"},
				},
			},
			runCmd: []interface{}{
				map[string]interface{}{"cmd": "git tag v1.0.0"},
				map[string]interface{}{"cmd": "git push --tags"},
			},
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
			name:     "task with platform-specific commands",
			taskName: "setup",
			taskMap: map[string]interface{}{
				"desc": "Setup environment",
				"run": []interface{}{
					map[string]interface{}{
						"cmd":       "./setup.sh",
						"platforms": []interface{}{"linux", "darwin"},
					},
					map[string]interface{}{
						"cmd":       "setup.bat",
						"platforms": []interface{}{"windows"},
					},
				},
			},
			runCmd: []interface{}{
				map[string]interface{}{
					"cmd":       "./setup.sh",
					"platforms": []interface{}{"linux", "darwin"},
				},
				map[string]interface{}{
					"cmd":       "setup.bat",
					"platforms": []interface{}{"windows"},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, task *Task) {
				if len(task.Commands) != 2 {
					t.Fatalf("got %d commands, want 2", len(task.Commands))
				}
				if task.Commands[0].Cmd != "./setup.sh" {
					t.Errorf("Commands[0].Cmd = %q, want %q", task.Commands[0].Cmd, "./setup.sh")
				}
				if len(task.Commands[0].Platforms) != 2 {
					t.Errorf("Commands[0].Platforms length = %d, want 2", len(task.Commands[0].Platforms))
				}
				if task.Commands[1].Cmd != "setup.bat" {
					t.Errorf("Commands[1].Cmd = %q, want %q", task.Commands[1].Cmd, "setup.bat")
				}
				if len(task.Commands[1].Platforms) != 1 || task.Commands[1].Platforms[0] != "windows" {
					t.Errorf("Commands[1].Platforms = %v, want [windows]", task.Commands[1].Platforms)
				}
			},
		},
		{
			name:     "task with invalid description type",
			taskName: "test",
			taskMap: map[string]interface{}{
				"desc": 123,
				"run": []interface{}{
					map[string]interface{}{"cmd": "echo test"},
				},
			},
			runCmd: []interface{}{
				map[string]interface{}{"cmd": "echo test"},
			},
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

func TestCommandShouldRunOnPlatform(t *testing.T) {
	tests := []struct {
		name     string
		command  Command
		platform string
		want     bool
	}{
		{
			name:     "no platforms specified - runs on any",
			command:  Command{Cmd: "echo hello"},
			platform: "linux",
			want:     true,
		},
		{
			name:     "empty platforms - runs on any",
			command:  Command{Cmd: "echo hello", Platforms: []string{}},
			platform: "darwin",
			want:     true,
		},
		{
			name:     "platform matches",
			command:  Command{Cmd: "echo hello", Platforms: []string{"linux", "darwin"}},
			platform: "linux",
			want:     true,
		},
		{
			name:     "platform does not match",
			command:  Command{Cmd: "echo hello", Platforms: []string{"linux", "darwin"}},
			platform: "windows",
			want:     false,
		},
		{
			name:     "single platform matches",
			command:  Command{Cmd: "echo hello", Platforms: []string{"windows"}},
			platform: "windows",
			want:     true,
		},
		{
			name:     "single platform does not match",
			command:  Command{Cmd: "echo hello", Platforms: []string{"windows"}},
			platform: "linux",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.command.ShouldRunOnPlatform(tt.platform)
			if got != tt.want {
				t.Errorf("ShouldRunOnPlatform(%q) = %v, want %v", tt.platform, got, tt.want)
			}
		})
	}
}
