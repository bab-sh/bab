package parser

import (
	"errors"
	"path/filepath"
	"strings"
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

	buildTask := result.Tasks["build"]
	if buildTask == nil || len(buildTask.Deps) != 1 || buildTask.Deps[0] != "clean" {
		t.Errorf("build should depend on 'clean', got %v", buildTask.Deps)
	}

	testTask := result.Tasks["test"]
	if testTask == nil || len(testTask.Deps) != 1 || testTask.Deps[0] != buildTask.Name {
		t.Errorf("test should depend on 'build', got %v", testTask.Deps)
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

func TestParseIncludeWithLogAndPrompt(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "main_log_prompt.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result.Tasks["ops:deploy"] == nil {
		t.Fatal("included task 'ops:deploy' not found")
	}
	if result.Tasks["ops:confirm"] == nil {
		t.Fatal("included task 'ops:confirm' not found")
	}

	deploy := result.Tasks["ops:deploy"]
	if len(deploy.Run) != 3 {
		t.Fatalf("expected 3 run items in ops:deploy, got %d", len(deploy.Run))
	}

	log1, ok := deploy.Run[0].(babfile.LogRun)
	if !ok {
		t.Fatal("expected LogRun for first item")
	}
	if log1.Log != "Starting deployment" {
		t.Errorf("unexpected log message: %q", log1.Log)
	}

	log2, ok := deploy.Run[2].(babfile.LogRun)
	if !ok {
		t.Fatal("expected LogRun for third item")
	}
	if log2.Level != babfile.LogLevelWarn {
		t.Errorf("expected warn level, got %q", log2.Level)
	}

	confirm := result.Tasks["ops:confirm"]
	if len(confirm.Run) != 2 {
		t.Fatalf("expected 2 run items in ops:confirm, got %d", len(confirm.Run))
	}

	prompt, ok := confirm.Run[0].(babfile.PromptRun)
	if !ok {
		t.Fatal("expected PromptRun for first item")
	}
	if prompt.Prompt != "proceed" {
		t.Errorf("expected prompt 'proceed', got %q", prompt.Prompt)
	}
	if prompt.Type != babfile.PromptTypeConfirm {
		t.Errorf("expected confirm type, got %q", prompt.Type)
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

func TestParseOutputGlobal(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "output_global.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result.GlobalOutput == nil {
		t.Fatal("expected GlobalOutput to be set")
	}
	if *result.GlobalOutput != false {
		t.Errorf("expected GlobalOutput = false, got %v", *result.GlobalOutput)
	}
}

func TestParseOutputTask(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "output_task.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["hello"]
	if task == nil {
		t.Fatal("task 'hello' not found")
	}
	if task.Output == nil {
		t.Fatal("expected Output to be set on task")
	}
	if *task.Output != false {
		t.Errorf("expected Output = false, got %v", *task.Output)
	}
}

func TestParseOutputRunItem(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		taskName string
		checkFn  func(t *testing.T, item babfile.RunItem)
	}{
		{
			name:     "command",
			file:     "output_command.yml",
			taskName: "hello",
			checkFn: func(t *testing.T, item babfile.RunItem) {
				cmd, ok := item.(babfile.CommandRun)
				if !ok {
					t.Fatal("expected CommandRun")
				}
				if cmd.Output == nil || *cmd.Output != false {
					t.Error("expected Output = false")
				}
			},
		},
		{
			name:     "task_ref",
			file:     "output_taskrun.yml",
			taskName: "main",
			checkFn: func(t *testing.T, item babfile.RunItem) {
				taskRef, ok := item.(babfile.TaskRun)
				if !ok {
					t.Fatal("expected TaskRun")
				}
				if taskRef.Output == nil || *taskRef.Output != false {
					t.Error("expected Output = false")
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

func TestParseOutputTrue(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "output_true.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result.GlobalOutput == nil {
		t.Fatal("expected GlobalOutput to be set")
	}
	if *result.GlobalOutput != true {
		t.Errorf("expected GlobalOutput = true, got %v", *result.GlobalOutput)
	}

	task := result.Tasks["hello"]
	if task == nil {
		t.Fatal("task 'hello' not found")
	}
	if task.Output == nil || *task.Output != true {
		t.Errorf("expected task Output = true")
	}

	cmd, ok := task.Run[0].(babfile.CommandRun)
	if !ok {
		t.Fatal("expected CommandRun")
	}
	if cmd.Output == nil || *cmd.Output != true {
		t.Errorf("expected command Output = true")
	}
}

func TestParsePromptInput(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "prompt_input.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["test"]
	if task == nil {
		t.Fatal("task 'test' not found")
	}
	if len(task.Run) != 1 {
		t.Fatalf("expected 1 run item, got %d", len(task.Run))
	}

	prompt, ok := task.Run[0].(babfile.PromptRun)
	if !ok {
		t.Fatal("expected PromptRun")
	}
	if prompt.Prompt != "username" {
		t.Errorf("expected prompt 'username', got %q", prompt.Prompt)
	}
	if prompt.Type != babfile.PromptTypeInput {
		t.Errorf("expected type 'input', got %q", prompt.Type)
	}
	if prompt.Message != "Enter your username" {
		t.Errorf("unexpected message: %q", prompt.Message)
	}
	if prompt.Default != "guest" {
		t.Errorf("expected default 'guest', got %q", prompt.Default)
	}
	if prompt.Placeholder != "username" {
		t.Errorf("expected placeholder 'username', got %q", prompt.Placeholder)
	}
}

func TestParsePromptSelect(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "prompt_select.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["test"]
	if task == nil {
		t.Fatal("task 'test' not found")
	}

	prompt, ok := task.Run[0].(babfile.PromptRun)
	if !ok {
		t.Fatal("expected PromptRun")
	}
	if prompt.Type != babfile.PromptTypeSelect {
		t.Errorf("expected type 'select', got %q", prompt.Type)
	}
	if len(prompt.Options) != 3 {
		t.Errorf("expected 3 options, got %d", len(prompt.Options))
	}
	if prompt.Default != "dev" {
		t.Errorf("expected default 'dev', got %q", prompt.Default)
	}
}

func TestParsePromptMultiselect(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "prompt_multiselect.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["test"]
	if task == nil {
		t.Fatal("task 'test' not found")
	}

	prompt, ok := task.Run[0].(babfile.PromptRun)
	if !ok {
		t.Fatal("expected PromptRun")
	}
	if prompt.Type != babfile.PromptTypeMultiselect {
		t.Errorf("expected type 'multiselect', got %q", prompt.Type)
	}
	if len(prompt.Options) != 3 {
		t.Errorf("expected 3 options, got %d", len(prompt.Options))
	}
	if len(prompt.Defaults) != 2 {
		t.Errorf("expected 2 defaults, got %d", len(prompt.Defaults))
	}
	if prompt.Min == nil || *prompt.Min != 1 {
		t.Errorf("expected min 1")
	}
	if prompt.Max == nil || *prompt.Max != 2 {
		t.Errorf("expected max 2")
	}
}

func TestParsePromptConfirm(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "prompt_confirm.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["test"]
	if task == nil {
		t.Fatal("task 'test' not found")
	}

	prompt, ok := task.Run[0].(babfile.PromptRun)
	if !ok {
		t.Fatal("expected PromptRun")
	}
	if prompt.Type != babfile.PromptTypeConfirm {
		t.Errorf("expected type 'confirm', got %q", prompt.Type)
	}
	if prompt.Default != "false" {
		t.Errorf("expected default 'false', got %q", prompt.Default)
	}
}

func TestParsePromptPassword(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "prompt_password.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["test"]
	if task == nil {
		t.Fatal("task 'test' not found")
	}

	prompt, ok := task.Run[0].(babfile.PromptRun)
	if !ok {
		t.Fatal("expected PromptRun")
	}
	if prompt.Type != babfile.PromptTypePassword {
		t.Errorf("expected type 'password', got %q", prompt.Type)
	}
	if prompt.Confirm == nil || *prompt.Confirm != true {
		t.Error("expected confirm = true")
	}
}

func TestParsePromptNumber(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "prompt_number.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["test"]
	if task == nil {
		t.Fatal("task 'test' not found")
	}

	prompt, ok := task.Run[0].(babfile.PromptRun)
	if !ok {
		t.Fatal("expected PromptRun")
	}
	if prompt.Type != babfile.PromptTypeNumber {
		t.Errorf("expected type 'number', got %q", prompt.Type)
	}
	if prompt.Default != "8080" {
		t.Errorf("expected default '8080', got %q", prompt.Default)
	}
	if prompt.Min == nil || *prompt.Min != 1024 {
		t.Errorf("expected min 1024")
	}
	if prompt.Max == nil || *prompt.Max != 65535 {
		t.Errorf("expected max 65535")
	}
}

func TestParsePromptMissingType(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "prompt_invalid_missing_type.yml"))
	if err == nil {
		t.Fatal("expected error for missing type")
	}
	var verrs *errs.ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
}

func TestParsePromptInvalidType(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "prompt_invalid_type.yml"))
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	var verrs *errs.ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
}

func TestParsePromptSelectMissingOptions(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "prompt_invalid_missing_options.yml"))
	if err == nil {
		t.Fatal("expected error for missing options")
	}
	var verrs *errs.ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
}

