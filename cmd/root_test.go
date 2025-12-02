package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCLI(t *testing.T) {
	cli := newCLI()
	if cli == nil {
		t.Fatal("newCLI() returned nil")
		return
	}
	if cli.verbose {
		t.Error("verbose should be false by default")
	}
	if cli.dryRun {
		t.Error("dryRun should be false by default")
	}
	if cli.listTasks {
		t.Error("listTasks should be false by default")
	}
	if cli.completion != "" {
		t.Errorf("completion should be empty by default, got %q", cli.completion)
	}
}

func TestCLI_buildCommand(t *testing.T) {
	cli := newCLI()
	cmd := cli.buildCommand()

	t.Run("command properties", func(t *testing.T) {
		if cmd.Use != "bab [task]" {
			t.Errorf("Use = %q, want %q", cmd.Use, "bab [task]")
		}
		if cmd.Short != "Custom commands for every project" {
			t.Errorf("Short = %q, want %q", cmd.Short, "Custom commands for every project")
		}
		if !cmd.SilenceErrors {
			t.Error("SilenceErrors should be true")
		}
		if !cmd.SilenceUsage {
			t.Error("SilenceUsage should be true")
		}
	})

	t.Run("flags exist", func(t *testing.T) {
		persistentFlags := []string{"verbose", "dry-run"}
		for _, name := range persistentFlags {
			if cmd.PersistentFlags().Lookup(name) == nil {
				t.Errorf("persistent flag %q not found", name)
			}
		}

		localFlags := []string{"list", "completion"}
		for _, name := range localFlags {
			if cmd.Flags().Lookup(name) == nil {
				t.Errorf("flag %q not found", name)
			}
		}
	})

	t.Run("flag shorthand", func(t *testing.T) {
		shorthands := map[string]string{
			"verbose":    "v",
			"dry-run":    "n",
			"list":       "l",
			"completion": "c",
		}
		for name, shorthand := range shorthands {
			flag := cmd.Flags().Lookup(name)
			if flag == nil {
				flag = cmd.PersistentFlags().Lookup(name)
			}
			if flag == nil {
				t.Errorf("flag %q not found", name)
				continue
			}
			if flag.Shorthand != shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", name, flag.Shorthand, shorthand)
			}
		}
	})
}

func TestCLI_runTask(t *testing.T) {
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
    run:
      - cmd: echo "Hello World"`,
			wantErr: false,
		},
		{
			name:     "task not found",
			taskName: "nonexistent",
			babfileYAML: `tasks:
  hello:
    run:
      - cmd: echo "Hello"`,
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

			cli := newCLI()
			cli.ctx = context.Background()
			cli.dryRun = true

			err := cli.runTask(tt.taskName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("runTask() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("runTask() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("runTask() unexpected error: %v", err)
			}
		})
	}
}

func TestCLI_run_dispatching(t *testing.T) {
	t.Run("completion flag takes priority", func(t *testing.T) {
		cli := newCLI()
		cmd := cli.buildCommand()
		cli.completion = "invalid"
		cli.listTasks = true

		err := cli.run(cmd, []string{"sometask"})

		if err == nil || !strings.Contains(err.Error(), "invalid shell") {
			t.Errorf("expected completion error, got: %v", err)
		}
	})

	t.Run("list flag takes priority over args", func(t *testing.T) {
		tmpDir := t.TempDir()
		babfilePath := filepath.Join(tmpDir, "Babfile")
		babfileYAML := `tasks:
  hello:
    run:
      - cmd: echo "Hello"`
		if err := os.WriteFile(babfilePath, []byte(babfileYAML), 0600); err != nil {
			t.Fatalf("failed to create test Babfile: %v", err)
		}

		oldDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(oldDir) }()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}

		cli := newCLI()
		cmd := cli.buildCommand()
		cli.listTasks = true
		cli.ctx = context.Background()

		err := cli.run(cmd, []string{"nonexistent"})

		if err != nil {
			t.Errorf("expected list to run successfully, got error: %v", err)
		}
	})
}
