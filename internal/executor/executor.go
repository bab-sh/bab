package executor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/bab-sh/bab/internal/registry"
	"github.com/charmbracelet/log"
)

// Executor executes tasks with configurable options.
type Executor struct {
	dryRun  bool
	verbose bool
}

// New creates a new Executor with the given options.
func New(options ...Option) *Executor {
	e := &Executor{}
	for _, opt := range options {
		opt(e)
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

// Execute runs the given task.
func (e *Executor) Execute(task *registry.Task) error {
	if task == nil {
		return fmt.Errorf("cannot execute nil task")
	}

	log.Info("▶ Running task", "name", task.Name)

	if task.Description != "" && e.verbose {
		log.Debug("Task description", "desc", task.Description)
	}

	for i, command := range task.Commands {
		if err := e.runCommand(command, i+1, len(task.Commands)); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	log.Info("Task completed", "name", task.Name)

	return nil
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
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
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
		return fmt.Errorf("failed to start command: %w", err)
	}

	go e.streamOutput(stdout, false)
	go e.streamOutput(stderr, true)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func (e *Executor) streamOutput(pipe io.ReadCloser, isStderr bool) {
	defer pipe.Close()
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
