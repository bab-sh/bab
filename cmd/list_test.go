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
    run: echo "Hello"
  world:
    run: echo "World"`,
			wantErr: false,
		},
		{
			name: "list tasks with descriptions",
			babfileYAML: `tasks:
  test:
    desc: Run tests
    run: go test ./...
  build:
    desc: Build the project
    run: go build`,
			wantErr: false,
		},
		{
			name: "list nested tasks",
			babfileYAML: `tasks:
  ci:
    test:
      run: echo "Testing"
    lint:
      run: echo "Linting"
  dev:
    start:
      desc: Start dev server
      run: echo "Starting"`,
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
    run: echo "Cleaning"
  build:
    deps: clean
    run: echo "Building"
  test:
    deps: build
    run: echo "Testing"`,
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

			oldDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(oldDir) }()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}

			cli := newCLI()
			err := cli.runList()

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

func TestCLI_runListNoBabfile(t *testing.T) {
	tmpDir := t.TempDir()

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	cli := newCLI()
	err := cli.runList()

	if err == nil {
		t.Error("runList() expected error when no Babfile exists, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "Babfile") {
		t.Errorf("runList() error = %q, want error containing 'Babfile'", err.Error())
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

func TestNodeStructure(t *testing.T) {
	t.Run("create empty node", func(t *testing.T) {
		n := &node{children: make(map[string]*node)}
		if n.desc != "" {
			t.Errorf("new node desc = %q, want empty", n.desc)
		}
		if n.children == nil {
			t.Error("new node children = nil, want empty map")
		}
		if len(n.children) != 0 {
			t.Errorf("new node children length = %d, want 0", len(n.children))
		}
	})

	t.Run("create node with description", func(t *testing.T) {
		n := &node{
			desc:     "Test description",
			children: make(map[string]*node),
		}
		if n.desc != "Test description" {
			t.Errorf("node desc = %q, want %q", n.desc, "Test description")
		}
	})

	t.Run("add child to node", func(t *testing.T) {
		parent := &node{children: make(map[string]*node)}
		child := &node{
			desc:     "Child node",
			children: make(map[string]*node),
		}
		parent.children["test"] = child

		if len(parent.children) != 1 {
			t.Errorf("parent children length = %d, want 1", len(parent.children))
		}
		if parent.children["test"].desc != "Child node" {
			t.Errorf("child desc = %q, want %q", parent.children["test"].desc, "Child node")
		}
	})
}
