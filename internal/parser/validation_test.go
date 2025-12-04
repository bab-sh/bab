package parser

import (
	"strings"
	"testing"

	"github.com/bab-sh/bab/internal/babfile"
)

func TestValidateDependencies(t *testing.T) {
	tests := []struct {
		name    string
		tasks   babfile.TaskMap
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid dependencies",
			tasks: babfile.TaskMap{
				"clean": &babfile.Task{
					Name:     "clean",
					Commands: []babfile.Command{{Cmd: "rm -rf dist"}},
				},
				"build": &babfile.Task{
					Name:         "build",
					Commands:     []babfile.Command{{Cmd: "go build"}},
					Dependencies: []string{"clean"},
				},
				"test": &babfile.Task{
					Name:         "test",
					Commands:     []babfile.Command{{Cmd: "go test"}},
					Dependencies: []string{"build"},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple valid dependencies",
			tasks: babfile.TaskMap{
				"lint": &babfile.Task{
					Name:     "lint",
					Commands: []babfile.Command{{Cmd: "golangci-lint run"}},
				},
				"test": &babfile.Task{
					Name:     "test",
					Commands: []babfile.Command{{Cmd: "go test"}},
				},
				"ci": &babfile.Task{
					Name:         "ci",
					Commands:     []babfile.Command{{Cmd: "echo done"}},
					Dependencies: []string{"lint", "test"},
				},
			},
			wantErr: false,
		},
		{
			name: "no dependencies",
			tasks: babfile.TaskMap{
				"build": &babfile.Task{
					Name:     "build",
					Commands: []babfile.Command{{Cmd: "go build"}},
				},
				"test": &babfile.Task{
					Name:     "test",
					Commands: []babfile.Command{{Cmd: "go test"}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid dependency - task does not exist",
			tasks: babfile.TaskMap{
				"build": &babfile.Task{
					Name:         "build",
					Commands:     []babfile.Command{{Cmd: "go build"}},
					Dependencies: []string{"nonexistent"},
				},
			},
			wantErr: true,
			errMsg:  "invalid dependency",
		},
		{
			name: "invalid dependency - one of multiple",
			tasks: babfile.TaskMap{
				"lint": &babfile.Task{
					Name:     "lint",
					Commands: []babfile.Command{{Cmd: "golangci-lint run"}},
				},
				"ci": &babfile.Task{
					Name:         "ci",
					Commands:     []babfile.Command{{Cmd: "echo done"}},
					Dependencies: []string{"lint", "nonexistent", "alsonothere"},
				},
			},
			wantErr: true,
			errMsg:  "invalid dependency",
		},
		{
			name:    "empty task map",
			tasks:   babfile.TaskMap{},
			wantErr: false,
		},
		{
			name: "nested task dependencies",
			tasks: babfile.TaskMap{
				"ci:test": &babfile.Task{
					Name:     "ci:test",
					Commands: []babfile.Command{{Cmd: "go test"}},
				},
				"ci:lint": &babfile.Task{
					Name:     "ci:lint",
					Commands: []babfile.Command{{Cmd: "golangci-lint run"}},
				},
				"ci:full": &babfile.Task{
					Name:         "ci:full",
					Commands:     []babfile.Command{{Cmd: "echo done"}},
					Dependencies: []string{"ci:test", "ci:lint"},
				},
			},
			wantErr: false,
		},
		{
			name: "self dependency should be allowed by this function",
			tasks: babfile.TaskMap{
				"test": &babfile.Task{
					Name:         "test",
					Commands:     []babfile.Command{{Cmd: "go test"}},
					Dependencies: []string{"test"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDependencies(tt.tasks)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateDependencies() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateDependencies() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateDependencies() unexpected error: %v", err)
			}
		})
	}
}