func TestParsePromptSelectDefaultNotInOptions(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "prompt_invalid_default_not_in_options.yml"))
	if err == nil {
		t.Fatal("expected error for default not in options")
	}
	var verrs *errs.ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
}

func TestParsePromptOptionsOnInput(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "prompt_invalid_options_on_input.yml"))
	if err == nil {
		t.Fatal("expected error for options on input type")
	}
	var verrs *errs.ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
}

func TestParsePromptConfirmOnInput(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "prompt_invalid_confirm_on_input.yml"))
	if err == nil {
		t.Fatal("expected error for confirm on input type")
	}
	var verrs *errs.ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
}

func TestParsePromptNumberInvalidDefault(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "prompt_invalid_number_default.yml"))
	if err == nil {
		t.Fatal("expected error for non-numeric default on number type")
	}
	var verrs *errs.ValidationErrors
	if !errors.As(err, &verrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
}

func TestParseDirGlobal(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "dir_global.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.GlobalDir != "./subdir" {
		t.Errorf("expected GlobalDir './subdir', got %q", result.GlobalDir)
	}
}

func TestParseDirTask(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "dir_task.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	task := result.Tasks["test"]
	if task == nil {
		t.Fatal("task 'test' not found")
	}
	if task.Dir != "./mydir" {
		t.Errorf("expected Dir './mydir', got %q", task.Dir)
	}
}

