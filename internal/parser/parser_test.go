package parser

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/errs"
)

func TestParseSimpleTask(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "simple.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(result.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(result.Tasks))
	}
	task := result.Tasks["hello"]
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
	result, err := Parse(filepath.Join("testdata", "multi_command.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	task := result.Tasks["build"]
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
	result, err := Parse(filepath.Join("testdata", "dependencies.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(result.Tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(result.Tasks))
	}

	clean := result.Tasks["clean"]
	if clean == nil || len(clean.Deps) != 0 {
		t.Errorf("clean should have no dependencies")
	}

	build := result.Tasks["build"]
	if build == nil || len(build.Deps) != 1 || build.Deps[0] != "clean" {
		t.Errorf("build should depend on 'clean', got %v", build.Deps)
	}

	test := result.Tasks["test"]
	if test == nil || len(test.Deps) != 1 || test.Deps[0] != "build" {
		t.Errorf("test should depend on 'build', got %v", test.Deps)
	}
}

func TestParseNestedTasks(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "nested.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(result.Tasks) != 4 {
		t.Errorf("expected 4 tasks, got %d", len(result.Tasks))
	}

	expected := []string{"ci:test", "ci:lint", "ci:full", "dev:start"}
	for _, name := range expected {
		if result.Tasks[name] == nil {
			t.Errorf("expected task %q not found", name)
		}
	}

	full := result.Tasks["ci:full"]
	if full == nil || len(full.Deps) != 2 {
		t.Errorf("ci:full should have 2 dependencies, got %v", full)
	}
}

func TestParseDescriptions(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "with_description.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	test := result.Tasks["test"]
	if test == nil || test.Desc != "Run all unit tests" {
		t.Errorf("test description wrong: %v", test)
	}

	coverage := result.Tasks["coverage"]
	if coverage == nil || coverage.Desc != "Generate coverage report" {
		t.Errorf("coverage description wrong: %v", coverage)
	}
}

func TestParseEmptyTasks(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "empty.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(result.Tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(result.Tasks))
	}
}

func TestParseInvalidYAML(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "invalid_yaml.yml"))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !errors.Is(err, errs.ErrInvalidYAML) {
		t.Errorf("expected errs.ErrInvalidYAML, got: %v", err)
	}
}

func TestParseInvalidDependency(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "invalid_deps.yml"))
	if err == nil {
		t.Fatal("expected error for invalid dependency")
	}
	if !errors.Is(err, errs.ErrTaskNotFound) {
		t.Errorf("expected errs.ErrTaskNotFound, got: %v", err)
	}
}

func TestParseMultipleErrors(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "multiple_errors.yml"))
	if err == nil {
		t.Fatal("expected error")
	}

	var verrs *errs.ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T: %v", err, err)
	}

	if len(verrs.Errors) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(verrs.Errors), verrs.Errors)
	}
}

func TestParseNonexistentFile(t *testing.T) {
	_, err := Parse("nonexistent.yml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !errors.Is(err, errs.ErrFileNotFound) {
		t.Errorf("expected errs.ErrFileNotFound, got: %v", err)
	}
}

func TestParseEmptyPath(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	if !errors.Is(err, errs.ErrPathEmpty) {
		t.Errorf("expected errs.ErrPathEmpty, got: %v", err)
	}
}

func TestParseWhitespacePath(t *testing.T) {
	_, err := Parse("   ")
	if err == nil {
		t.Fatal("expected error for whitespace path")
	}
	if !errors.Is(err, errs.ErrPathEmpty) {
		t.Errorf("expected errs.ErrPathEmpty, got: %v", err)
	}
}

func TestParseSimpleInclude(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(result.Tasks) != 4 {
		t.Errorf("expected 4 tasks, got %d: %v", len(result.Tasks), result.Tasks.Names())
	}

	if result.Tasks["setup"] == nil {
		t.Error("local task 'setup' not found")
	}
	if result.Tasks["all"] == nil {
		t.Error("local task 'all' not found")
	}
	if result.Tasks["gen:build"] == nil {
		t.Error("included task 'gen:build' not found")
	}
	if result.Tasks["gen:test"] == nil {
		t.Error("included task 'gen:test' not found")
	}
}

