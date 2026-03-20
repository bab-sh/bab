package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/interpolate"
	"github.com/bab-sh/bab/internal/output"
	"github.com/bab-sh/bab/internal/tui"
	"github.com/charmbracelet/log"
	"golang.org/x/term"
)

type ParallelContext struct {
	Program *tea.Program
	Path    []int
}

func pathToKey(path []int) string {
	parts := make([]string, len(path))
	for i, v := range path {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, ".")
}

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

func (r *Runner) executeParallel(ctx context.Context, pr babfile.ParallelRun, task *babfile.Task, tasks babfile.TaskMap, state *syncState, taskVars map[string]string, taskEnv map[string]string, overrideSilent, overrideOutput *bool, stdout, stderr io.Writer, parentNoColor bool, pctx *ParallelContext) error {
	labels := make([]string, len(pr.Items))
	maxLabelLen := 0
	for i := range pr.Items {
		labels[i] = pr.ItemLabel(i)
		if len(labels[i]) > maxLabelLen {
			maxLabelLen = len(labels[i])
		}
	}

	noColor := parentNoColor || !pr.UseColor()

	if pr.Silent != nil {
		overrideSilent = pr.Silent
	}
	if pr.Output != nil {
		overrideOutput = pr.Output
	}

	if err := r.preResolveDeps(ctx, pr.Items, tasks, state, overrideSilent, overrideOutput, stdout, stderr, noColor, pctx); err != nil {
		return err
	}

	if r.DryRun {
		for i, item := range pr.Items {
			log.Info("Would run in parallel", "index", i, "label", labels[i], "item", fmt.Sprintf("%T", item))
		}
		return nil
	}

	isTerminal := term.IsTerminal(int(os.Stderr.Fd()))
	hasParentGrouped := pctx != nil && pctx.Program != nil
	useGroupedTUI := pr.Mode == babfile.ParallelGrouped && isTerminal && (stdout == nil || hasParentGrouped)

	if pr.Mode == babfile.ParallelGrouped && !useGroupedTUI {
		log.Debug("Grouped parallel downgraded to interleaved", "reason", "nested inside non-grouped parent")
	}

	if useGroupedTUI {
		return r.executeParallelGrouped(ctx, pr, task, tasks, state, labels, taskVars, taskEnv, overrideSilent, overrideOutput, noColor, pctx)
	}
	return r.executeParallelInterleaved(ctx, pr, task, tasks, state, labels, maxLabelLen, taskVars, taskEnv, overrideSilent, overrideOutput, noColor, stdout, stderr)
}