func TestParseDirCommand(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "dir_command.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	task := result.Tasks["test"]
	if task == nil {
		t.Fatal("task 'test' not found")
	}
	cmd, ok := task.Run[0].(babfile.CommandRun)
	if !ok {
		t.Fatal("expected CommandRun")
	}
	if cmd.Dir != "./cmddir" {
		t.Errorf("expected Dir './cmddir', got %q", cmd.Dir)
	}
}

func TestParseDirCascade(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "dir_cascade.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.GlobalDir != "./global" {
		t.Errorf("expected GlobalDir './global', got %q", result.GlobalDir)
	}
	task := result.Tasks["test"]
	if task == nil {
		t.Fatal("task 'test' not found")
	}
	if task.Dir != "./task" {
		t.Errorf("expected task Dir './task', got %q", task.Dir)
	}
	if len(task.Run) < 2 {
		t.Fatal("expected at least 2 run items")
	}
	cmd2, ok := task.Run[1].(babfile.CommandRun)
	if !ok {
		t.Fatal("expected CommandRun for second item")
	}
	if cmd2.Dir != "./cmd" {
		t.Errorf("expected cmd Dir './cmd', got %q", cmd2.Dir)
	}
}

func TestParseTaskSourcePath(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "simple.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	task := result.Tasks["hello"]
	if task == nil {
		t.Fatal("task 'hello' not found")
	}
	if task.SourcePath == "" {
		t.Error("expected SourcePath to be set")
	}
	if !filepath.IsAbs(task.SourcePath) {
		t.Errorf("expected absolute SourcePath, got %q", task.SourcePath)
	}
}

func TestParseIncludeSourcePath(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	local := result.Tasks["setup"]
	if local == nil {
		t.Fatal("task 'setup' not found")
	}
	if !strings.HasSuffix(local.SourcePath, "main.yml") {
		t.Errorf("local task SourcePath should end with main.yml, got %q", local.SourcePath)
	}

	included := result.Tasks["gen:build"]
	if included == nil {
		t.Fatal("task 'gen:build' not found")
	}
	if !strings.HasSuffix(included.SourcePath, "simple.yml") {
		t.Errorf("included task SourcePath should end with simple.yml, got %q", included.SourcePath)
	}
}

func TestParseDeepIncludeSourcePath(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "deep", "main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	root := result.Tasks["root"]
	if root == nil {
		t.Fatal("task 'root' not found")
	}
	if !strings.HasSuffix(root.SourcePath, "main.yml") {
		t.Errorf("root task SourcePath wrong: %q", root.SourcePath)
	}

	first := result.Tasks["first:build"]
	if first == nil {
		t.Fatal("task 'first:build' not found")
	}
	if !strings.HasSuffix(first.SourcePath, "first.yml") {
		t.Errorf("first:build task SourcePath wrong: %q", first.SourcePath)
	}

	second := result.Tasks["first:second:compile"]
	if second == nil {
		t.Fatal("task 'first:second:compile' not found")
	}
	if !strings.HasSuffix(second.SourcePath, "second.yml") {
		t.Errorf("first:second:compile task SourcePath wrong: %q", second.SourcePath)
	}
}