func TestParseIncludeWithDeps(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	all := result.Tasks["all"]
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
	if !errors.Is(err, errs.ErrCircularDep) {
		t.Errorf("expected errs.ErrCircularDep, got: %v", err)
	}
}

func TestParseNestedTaskNames(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "nested.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result.Tasks["docker:build"] == nil {
		t.Error("task 'docker:build' not found")
	}
	if result.Tasks["docker:push"] == nil {
		t.Error("task 'docker:push' not found")
	}
}

func TestParseTaskRun(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "task_run.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(result.Tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(result.Tasks))
	}

	main := result.Tasks["main"]
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
	_, err := Parse(filepath.Join("testdata", "task_run_cycle.yml"))
	if err == nil {
		t.Fatal("expected error for circular task run")
	}
	if !errors.Is(err, errs.ErrCircularDep) {
		t.Errorf("expected errs.ErrCircularDep, got: %v", err)
	}
}

func TestParseTaskRunNotFound(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "task_run_notfound.yml"))
	if err == nil {
		t.Fatal("expected error for missing task reference")
	}
	if !errors.Is(err, errs.ErrTaskNotFound) {
		t.Errorf("expected errs.ErrTaskNotFound, got: %v", err)
	}
}

func TestParseDeepNestedInclude(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "deep", "main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	expectedTasks := []string{
		"root",
		"first:build",
		"first:second:compile",
		"first:second:test",
	}

	if len(result.Tasks) != len(expectedTasks) {
		t.Errorf("expected %d tasks, got %d: %v", len(expectedTasks), len(result.Tasks), result.Tasks.Names())
	}

	for _, name := range expectedTasks {
		if result.Tasks[name] == nil {
			t.Errorf("expected task %q not found", name)
		}
	}

	firstBuild := result.Tasks["first:build"]
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

func TestParseGlobalEnv(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "env_global.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(result.GlobalEnv) != 3 {
		t.Errorf("expected 3 global env vars, got %d", len(result.GlobalEnv))
	}
	if result.GlobalEnv["NODE_ENV"] != "production" {
		t.Errorf("NODE_ENV = %q, want %q", result.GlobalEnv["NODE_ENV"], "production")
	}
	if result.GlobalEnv["PORT"] != "3000" {
		t.Errorf("PORT = %q, want %q", result.GlobalEnv["PORT"], "3000")
	}
	if result.GlobalEnv["DEBUG"] != "false" {
		t.Errorf("DEBUG = %q, want %q", result.GlobalEnv["DEBUG"], "false")
	}
}

func TestParseTaskEnv(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "env_task.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["build"]
	if task == nil {
		t.Fatal("task 'build' not found")
	}
	if len(task.Env) != 2 {
		t.Errorf("expected 2 task env vars, got %d", len(task.Env))
	}
	if task.Env["BUILD_TYPE"] != "release" {
		t.Errorf("BUILD_TYPE = %q, want %q", task.Env["BUILD_TYPE"], "release")
	}
	if task.Env["OPTIMIZE"] != "true" {
		t.Errorf("OPTIMIZE = %q, want %q", task.Env["OPTIMIZE"], "true")
	}
}

func TestParseCommandEnv(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "env_command.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["build"]
	if task == nil {
		t.Fatal("task 'build' not found")
	}
	if len(task.Run) != 2 {
		t.Fatalf("expected 2 run items, got %d", len(task.Run))
	}

	cmd1, ok := task.Run[0].(babfile.CommandRun)
	if !ok {
		t.Fatal("expected CommandRun for first item")
	}
	if len(cmd1.Env) != 2 {
		t.Errorf("expected 2 command env vars, got %d", len(cmd1.Env))
	}
	if cmd1.Env["DEBUG"] != "true" {
		t.Errorf("DEBUG = %q, want %q", cmd1.Env["DEBUG"], "true")
	}
	if cmd1.Env["LOG_LEVEL"] != "verbose" {
		t.Errorf("LOG_LEVEL = %q, want %q", cmd1.Env["LOG_LEVEL"], "verbose")
	}

	cmd2, ok := task.Run[1].(babfile.CommandRun)
	if !ok {
		t.Fatal("expected CommandRun for second item")
	}
	if len(cmd2.Env) != 0 {
		t.Errorf("expected 0 command env vars for second command, got %d", len(cmd2.Env))
	}
}

