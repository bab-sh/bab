package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/condition"
	"github.com/bab-sh/bab/internal/errs"
	"github.com/bab-sh/bab/internal/finder"
	"github.com/bab-sh/bab/internal/interpolate"
	"github.com/bab-sh/bab/internal/output"
	"github.com/bab-sh/bab/internal/parser"
	"github.com/bab-sh/bab/internal/tui"
	"github.com/charmbracelet/log"
)

type status int

const (
	_ status = iota
	running
	done
)

type Runner struct {
	DryRun       bool
	Babfile      string
	BabfilePath  string
	GlobalVars   map[string]string
	GlobalEnv    map[string]string
	GlobalSilent *bool
	GlobalOutput *bool
	GlobalDir    string
	Aliases      map[string]string
}

func New(dryRun bool, babfile string) *Runner {
	return &Runner{DryRun: dryRun, Babfile: babfile}
}

func LoadTasks(customPath string) (*parser.ParseResult, error) {
	var path string
	if customPath != "" {
		path = customPath
	} else {
		p, err := finder.FindBabfile()
		if err != nil {
			return nil, err
		}
		path = p
	}
	return parser.Parse(path)
}

func (r *Runner) Run(ctx context.Context, taskName string) error {
	result, err := LoadTasks(r.Babfile)
	if err != nil {
		return err
	}

	r.BabfilePath = result.Path
	r.Aliases = result.Aliases

	resolvedVars, err := interpolate.ResolveVarsWithLocation(result.GlobalVars, nil, r.BabfilePath, 0)
	if err != nil {
		return fmt.Errorf("resolving global variables: %w", err)
	}
	r.GlobalVars = resolvedVars
	r.GlobalEnv = result.GlobalEnv
	r.GlobalSilent = result.GlobalSilent
	r.GlobalOutput = result.GlobalOutput
	r.GlobalDir = result.GlobalDir

	resolvedName := r.resolveTaskName(taskName)

	return r.RunWithTasks(ctx, resolvedName, result.Tasks)
}

func (r *Runner) resolveTaskName(name string) string {
	if r.Aliases == nil {
		return name
	}
	if actual, isAlias := r.Aliases[name]; isAlias {
		return actual
	}
	return name
}

func (r *Runner) RunWithTasks(ctx context.Context, taskName string, tasks babfile.TaskMap) error {
	state := make(map[string]status)
	return r.runTask(ctx, taskName, tasks, state, true, nil, nil)
}

func isSilent(item, task, global *bool) bool {
	if item != nil {
		return *item
	}
	if task != nil {
		return *task
	}
	if global != nil {
		return *global
	}
	return false
}

func isOutput(item, task, global *bool) bool {
	if item != nil {
		return *item
	}
	if task != nil {
		return *task
	}
	if global != nil {
		return *global
	}
	return true
}

func (r *Runner) runTask(ctx context.Context, name string, tasks babfile.TaskMap, state map[string]status, isMain bool, overrideSilent, overrideOutput *bool) error {
	switch state[name] {
	case done:
		return nil
	case running:
		return &errs.CircularDepError{
			Type:  "dependency",
			Chain: buildChainSlice(name, tasks, state),
		}
	}

	task, ok := tasks[name]
	if !ok {
		available := tasks.Names()
		for alias := range r.Aliases {
			available = append(available, alias)
		}
		return &errs.TaskNotFoundError{
			TaskName:  name,
			Available: available,
		}
	}

	if task.When != "" {
		whenCtx := interpolate.NewContextWithLocation(r.GlobalVars, r.BabfilePath, task.Line)
		result, err := condition.Evaluate(task.When, whenCtx)
		if err != nil {
			return fmt.Errorf("task %q: evaluating when condition: %w", name, err)
		}
		if !result.ShouldRun {
			log.Debug("Skipping task", "task", name, "reason", "when condition", "detail", result.Reason)
			state[name] = done
			return nil
		}
	}

	state[name] = running

	if !r.DryRun && !isSilent(overrideSilent, task.Silent, r.GlobalSilent) {
		if isMain {
			output.Task(name)
		} else {
			output.Dep(name)
		}
	}

	for _, dep := range task.Deps {
		log.Debug("Running dependency", "task", name, "dep", dep)
		if err := r.runTask(ctx, dep, tasks, state, false, nil, nil); err != nil {
			return fmt.Errorf("dependency %q failed: %w", dep, err)
		}
	}

	if len(task.Run) > 0 {
		if err := r.executeTask(ctx, task, tasks, state, overrideOutput); err != nil {
			return err
		}
	}

	state[name] = done
	return nil
}