func TestParseIncludePreservesDir(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "dir_main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["sub:build"]
	if task == nil {
		t.Fatal("task 'sub:build' not found")
	}
	if task.Dir != "./subtask" {
		t.Errorf("expected Dir './subtask', got %q", task.Dir)
	}

	if len(task.Run) < 2 {
		t.Fatal("expected at least 2 run items")
	}
	cmd, ok := task.Run[1].(babfile.CommandRun)
	if !ok {
		t.Fatal("expected CommandRun for second item")
	}
	if cmd.Dir != "./subcmd" {
		t.Errorf("expected cmd Dir './subcmd', got %q", cmd.Dir)
	}
}

func TestParseSingleAlias(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "alias_single.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["greet"]
	if task == nil {
		t.Fatal("task 'greet' not found")
	}
	if task.Alias != "g" {
		t.Errorf("expected alias 'g', got %q", task.Alias)
	}
	if result.Aliases["g"] != "greet" {
		t.Errorf("expected Aliases['g'] = 'greet', got %q", result.Aliases["g"])
	}
}

func TestParseMultipleAliases(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "alias_multiple.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["build"]
	if task == nil {
		t.Fatal("task 'build' not found")
	}
	if len(task.Aliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(task.Aliases))
	}
	if task.Aliases[0] != "b" || task.Aliases[1] != "bld" {
		t.Errorf("expected aliases [b, bld], got %v", task.Aliases)
	}
	if result.Aliases["b"] != "build" {
		t.Errorf("expected Aliases['b'] = 'build', got %q", result.Aliases["b"])
	}
	if result.Aliases["bld"] != "build" {
		t.Errorf("expected Aliases['bld'] = 'build', got %q", result.Aliases["bld"])
	}
}

func TestParseBothAliasAndAliases(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "alias_both.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["deploy"]
	if task == nil {
		t.Fatal("task 'deploy' not found")
	}

	allAliases := task.GetAllAliases()
	if len(allAliases) != 3 {
		t.Errorf("expected 3 total aliases, got %d: %v", len(allAliases), allAliases)
	}
	for _, alias := range []string{"d", "dep", "ship"} {
		if result.Aliases[alias] != "deploy" {
			t.Errorf("expected Aliases[%q] = 'deploy', got %q", alias, result.Aliases[alias])
		}
	}
}

func TestParseAliasConflictsWithTaskName(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "alias_conflict.yml"))
	if err == nil {
		t.Fatal("expected error for alias conflicting with task name")
	}
	if !errors.Is(err, errs.ErrAliasConflict) {
		t.Errorf("expected ErrAliasConflict, got %v", err)
	}
}

func TestParseDuplicateAlias(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "alias_duplicate.yml"))
	if err == nil {
		t.Fatal("expected error for duplicate alias")
	}
	if !errors.Is(err, errs.ErrDuplicateAlias) {
		t.Errorf("expected ErrDuplicateAlias, got %v", err)
	}
}

func TestParseIncludeWithAliases(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "alias_main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if result.Aliases["r"] != "run" {
		t.Errorf("expected Aliases['r'] = 'run', got %q", result.Aliases["r"])
	}
	if result.Aliases["gen:b"] != "gen:build" {
		t.Errorf("expected Aliases['gen:b'] = 'gen:build', got %q", result.Aliases["gen:b"])
	}
	if result.Aliases["gen:bld"] != "gen:build" {
		t.Errorf("expected Aliases['gen:bld'] = 'gen:build', got %q", result.Aliases["gen:bld"])
	}
	genBuild := result.Tasks["gen:build"]
	if genBuild == nil {
		t.Fatal("task 'gen:build' not found")
	}
	if genBuild.Alias != "gen:b" {
		t.Errorf("expected task alias 'gen:b', got %q", genBuild.Alias)
	}
}

func TestParseEmptyAliasesIgnored(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "alias_empty.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(result.Aliases) != 1 {
		t.Errorf("expected 1 alias, got %d: %v", len(result.Aliases), result.Aliases)
	}
	if result.Aliases["l"] != "lint" {
		t.Errorf("expected Aliases['l'] = 'lint', got %q", result.Aliases["l"])
	}
}