func TestParseEnvInheritance(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "env_inheritance.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result.GlobalEnv["LEVEL"] != "global" {
		t.Errorf("global LEVEL = %q, want %q", result.GlobalEnv["LEVEL"], "global")
	}
	if result.GlobalEnv["GLOBAL_ONLY"] != "global-only" {
		t.Errorf("GLOBAL_ONLY = %q, want %q", result.GlobalEnv["GLOBAL_ONLY"], "global-only")
	}

	task := result.Tasks["test"]
	if task == nil {
		t.Fatal("task 'test' not found")
	}
	if task.Env["LEVEL"] != "task" {
		t.Errorf("task LEVEL = %q, want %q", task.Env["LEVEL"], "task")
	}
	if task.Env["TASK_ONLY"] != "task-only" {
		t.Errorf("TASK_ONLY = %q, want %q", task.Env["TASK_ONLY"], "task-only")
	}

	cmd, ok := task.Run[0].(babfile.CommandRun)
	if !ok {
		t.Fatal("expected CommandRun")
	}
	if cmd.Env["LEVEL"] != "command" {
		t.Errorf("command LEVEL = %q, want %q", cmd.Env["LEVEL"], "command")
	}
	if cmd.Env["CMD_ONLY"] != "cmd-only" {
		t.Errorf("CMD_ONLY = %q, want %q", cmd.Env["CMD_ONLY"], "cmd-only")
	}
}

func TestParseLogSimple(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "log_simple.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["deploy"]
	if task == nil {
		t.Fatal("task 'deploy' not found")
	}
	if len(task.Run) != 3 {
		t.Fatalf("expected 3 run items, got %d", len(task.Run))
	}

	log1, ok := task.Run[0].(babfile.LogRun)
	if !ok {
		t.Fatal("expected LogRun for first item")
	}
	if log1.Log != "Starting deployment..." {
		t.Errorf("unexpected message: %q", log1.Log)
	}
	if log1.Level != babfile.LogLevelInfo {
		t.Errorf("expected info level (default), got %q", log1.Level)
	}

	_, ok = task.Run[1].(babfile.CommandRun)
	if !ok {
		t.Fatal("expected CommandRun for second item")
	}

	log2, ok := task.Run[2].(babfile.LogRun)
	if !ok {
		t.Fatal("expected LogRun for third item")
	}
	if log2.Log != "Deployment complete!" {
		t.Errorf("unexpected message: %q", log2.Log)
	}
}

func TestParseLogExpanded(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "log_expanded.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["build"]
	if task == nil {
		t.Fatal("task 'build' not found")
	}
	if len(task.Run) != 4 {
		t.Fatalf("expected 4 run items, got %d", len(task.Run))
	}

	log1, ok := task.Run[0].(babfile.LogRun)
	if !ok {
		t.Fatal("expected LogRun for first item")
	}
	if log1.Log != "Build started" {
		t.Errorf("unexpected message: %q", log1.Log)
	}
	if log1.Level != babfile.LogLevelInfo {
		t.Errorf("expected info level, got %q", log1.Level)
	}

	log2, ok := task.Run[2].(babfile.LogRun)
	if !ok {
		t.Fatal("expected LogRun for third item")
	}
	if log2.Level != babfile.LogLevelDebug {
		t.Errorf("expected debug level, got %q", log2.Level)
	}

	log3, ok := task.Run[3].(babfile.LogRun)
	if !ok {
		t.Fatal("expected LogRun for fourth item")
	}
	if log3.Level != babfile.LogLevelWarn {
		t.Errorf("expected warn level, got %q", log3.Level)
	}
}

func TestParseLogPlatforms(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "log_platforms.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["setup"]
	if task == nil {
		t.Fatal("task 'setup' not found")
	}
	if len(task.Run) != 3 {
		t.Fatalf("expected 3 run items, got %d", len(task.Run))
	}

	log1, ok := task.Run[0].(babfile.LogRun)
	if !ok {
		t.Fatal("expected LogRun for first item")
	}
	if len(log1.Platforms) != 1 || log1.Platforms[0] != babfile.PlatformDarwin {
		t.Errorf("expected darwin platform, got %v", log1.Platforms)
	}

	log2, ok := task.Run[1].(babfile.LogRun)
	if !ok {
		t.Fatal("expected LogRun for second item")
	}
	if len(log2.Platforms) != 1 || log2.Platforms[0] != babfile.PlatformLinux {
		t.Errorf("expected linux platform, got %v", log2.Platforms)
	}

	log3, ok := task.Run[2].(babfile.LogRun)
	if !ok {
		t.Fatal("expected LogRun for third item")
	}
	if len(log3.Platforms) != 0 {
		t.Errorf("expected no platforms, got %v", log3.Platforms)
	}
}

