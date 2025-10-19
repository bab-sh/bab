package parser

import (
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
					Commands: []string{"rm -rf dist"},
				},
				"build": &Task{
					Name:         "build",
					Commands:     []string{"go build"},
					Dependencies: []string{"clean"},
				},
				"test": &Task{
					Name:         "test",
					Commands:     []string{"go test"},
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
					Commands: []string{"golangci-lint run"},
				},
				"test": &Task{
					Name:     "test",
					Commands: []string{"go test"},
				},
				"ci": &Task{
					Name:         "ci",
					Commands:     []string{"echo done"},
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
					Commands: []string{"go build"},
				},
				"test": &Task{
					Name:     "test",
					Commands: []string{"go test"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid dependency - task does not exist",
			tasks: TaskMap{
				"build": &Task{
					Name:         "build",
					Commands:     []string{"go build"},
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
					Commands: []string{"golangci-lint run"},
				},
				"ci": &Task{
					Name:         "ci",
					Commands:     []string{"echo done"},
					Dependencies: []string{"lint", "nonexistent", "alsonothere"},
				},
			},
			wantErr: true,
			errMsg:  "invalid dependency",
		},
		{
			name: "empty task map",
			tasks: TaskMap{},
			wantErr: false,
		},
		{
			name: "nested task dependencies",
			tasks: TaskMap{
				"ci:test": &Task{
					Name:     "ci:test",
					Commands: []string{"go test"},
				},
				"ci:lint": &Task{
					Name:     "ci:lint",
					Commands: []string{"golangci-lint run"},
				},
				"ci:full": &Task{
					Name:         "ci:full",
					Commands:     []string{"echo done"},
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
					Commands:     []string{"go test"},
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
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
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