func (r *Runner) executeTask(ctx context.Context, task *babfile.Task, tasks babfile.TaskMap, state map[string]status, overrideOutput *bool) error {
	shell, shellArg := shellCommand()
	platform := runtime.GOOS
	executed := 0
	skippedByCondition := 0

	taskVars, err := interpolate.ResolveVarsWithLocation(task.Vars, r.GlobalVars, r.BabfilePath, task.Line)
	if err != nil {
		return err
	}

	taskCtx := interpolate.NewContextWithLocation(taskVars, r.BabfilePath, task.Line)

	taskEnv, err := r.interpolateEnv(babfile.MergeEnvMaps(r.GlobalEnv, task.Env), taskCtx)
	if err != nil {
		return err
	}

	log.Debug("Executing task", "name", task.Name, "runItems", len(task.Run), "dryRun", r.DryRun, "envVars", len(taskEnv), "vars", len(taskVars))

	for i, item := range task.Run {
		select {
		case <-ctx.Done():
			return fmt.Errorf("cancelled: %w", ctx.Err())
		default:
		}

		if !item.ShouldRunOnPlatform(platform) {
			log.Debug("Skipping run item", "task", task.Name, "index", i+1, "reason", "platform")
			continue
		}

		if shouldSkip, err := r.shouldSkipRunItem(item, taskVars, task.Name, i+1); err != nil {
			return err
		} else if shouldSkip {
			skippedByCondition++
			continue
		}

		switch v := item.(type) {
		case babfile.CommandRun:
			if strings.TrimSpace(v.Cmd) == "" {
				return fmt.Errorf("task %q command %d is empty", task.Name, i+1)
			}

			cmdCtx := interpolate.NewContextWithLocation(taskVars, r.BabfilePath, v.Line)
			interpolatedCmd, err := interpolate.Interpolate(v.Cmd, cmdCtx)
			if err != nil {
				return err
			}

			cmdEnv, err := r.interpolateEnv(babfile.MergeEnvMaps(taskEnv, v.Env), cmdCtx)
			if err != nil {
				return err
			}

			cmdDir, err := r.resolveDir(task, v.Dir, cmdCtx)
			if err != nil {
				return fmt.Errorf("task %q: command %d: %w", task.Name, i+1, err)
			}

			if r.DryRun {
				log.Info("Would run", "cmd", interpolatedCmd, "env", len(cmdEnv), "dir", cmdDir)
			} else {
				if !isSilent(v.Silent, task.Silent, r.GlobalSilent) {
					output.Cmd(interpolatedCmd)
				}
				taskOutput := task.Output
				if overrideOutput != nil {
					taskOutput = overrideOutput
				}
				showOutput := isOutput(v.Output, taskOutput, r.GlobalOutput)
				if err := runCommand(ctx, shell, shellArg, interpolatedCmd, cmdEnv, showOutput, cmdDir); err != nil {
					return fmt.Errorf("task %q: command %d failed: %w", task.Name, i+1, err)
				}
			}

		case babfile.TaskRun:
			if r.DryRun {
				log.Info("Would run task", "task", v.Task)
			} else {
				log.Debug("Running task", "task", v.Task)
				if err := r.runTask(ctx, v.Task, tasks, state, false, v.Silent, v.Output); err != nil {
					return fmt.Errorf("task %q failed: %w", v.Task, err)
				}
			}

		case babfile.LogRun:
			logCtx := interpolate.NewContextWithLocation(taskVars, r.BabfilePath, v.Line)
			interpolatedLog, err := interpolate.Interpolate(v.Log, logCtx)
			if err != nil {
				return err
			}

			if r.DryRun {
				log.Info("Would log", "msg", interpolatedLog, "level", v.Level)
			} else {
				executeLog(babfile.LogRun{Log: interpolatedLog, Level: v.Level})
			}

		case babfile.PromptRun:
			promptCtx := interpolate.NewContextWithLocation(taskVars, r.BabfilePath, v.Line)
			interpolatedMsg, err := interpolate.Interpolate(v.Message, promptCtx)
			if err != nil {
				return err
			}

			if r.DryRun {
				log.Info("Would prompt", "var", v.Prompt, "type", v.Type, "message", interpolatedMsg)
			} else {
				result, err := tui.RunPrompt(v, interpolatedMsg)
				if err != nil {
					return fmt.Errorf("task %q: prompt %q: %w", task.Name, v.Prompt, err)
				}
				taskVars[v.Prompt] = result
				log.Debug("Prompt result stored", "var", v.Prompt, "value", result)
			}
		}

		executed++
	}

	if executed == 0 {
		if skippedByCondition > 0 {
			log.Debug("All run items skipped by condition", "task", task.Name, "skipped", skippedByCondition)
		} else {
			return fmt.Errorf("task %q has no run items for platform %q", task.Name, platform)
		}
	}

	return nil
}

