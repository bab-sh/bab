package parser

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bab-sh/bab/internal/babfile"
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
	if len(task.Run) != 1 {
		t.Errorf("expected 1 run item, got %d", len(task.Run))
	}
	cmd, ok := task.Run[0].(babfile.CommandRun)
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
	if task.Desc != "Build the project" {
		t.Errorf("expected description 'Build the project', got %q", task.Desc)
	}
	if len(task.Run) != 3 {
		t.Fatalf("expected 3 run items, got %d", len(task.Run))
	}
	expected := []string{`echo "Compiling..."`, `echo "Linking..."`, `echo "Done!"`}
	for i, want := range expected {
		cmd, ok := task.Run[i].(babfile.CommandRun)
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
	if clean == nil || len(clean.Deps) != 0 {
		t.Errorf("clean should have no dependencies")
	}

	build := tasks["build"]
	if build == nil || len(build.Deps) != 1 || build.Deps[0] != "clean" {
		t.Errorf("build should depend on 'clean', got %v", build.Deps)
	}

	test := tasks["test"]
	if test == nil || len(test.Deps) != 1 || test.Deps[0] != "build" {
		t.Errorf("test should depend on 'build', got %v", test.Deps)
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
	if full == nil || len(full.Deps) != 2 {
		t.Errorf("ci:full should have 2 dependencies, got %v", full)
	}
}

func TestParseDescriptions(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "with_description.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	test := tasks["test"]
	if test == nil || test.Desc != "Run all unit tests" {
		t.Errorf("test description wrong: %v", test)
	}

	coverage := tasks["coverage"]
	if coverage == nil || coverage.Desc != "Generate coverage report" {
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
	if !errors.Is(err, ErrInvalidYAML) {
		t.Errorf("expected ErrInvalidYAML, got: %v", err)
	}
}

func TestParseInvalidDependency(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "invalid_deps.yml"))
	if err == nil {
		t.Fatal("expected error for invalid dependency")
	}
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got: %v", err)
	}
}

func TestParseNonexistentFile(t *testing.T) {
	_, err := Parse("nonexistent.yml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("expected ErrFileNotFound, got: %v", err)
	}
}

func TestParseEmptyPath(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !errors.Is(err, ErrPathEmpty) {
		t.Errorf("expected ErrPathEmpty, got: %v", err)
	}
}

func TestParseWhitespacePath(t *testing.T) {
	_, err := Parse("   ")
	if err == nil {
		t.Fatal("expected error for whitespace path")
	}
	if !errors.Is(err, ErrPathEmpty) {
		t.Errorf("expected ErrPathEmpty, got: %v", err)
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

	if len(all.Deps) != 2 {
		t.Errorf("expected 2 dependencies, got %d: %v", len(all.Deps), all.Deps)
	}

	hasDep := func(deps []string, name string) bool {
		for _, d := range deps {
			if d == name {
				return true
			}
		}
		return false
	}

	if !hasDep(all.Deps, "setup") {
		t.Error("missing dependency 'setup'")
	}
	if !hasDep(all.Deps, "gen:build") {
		t.Error("missing dependency 'gen:build'")
	}
}

func TestParseCircularInclude(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "includes", "circular_a.yml"))
	if err == nil {
		t.Fatal("expected error for circular include")
	}
	if !errors.Is(err, ErrCircularDep) {
		t.Errorf("expected ErrCircularDep, got: %v", err)
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
	if len(main.Run) != 3 {
		t.Fatalf("expected 3 run items, got %d", len(main.Run))
	}

	cmd, ok := main.Run[0].(babfile.CommandRun)
	if !ok {
		t.Fatal("expected CommandRun for first item")
	}
	if cmd.Cmd != `echo "main command"` {
		t.Errorf("unexpected command: %q", cmd.Cmd)
	}

	taskRef, ok := main.Run[1].(babfile.TaskRun)
	if !ok {
		t.Fatal("expected TaskRun for second item")
	}
	if taskRef.Task != "helper" {
		t.Errorf("expected task ref 'helper', got %q", taskRef.Task)
	}

	taskRef2, ok := main.Run[2].(babfile.TaskRun)
	if !ok {
		t.Fatal("expected TaskRun for third item")
	}
	if taskRef2.Task != "nested:task" {
		t.Errorf("expected task ref 'nested:task', got %q", taskRef2.Task)
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
	if !errors.Is(err, ErrCircularDep) {
		t.Errorf("expected ErrCircularDep, got: %v", err)
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
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got: %v", err)
	}
}

func TestParseDeepNestedInclude(t *testing.T) {
	tasks, err := Parse(filepath.Join("testdata", "includes", "deep", "main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	expectedTasks := []string{
		"root",
		"first:build",
		"first:second:compile",
		"first:second:test",
	}

	if len(tasks) != len(expectedTasks) {
		t.Errorf("expected %d tasks, got %d: %v", len(expectedTasks), len(tasks), tasks.Names())
	}

	for _, name := range expectedTasks {
		if tasks[name] == nil {
			t.Errorf("expected task %q not found", name)
		}
	}

	firstBuild := tasks["first:build"]
	if firstBuild == nil {
		t.Fatal("task 'first:build' not found")
	}
	if len(firstBuild.Deps) != 1 || firstBuild.Deps[0] != "first:second:compile" {
		t.Errorf("expected dep 'first:second:compile', got %v", firstBuild.Deps)
	}

	if len(firstBuild.Run) < 2 {
		t.Fatalf("expected at least 2 run items, got %d", len(firstBuild.Run))
	}
	taskRun, ok := firstBuild.Run[1].(babfile.TaskRun)
	if !ok {
		t.Fatal("expected second run item to be TaskRun")
	}
	if taskRun.Task != "first:second:test" {
		t.Errorf("expected task ref 'first:second:test', got %q", taskRun.Task)
	}
}
