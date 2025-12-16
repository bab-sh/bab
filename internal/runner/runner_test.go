package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bab-sh/bab/internal/babfile"
)

func TestRunSimpleTask(t *testing.T) {
	tasks := babfile.TaskMap{
		"hello": &babfile.Task{
			Name: "hello",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo hello"}},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "hello", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunWithDependencies(t *testing.T) {
	tasks := babfile.TaskMap{
		"build": &babfile.Task{
			Name: "build",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo building"}},
		},
		"test": &babfile.Task{
			Name: "test",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo testing"}},
			Deps: []string{"build"},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "test", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunTaskNotFound(t *testing.T) {
	tasks := babfile.TaskMap{
		"hello": &babfile.Task{
			Name: "hello",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo hello"}},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "nonexistent", tasks)
	if err == nil {
		t.Fatal("expected error for nonexistent task")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestRunCircularDependency(t *testing.T) {
	tasks := babfile.TaskMap{
		"a": &babfile.Task{
			Name: "a",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo a"}},
			Deps: []string{"b"},
		},
		"b": &babfile.Task{
			Name: "b",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo b"}},
			Deps: []string{"a"},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "a", tasks)
	if err == nil {
		t.Fatal("expected error for circular dependency")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected 'circular' error, got: %v", err)
	}
}

func TestLoadTasks(t *testing.T) {
	tmpDir := t.TempDir()
	babfilePath := filepath.Join(tmpDir, "Babfile.yml")

	yaml := `tasks:
  hello:
    run:
      - cmd: echo "Hello"
  world:
    run:
      - cmd: echo "World"`

	if err := os.WriteFile(babfilePath, []byte(yaml), 0600); err != nil {
		t.Fatalf("failed to create Babfile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	result, err := LoadTasks("")
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}
	if len(result.Tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(result.Tasks))
	}
}

func TestLoadTasksWithCustomPath(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "custom-tasks.yml")

	yaml := `tasks:
  custom:
    run:
      - cmd: echo "Custom"`

	if err := os.WriteFile(customPath, []byte(yaml), 0600); err != nil {
		t.Fatalf("failed to create custom Babfile: %v", err)
	}

	result, err := LoadTasks(customPath)
	if err != nil {
		t.Fatalf("LoadTasks(customPath) error: %v", err)
	}
	if len(result.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(result.Tasks))
	}
	if _, ok := result.Tasks["custom"]; !ok {
		t.Error("expected 'custom' task to exist")
	}
}

func TestLoadTasksWithNonexistentPath(t *testing.T) {
	_, err := LoadTasks("/nonexistent/path/Babfile.yml")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestNew(t *testing.T) {
	r := New(false, "")
	if r.DryRun {
		t.Error("DryRun should be false")
	}
	if r.Babfile != "" {
		t.Error("Babfile should be empty")
	}

	r = New(true, "/custom/path")
	if !r.DryRun {
		t.Error("DryRun should be true")
	}
	if r.Babfile != "/custom/path" {
		t.Errorf("Babfile = %q, want %q", r.Babfile, "/custom/path")
	}
}

func TestRunTaskWithTaskRef(t *testing.T) {
	tasks := babfile.TaskMap{
		"main": &babfile.Task{
			Name: "main",
			Run: []babfile.RunItem{
				babfile.CommandRun{Cmd: "echo main"},
				babfile.TaskRun{Task: "helper"},
			},
		},
		"helper": &babfile.Task{
			Name: "helper",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo helper"}},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "main", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunTaskRefCircular(t *testing.T) {
	tasks := babfile.TaskMap{
		"a": &babfile.Task{
			Name: "a",
			Run:  []babfile.RunItem{babfile.TaskRun{Task: "b"}},
		},
		"b": &babfile.Task{
			Name: "b",
			Run:  []babfile.RunItem{babfile.TaskRun{Task: "a"}},
		},
	}

	r := New(false, "")
	err := r.RunWithTasks(context.Background(), "a", tasks)
	if err == nil {
		t.Fatal("expected error for circular task run")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected 'circular' error, got: %v", err)
	}
}

func TestRunWithGlobalEnv(t *testing.T) {
	tasks := babfile.TaskMap{
		"hello": &babfile.Task{
			Name: "hello",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo $FOO"}},
		},
	}

	r := New(true, "")
	r.GlobalEnv = map[string]string{"FOO": "bar"}
	err := r.RunWithTasks(context.Background(), "hello", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunWithTaskEnv(t *testing.T) {
	tasks := babfile.TaskMap{
		"hello": &babfile.Task{
			Name: "hello",
			Env:  map[string]string{"FOO": "task-value"},
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo $FOO"}},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "hello", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunWithCommandEnv(t *testing.T) {
	tasks := babfile.TaskMap{
		"hello": &babfile.Task{
			Name: "hello",
			Run: []babfile.RunItem{
				babfile.CommandRun{
					Cmd: "echo $FOO",
					Env: map[string]string{"FOO": "cmd-value"},
				},
			},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "hello", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunEnvPrecedence(t *testing.T) {
	tasks := babfile.TaskMap{
		"test": &babfile.Task{
			Name: "test",
			Env:  map[string]string{"FOO": "task", "BAR": "task"},
			Run: []babfile.RunItem{
				babfile.CommandRun{
					Cmd: "echo $FOO $BAR $BAZ",
					Env: map[string]string{"FOO": "cmd"},
				},
			},
		},
	}

	r := New(true, "")
	r.GlobalEnv = map[string]string{"FOO": "global", "BAR": "global", "BAZ": "global"}
	err := r.RunWithTasks(context.Background(), "test", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunWithLogItem(t *testing.T) {
	tasks := babfile.TaskMap{
		"deploy": &babfile.Task{
			Name: "deploy",
			Run: []babfile.RunItem{
				babfile.LogRun{Log: "Starting deployment...", Level: babfile.LogLevelInfo},
				babfile.CommandRun{Cmd: "echo deploying"},
				babfile.LogRun{Log: "Deployment complete!", Level: babfile.LogLevelInfo},
			},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "deploy", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunWithLogAllLevels(t *testing.T) {
	tasks := babfile.TaskMap{
		"test": &babfile.Task{
			Name: "test",
			Run: []babfile.RunItem{
				babfile.LogRun{Log: "Debug message", Level: babfile.LogLevelDebug},
				babfile.LogRun{Log: "Info message", Level: babfile.LogLevelInfo},
				babfile.LogRun{Log: "Warning message", Level: babfile.LogLevelWarn},
				babfile.LogRun{Log: "Error message", Level: babfile.LogLevelError},
			},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "test", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunLogPlatformSkip(t *testing.T) {
	tasks := babfile.TaskMap{
		"test": &babfile.Task{
			Name: "test",
			Run: []babfile.RunItem{
				babfile.LogRun{
					Log:       "Platform-specific log",
					Level:     babfile.LogLevelInfo,
					Platforms: []babfile.Platform{"nonexistent"},
				},
			},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "test", tasks)
	if err == nil {
		t.Fatal("expected error when no run items match platform")
	}
	if !strings.Contains(err.Error(), "no run items for platform") {
		t.Errorf("expected 'no run items for platform' error, got: %v", err)
	}
}

func TestIsSilent(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name     string
		item     *bool
		task     *bool
		global   *bool
		expected bool
	}{
		{"all nil", nil, nil, nil, false},
		{"global true", nil, nil, &trueVal, true},
		{"global false", nil, nil, &falseVal, false},
		{"task true overrides global false", nil, &trueVal, &falseVal, true},
		{"task false overrides global true", nil, &falseVal, &trueVal, false},
		{"item true overrides task false", &trueVal, &falseVal, nil, true},
		{"item false overrides task true", &falseVal, &trueVal, nil, false},
		{"item true overrides all false", &trueVal, &falseVal, &falseVal, true},
		{"item false overrides all true", &falseVal, &trueVal, &trueVal, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSilent(tt.item, tt.task, tt.global)
			if result != tt.expected {
				t.Errorf("isSilent(%v, %v, %v) = %v, want %v",
					tt.item, tt.task, tt.global, result, tt.expected)
			}
		})
	}
}

func TestRunWithGlobalSilent(t *testing.T) {
	trueVal := true
	tasks := babfile.TaskMap{
		"hello": &babfile.Task{
			Name: "hello",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo hello"}},
		},
	}

	r := New(true, "")
	r.GlobalSilent = &trueVal
	err := r.RunWithTasks(context.Background(), "hello", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunWithTaskSilent(t *testing.T) {
	trueVal := true
	tasks := babfile.TaskMap{
		"hello": &babfile.Task{
			Name:   "hello",
			Silent: &trueVal,
			Run:    []babfile.RunItem{babfile.CommandRun{Cmd: "echo hello"}},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "hello", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunWithCommandSilent(t *testing.T) {
	trueVal := true
	tasks := babfile.TaskMap{
		"hello": &babfile.Task{
			Name: "hello",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo hello", Silent: &trueVal}},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "hello", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestRunWithTaskRefSilent(t *testing.T) {
	trueVal := true
	tasks := babfile.TaskMap{
		"main": &babfile.Task{
			Name: "main",
			Run:  []babfile.RunItem{babfile.TaskRun{Task: "helper", Silent: &trueVal}},
		},
		"helper": &babfile.Task{
			Name: "helper",
			Run:  []babfile.RunItem{babfile.CommandRun{Cmd: "echo helper"}},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "main", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}

func TestSilentInheritance(t *testing.T) {
	trueVal := true
	falseVal := false
	tasks := babfile.TaskMap{
		"test": &babfile.Task{
			Name:   "test",
			Silent: &trueVal,
			Run: []babfile.RunItem{
				babfile.CommandRun{Cmd: "echo cmd1", Silent: &falseVal},
				babfile.CommandRun{Cmd: "echo cmd2"},
			},
		},
	}

	r := New(true, "")
	err := r.RunWithTasks(context.Background(), "test", tasks)
	if err != nil {
		t.Errorf("RunWithTasks() error: %v", err)
	}
}
