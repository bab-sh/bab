package parser

import (
	"path/filepath"
	"testing"
)

func TestParseIncludes(t *testing.T) {
	tests := []struct {
		name     string
		rootMap  map[string]interface{}
		basePath string
		wantLen  int
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "no includes section",
			rootMap:  map[string]interface{}{"task": map[string]interface{}{"run": "echo test"}},
			basePath: "/project/Babfile.yml",
			wantLen:  0,
			wantErr:  false,
		},
		{
			name: "single include",
			rootMap: map[string]interface{}{
				"includes": map[string]interface{}{
					"gen": map[string]interface{}{
						"babfile": "./lib/tasks.yml",
					},
				},
			},
			basePath: "/project/Babfile.yml",
			wantLen:  1,
			wantErr:  false,
		},
		{
			name: "multiple includes",
			rootMap: map[string]interface{}{
				"includes": map[string]interface{}{
					"gen":    map[string]interface{}{"babfile": "./gen.yml"},
					"deploy": map[string]interface{}{"babfile": "./deploy.yml"},
					"test":   map[string]interface{}{"babfile": "./test.yml"},
				},
			},
			basePath: "/project/Babfile.yml",
			wantLen:  3,
			wantErr:  false,
		},
		{
			name: "absolute path",
			rootMap: map[string]interface{}{
				"includes": map[string]interface{}{
					"gen": map[string]interface{}{
						"babfile": "/absolute/path/tasks.yml",
					},
				},
			},
			basePath: "/project/Babfile.yml",
			wantLen:  1,
			wantErr:  false,
		},
		{
			name: "includes not a map",
			rootMap: map[string]interface{}{
				"includes": "not a map",
			},
			basePath: "/project/Babfile.yml",
			wantErr:  true,
			errMsg:   "'includes' must be a map",
		},
		{
			name: "include entry not a map",
			rootMap: map[string]interface{}{
				"includes": map[string]interface{}{
					"gen": "./tasks.yml",
				},
			},
			basePath: "/project/Babfile.yml",
			wantErr:  true,
			errMsg:   "must be a map with 'babfile' key",
		},
		{
			name: "missing babfile key",
			rootMap: map[string]interface{}{
				"includes": map[string]interface{}{
					"gen": map[string]interface{}{
						"path": "./tasks.yml",
					},
				},
			},
			basePath: "/project/Babfile.yml",
			wantErr:  true,
			errMsg:   "missing required 'babfile' key",
		},
		{
			name: "empty babfile value",
			rootMap: map[string]interface{}{
				"includes": map[string]interface{}{
					"gen": map[string]interface{}{
						"babfile": "",
					},
				},
			},
			basePath: "/project/Babfile.yml",
			wantErr:  true,
			errMsg:   "invalid 'babfile' value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			includes, err := parseIncludes(tt.rootMap, tt.basePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseIncludes() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("parseIncludes() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("parseIncludes() unexpected error: %v", err)
				return
			}

			if len(includes) != tt.wantLen {
				t.Errorf("parseIncludes() returned %d includes, want %d", len(includes), tt.wantLen)
			}
		})
	}
}

func TestParseIncludesPathResolution(t *testing.T) {
	rootMap := map[string]interface{}{
		"includes": map[string]interface{}{
			"gen": map[string]interface{}{
				"babfile": "./lib/tasks.yml",
			},
		},
	}

	includes, err := parseIncludes(rootMap, "/project/Babfile.yml")
	if err != nil {
		t.Fatalf("parseIncludes() unexpected error: %v", err)
	}

	path, exists := includes["gen"]
	if !exists {
		t.Fatal("expected 'gen' include not found")
	}

	expected := filepath.Clean("/project/lib/tasks.yml")
	if path != expected {
		t.Errorf("parseIncludes() path = %q, want %q", path, expected)
	}
}

