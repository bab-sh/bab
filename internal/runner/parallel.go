package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/interpolate"
	"github.com/bab-sh/bab/internal/output"
	"github.com/bab-sh/bab/internal/tui"
	"github.com/charmbracelet/log"
	"golang.org/x/term"
)

type syncState struct {
	mu    sync.Mutex
	state map[string]status
}

func (s *syncState) get(name string) status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state[name]
}

func (s *syncState) set(name string, st status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[name] = st
}

func (s *syncState) claim(name string) status {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.state[name]
	if current == 0 {
		s.state[name] = running
	}
	return current
}

func (s *syncState) snapshot() map[string]status {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make(map[string]status, len(s.state))
	for k, v := range s.state {
		cp[k] = v
	}
	return cp
}

func (r *Runner) executeParallel(ctx context.Context, pr babfile.ParallelRun, task *babfile.Task, tasks babfile.TaskMap, state *syncState, taskVars map[string]string, taskEnv map[string]string, overrideSilent, overrideOutput *bool) error {
	labels := make([]string, len(pr.Items))
	maxLabelLen := 0
	for i := range pr.Items {
		labels[i] = pr.ItemLabel(i)
		if len(labels[i]) > maxLabelLen {
			maxLabelLen = len(labels[i])
		}
	}

	if err := r.preResolveDeps(ctx, pr.Items, tasks, state, overrideSilent, overrideOutput); err != nil {
		return err
	}

	if r.DryRun {
		for i, item := range pr.Items {
			log.Info("Would run in parallel", "index", i, "label", labels[i], "item", fmt.Sprintf("%T", item))
		}
		return nil
	}

	noColor := !pr.UseColor()
	useGroupedTUI := pr.Mode == babfile.ParallelGrouped && term.IsTerminal(int(os.Stderr.Fd()))

	if useGroupedTUI {
		return r.executeParallelGrouped(ctx, pr, task, tasks, state, labels, taskVars, taskEnv, overrideSilent, overrideOutput, noColor)
	}
	return r.executeParallelInterleaved(ctx, pr, task, tasks, state, labels, maxLabelLen, taskVars, taskEnv, overrideSilent, overrideOutput, noColor)
}

