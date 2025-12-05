package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSimpleTask(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "simple.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
	task := tasks["hello"]
	if task == nil {
		t.Fatal("task 'hello' not found")
	}
	if task.Name != "hello" {
		t.Errorf("expected name 'hello', got %q", task.Name)
	}
	if len(task.RunItems) != 1 {
		t.Errorf("expected 1 run item, got %d", len(task.RunItems))
	}
	cmd, ok := task.RunItems[0].(CommandRun)
	if !ok {
		t.Fatal("expected CommandRun")
	}
	if cmd.Cmd != `echo "Hello World"` {
		t.Errorf("unexpected command: %q", cmd.Cmd)
	}
}

func TestParseMultiCommand(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "multi_command.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	task := tasks["build"]
	if task == nil {
		t.Fatal("task 'build' not found")
	}
	if task.Description != "Build the project" {
		t.Errorf("expected description 'Build the project', got %q", task.Description)
	}
	if len(task.RunItems) != 3 {
		t.Fatalf("expected 3 run items, got %d", len(task.RunItems))
	}
	expected := []string{`echo "Compiling..."`, `echo "Linking..."`, `echo "Done!"`}
	for i, want := range expected {
		cmd, ok := task.RunItems[i].(CommandRun)
		if !ok {
			t.Fatalf("run item[%d]: expected CommandRun", i)
		}
		if cmd.Cmd != want {
			t.Errorf("command[%d]: expected %q, got %q", i, want, cmd.Cmd)
		}
	}
}

func TestParseDependencies(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "dependencies.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}

	clean := tasks["clean"]
	if clean == nil || len(clean.Dependencies) != 0 {
		t.Errorf("clean should have no dependencies")
	}

	build := tasks["build"]
	if build == nil || len(build.Dependencies) != 1 || build.Dependencies[0] != "clean" {
		t.Errorf("build should depend on 'clean', got %v", build.Dependencies)
	}

	test := tasks["test"]
	if test == nil || len(test.Dependencies) != 1 || test.Dependencies[0] != "build" {
		t.Errorf("test should depend on 'build', got %v", test.Dependencies)
	}
}

func TestParseNestedTasks(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "nested.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(tasks) != 4 {
		t.Errorf("expected 4 tasks, got %d", len(tasks))
	}

	expected := []string{"ci:test", "ci:lint", "ci:full", "dev:start"}
	for _, name := range expected {
		if tasks[name] == nil {
			t.Errorf("expected task %q not found", name)
		}
	}

	full := tasks["ci:full"]
	if full == nil || len(full.Dependencies) != 2 {
		t.Errorf("ci:full should have 2 dependencies, got %v", full)
	}
}

func TestParseDescriptions(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "with_description.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	test := tasks["test"]
	if test == nil || test.Description != "Run all unit tests" {
		t.Errorf("test description wrong: %v", test)
	}

	coverage := tasks["coverage"]
	if coverage == nil || coverage.Description != "Generate coverage report" {
		t.Errorf("coverage description wrong: %v", coverage)
	}
}

func TestParseEmptyTasks(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "empty.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestParseInvalidYAML(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "invalid_yaml.yml"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "invalid YAML") {
		t.Errorf("expected 'invalid YAML' error, got: %v", err)
	}
}

func TestParseInvalidDependency(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "invalid_deps.yml"))
	if err == nil {
		t.Fatal("expected error for invalid dependency")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestParseNonexistentFile(t *testing.T) {
	_, err := Parse("nonexistent.yml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("expected 'failed to read file' error, got: %v", err)
	}
}

func TestParseEmptyPath(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' error, got: %v", err)
	}
}

func TestParseWhitespacePath(t *testing.T) {
	_, err := Parse("   ")
	if err == nil {
		t.Fatal("expected error for whitespace path")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' error, got: %v", err)
	}
}

func TestParseSimpleInclude(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "includes", "main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(tasks) != 4 {
		t.Errorf("expected 4 tasks, got %d: %v", len(tasks), tasks.Names())
	}

	if tasks["setup"] == nil {
		t.Error("local task 'setup' not found")
	}
	if tasks["all"] == nil {
		t.Error("local task 'all' not found")
	}
	if tasks["gen:build"] == nil {
		t.Error("included task 'gen:build' not found")
	}
	if tasks["gen:test"] == nil {
		t.Error("included task 'gen:test' not found")
	}
}

func TestParseIncludeWithDeps(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "includes", "main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	all := tasks["all"]
	if all == nil {
		t.Fatal("task 'all' not found")
	}

	if len(all.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d: %v", len(all.Dependencies), all.Dependencies)
	}

	hasDep := func(deps []string, name string) bool {
		for _, d := range deps {
			if d == name {
				return true
			}
		}
		return false
	}

	if !hasDep(all.Dependencies, "setup") {
		t.Error("missing dependency 'setup'")
	}
	if !hasDep(all.Dependencies, "gen:build") {
		t.Error("missing dependency 'gen:build'")
	}
}

func TestParseCircularInclude(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "includes", "circular_a.yml"))
	if err == nil {
		t.Fatal("expected error for circular include")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected 'circular' error, got: %v", err)
	}
}

func TestParseNestedTaskNames(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "includes", "nested.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if tasks["docker:build"] == nil {
		t.Error("task 'docker:build' not found")
	}
	if tasks["docker:push"] == nil {
		t.Error("task 'docker:push' not found")
	}
}

func TestParseTaskRun(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "task_run.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}

	main := tasks["main"]
	if main == nil {
		t.Fatal("task 'main' not found")
	}
	if len(main.RunItems) != 3 {
		t.Fatalf("expected 3 run items, got %d", len(main.RunItems))
	}

	// First item: cmd
	cmd, ok := main.RunItems[0].(CommandRun)
	if !ok {
		t.Fatal("expected CommandRun for first item")
	}
	if cmd.Cmd != `echo "main command"` {
		t.Errorf("unexpected command: %q", cmd.Cmd)
	}

	// Second item: task reference
	taskRef, ok := main.RunItems[1].(TaskRun)
	if !ok {
		t.Fatal("expected TaskRun for second item")
	}
	if taskRef.TaskRef != "helper" {
		t.Errorf("expected task ref 'helper', got %q", taskRef.TaskRef)
	}

	// Third item: task reference with colon
	taskRef2, ok := main.RunItems[2].(TaskRun)
	if !ok {
		t.Fatal("expected TaskRun for third item")
	}
	if taskRef2.TaskRef != "nested:task" {
		t.Errorf("expected task ref 'nested:task', got %q", taskRef2.TaskRef)
	}
}

func TestParseTaskRunCycle(t *testing.T) {
	yaml := `tasks:
  a:
    run:
      - task: b
  b:
    run:
      - task: a`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cycle.yml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Parse(path)
	if err == nil {
		t.Fatal("expected error for circular task run")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("expected 'circular' error, got: %v", err)
	}
}

func TestParseTaskRunNotFound(t *testing.T) {
	yaml := `tasks:
  main:
    run:
      - task: nonexistent`

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "notfound.yml")
	if err := os.WriteFile(path, []byte(yaml), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Parse(path)
	if err == nil {
		t.Fatal("expected error for missing task reference")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}
