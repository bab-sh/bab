package executor

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/bab-sh/bab/internal/parser"
)

func TestGetShellCommand(t *testing.T) {
	shell, arg := getShellCommand()

	if runtime.GOOS == "windows" {
		if shell != windowsShell {
			t.Errorf("getShellCommand() shell = %q, want %q on Windows", shell, windowsShell)
		}
		if arg != windowsArg {
			t.Errorf("getShellCommand() arg = %q, want %q on Windows", arg, windowsArg)
		}
	} else {
		if shell != unixShell {
			t.Errorf("getShellCommand() shell = %q, want %q on Unix", shell, unixShell)
		}
		if arg != unixArg {
			t.Errorf("getShellCommand() arg = %q, want %q on Unix", arg, unixArg)
		}
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		task    *parser.Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid task with single command",
			task: &parser.Task{
				Name:     "test",
				Commands: []parser.Command{{Cmd: "echo hello"}},
			},
			wantErr: false,
		},
		{
			name: "valid task with multiple commands",
			task: &parser.Task{
				Name: "build",
				Commands: []parser.Command{
					{Cmd: "echo step1"},
					{Cmd: "echo step2"},
					{Cmd: "echo step3"},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil task",
			task:    nil,
			wantErr: true,
			errMsg:  "task cannot be nil",
		},
		{
			name: "task with no commands",
			task: &parser.Task{
				Name:     "empty",
				Commands: []parser.Command{},
			},
			wantErr: true,
			errMsg:  "has no commands to execute",
		},
		{
			name: "task with empty command string",
			task: &parser.Task{
				Name:     "invalid",
				Commands: []parser.Command{{Cmd: ""}},
			},
			wantErr: true,
			errMsg:  "command cannot be empty",
		},
		{
			name: "task with whitespace-only command",
			task: &parser.Task{
				Name:     "invalid",
				Commands: []parser.Command{{Cmd: "   "}},
			},
			wantErr: true,
			errMsg:  "command cannot be",
		},
		{
			name: "task with valid and invalid commands",
			task: &parser.Task{
				Name:     "mixed",
				Commands: []parser.Command{{Cmd: "echo valid"}, {Cmd: ""}},
			},
			wantErr: true,
			errMsg:  "command cannot be empty",
		},
		{
			name: "task with failing command",
			task: &parser.Task{
				Name:     "fail",
				Commands: []parser.Command{{Cmd: "exit 1"}},
			},
			wantErr: true,
			errMsg:  "command 1 failed",
		},
		{
			name: "task with nonexistent command",
			task: &parser.Task{
				Name:     "nonexistent",
				Commands: []parser.Command{{Cmd: "this-command-definitely-does-not-exist-12345"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := Execute(ctx, tt.task)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Execute() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Execute() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
			}
		})
	}
}

func TestExecutePlatformFiltering(t *testing.T) {
	currentPlatform := runtime.GOOS

	tests := []struct {
		name    string
		task    *parser.Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "command matches current platform",
			task: &parser.Task{
				Name: "test",
				Commands: []parser.Command{
					{Cmd: "echo hello", Platforms: []string{currentPlatform}},
				},
			},
			wantErr: false,
		},
		{
			name: "command without platform filter runs",
			task: &parser.Task{
				Name: "test",
				Commands: []parser.Command{
					{Cmd: "echo hello"},
				},
			},
			wantErr: false,
		},
		{
			name: "all commands filtered out",
			task: &parser.Task{
				Name: "test",
				Commands: []parser.Command{
					{Cmd: "echo hello", Platforms: []string{"nonexistent_platform"}},
				},
			},
			wantErr: true,
			errMsg:  "has no commands for platform",
		},
		{
			name: "mixed commands - some filtered some not",
			task: &parser.Task{
				Name: "test",
				Commands: []parser.Command{
					{Cmd: "echo filtered", Platforms: []string{"nonexistent_platform"}},
					{Cmd: "echo runs"},
				},
			},
			wantErr: false,
		},
		{
			name: "platform-specific with multiple platforms",
			task: &parser.Task{
				Name: "test",
				Commands: []parser.Command{
					{Cmd: "echo hello", Platforms: []string{"linux", "darwin", "windows"}},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := Execute(ctx, tt.task)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Execute() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Execute() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
			}
		})
	}
}

func TestExecuteWithContext(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		task := &parser.Task{
			Name:     "test",
			Commands: []parser.Command{{Cmd: "sleep 10"}},
		}

		err := Execute(ctx, task)
		if err == nil {
			t.Error("Execute() expected error due to cancelled context, got nil")
			return
		}

		if !strings.Contains(err.Error(), "cancelled") {
			t.Errorf("Execute() error = %q, want error containing 'cancelled'", err.Error())
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		task := &parser.Task{
			Name:     "test",
			Commands: []parser.Command{{Cmd: "sleep 10"}},
		}

		err := Execute(ctx, task)
		if err == nil {
			t.Error("Execute() expected timeout error, got nil")
		}
	})
}