func (r *Runner) executeParallelInterleaved(ctx context.Context, pr babfile.ParallelRun, task *babfile.Task, tasks babfile.TaskMap, state *syncState, labels []string, maxLabelLen int, taskVars map[string]string, taskEnv map[string]string, overrideSilent, overrideOutput *bool, noColor bool, parentOut, parentErr io.Writer) error {
	var sem chan struct{}
	if pr.Limit > 0 {
		sem = make(chan struct{}, pr.Limit)
	}

	outDest, errDest := io.Writer(os.Stdout), io.Writer(os.Stderr)
	if parentOut != nil {
		outDest = parentOut
	}
	if parentErr != nil {
		errDest = parentErr
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

			pw := NewPrefixWriter(labels[idx], maxLabelLen, colorForPath([]int{idx}), outDest, &mu, noColor)
			pwErr := NewPrefixWriter(labels[idx], maxLabelLen, colorForPath([]int{idx}), errDest, &mu, noColor)

			err := r.executeRunItem(ctx, runItem, task, tasks, state, taskVars, taskEnv, overrideSilent, overrideOutput, pw, pwErr, noColor, nil)
			_ = pw.Flush()
			_ = pwErr.Flush()
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

func (r *Runner) executeParallelGrouped(ctx context.Context, pr babfile.ParallelRun, task *babfile.Task, tasks babfile.TaskMap, state *syncState, labels []string, taskVars map[string]string, taskEnv map[string]string, overrideSilent, overrideOutput *bool, noColor bool, pctx *ParallelContext) error {
	var program *tea.Program
	var ownsProgram bool
	var basePath []int

	if pctx != nil && pctx.Program != nil {
		program = pctx.Program
		basePath = pctx.Path
		parentKey := pathToKey(basePath)
		for i, label := range labels {
			childPath := make([]int, len(basePath)+1)
			copy(childPath, basePath)
			childPath[len(basePath)] = i
			program.Send(tui.ItemRegisterMsg{
				Key:    pathToKey(childPath),
				Parent: parentKey,
				Label:  label,
				Color:  colorForPath(childPath),
			})
		}
	} else {
		ownsProgram = true

		tuiItems := make([]tui.ParallelItem, len(pr.Items))
		for i, label := range labels {
			tuiItems[i] = tui.ParallelItem{Label: label, Color: colorForPath([]int{i})}
		}

		workCtx, workCancel := context.WithCancel(ctx)
		defer workCancel()
		ctx = workCtx

		var err error
		program, err = tui.RunParallel(tuiItems, workCancel)
		if err != nil {
			return fmt.Errorf("failed to start parallel TUI: %w", err)
		}

		log.SetOutput(io.Discard)
		defer log.SetOutput(os.Stderr)
	}

	var sem chan struct{}
	if pr.Limit > 0 {
		sem = make(chan struct{}, pr.Limit)
	}

	var wg sync.WaitGroup
	var firstErr error
	var firstErrOnce sync.Once
	itemErrs := make([]error, len(pr.Items))

	for i, item := range pr.Items {
		wg.Add(1)

		go func(idx int, runItem babfile.RunItem) {
			defer wg.Done()

			childPath := make([]int, len(basePath)+1)
			copy(childPath, basePath)
			childPath[len(basePath)] = idx
			childKey := pathToKey(childPath)

			if sem != nil {
				select {
				case sem <- struct{}{}:
				case <-ctx.Done():
					itemErrs[idx] = context.Canceled

					program.Send(tui.ItemDoneMsg{Key: childKey, Err: context.Canceled})
					return
				}
				defer func() { <-sem }()
			}

			program.Send(tui.ItemStartMsg{Key: childKey})

			childPctx := &ParallelContext{
				Program: program,
				Path:    childPath,
			}

			lw := NewKeyLineWriter(childKey, program, noColor)

			err := r.executeRunItem(ctx, runItem, task, tasks, state, taskVars, taskEnv, overrideSilent, overrideOutput, lw, lw, noColor, childPctx)
			lw.Flush()
			itemErrs[idx] = err

			program.Send(tui.ItemDoneMsg{Key: childKey, Err: err})

			if err != nil {
				firstErrOnce.Do(func() {
					firstErr = fmt.Errorf("parallel item %q failed: %w", labels[idx], err)
				})
			}
		}(i, item)
	}

	wg.Wait()

	if pctx != nil {
		program.Send(tui.ItemClearChildrenMsg{Key: pathToKey(basePath)})
	}

	if ownsProgram {
		program.Send(tui.AllDoneMsg{})
		program.Wait()
	}

	if !isSilent(overrideSilent, task.Silent, r.GlobalSilent) {
		if ownsProgram {
			output.ParallelDone(labels, itemErrs)
		} else if pctx != nil {
			parentKey := pathToKey(pctx.Path)
			summary := output.RenderParallelDone(labels, itemErrs)
			program.Send(tui.ItemOutputMsg{Key: parentKey, Line: summary})
		}
	}

	return firstErr
}

func (r *Runner) executeRunItem(ctx context.Context, item babfile.RunItem, task *babfile.Task, tasks babfile.TaskMap, state *syncState, taskVars map[string]string, taskEnv map[string]string, overrideSilent, overrideOutput *bool, stdout, stderr io.Writer, noColor bool, pctx *ParallelContext) error {
	switch v := item.(type) {
	case babfile.CommandRun:
		shell, shellArg := shellCommand()
		return r.executeCommand(ctx, v, task, shell, shellArg, taskVars, taskEnv, overrideSilent, overrideOutput, stdout, stderr, noColor)

	case babfile.TaskRun:
		if r.DryRun {
			log.Info("Would run task", "task", v.Task)
			return nil
		}
		effSilent := firstNonNil(v.Silent, overrideSilent)
		effOutput := firstNonNil(v.Output, overrideOutput)
		return r.runTask(ctx, v.Task, tasks, state, false, effSilent, effOutput, stdout, stderr, noColor, pctx)

	case babfile.LogRun:
		logCtx := interpolate.NewContextWithLocation(taskVars, r.BabfilePath, v.Line)
		interpolatedLog, err := interpolate.Interpolate(v.Log, logCtx)
		if err != nil {
			return err
		}
		switch {
		case r.DryRun:
			log.Info("Would log", "msg", interpolatedLog, "level", v.Level)
		case stdout != nil:
			_, _ = fmt.Fprintln(stdout, output.RenderLog(interpolatedLog, v.Level))
		default:
			executeLog(babfile.LogRun{Log: interpolatedLog, Level: v.Level})
		}
		return nil

	default:
		return fmt.Errorf("unsupported run item type: %T", item)
	}
}

func (r *Runner) preResolveDeps(ctx context.Context, items []babfile.RunItem, tasks babfile.TaskMap, state *syncState, overrideSilent, overrideOutput *bool, stdout, stderr io.Writer, noColor bool, pctx *ParallelContext) error {
	seen := make(map[string]bool)
	for _, item := range items {
		tr, ok := item.(babfile.TaskRun)
		if !ok {
			continue
		}
		task, exists := tasks[tr.Task]
		if !exists {
			return fmt.Errorf("parallel item references unknown task %q", tr.Task)
		}
		for _, dep := range task.Deps {
			if seen[dep] {
				continue
			}
			seen[dep] = true
			if state.get(dep) == done {
				continue
			}
			if err := r.runTask(ctx, dep, tasks, state, false, overrideSilent, overrideOutput, stdout, stderr, noColor, pctx); err != nil {
				return fmt.Errorf("parallel pre-dependency %q failed: %w", dep, err)
			}
		}
	}
	return nil
}
