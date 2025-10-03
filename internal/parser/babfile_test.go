package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bab-sh/bab/internal/registry"
)

func TestNew(t *testing.T) {
	reg := registry.New()
	p := New(reg)

	if p == nil {
		t.Fatal("New() returned nil")
	}

	if p.registry == nil {
		t.Error("Parser registry should not be nil")
	}
}

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTasks []string
		wantErr   bool
	}{
		{
			name: "simple task",
			input: `
build:
  desc: Build the project
  run: go build
`,
			wantTasks: []string{"build"},
			wantErr:   false,
		},
		{
			name: "multiple tasks",
			input: `
build:
  desc: Build the project
  run: go build

test:
  desc: Run tests
  run: go test
`,
			wantTasks: []string{"build", "test"},
			wantErr:   false,
		},
		{
			name: "nested tasks",
			input: `
dev:
  start:
    desc: Start dev server
    run: npm run dev
  watch:
    desc: Watch files
    run: npm run watch
`,
			wantTasks: []string{"dev:start", "dev:watch"},
			wantErr:   false,
		},
		{
			name: "task with multiple commands",
			input: `
deploy:
  desc: Deploy application
  run:
    - npm run build
    - npm run deploy
`,
			wantTasks: []string{"deploy"},
			wantErr:   false,
		},
		{
			name: "task without description",
			input: `
build:
  run: go build
`,
			wantTasks: []string{"build"},
			wantErr:   false,
		},
		{
			name: "mixed root and nested tasks",
			input: `
build:
  desc: Build
  run: go build

dev:
  start:
    desc: Start dev
    run: npm run dev

test:
  desc: Test
  run: go test
`,
			wantTasks: []string{"build", "dev:start", "test"},
			wantErr:   false,
		},
		{
			name: "invalid yaml",
			input: `
build: [
  desc: Build
`,
			wantErr: true,
		},
		{
			name: "task without run command",
			input: `
build:
  desc: Build the project
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := registry.New()
			p := New(reg)

			reader := strings.NewReader(tt.input)
			err := p.Parse(reader)

			if tt.wantErr {
				if err == nil {
					t.Error("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}

			// Verify expected tasks were registered
			tasks := reg.List()
			if len(tasks) != len(tt.wantTasks) {
				t.Errorf("Parse() registered %d tasks, want %d", len(tasks), len(tt.wantTasks))
			}

			for _, wantName := range tt.wantTasks {
				task, err := reg.Get(wantName)
				if err != nil {
					t.Errorf("Parse() task %q not found: %v", wantName, err)
					continue
				}
				if task.Name != wantName {
					t.Errorf("Parse() task name = %q, want %q", task.Name, wantName)
				}
				if len(task.Commands) == 0 {
					t.Errorf("Parse() task %q has no commands", wantName)
				}
			}
		})
	}
}

func TestParser_ParseFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		filename  string
		content   string
		wantTasks []string
		wantErr   bool
	}{
		{
			name:     "valid babfile",
			filename: "Babfile",
			content: `
build:
  desc: Build
  run: go build

test:
  desc: Test
  run: go test
`,
			wantTasks: []string{"build", "test"},
			wantErr:   false,
		},
		{
			name:     "non-existent file",
			filename: "nonexistent.yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string

			if tt.content != "" {
				// Create test file
				filePath = filepath.Join(tmpDir, tt.filename)
				if err := os.WriteFile(filePath, []byte(tt.content), 0600); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			} else {
				filePath = filepath.Join(tmpDir, tt.filename)
			}

			reg := registry.New()
			p := New(reg)

			err := p.ParseFile(filePath)

			if tt.wantErr {
				if err == nil {
					t.Error("ParseFile() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseFile() unexpected error: %v", err)
			}

			// Verify tasks
			tasks := reg.List()
			if len(tasks) != len(tt.wantTasks) {
				t.Errorf("ParseFile() registered %d tasks, want %d", len(tasks), len(tt.wantTasks))
			}
		})
	}
}

func TestParser_parseNode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTasks map[string][]string // task name -> commands
		wantErr   bool
	}{
		{
			name: "simple task node",
			input: `
build:
  desc: Build
  run: go build
`,
			wantTasks: map[string][]string{
				"build": {"go build"},
			},
			wantErr: false,
		},
		{
			name: "task with multiple commands",
			input: `
deploy:
  run:
    - build
    - test
    - deploy
`,
			wantTasks: map[string][]string{
				"deploy": {"build", "test", "deploy"},
			},
			wantErr: false,
		},
		{
			name: "deeply nested tasks",
			input: `
dev:
  server:
    start:
      desc: Start server
      run: npm start
`,
			wantTasks: map[string][]string{
				"dev:server:start": {"npm start"},
			},
			wantErr: false,
		},
		{
			name: "very deep nesting (4 levels)",
			input: `
build:
  platforms:
    linux:
      amd64:
        desc: Build for Linux AMD64
        run: GOOS=linux GOARCH=amd64 go build
`,
			wantTasks: map[string][]string{
				"build:platforms:linux:amd64": {"GOOS=linux GOARCH=amd64 go build"},
			},
			wantErr: false,
		},
		{
			name: "mixed depth nesting",
			input: `
test:
  unit:
    desc: Unit tests
    run: go test ./...
  integration:
    api:
      desc: API integration tests
      run: go test -tags=integration ./tests/api
`,
			wantTasks: map[string][]string{
				"test:unit":            {"go test ./..."},
				"test:integration:api": {"go test -tags=integration ./tests/api"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := registry.New()
			p := New(reg)

			reader := strings.NewReader(tt.input)
			err := p.Parse(reader)

			if tt.wantErr {
				if err == nil {
					t.Error("parseNode() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseNode() unexpected error: %v", err)
			}

			// Verify tasks and their commands
			for taskName, wantCmds := range tt.wantTasks {
				task, err := reg.Get(taskName)
				if err != nil {
					t.Errorf("parseNode() task %q not found: %v", taskName, err)
					continue
				}

				if len(task.Commands) != len(wantCmds) {
					t.Errorf("parseNode() task %q has %d commands, want %d",
						taskName, len(task.Commands), len(wantCmds))
					continue
				}

				for i, wantCmd := range wantCmds {
					if task.Commands[i] != wantCmd {
						t.Errorf("parseNode() task %q command[%d] = %q, want %q",
							taskName, i, task.Commands[i], wantCmd)
					}
				}
			}
		})
	}
}

func TestIsTask(t *testing.T) {
	tests := []struct {
		name string
		node map[string]interface{}
		want bool
	}{
		{
			name: "has desc only",
			node: map[string]interface{}{
				"desc": "description",
			},
			want: true,
		},
		{
			name: "has run only",
			node: map[string]interface{}{
				"run": "command",
			},
			want: true,
		},
		{
			name: "has both desc and run",
			node: map[string]interface{}{
				"desc": "description",
				"run":  "command",
			},
			want: true,
		},
		{
			name: "has neither desc nor run",
			node: map[string]interface{}{
				"other": "value",
			},
			want: false,
		},
		{
			name: "empty node",
			node: map[string]interface{}{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTask(tt.node)
			if got != tt.want {
				t.Errorf("isTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertMap(t *testing.T) {
	tests := []struct {
		name  string
		input map[interface{}]interface{}
		want  map[string]interface{}
	}{
		{
			name: "simple map",
			input: map[interface{}]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "nested map",
			input: map[interface{}]interface{}{
				"outer": map[interface{}]interface{}{
					"inner": "value",
				},
			},
			want: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			},
		},
		{
			name:  "empty map",
			input: map[interface{}]interface{}{},
			want:  map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMap(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("convertMap() returned map with %d keys, want %d", len(got), len(tt.want))
			}

			for key, wantVal := range tt.want {
				gotVal, exists := got[key]
				if !exists {
					t.Errorf("convertMap() missing key %q", key)
					continue
				}

				// For nested maps, check recursively
				if wantMap, ok := wantVal.(map[string]interface{}); ok {
					gotMap, ok := gotVal.(map[string]interface{})
					if !ok {
						t.Errorf("convertMap() key %q value is not a map", key)
						continue
					}
					if len(gotMap) != len(wantMap) {
						t.Errorf("convertMap() nested map has %d keys, want %d", len(gotMap), len(wantMap))
					}
				} else if gotVal != wantVal {
					t.Errorf("convertMap() key %q = %v, want %v", key, gotVal, wantVal)
				}
			}
		})
	}
}
