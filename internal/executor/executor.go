// Package executor provides functionality for executing tasks with configurable options.
package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	baberrors "github.com/bab-sh/bab/internal/errors"
	"github.com/bab-sh/bab/internal/history"
	"github.com/bab-sh/bab/internal/registry"
	"github.com/charmbracelet/log"
)

// Executor executes tasks with configurable options.
type Executor struct {
	dryRun         bool
	verbose        bool
	projectRoot    string
	historyManager *history.Manager
	ctx            context.Context
}

// New creates a new Executor with the given options.
func New(options ...Option) *Executor {
	e := &Executor{
		ctx: context.Background(),
	}
	for _, opt := range options {
		opt(e)
	}

	if e.projectRoot != "" {
		historyManager, err := history.NewManager(e.projectRoot)
		if err != nil {
			log.Debug("Failed to initialize history manager", "error", err)
		} else {
			e.historyManager = historyManager
		}
	}

	return e
}

// Option is a functional option for configuring the Executor.
type Option func(*Executor)

// WithDryRun enables dry-run mode (shows commands without executing).
func WithDryRun(dryRun bool) Option {
	return func(e *Executor) {
		e.dryRun = dryRun
	}
}

// WithVerbose enables verbose output.
func WithVerbose(verbose bool) Option {
	return func(e *Executor) {
		e.verbose = verbose
	}
}

func WithProjectRoot(projectRoot string) Option {
	return func(e *Executor) {
		e.projectRoot = projectRoot
	}
}

func WithContext(ctx context.Context) Option {
	return func(e *Executor) {
		e.ctx = ctx
	}
}

// Execute runs the given task.
func (e *Executor) Execute(task *registry.Task) error {
	if task == nil {
		return fmt.Errorf("cannot execute nil task")
	}

	startTime := time.Now()
	log.Info("▶ Running task", "name", task.Name)

	if task.Description != "" && e.verbose {
		log.Debug("Task description", "desc", task.Description)
	}

	var execErr error
	for i, command := range task.Commands {
		if err := e.runCommand(command, i+1, len(task.Commands)); err != nil {
			execErr = fmt.Errorf("command failed: %w", err)
			break
		}
	}

	// Record history entry
	e.recordHistory(task, startTime, execErr)

	if execErr != nil {
		return execErr
	}

	log.Info("Task completed", "name", task.Name)

	return nil
}

func (e *Executor) recordHistory(task *registry.Task, startTime time.Time, execErr error) {
	if e.historyManager == nil {
		return
	}

	workDir, err := os.Getwd()
	if err != nil {
		workDir = ""
	}

	duration := time.Since(startTime)
	status := history.StatusSuccess
	if execErr != nil {
		status = history.StatusFailure
	}

	entry := history.NewEntry(task.Name, task.Description, workDir, status, duration, execErr)

	if err := e.historyManager.Record(entry); err != nil {
		log.Debug("Failed to record history", "error", err)
	}
}

func (e *Executor) runCommand(command string, current, total int) error {
	if e.verbose || e.dryRun {
		log.Debug("Command", "step", fmt.Sprintf("[%d/%d]", current, total), "cmd", command)
	}

	if e.dryRun {
		return nil
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(e.ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(e.ctx, "sh", "-c", command)
	}

	cmd.Stdin = os.Stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return baberrors.NewExecutionError("", command, fmt.Errorf("failed to start: %w", err))
	}

	go e.streamOutput(stdout, false)
	go e.streamOutput(stderr, true)

	if err := cmd.Wait(); err != nil {
		return baberrors.NewExecutionError("", command, err)
	}

	return nil
}

func (e *Executor) streamOutput(pipe io.ReadCloser, isStderr bool) {
	scanner := bufio.NewScanner(pipe)

	for scanner.Scan() {
		line := scanner.Text()
		if isStderr {
			log.Error(line)
		} else {
			fmt.Printf("  %s\n", line)
		}
	}
}
