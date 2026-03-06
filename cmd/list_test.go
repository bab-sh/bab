package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_runList(t *testing.T) {
	tests := []struct {
		name        string
		babfileYAML string
		wantErr     bool
		errMsg      string
	}{
		{
			name: "list simple tasks",
			babfileYAML: `tasks:
  hello:
    run:
      - cmd: echo "Hello"
  world:
    run:
      - cmd: echo "World"`,
			wantErr: false,
		},
		{
			name: "list tasks with descriptions",
			babfileYAML: `tasks:
  test:
    desc: Run tests
    run:
      - cmd: go test ./...
  build:
    desc: Build the project
    run:
      - cmd: go build`,
			wantErr: false,
		},
		{
			name: "list tasks with colons in names",
			babfileYAML: `tasks:
  ci:test:
    run:
      - cmd: echo "Testing"
  ci:lint:
    run:
      - cmd: echo "Linting"
  dev:start:
    desc: Start dev server
    run:
      - cmd: echo "Starting"`,
			wantErr: false,
		},
		{
			name:        "empty babfile",
			babfileYAML: ``,
			wantErr:     true,
			errMsg:      "Babfile",
		},
		{
			name: "tasks with dependencies",
			babfileYAML: `tasks:
  clean:
    run:
      - cmd: echo "Cleaning"
  build:
    deps: [clean]
    run:
      - cmd: echo "Building"
  test:
    deps: [build]
    run:
      - cmd: echo "Testing"`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			babfilePath := filepath.Join(tmpDir, "Babfile")

			if tt.babfileYAML != "" {
				if err := os.WriteFile(babfilePath, []byte(tt.babfileYAML), 0600); err != nil {
					t.Fatalf("failed to create test Babfile: %v", err)
				}
			}

			oldDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get working directory: %v", err)
			}
			defer func() { _ = os.Chdir(oldDir) }()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}

			cli := newCLI()
			err = cli.runList()

			if tt.wantErr {
				if err == nil {
					t.Errorf("runList() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("runList() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("runList() unexpected error: %v", err)
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]*node
		want  []string
	}{
		{
			name: "sorted alphabetically",
			input: map[string]*node{
				"z": {},
				"a": {},
				"m": {},
			},
			want: []string{"a", "m", "z"},
		},
		{
			name:  "empty map",
			input: map[string]*node{},
			want:  []string{},
		},
		{
			name: "single item",
			input: map[string]*node{
				"test": {},
			},
			want: []string{"test"},
		},
		{
			name: "numbers sorted as strings",
			input: map[string]*node{
				"3": {},
				"1": {},
				"2": {},
			},
			want: []string{"1", "2", "3"},
		},
		{
			name: "mixed case",
			input: map[string]*node{
				"Zulu":  {},
				"alpha": {},
				"Bravo": {},
			},
			want: []string{"Bravo", "Zulu", "alpha"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortedKeys(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("sortedKeys() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i, key := range got {
				if key != tt.want[i] {
					t.Errorf("sortedKeys()[%d] = %q, want %q", i, key, tt.want[i])
				}
			}
		})
	}
}
