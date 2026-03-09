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

func assertBoolPtr(t *testing.T, label string, got *bool, want bool) {
	t.Helper()
	if got == nil {
		t.Fatalf("expected %s to be set", label)
	}
	if *got != want {
		t.Errorf("expected %s = %v, got %v", label, want, *got)
	}
}

type boolFlagAccessors struct {
	global       func(*ParseResult) *bool
	task         func(*babfile.Task) *bool
	cmdRunItem   func(babfile.CommandRun) *bool
	taskRunItem  func(babfile.TaskRun) *bool
	globalFile   string
	taskFile     string
	allFalseFile string
	allTrueFile  string
	cmdFile      string
	taskRefFile  string
	globalVal    bool
	taskVal      bool
}

func runParseBoolFlagSuite(t *testing.T, a boolFlagAccessors) {
	t.Helper()

	t.Run("global", func(t *testing.T) {
		result, err := Parse(filepath.Join("testdata", a.globalFile))
		if err != nil {
			t.Fatalf("Parse() error: %v", err)
		}
		assertBoolPtr(t, "global flag", a.global(result), a.globalVal)
	})

	t.Run("task", func(t *testing.T) {
		result, err := Parse(filepath.Join("testdata", a.taskFile))
		if err != nil {
			t.Fatalf("Parse() error: %v", err)
		}
		task := result.Tasks["hello"]
		if task == nil {
			t.Fatal("task 'hello' not found")
		}
		assertBoolPtr(t, "task flag", a.task(task), a.taskVal)
	})

	allLevelsFile := a.allFalseFile
	allLevelsVal := false
	if a.allTrueFile != "" {
		allLevelsFile = a.allTrueFile
		allLevelsVal = true
	}
	t.Run("all_levels", func(t *testing.T) {
		result, err := Parse(filepath.Join("testdata", allLevelsFile))
		if err != nil {
			t.Fatalf("Parse() error: %v", err)
		}
		assertBoolPtr(t, "global flag", a.global(result), allLevelsVal)
		task := result.Tasks["hello"]
		if task == nil {
			t.Fatal("task 'hello' not found")
		}
		assertBoolPtr(t, "task flag", a.task(task), allLevelsVal)
		cmd, ok := task.Run[0].(babfile.CommandRun)
		if !ok {
			t.Fatal("expected CommandRun")
		}
		assertBoolPtr(t, "cmd flag", a.cmdRunItem(cmd), allLevelsVal)
	})

	t.Run("run_item_command", func(t *testing.T) {
		result, err := Parse(filepath.Join("testdata", a.cmdFile))
		if err != nil {
			t.Fatalf("Parse() error: %v", err)
		}
		task := result.Tasks["hello"]
		if task == nil {
			t.Fatal("task 'hello' not found")
		}
		cmd, ok := task.Run[0].(babfile.CommandRun)
		if !ok {
			t.Fatal("expected CommandRun")
		}
		assertBoolPtr(t, "cmd flag", a.cmdRunItem(cmd), !allLevelsVal)
	})

	t.Run("run_item_task_ref", func(t *testing.T) {
		result, err := Parse(filepath.Join("testdata", a.taskRefFile))
		if err != nil {
			t.Fatalf("Parse() error: %v", err)
		}
		task := result.Tasks["main"]
		if task == nil {
			t.Fatal("task 'main' not found")
		}
		taskRef, ok := task.Run[0].(babfile.TaskRun)
		if !ok {
			t.Fatal("expected TaskRun")
		}
		assertBoolPtr(t, "taskRef flag", a.taskRunItem(taskRef), !allLevelsVal)
	})
}

func TestParseSilent(t *testing.T) {
	runParseBoolFlagSuite(t, boolFlagAccessors{
		global:       func(r *ParseResult) *bool { return r.GlobalSilent },
		task:         func(task *babfile.Task) *bool { return task.Silent },
		cmdRunItem:   func(cmd babfile.CommandRun) *bool { return cmd.Silent },
		taskRunItem:  func(tr babfile.TaskRun) *bool { return tr.Silent },
		globalFile:   "silent_global.yml",
		taskFile:     "silent_task.yml",
		allFalseFile: "silent_false.yml",
		cmdFile:      "silent_command.yml",
		taskRefFile:  "silent_taskrun.yml",
		globalVal:    true,
		taskVal:      true,
	})
}

