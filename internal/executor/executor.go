package executor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/bab/bab/internal/registry"
	"github.com/fatih/color"
)

type Executor struct {
	dryRun  bool
	verbose bool
}

func New(options ...Option) *Executor {
	e := &Executor{}
	for _, opt := range options {
		opt(e)
	}
	return e
}

type Option func(*Executor)

func WithDryRun(dryRun bool) Option {
	return func(e *Executor) {
		e.dryRun = dryRun
	}
}

func WithVerbose(verbose bool) Option {
	return func(e *Executor) {
		e.verbose = verbose
	}
}

func (e *Executor) Execute(task *registry.Task) error {
	if task == nil {
		return fmt.Errorf("cannot execute nil task")
	}

	header := color.New(color.FgCyan, color.Bold)
	header.Printf("\n▶ Running task: %s\n", task.Name)

	if task.Description != "" && e.verbose {
		desc := color.New(color.FgGreen)
		desc.Printf("  %s\n", task.Description)
	}

	for i, command := range task.Commands {
		if err := e.runCommand(command, i+1, len(task.Commands)); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	success := color.New(color.FgGreen, color.Bold)
	success.Printf("✓ Task completed: %s\n\n", task.Name)

	return nil
}

func (e *Executor) runCommand(command string, current, total int) error {
	cmdColor := color.New(color.FgYellow)
	if e.verbose || e.dryRun {
		cmdColor.Printf("  [%d/%d] $ %s\n", current, total, command)
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
		errColor := color.New(color.FgRed, color.Bold)
		errColor.Printf("  ✗ Command failed with exit code: %v\n", err)
		return err
	}

	return nil
}

func (e *Executor) streamOutput(pipe io.ReadCloser, isStderr bool) {
	defer pipe.Close()
	scanner := bufio.NewScanner(pipe)

	prefix := "  "
	if isStderr {
		prefix = color.RedString("  ")
	}

	for scanner.Scan() {
		fmt.Printf("%s%s\n", prefix, scanner.Text())
	}
}