func getItemLine(item babfile.RunItem) int {
	switch v := item.(type) {
	case babfile.CommandRun:
		return v.Line
	case babfile.TaskRun:
		return v.Line
	case babfile.LogRun:
		return v.Line
	case babfile.PromptRun:
		return v.Line
	default:
		return 0
	}
}

func (r *Runner) shouldSkipRunItem(item babfile.RunItem, taskVars map[string]string, taskName string, index int) (bool, error) {
	whenCond := item.GetWhen()
	if whenCond == "" {
		return false, nil
	}

	itemCtx := interpolate.NewContextWithLocation(taskVars, r.BabfilePath, getItemLine(item))
	result, err := condition.Evaluate(whenCond, itemCtx)
	if err != nil {
		return false, fmt.Errorf("task %q: run[%d]: evaluating when condition: %w", taskName, index, err)
	}

	if !result.ShouldRun {
		log.Debug("Skipping run item", "task", taskName, "index", index, "reason", "when condition", "detail", result.Reason)
		return true, nil
	}

	return false, nil
}

func (r *Runner) interpolateEnv(env map[string]string, ctx *interpolate.Context) (map[string]string, error) {
	if len(env) == 0 {
		return env, nil
	}

	result := make(map[string]string, len(env))
	for k, v := range env {
		interpolated, err := interpolate.Interpolate(v, ctx)
		if err != nil {
			return nil, err
		}
		result[k] = interpolated
	}
	return result, nil
}

func (r *Runner) resolveDir(task *babfile.Task, cmdDir string, ctx *interpolate.Context) (string, error) {
	baseDir := filepath.Dir(task.SourcePath)

	dir := cmdDir
	if dir == "" {
		dir = task.Dir
	}
	if dir == "" {
		dir = r.GlobalDir
	}

	if dir == "" {
		return baseDir, nil
	}

	interpolatedDir, err := interpolate.Interpolate(dir, ctx)
	if err != nil {
		return "", fmt.Errorf("interpolating dir: %w", err)
	}

	if !filepath.IsAbs(interpolatedDir) {
		interpolatedDir = filepath.Join(baseDir, interpolatedDir)
	}

	interpolatedDir = filepath.Clean(interpolatedDir)

	info, err := os.Stat(interpolatedDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory does not exist: %s", interpolatedDir)
		}
		return "", fmt.Errorf("accessing directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("path is not a directory: %s", interpolatedDir)
	}

	return interpolatedDir, nil
}

func shellCommand() (string, string) {
	if runtime.GOOS == "windows" {
		return "cmd", "/C"
	}
	return "sh", "-c"
}

func runCommand(ctx context.Context, shell, shellArg, command string, env map[string]string, showOutput bool, dir string) error {
	cmd := exec.CommandContext(ctx, shell, shellArg, command)
	if showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}

	if len(env) > 0 {
		cmd.Env = append(os.Environ(), babfile.MergeEnv(env)...)
	}

	if dir != "" {
		cmd.Dir = dir
	}

	err := cmd.Run()
	if err != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

func executeLog(l babfile.LogRun) {
	switch l.Level {
	case babfile.LogLevelDebug:
		log.Debug(l.Log)
	case babfile.LogLevelInfo:
		log.Info(l.Log)
	case babfile.LogLevelWarn:
		log.Warn(l.Log)
	case babfile.LogLevelError:
		log.Error(l.Log)
	default:
		log.Info(l.Log)
	}
}

func buildChainSlice(current string, tasks babfile.TaskMap, state map[string]status) []string {
	chain := []string{current}
	seen := make(map[string]bool)

	for {
		last := chain[len(chain)-1]
		if seen[last] {
			break
		}
		seen[last] = true

		task, ok := tasks[last]
		if !ok || len(task.Deps) == 0 {
			break
		}

		for _, dep := range task.Deps {
			if state[dep] == running {
				chain = append(chain, dep)
				if dep == current {
					return chain
				}
				break
			}
		}
	}
	return chain
}