func TestParseOutput(t *testing.T) {
	runParseBoolFlagSuite(t, boolFlagAccessors{
		global:      func(r *ParseResult) *bool { return r.GlobalOutput },
		task:        func(task *babfile.Task) *bool { return task.Output },
		cmdRunItem:  func(cmd babfile.CommandRun) *bool { return cmd.Output },
		taskRunItem: func(tr babfile.TaskRun) *bool { return tr.Output },
		globalFile:  "output_global.yml",
		taskFile:    "output_task.yml",
		allTrueFile: "output_true.yml",
		cmdFile:     "output_command.yml",
		taskRefFile: "output_taskrun.yml",
		globalVal:   false,
		taskVal:     false,
	})
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

func TestParsePromptErrors(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"missing type", "prompt_invalid_missing_type.yml"},
		{"invalid type", "prompt_invalid_type.yml"},
		{"select missing options", "prompt_invalid_missing_options.yml"},
		{"select default not in options", "prompt_invalid_default_not_in_options.yml"},
		{"options on input", "prompt_invalid_options_on_input.yml"},
		{"confirm on input", "prompt_invalid_confirm_on_input.yml"},
		{"number invalid default", "prompt_invalid_number_default.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(filepath.Join("testdata", tt.file))
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
			var verrs *errs.ValidationErrors
			if !errors.As(err, &verrs) {
				t.Fatalf("expected ValidationErrors, got %T", err)
			}
		})
	}
}

func TestParseDir(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		checkFn func(t *testing.T, result *ParseResult)
	}{
		{
			name: "global",
			file: "dir_global.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
				if result.GlobalDir != "./subdir" {
					t.Errorf("expected GlobalDir './subdir', got %q", result.GlobalDir)
				}
			},
		},
		{
			name: "task",
			file: "dir_task.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
				task := result.Tasks["test"]
				if task == nil {
					t.Fatal("task 'test' not found")
				}
				if task.Dir != "./mydir" {
					t.Errorf("expected Dir './mydir', got %q", task.Dir)
				}
			},
		},
		{
			name: "command",
			file: "dir_command.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
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
			},
		},
		{
			name: "cascade",
			file: "dir_cascade.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(filepath.Join("testdata", tt.file))
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}
			tt.checkFn(t, result)
		})
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

func TestParseAlias(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		checkFn func(t *testing.T, result *ParseResult)
	}{
		{
			name: "single",
			file: "alias_single.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
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
			},
		},
		{
			name: "multiple",
			file: "alias_multiple.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
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
			},
		},
		{
			name: "both alias and aliases",
			file: "alias_both.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
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
			},
		},
		{
			name: "empty aliases ignored",
			file: "alias_empty.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
				if len(result.Aliases) != 1 {
					t.Errorf("expected 1 alias, got %d: %v", len(result.Aliases), result.Aliases)
				}
				if result.Aliases["l"] != "lint" {
					t.Errorf("expected Aliases['l'] = 'lint', got %q", result.Aliases["l"])
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(filepath.Join("testdata", tt.file))
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}
			tt.checkFn(t, result)
		})
	}
}

