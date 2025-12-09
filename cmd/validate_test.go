package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_runValidate(t *testing.T) {
	tests := []struct {
		name        string
		babfileYAML string
		wantErr     bool
		errMsg      string
	}{
		{
			name: "valid babfile",
			babfileYAML: `tasks:
  hello:
    run:
      - cmd: echo "Hello World"`,
			wantErr: false,
		},
		{
			name: "valid babfile with dependencies",
			babfileYAML: `tasks:
  build:
    deps: [clean]
    run:
      - cmd: go build
  clean:
    run:
      - cmd: rm -rf build`,
			wantErr: false,
		},
		{
			name:        "invalid yaml",
			babfileYAML: `tasks: [invalid`,
			wantErr:     true,
			errMsg:      "did not find expected",
		},
		{
			name: "missing dependency",
			babfileYAML: `tasks:
  build:
    deps: [nonexistent]
    run:
      - cmd: go build`,
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

			err := cli.runValidate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("runValidate() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("runValidate() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("runValidate() unexpected error: %v", err)
			}
		})
	}
}

func TestCLI_runValidate_NoBabfile(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cli := newCLI()
	cli.ctx = context.Background()

	err := cli.runValidate()

	if err == nil {
		t.Error("runValidate() expected error when no Babfile exists")
	}
}
