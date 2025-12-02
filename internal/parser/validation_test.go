package parser

import (
	"strings"
	"testing"
)

func TestValidateDependencies(t *testing.T) {
	tests := []struct {
		name    string
		tasks   TaskMap
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid dependencies",
			tasks: TaskMap{
				"clean": &Task{
					Name:     "clean",
					Commands: []Command{{Cmd: "rm -rf dist"}},
				},
				"build": &Task{
					Name:         "build",
					Commands:     []Command{{Cmd: "go build"}},
					Dependencies: []string{"clean"},
				},
				"test": &Task{
					Name:         "test",
					Commands:     []Command{{Cmd: "go test"}},
					Dependencies: []string{"build"},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple valid dependencies",
			tasks: TaskMap{
				"lint": &Task{
					Name:     "lint",
					Commands: []Command{{Cmd: "golangci-lint run"}},
				},
				"test": &Task{
					Name:     "test",
					Commands: []Command{{Cmd: "go test"}},
				},
				"ci": &Task{
					Name:         "ci",
					Commands:     []Command{{Cmd: "echo done"}},
					Dependencies: []string{"lint", "test"},
				},
			},
			wantErr: false,
		},
		{
			name: "no dependencies",
			tasks: TaskMap{
				"build": &Task{
					Name:     "build",
					Commands: []Command{{Cmd: "go build"}},
				},
				"test": &Task{
					Name:     "test",
					Commands: []Command{{Cmd: "go test"}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid dependency - task does not exist",
			tasks: TaskMap{
				"build": &Task{
					Name:         "build",
					Commands:     []Command{{Cmd: "go build"}},
					Dependencies: []string{"nonexistent"},
				},
			},
			wantErr: true,
			errMsg:  "invalid dependency",
		},
		{
			name: "invalid dependency - one of multiple",
			tasks: TaskMap{
				"lint": &Task{
					Name:     "lint",
					Commands: []Command{{Cmd: "golangci-lint run"}},
				},
				"ci": &Task{
					Name:         "ci",
					Commands:     []Command{{Cmd: "echo done"}},
					Dependencies: []string{"lint", "nonexistent", "alsonothere"},
				},
			},
			wantErr: true,
			errMsg:  "invalid dependency",
		},
		{
			name:    "empty task map",
			tasks:   TaskMap{},
			wantErr: false,
		},
		{
			name: "nested task dependencies",
			tasks: TaskMap{
				"ci:test": &Task{
					Name:     "ci:test",
					Commands: []Command{{Cmd: "go test"}},
				},
				"ci:lint": &Task{
					Name:     "ci:lint",
					Commands: []Command{{Cmd: "golangci-lint run"}},
				},
				"ci:full": &Task{
					Name:         "ci:full",
					Commands:     []Command{{Cmd: "echo done"}},
					Dependencies: []string{"ci:test", "ci:lint"},
				},
			},
			wantErr: false,
		},
		{
			name: "self dependency should be allowed by this function",
			tasks: TaskMap{
				"test": &Task{
					Name:         "test",
					Commands:     []Command{{Cmd: "go test"}},
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
