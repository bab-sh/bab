package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestCLI_runCompletion(t *testing.T) {
	tests := []struct {
		shell       string
		wantErr     bool
		errMsg      string
		wantContain string
	}{
		{
			shell:       "bash",
			wantErr:     false,
			wantContain: "bash completion",
		},
		{
			shell:       "zsh",
			wantErr:     false,
			wantContain: "compdef",
		},
		{
			shell:       "fish",
			wantErr:     false,
			wantContain: "complete",
		},
		{
			shell:       "powershell",
			wantErr:     false,
			wantContain: "Register-ArgumentCompleter",
		},
		{
			shell:   "invalid",
			wantErr: true,
			errMsg:  "invalid shell",
		},
		{
			shell:   "",
			wantErr: true,
			errMsg:  "invalid shell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			cli := newCLI()
			cmd := cli.buildCommand()
			cli.completion = tt.shell

			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed to create pipe: %v", err)
			}
			os.Stdout = w

			runErr := cli.runCompletion(cmd)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			if tt.wantErr {
				if runErr == nil {
					t.Errorf("expected error for shell %q, got nil", tt.shell)
					return
				}
				if !strings.Contains(runErr.Error(), tt.errMsg) {
					t.Errorf("error = %q, want containing %q", runErr.Error(), tt.errMsg)
				}
				return
			}

			if runErr != nil {
				t.Errorf("unexpected error: %v", runErr)
				return
			}

			if len(output) == 0 {
				t.Error("expected completion output, got empty")
				return
			}

			if tt.wantContain != "" && !strings.Contains(strings.ToLower(output), strings.ToLower(tt.wantContain)) {
				t.Errorf("output should contain %q for %s shell", tt.wantContain, tt.shell)
			}
		})
	}
}

func TestCLI_runCompletion_allShells(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell+"_generates_output", func(t *testing.T) {
			cli := newCLI()
			cmd := cli.buildCommand()
			cli.completion = shell

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cli.runCompletion(cmd)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)

			if err != nil {
				t.Errorf("%s completion failed: %v", shell, err)
				return
			}

			if buf.Len() < 100 {
				t.Errorf("%s completion output too short (%d bytes)", shell, buf.Len())
			}
		})
	}
}