func TestDryRun(t *testing.T) {
	tests := []struct {
		name    string
		task    *parser.Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid task with single command",
			task: &parser.Task{
				Name:        "test",
				Description: "Test task",
				Commands:    []parser.Command{{Cmd: "echo hello"}},
			},
			wantErr: false,
		},
		{
			name: "valid task with multiple commands",
			task: &parser.Task{
				Name: "build",
				Commands: []parser.Command{
					{Cmd: "echo step1"},
					{Cmd: "echo step2"},
					{Cmd: "echo step3"},
				},
			},
			wantErr: false,
		},
		{
			name: "task with description and dependencies",
			task: &parser.Task{
				Name:         "deploy",
				Description:  "Deploy application",
				Commands:     []parser.Command{{Cmd: "kubectl apply"}},
				Dependencies: []string{"build", "test"},
			},
			wantErr: false,
		},
		{
			name:    "nil task",
			task:    nil,
			wantErr: true,
			errMsg:  "task cannot be nil",
		},
		{
			name: "task with no commands",
			task: &parser.Task{
				Name:     "empty",
				Commands: []parser.Command{},
			},
			wantErr: true,
			errMsg:  "has no commands to execute",
		},
		{
			name: "task without description",
			task: &parser.Task{
				Name:     "nodesc",
				Commands: []parser.Command{{Cmd: "echo test"}},
			},
			wantErr: false,
		},
		{
			name: "task without dependencies",
			task: &parser.Task{
				Name:     "nodeps",
				Commands: []parser.Command{{Cmd: "echo test"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := DryRun(ctx, tt.task)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DryRun() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("DryRun() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("DryRun() unexpected error: %v", err)
			}
		})
	}
}

func TestDryRunPlatformFiltering(t *testing.T) {
	tests := []struct {
		name    string
		task    *parser.Task
		wantErr bool
		errMsg  string
	}{
		{
			name: "all commands filtered out in dry run",
			task: &parser.Task{
				Name: "test",
				Commands: []parser.Command{
					{Cmd: "echo hello", Platforms: []string{"nonexistent_platform"}},
				},
			},
			wantErr: true,
			errMsg:  "has no commands for platform",
		},
		{
			name: "mixed commands in dry run - some filtered",
			task: &parser.Task{
				Name: "test",
				Commands: []parser.Command{
					{Cmd: "echo filtered", Platforms: []string{"nonexistent_platform"}},
					{Cmd: "echo runs"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := DryRun(ctx, tt.task)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DryRun() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("DryRun() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("DryRun() unexpected error: %v", err)
			}
		})
	}
}

func TestDryRunWithContext(t *testing.T) {
	t.Run("context cancellation during dry run", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		task := &parser.Task{
			Name: "test",
			Commands: []parser.Command{
				{Cmd: "echo 1"},
				{Cmd: "echo 2"},
				{Cmd: "echo 3"},
			},
		}

		err := DryRun(ctx, task)
		if err == nil {
			t.Error("DryRun() expected error due to cancelled context, got nil")
			return
		}

		if !strings.Contains(err.Error(), "cancelled") {
			t.Errorf("DryRun() error = %q, want error containing 'cancelled'", err.Error())
		}
	})
}