func TestParseAliasErrors(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		wantErr error
	}{
		{"conflicts with task name", "alias_conflict.yml", errs.ErrAliasConflict},
		{"duplicate alias", "alias_duplicate.yml", errs.ErrDuplicateAlias},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(filepath.Join("testdata", tt.file))
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
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

func TestParseIncludeInheritsGlobalDir(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "globals_main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["sub:no-overrides"]
	if task == nil {
		t.Fatal("task 'sub:no-overrides' not found")
	}
	if task.Dir != "./subglobal" {
		t.Errorf("expected Dir './subglobal' from included root, got %q", task.Dir)
	}

	task = result.Tasks["sub:with-overrides"]
	if task == nil {
		t.Fatal("task 'sub:with-overrides' not found")
	}
	if task.Dir != "./subtask" {
		t.Errorf("expected Dir './subtask' (task-level override), got %q", task.Dir)
	}
}

func TestParseIncludeInheritsGlobalVars(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "globals_main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["sub:no-overrides"]
	if task == nil {
		t.Fatal("task 'sub:no-overrides' not found")
	}
	if task.Vars["sub_version"] != "2.0" {
		t.Errorf("expected var sub_version='2.0' from included root, got %q", task.Vars["sub_version"])
	}

	task = result.Tasks["sub:with-overrides"]
	if task == nil {
		t.Fatal("task 'sub:with-overrides' not found")
	}
	if task.Vars["sub_version"] != "3.0" {
		t.Errorf("expected var sub_version='3.0' (task-level override), got %q", task.Vars["sub_version"])
	}
}

func TestParseIncludeInheritsGlobalEnv(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "globals_main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["sub:no-overrides"]
	if task == nil {
		t.Fatal("task 'sub:no-overrides' not found")
	}
	if task.Env["SUB_ENV"] != "from-sub" {
		t.Errorf("expected env SUB_ENV='from-sub' from included root, got %q", task.Env["SUB_ENV"])
	}

	task = result.Tasks["sub:with-overrides"]
	if task == nil {
		t.Fatal("task 'sub:with-overrides' not found")
	}
	if task.Env["SUB_ENV"] != "from-task" {
		t.Errorf("expected env SUB_ENV='from-task' (task-level override), got %q", task.Env["SUB_ENV"])
	}
}

func TestParseIncludeInheritsGlobalSilentOutput(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "globals_main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["sub:no-overrides"]
	if task == nil {
		t.Fatal("task 'sub:no-overrides' not found")
	}
	if task.Silent == nil || !*task.Silent {
		t.Error("expected Silent=true from included root")
	}
	if task.Output == nil || *task.Output {
		t.Error("expected Output=false from included root")
	}

	task = result.Tasks["sub:with-overrides"]
	if task == nil {
		t.Fatal("task 'sub:with-overrides' not found")
	}
	if task.Silent == nil || *task.Silent {
		t.Error("expected Silent=false (task-level override)")
	}
	if task.Output == nil || !*task.Output {
		t.Error("expected Output=true (task-level override)")
	}
}

func TestParseIncludePreservesTaskRunFields(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "includes", "globals_main.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["sub:with-taskref"]
	if task == nil {
		t.Fatal("task 'sub:with-taskref' not found")
	}

	if len(task.Run) != 1 {
		t.Fatalf("expected 1 run item, got %d", len(task.Run))
	}

	tr, ok := task.Run[0].(babfile.TaskRun)
	if !ok {
		t.Fatal("expected TaskRun")
	}
	if tr.Task != "sub:no-overrides" {
		t.Errorf("expected task ref 'sub:no-overrides', got %q", tr.Task)
	}
	if tr.Silent == nil || !*tr.Silent {
		t.Error("expected Silent=true preserved on TaskRun")
	}
	if tr.When != "${{ sub_version }}" {
		t.Errorf("expected When preserved on TaskRun, got %q", tr.When)
	}
}

func TestParseInvalidPlatform(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "invalid_platform.yml"))
	if err == nil {
		t.Fatal("expected error for invalid platform")
	}
	if !strings.Contains(err.Error(), "invalid platform") {
		t.Errorf("expected error about invalid platform, got: %v", err)
	}
}

func TestParseInvalidDirGlobal(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "invalid_dir_global.yml"))
	if err == nil {
		t.Fatal("expected error for non-scalar global dir")
	}
	if !strings.Contains(err.Error(), "dir must be a string") {
		t.Errorf("expected error about dir type, got: %v", err)
	}
}