func (r *Runner) executeParallelInterleaved(ctx context.Context, pr babfile.ParallelRun, task *babfile.Task, tasks babfile.TaskMap, state *syncState, labels []string, maxLabelLen int, taskVars map[string]string, taskEnv map[string]string, overrideSilent, overrideOutput *bool, noColor bool) error {
	var sem chan struct{}
	if pr.Limit > 0 {
		sem = make(chan struct{}, pr.Limit)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	var firstErrOnce sync.Once

	for i, item := range pr.Items {
		wg.Add(1)

		go func(idx int, runItem babfile.RunItem) {
			defer wg.Done()

			if sem != nil {
				select {
				case sem <- struct{}{}:
				case <-ctx.Done():
					return
				}
				defer func() { <-sem }()
			}

			pw := NewPrefixWriter(labels[idx], maxLabelLen, colorForIndex(idx), os.Stdout, &mu, noColor)
			pwErr := NewPrefixWriter(labels[idx], maxLabelLen, colorForIndex(idx), os.Stderr, &mu, noColor)
			defer func() {
				_ = pw.Flush()
				_ = pwErr.Flush()
			}()

			err := r.executeRunItem(ctx, runItem, task, tasks, state, taskVars, taskEnv, overrideSilent, overrideOutput, pw, pwErr, noColor)
			if err != nil {
				firstErrOnce.Do(func() {
					firstErr = fmt.Errorf("parallel item %q failed: %w", labels[idx], err)
				})
			}
		}(i, item)
	}

	wg.Wait()
	return firstErr
}

func (r *Runner) executeParallelGrouped(ctx context.Context, pr babfile.ParallelRun, task *babfile.Task, tasks babfile.TaskMap, state *syncState, labels []string, taskVars map[string]string, taskEnv map[string]string, overrideSilent, overrideOutput *bool, noColor bool) error {
	tuiItems := make([]tui.ParallelItem, len(pr.Items))
	for i, label := range labels {
		tuiItems[i] = tui.ParallelItem{
			Label: label,
			Color: colorForIndex(i),
		}
	}

	workCtx, workCancel := context.WithCancel(ctx)
	defer workCancel()

	program, err := tui.RunParallel(tuiItems, workCancel)
	if err != nil {
		return fmt.Errorf("failed to start parallel TUI: %w", err)
	}

	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	var sem chan struct{}
	if pr.Limit > 0 {
		sem = make(chan struct{}, pr.Limit)
	}

	var wg sync.WaitGroup
	var firstErr error
	var firstErrOnce sync.Once
	itemErrs := make([]error, len(pr.Items))
	itemDone := make([]bool, len(pr.Items))

	for i, item := range pr.Items {
		wg.Add(1)

		go func(idx int, runItem babfile.RunItem) {
			defer wg.Done()

			if sem != nil {
				select {
				case sem <- struct{}{}:
				case <-workCtx.Done():
					return
				}
				defer func() { <-sem }()
			}

			lw := NewLineWriter(idx, program, noColor)
			defer lw.Flush()

			err := r.executeRunItem(workCtx, runItem, task, tasks, state, taskVars, taskEnv, overrideSilent, overrideOutput, lw, lw, noColor)
			itemErrs[idx] = err
			itemDone[idx] = true

			program.Send(tui.ItemDoneMsg{Index: idx, Err: err})

			if err != nil {
				firstErrOnce.Do(func() {
					firstErr = fmt.Errorf("parallel item %q failed: %w", labels[idx], err)
				})
			}
		}(i, item)
	}

	wg.Wait()
	if workCtx.Err() != nil {
		for i := range itemErrs {
			if !itemDone[i] {
				itemErrs[i] = context.Canceled
			}
		}
	}

	program.Send(tui.AllDoneMsg{})
	program.Wait()

	if !isSilent(overrideSilent, task.Silent, r.GlobalSilent) {
		output.ParallelDone(labels, itemErrs)
	}

	return firstErr
}

func (r *Runner) executeRunItem(ctx context.Context, item babfile.RunItem, task *babfile.Task, tasks babfile.TaskMap, state *syncState, taskVars map[string]string, taskEnv map[string]string, overrideSilent, overrideOutput *bool, stdout, stderr io.Writer, noColor bool) error {
	shell, shellArg := shellCommand()

	switch v := item.(type) {
	case babfile.CommandRun:
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
			return err
		}

		if !isSilent(v.Silent, overrideSilent, task.Silent, r.GlobalSilent) && stderr != nil {
			_, _ = fmt.Fprintln(stderr, output.RenderCmd(interpolatedCmd))
		}

		showOutput := isOutput(v.Output, overrideOutput, task.Output, r.GlobalOutput)
		if showOutput {
			return runCommandWithWriters(ctx, shell, shellArg, interpolatedCmd, cmdEnv, stdout, stderr, false, noColor, cmdDir)
		}
		return runCommandWithWriters(ctx, shell, shellArg, interpolatedCmd, cmdEnv, nil, nil, false, noColor, cmdDir)

	case babfile.TaskRun:
		return r.runTask(ctx, v.Task, tasks, state, false, v.Silent, v.Output, stdout, stderr, noColor)

	case babfile.LogRun:
		logCtx := interpolate.NewContextWithLocation(taskVars, r.BabfilePath, v.Line)
		interpolatedLog, err := interpolate.Interpolate(v.Log, logCtx)
		if err != nil {
			return err
		}
		if stdout != nil {
			_, _ = fmt.Fprintln(stdout, output.RenderLog(interpolatedLog, v.Level))
		}
		return nil

	default:
		return fmt.Errorf("unsupported run item type in parallel: %T", item)
	}
}

func (r *Runner) preResolveDeps(ctx context.Context, items []babfile.RunItem, tasks babfile.TaskMap, state *syncState, overrideSilent, overrideOutput *bool) error {
	seen := make(map[string]bool)
	for _, item := range items {
		tr, ok := item.(babfile.TaskRun)
		if !ok {
			continue
		}
		task, exists := tasks[tr.Task]
		if !exists {
			continue
		}
		for _, dep := range task.Deps {
			if seen[dep] {
				continue
			}
			seen[dep] = true
			if state.get(dep) == done {
				continue
			}
			if err := r.runTask(ctx, dep, tasks, state, false, overrideSilent, overrideOutput, nil, nil, false); err != nil {
				return fmt.Errorf("parallel pre-dependency %q failed: %w", dep, err)
			}
		}
	}
	return nil
}
