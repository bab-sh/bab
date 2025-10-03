package executor

import (
	"runtime"
	"testing"

	"github.com/bab-sh/bab/internal/registry"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		check   func(*testing.T, *Executor)
	}{
		{
			name:    "default executor",
			options: nil,
			check: func(t *testing.T, e *Executor) {
				if e.dryRun {
					t.Error("default executor should not have dryRun enabled")
				}
				if e.verbose {
					t.Error("default executor should not have verbose enabled")
				}
			},
		},
		{
			name:    "with dry-run",
			options: []Option{WithDryRun(true)},
			check: func(t *testing.T, e *Executor) {
				if !e.dryRun {
					t.Error("executor should have dryRun enabled")
				}
			},
		},
		{
			name:    "with verbose",
			options: []Option{WithVerbose(true)},
			check: func(t *testing.T, e *Executor) {
				if !e.verbose {
					t.Error("executor should have verbose enabled")
				}
			},
		},
		{
			name:    "with multiple options",
			options: []Option{WithDryRun(true), WithVerbose(true)},
			check: func(t *testing.T, e *Executor) {
				if !e.dryRun {
					t.Error("executor should have dryRun enabled")
				}
				if !e.verbose {
					t.Error("executor should have verbose enabled")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := New(tt.options...)
			if executor == nil {
				t.Fatal("New() returned nil")
			}
			if tt.check != nil {
				tt.check(t, executor)
			}
		})
	}
}

func TestExecutor_Execute(t *testing.T) {
	tests := []struct {
		name    string
		task    *registry.Task
		dryRun  bool
		wantErr bool
	}{
		{
			name:    "nil task",
			task:    nil,
			wantErr: true,
		},
		{
			name: "simple echo command",
			task: &registry.Task{
				Name:        "echo",
				Description: "Echo test",
				Commands:    []string{"echo 'test'"},
			},
			wantErr: false,
		},
		{
			name: "multiple commands",
			task: &registry.Task{
				Name:     "multi",
				Commands: []string{"echo 'first'", "echo 'second'"},
			},
			wantErr: false,
		},
		{
			name: "dry-run mode",
			task: &registry.Task{
				Name:     "dryrun",
				Commands: []string{"echo 'dry run test'"},
			},
			dryRun:  true,
			wantErr: false,
		},
		{
			name: "failing command",
			task: &registry.Task{
				Name:     "fail",
				Commands: []string{"exit 1"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []Option
			if tt.dryRun {
				opts = append(opts, WithDryRun(true))
			}

			executor := New(opts...)
			err := executor.Execute(tt.task)

			if tt.wantErr && err == nil {
				t.Error("Execute() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
			}
		})
	}
}

func TestExecutor_runCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		dryRun  bool
		wantErr bool
	}{
		{
			name:    "simple command",
			command: "echo 'test'",
			wantErr: false,
		},
		{
			name:    "dry-run does not execute",
			command: "echo 'test'",
			dryRun:  true,
			wantErr: false,
		},
		{
			name:    "failing command",
			command: "exit 1",
			wantErr: true,
		},
		{
			name:    "non-existent command",
			command: "nonexistentcommand12345",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []Option
			if tt.dryRun {
				opts = append(opts, WithDryRun(true))
			}

			executor := New(opts...)
			err := executor.runCommand(tt.command, 1, 1)

			if tt.wantErr && err == nil {
				t.Error("runCommand() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("runCommand() unexpected error: %v", err)
			}
		})
	}
}

func TestWithDryRun(t *testing.T) {
	tests := []struct {
		name   string
		dryRun bool
	}{
		{"enable dry-run", true},
		{"disable dry-run", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Executor{}
			opt := WithDryRun(tt.dryRun)
			opt(e)

			if e.dryRun != tt.dryRun {
				t.Errorf("WithDryRun(%v) set dryRun = %v", tt.dryRun, e.dryRun)
			}
		})
	}
}

func TestWithVerbose(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{"enable verbose", true},
		{"disable verbose", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Executor{}
			opt := WithVerbose(tt.verbose)
			opt(e)

			if e.verbose != tt.verbose {
				t.Errorf("WithVerbose(%v) set verbose = %v", tt.verbose, e.verbose)
			}
		})
	}
}

func TestExecutor_platformSpecificExecution(t *testing.T) {
	// Test that commands execute through the correct shell based on platform
	var expectedCommand string
	if runtime.GOOS == "windows" {
		expectedCommand = "echo %OS%"
	} else {
		expectedCommand = "echo $SHELL"
	}

	task := &registry.Task{
		Name:     "platform-test",
		Commands: []string{expectedCommand},
	}

	executor := New()
	err := executor.Execute(task)
	if err != nil {
		t.Errorf("Execute() platform-specific command failed: %v", err)
	}
}