func TestParseParallelBasic(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "parallel_basic.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["build:all"]
	if task == nil {
		t.Fatal("task 'build:all' not found")
	}
	if len(task.Run) != 1 {
		t.Fatalf("expected 1 run item, got %d", len(task.Run))
	}

	pr, ok := task.Run[0].(babfile.ParallelRun)
	if !ok {
		t.Fatal("expected ParallelRun")
	}
	if len(pr.Items) != 3 {
		t.Fatalf("expected 3 parallel items, got %d", len(pr.Items))
	}
	if pr.Mode != babfile.ParallelInterleaved {
		t.Errorf("expected mode 'interleaved', got %q", pr.Mode)
	}
	if pr.Limit != 2 {
		t.Errorf("expected limit 2, got %d", pr.Limit)
	}

	if _, ok := pr.Items[0].(babfile.TaskRun); !ok {
		t.Error("expected first item to be TaskRun")
	}
	if _, ok := pr.Items[1].(babfile.TaskRun); !ok {
		t.Error("expected second item to be TaskRun")
	}
	if _, ok := pr.Items[2].(babfile.CommandRun); !ok {
		t.Error("expected third item to be CommandRun")
	}

	if pr.Labels[2] != "libs" {
		t.Errorf("expected label 'libs', got %q", pr.Labels[2])
	}
}

func TestParseParallelGrouped(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "parallel_grouped.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["test:all"]
	if task == nil {
		t.Fatal("task 'test:all' not found")
	}

	pr, ok := task.Run[0].(babfile.ParallelRun)
	if !ok {
		t.Fatal("expected ParallelRun")
	}
	if pr.Mode != babfile.ParallelGrouped {
		t.Errorf("expected mode 'grouped', got %q", pr.Mode)
	}
	if len(pr.Items) != 2 {
		t.Errorf("expected 2 parallel items, got %d", len(pr.Items))
	}
}

func TestParseParallelRejectsPrompt(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "parallel_prompt.yml"))
	if err == nil {
		t.Fatal("expected error for prompt inside parallel")
	}
	if !strings.Contains(err.Error(), "prompt items cannot be used inside parallel blocks") {
		t.Errorf("expected prompt rejection error, got: %v", err)
	}
}

func TestParseParallelRejectsNested(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "parallel_nested.yml"))
	if err == nil {
		t.Fatal("expected error for nested parallel")
	}
	if !strings.Contains(err.Error(), "parallel items cannot be nested") {
		t.Errorf("expected nesting rejection error, got: %v", err)
	}
}

func TestParseParallelInvalidMode(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "parallel_invalid_mode.yml"))
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Errorf("expected invalid mode error, got: %v", err)
	}
}

func TestParseParallelItemLabel(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "parallel_basic.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["build:all"]
	pr := task.Run[0].(babfile.ParallelRun)

	if label := pr.ItemLabel(0); label != "build:frontend" {
		t.Errorf("expected label 'build:frontend', got %q", label)
	}

	if label := pr.ItemLabel(2); label != "libs" {
		t.Errorf("expected label 'libs', got %q", label)
	}
}

func TestParseArgsSimple(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "args_simple.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["greet"]
	if task == nil {
		t.Fatal("task 'greet' not found")
	}
	if task.Args == nil {
		t.Fatal("expected args to be set")
	}
	if len(task.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(task.Args))
	}

	nameArg, ok := task.Args["name"]
	if !ok {
		t.Fatal("expected 'name' arg")
	}
	if nameArg.Default != nil {
		t.Error("expected 'name' arg to be required (nil default)")
	}

	greetingArg, ok := task.Args["greeting"]
	if !ok {
		t.Fatal("expected 'greeting' arg")
	}
	if greetingArg.Default == nil {
		t.Fatal("expected 'greeting' arg to have a default")
	}
	if *greetingArg.Default != "Hello" {
		t.Errorf("expected 'greeting' default 'Hello', got %q", *greetingArg.Default)
	}
}

