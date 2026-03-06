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
	"time"

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
	state := &syncState{state: make(map[string]status)}
	return r.runTask(ctx, taskName, tasks, state, true, nil, nil, nil, nil)
}

func isSilent(vals ...*bool) bool {
	for _, v := range vals {
		if v != nil {
			return *v
		}
	}
	return false
}

func isOutput(vals ...*bool) bool {
	for _, v := range vals {
		if v != nil {
			return *v
		}
	}
	return true
}

func (r *Runner) runTask(ctx context.Context, name string, tasks babfile.TaskMap, state *syncState, isMain bool, overrideSilent, overrideOutput *bool, stdout, stderr io.Writer) error {
	switch state.claim(name) {
	case done:
		return nil
	case running:
		return &errs.CircularDepError{
			Type:  "dependency",
			Chain: buildChainSlice(name, tasks, state.snapshot()),
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
			state.set(name, done)
			return nil
		}
	}

	if !r.DryRun && !isSilent(overrideSilent, task.Silent, r.GlobalSilent) {
		if stderr != nil {
			if isMain {
				_, _ = fmt.Fprintln(stderr, output.RenderTask(name))
			} else {
				_, _ = fmt.Fprintln(stderr, output.RenderDep(name))
			}
		} else {
			if isMain {
				output.Task(name)
			} else {
				output.Dep(name)
			}
		}
	}

	for _, dep := range task.Deps {
		log.Debug("Running dependency", "task", name, "dep", dep)
		if err := r.runTask(ctx, dep, tasks, state, false, nil, nil, stdout, stderr); err != nil {
			return fmt.Errorf("dependency %q failed: %w", dep, err)
		}
	}

	if len(task.Run) > 0 {
		if err := r.executeTask(ctx, task, tasks, state, overrideSilent, overrideOutput, stdout, stderr); err != nil {
			return err
		}
	}

	state.set(name, done)
	return nil
}

func (r *Runner) executeTask(ctx context.Context, task *babfile.Task, tasks babfile.TaskMap, state *syncState, overrideSilent, overrideOutput *bool, stdout, stderr io.Writer) error {
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
			if err := r.executeCommand(ctx, v, task, i+1, shell, shellArg, taskVars, taskEnv, overrideSilent, overrideOutput, stdout, stderr); err != nil {
				return err
			}

		case babfile.TaskRun:
			if r.DryRun {
				log.Info("Would run task", "task", v.Task)
			} else {
				log.Debug("Running task", "task", v.Task)
				if err := r.runTask(ctx, v.Task, tasks, state, false, v.Silent, v.Output, stdout, stderr); err != nil {
					return fmt.Errorf("task %q failed: %w", v.Task, err)
				}
			}

		case babfile.ParallelRun:
			if r.DryRun {
				log.Info("Would run parallel", "items", len(v.Items), "mode", v.Mode)
			} else {
				if err := r.executeParallel(ctx, v, task, tasks, state, taskVars, taskEnv, overrideSilent, overrideOutput); err != nil {
					return err
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

			interpolated, err := interpolatePrompt(v, promptCtx)
			if err != nil {
				return err
			}

			if r.DryRun {
				log.Info("Would prompt", "var", interpolated.Prompt, "type", interpolated.Type, "message", interpolated.Message)
			} else {
				result, err := tui.RunPrompt(ctx, interpolated, interpolated.Message)
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

func (r *Runner) executeCommand(ctx context.Context, v babfile.CommandRun, task *babfile.Task, cmdIndex int, shell, shellArg string, taskVars map[string]string, taskEnv map[string]string, overrideSilent, overrideOutput *bool, stdout, stderr io.Writer) error {
	if strings.TrimSpace(v.Cmd) == "" {
		return fmt.Errorf("task %q command %d is empty", task.Name, cmdIndex)
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
		return fmt.Errorf("task %q: command %d: %w", task.Name, cmdIndex, err)
	}

	if r.DryRun {
		log.Info("Would run", "cmd", interpolatedCmd, "env", len(cmdEnv), "dir", cmdDir)
		return nil
	}

	if !isSilent(v.Silent, overrideSilent, task.Silent, r.GlobalSilent) {
		if stderr != nil {
			_, _ = fmt.Fprintln(stderr, output.RenderCmd(interpolatedCmd))
		} else {
			output.Cmd(interpolatedCmd)
		}
	}
	showOutput := isOutput(v.Output, overrideOutput, task.Output, r.GlobalOutput)
	if stdout != nil {
		var outW, errW io.Writer
		if showOutput {
			outW, errW = stdout, stderr
		}
		if err := runCommandWithWriters(ctx, shell, shellArg, interpolatedCmd, cmdEnv, outW, errW, false, cmdDir); err != nil {
			return fmt.Errorf("task %q: command %d failed: %w", task.Name, cmdIndex, err)
		}
	} else {
		if err := runCommand(ctx, shell, shellArg, interpolatedCmd, cmdEnv, showOutput, cmdDir); err != nil {
			return fmt.Errorf("task %q: command %d failed: %w", task.Name, cmdIndex, err)
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

func interpolatePrompt(p babfile.PromptRun, ctx *interpolate.Context) (babfile.PromptRun, error) {
	msg, err := interpolate.Interpolate(p.Message, ctx)
	if err != nil {
		return p, err
	}
	p.Message = msg

	if p.Default != "" {
		p.Default, err = interpolate.Interpolate(p.Default, ctx)
		if err != nil {
			return p, err
		}
	}

	if p.Placeholder != "" {
		p.Placeholder, err = interpolate.Interpolate(p.Placeholder, ctx)
		if err != nil {
			return p, err
		}
	}

	if len(p.Options) > 0 {
		options := make([]string, len(p.Options))
		for i, opt := range p.Options {
			options[i], err = interpolate.Interpolate(opt, ctx)
			if err != nil {
				return p, err
			}
		}
		p.Options = options
	}

	if len(p.Defaults) > 0 {
		defaults := make([]string, len(p.Defaults))
		for i, d := range p.Defaults {
			defaults[i], err = interpolate.Interpolate(d, ctx)
			if err != nil {
				return p, err
			}
		}
		p.Defaults = defaults
	}

	return p, nil
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
	var stdout, stderr io.Writer
	if showOutput {
		stdout = os.Stdout
		stderr = os.Stderr
	}
	return runCommandWithWriters(ctx, shell, shellArg, command, env, stdout, stderr, showOutput, dir)
}

func runCommandWithWriters(ctx context.Context, shell, shellArg, command string, env map[string]string, stdout, stderr io.Writer, connectStdin bool, dir string) error {
	cmd := exec.CommandContext(ctx, shell, shellArg, command)
	cmd.SysProcAttr = sysProcAttr()
	cmd.Cancel = func() error {
		return signalProcessGroup(cmd)
	}
	cmd.WaitDelay = 3 * time.Second

	if stdout != nil {
		cmd.Stdout = stdout
	} else {
		cmd.Stdout = io.Discard
	}
	if stderr != nil {
		cmd.Stderr = stderr
	} else {
		cmd.Stderr = io.Discard
	}
	if connectStdin {
		cmd.Stdin = os.Stdin
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