func TestParseLogInvalidLevel(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "log_invalid_level.yml"))
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestParseLogWithCmd(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "log_with_cmd.yml"))
	if err == nil {
		t.Fatal("expected error for log with cmd")
	}
}

func TestParseLogAllLevels(t *testing.T) {
	levels := []struct {
		file     string
		expected babfile.LogLevel
	}{
		{"log_level_debug.yml", babfile.LogLevelDebug},
		{"log_level_info.yml", babfile.LogLevelInfo},
		{"log_level_warn.yml", babfile.LogLevelWarn},
		{"log_level_error.yml", babfile.LogLevelError},
	}

	for _, tc := range levels {
		t.Run(tc.file, func(t *testing.T) {
			result, err := Parse(filepath.Join("testdata", tc.file))
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			task := result.Tasks["test"]
			if task == nil {
				t.Fatal("task not found")
			}

			logRun, ok := task.Run[0].(babfile.LogRun)
			if !ok {
				t.Fatal("expected LogRun")
			}
			if logRun.Level != tc.expected {
				t.Errorf("expected level %q, got %q", tc.expected, logRun.Level)
			}
		})
	}
}

func TestParseSilentGlobal(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "silent_global.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result.GlobalSilent == nil {
		t.Fatal("expected GlobalSilent to be set")
	}
	if *result.GlobalSilent != true {
		t.Errorf("expected GlobalSilent = true, got %v", *result.GlobalSilent)
	}
}

func TestParseSilentTask(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "silent_task.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["hello"]
	if task == nil {
		t.Fatal("task 'hello' not found")
	}
	if task.Silent == nil {
		t.Fatal("expected Silent to be set on task")
	}
	if *task.Silent != true {
		t.Errorf("expected Silent = true, got %v", *task.Silent)
	}
}

func TestParseSilentRunItem(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		taskName string
		checkFn  func(t *testing.T, item babfile.RunItem)
	}{
		{
			name:     "command",
			file:     "silent_command.yml",
			taskName: "hello",
			checkFn: func(t *testing.T, item babfile.RunItem) {
				cmd, ok := item.(babfile.CommandRun)
				if !ok {
					t.Fatal("expected CommandRun")
				}
				if cmd.Silent == nil || !*cmd.Silent {
					t.Error("expected Silent = true")
				}
			},
		},
		{
			name:     "task_ref",
			file:     "silent_taskrun.yml",
			taskName: "main",
			checkFn: func(t *testing.T, item babfile.RunItem) {
				taskRef, ok := item.(babfile.TaskRun)
				if !ok {
					t.Fatal("expected TaskRun")
				}
				if taskRef.Silent == nil || !*taskRef.Silent {
					t.Error("expected Silent = true")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Parse(filepath.Join("testdata", tc.file))
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}

			task := result.Tasks[tc.taskName]
			if task == nil {
				t.Fatalf("task %q not found", tc.taskName)
			}
			if len(task.Run) < 1 {
				t.Fatal("expected at least 1 run item")
			}

			tc.checkFn(t, task.Run[0])
		})
	}
}

func TestParseSilentFalse(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "silent_false.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result.GlobalSilent == nil {
		t.Fatal("expected GlobalSilent to be set")
	}
	if *result.GlobalSilent != false {
		t.Errorf("expected GlobalSilent = false, got %v", *result.GlobalSilent)
	}

	task := result.Tasks["hello"]
	if task == nil {
		t.Fatal("task 'hello' not found")
	}
	if task.Silent == nil || *task.Silent != false {
		t.Errorf("expected task Silent = false")
	}

	cmd, ok := task.Run[0].(babfile.CommandRun)
	if !ok {
		t.Fatal("expected CommandRun")
	}
	if cmd.Silent == nil || *cmd.Silent != false {
		t.Errorf("expected command Silent = false")
	}
}