func TestParseArgsTaskRun(t *testing.T) {
	result, err := Parse(filepath.Join("testdata", "args_taskrun.yml"))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	task := result.Tasks["welcome"]
	if task == nil {
		t.Fatal("task 'welcome' not found")
	}
	if len(task.Run) != 1 {
		t.Fatalf("expected 1 run item, got %d", len(task.Run))
	}

	tr, ok := task.Run[0].(babfile.TaskRun)
	if !ok {
		t.Fatal("expected TaskRun")
	}
	if tr.Task != "greet" {
		t.Errorf("expected task 'greet', got %q", tr.Task)
	}
	if len(tr.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(tr.Args))
	}
	if tr.Args["name"] != "World" {
		t.Errorf("expected name='World', got %q", tr.Args["name"])
	}
	if tr.Args["greeting"] != "Hi" {
		t.Errorf("expected greeting='Hi', got %q", tr.Args["greeting"])
	}
}

func TestParseArgsMissingRequired(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "args_missing_required.yml"))
	if err == nil {
		t.Fatal("expected error for missing required arg")
	}
	if !strings.Contains(err.Error(), "required argument") {
		t.Errorf("expected 'required argument' error, got: %v", err)
	}
}

func TestParseArgsUnknown(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "args_unknown.yml"))
	if err == nil {
		t.Fatal("expected error for unknown arg")
	}
	if !strings.Contains(err.Error(), "unknown argument") {
		t.Errorf("expected 'unknown argument' error, got: %v", err)
	}
}

func TestParseArgsNoArgsTask(t *testing.T) {
	_, err := Parse(filepath.Join("testdata", "args_no_args_task.yml"))
	if err == nil {
		t.Fatal("expected error for args on task without args")
	}
	if !strings.Contains(err.Error(), "does not accept arguments") {
		t.Errorf("expected 'does not accept arguments' error, got: %v", err)
	}
}

func TestParseWhen(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		checkFn func(t *testing.T, result *ParseResult)
	}{
		{
			name: "task level when",
			file: "when_task.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
				task := result.Tasks["deploy"]
				if task == nil {
					t.Fatal("task 'deploy' not found")
				}
				if task.When != "${{ confirm }} == 'true'" {
					t.Errorf("expected task When = %q, got %q", "${{ confirm }} == 'true'", task.When)
				}
			},
		},
		{
			name: "command level when",
			file: "when_command.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
				task := result.Tasks["test"]
				if task == nil {
					t.Fatal("task 'test' not found")
				}
				if len(task.Run) != 2 {
					t.Fatalf("expected 2 run items, got %d", len(task.Run))
				}
				cmd0 := task.Run[0].(babfile.CommandRun)
				if cmd0.GetWhen() != "${{ confirm }} == 'true'" {
					t.Errorf("cmd[0] When = %q, want %q", cmd0.GetWhen(), "${{ confirm }} == 'true'")
				}
				cmd1 := task.Run[1].(babfile.CommandRun)
				if cmd1.GetWhen() != "${{ confirm }} != 'true'" {
					t.Errorf("cmd[1] When = %q, want %q", cmd1.GetWhen(), "${{ confirm }} != 'true'")
				}
			},
		},
		{
			name: "all run item types",
			file: "when_all_types.yml",
			checkFn: func(t *testing.T, result *ParseResult) {
				task := result.Tasks["test"]
				if task == nil {
					t.Fatal("task 'test' not found")
				}
				if task.When != "${{ enabled }}" {
					t.Errorf("task When = %q, want %q", task.When, "${{ enabled }}")
				}
				if len(task.Run) != 4 {
					t.Fatalf("expected 4 run items, got %d", len(task.Run))
				}
				wantWhens := []string{
					"${{ run_cmd }} == 'true'",
					"${{ run_task }} == 'true'",
					"${{ run_log }} == 'true'",
					"${{ run_prompt }} == 'true'",
				}
				for i, want := range wantWhens {
					if got := task.Run[i].GetWhen(); got != want {
						t.Errorf("run[%d].GetWhen() = %q, want %q", i, got, want)
					}
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(filepath.Join("testdata", tt.file))
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}
			tt.checkFn(t, result)
		})
	}
}