func TestMergeTasks(t *testing.T) {
	tests := []struct {
		name      string
		parent    TaskMap
		included  TaskMap
		namespace string
		wantErr   bool
		errMsg    string
		wantTasks []string
	}{
		{
			name:   "merge single task",
			parent: TaskMap{},
			included: TaskMap{
				"build": &Task{Name: "build", Commands: []string{"go build"}},
			},
			namespace: "gen",
			wantErr:   false,
			wantTasks: []string{"gen:build"},
		},
		{
			name:   "merge multiple tasks",
			parent: TaskMap{},
			included: TaskMap{
				"build": &Task{Name: "build", Commands: []string{"go build"}},
				"test":  &Task{Name: "test", Commands: []string{"go test"}},
			},
			namespace: "gen",
			wantErr:   false,
			wantTasks: []string{"gen:build", "gen:test"},
		},
		{
			name: "merge with existing parent tasks",
			parent: TaskMap{
				"setup": &Task{Name: "setup", Commands: []string{"echo setup"}},
			},
			included: TaskMap{
				"build": &Task{Name: "build", Commands: []string{"go build"}},
			},
			namespace: "gen",
			wantErr:   false,
			wantTasks: []string{"setup", "gen:build"},
		},
		{
			name: "task collision",
			parent: TaskMap{
				"gen:build": &Task{Name: "gen:build", Commands: []string{"existing"}},
			},
			included: TaskMap{
				"build": &Task{Name: "build", Commands: []string{"go build"}},
			},
			namespace: "gen",
			wantErr:   true,
			errMsg:    "task name collision",
		},
		{
			name:      "empty included tasks",
			parent:    TaskMap{"setup": &Task{Name: "setup", Commands: []string{"echo"}}},
			included:  TaskMap{},
			namespace: "gen",
			wantErr:   false,
			wantTasks: []string{"setup"},
		},
		{
			name:   "preserves dependencies",
			parent: TaskMap{},
			included: TaskMap{
				"build": &Task{
					Name:         "build",
					Commands:     []string{"go build"},
					Dependencies: []string{"gen:lint", "setup"},
				},
			},
			namespace: "gen",
			wantErr:   false,
			wantTasks: []string{"gen:build"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mergeTasks(tt.parent, tt.included, tt.namespace)

			if tt.wantErr {
				if err == nil {
					t.Errorf("mergeTasks() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("mergeTasks() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("mergeTasks() unexpected error: %v", err)
				return
			}

			for _, taskName := range tt.wantTasks {
				if _, exists := tt.parent[taskName]; !exists {
					t.Errorf("mergeTasks() expected task %q not found", taskName)
				}
			}
		})
	}
}

func TestMergeTasksPreservesDependencies(t *testing.T) {
	parent := TaskMap{}
	included := TaskMap{
		"build": &Task{
			Name:         "build",
			Commands:     []string{"go build"},
			Dependencies: []string{"gen:lint", "setup"},
		},
	}

	err := mergeTasks(parent, included, "gen")
	if err != nil {
		t.Fatalf("mergeTasks() unexpected error: %v", err)
	}

	task := parent["gen:build"]
	if task == nil {
		t.Fatal("expected task 'gen:build' not found")
		return
	}

	if len(task.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(task.Dependencies))
	}

	expectedDeps := []string{"gen:lint", "setup"}
	for i, dep := range expectedDeps {
		if task.Dependencies[i] != dep {
			t.Errorf("dependency[%d] = %q, want %q", i, task.Dependencies[i], dep)
		}
	}
}

func TestParseWithSingleInclude(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "includes", "main.yml"))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	expectedTasks := []string{"setup", "all", "gen:build", "gen:test"}
	for _, name := range expectedTasks {
		if tasks[name] == nil {
			t.Errorf("expected task %q not found", name)
		}
	}

	if len(tasks) != len(expectedTasks) {
		t.Errorf("expected %d tasks, got %d", len(expectedTasks), len(tasks))
	}
}

func TestParseWithMultipleIncludes(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "includes", "multi.yml"))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	expectedTasks := []string{
		"deploy",
		"gen:build", "gen:test",
		"docker:docker:build", "docker:docker:push",
	}

	for _, name := range expectedTasks {
		if tasks[name] == nil {
			t.Errorf("expected task %q not found", name)
		}
	}
}

func TestParseWithNestedIncludes(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "includes", "recursive.yml"))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	expectedTasks := []string{
		"root",
		"level1:task1",
		"level1:level2:task2",
	}

	for _, name := range expectedTasks {
		if tasks[name] == nil {
			t.Errorf("expected task %q not found", name)
		}
	}
}

func TestParseCircularInclude(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "includes", "circular_a.yml"))
	if err == nil {
		t.Fatal("Parse() expected error for circular include, got nil")
	}

	if !contains(err.Error(), "circular include detected") {
		t.Errorf("Parse() error = %q, want error containing 'circular include detected'", err.Error())
	}
}

func TestParseInvalidIncludes(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		errMsg string
	}{
		{
			name:   "includes not a map",
			file:   filepath.Join("testdata", "includes", "invalid", "not_map.yml"),
			errMsg: "'includes' must be a map",
		},
		{
			name:   "missing babfile key",
			file:   filepath.Join("testdata", "includes", "invalid", "no_babfile.yml"),
			errMsg: "missing required 'babfile' key",
		},
		{
			name:   "empty babfile path",
			file:   filepath.Join("testdata", "includes", "invalid", "empty_path.yml"),
			errMsg: "invalid 'babfile' value",
		},
		{
			name:   "include entry not a map",
			file:   filepath.Join("testdata", "includes", "invalid", "entry_not_map.yml"),
			errMsg: "must be a map with 'babfile' key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.file)
			if err == nil {
				t.Errorf("Parse() expected error containing %q, got nil", tt.errMsg)
				return
			}
			if !contains(err.Error(), tt.errMsg) {
				t.Errorf("Parse() error = %q, want error containing %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestParseIncludeFileNotFound(t *testing.T) {
	rootMap := map[string]interface{}{
		"includes": map[string]interface{}{
			"gen": map[string]interface{}{
				"babfile": "./nonexistent.yml",
			},
		},
		"task": map[string]interface{}{
			"run": "echo test",
		},
	}

	includes, err := parseIncludes(rootMap, "/project/Babfile.yml")
	if err != nil {
		t.Fatalf("parseIncludes() unexpected error: %v", err)
	}

	ctx := NewParseContext()
	ctx.stack = append(ctx.stack, "/project/Babfile.yml")
	tasks := TaskMap{}

	err = resolveInclude("gen", includes["gen"], tasks, ctx)
	if err == nil {
		t.Fatal("resolveInclude() expected error for nonexistent file, got nil")
	}

	if !contains(err.Error(), "failed to parse included babfile") {
		t.Errorf("resolveInclude() error = %q, want error containing 'failed to parse included babfile'", err.Error())
	}
}

func TestNewParseContext(t *testing.T) {
	ctx := NewParseContext()

	if ctx.visited == nil {
		t.Error("NewParseContext() visited map is nil")
	}

	if ctx.stack == nil {
		t.Error("NewParseContext() stack is nil")
	}

	if len(ctx.visited) != 0 {
		t.Errorf("NewParseContext() visited map should be empty, got %d entries", len(ctx.visited))
	}

	if len(ctx.stack) != 0 {
		t.Errorf("NewParseContext() stack should be empty, got %d entries", len(ctx.stack))
	}
}
