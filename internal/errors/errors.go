// Package errors provides custom error types for the bab task runner.
package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrTaskNotFound is returned when a requested task cannot be found in the registry.
	ErrTaskNotFound = errors.New("task not found")
	// ErrInvalidBabfile is returned when a Babfile has invalid syntax or structure.
	ErrInvalidBabfile = errors.New("invalid babfile")
	// ErrNoBabfile is returned when no Babfile can be found in the project.
	ErrNoBabfile = errors.New("no babfile found")
	// ErrEmptyTaskName is returned when a task name is empty.
	ErrEmptyTaskName = errors.New("task name cannot be empty")
	// ErrNoCommands is returned when a task has no commands to execute.
	ErrNoCommands = errors.New("task has no commands")
	// ErrTaskAlreadyExists is returned when attempting to register a task that already exists.
	ErrTaskAlreadyExists = errors.New("task already exists")
	// ErrExecutionFailed is returned when a task execution fails.
	ErrExecutionFailed = errors.New("task execution failed")
	// ErrParseError is returned when parsing a Babfile fails.
	ErrParseError = errors.New("parse error")
)

// TaskNotFoundError represents an error when a specific task cannot be found.
type TaskNotFoundError struct {
	TaskName string
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("task '%s' not found", e.TaskName)
}

// Is checks if the target error is ErrTaskNotFound.
func (e *TaskNotFoundError) Is(target error) bool {
	return target == ErrTaskNotFound
}

// NewTaskNotFoundError creates a new TaskNotFoundError with the given task name.
func NewTaskNotFoundError(taskName string) error {
	return &TaskNotFoundError{TaskName: taskName}
}

// InvalidBabfileError represents an error when a Babfile is invalid.
type InvalidBabfileError struct {
	Path   string
	Reason string
}

func (e *InvalidBabfileError) Error() string {
	return fmt.Sprintf("invalid babfile '%s': %s", e.Path, e.Reason)
}

// Is checks if the target error is ErrInvalidBabfile.
func (e *InvalidBabfileError) Is(target error) bool {
	return target == ErrInvalidBabfile
}

// NewInvalidBabfileError creates a new InvalidBabfileError with the given path and reason.
func NewInvalidBabfileError(path, reason string) error {
	return &InvalidBabfileError{Path: path, Reason: reason}
}

// TaskValidationError represents an error when a task fails validation.
type TaskValidationError struct {
	TaskName string
	Field    string
	Reason   string
}

func (e *TaskValidationError) Error() string {
	return fmt.Sprintf("task '%s' validation failed: %s - %s", e.TaskName, e.Field, e.Reason)
}

// NewTaskValidationError creates a new TaskValidationError with the given task name, field, and reason.
func NewTaskValidationError(taskName, field, reason string) error {
	return &TaskValidationError{
		TaskName: taskName,
		Field:    field,
		Reason:   reason,
	}
}

// ExecutionError represents an error that occurred during task execution.
type ExecutionError struct {
	TaskName string
	Command  string
	Err      error
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf("execution failed for task '%s' (command: %s): %v", e.TaskName, e.Command, e.Err)
}

func (e *ExecutionError) Unwrap() error {
	return e.Err
}

// Is checks if the target error is ErrExecutionFailed.
func (e *ExecutionError) Is(target error) bool {
	return target == ErrExecutionFailed
}

// NewExecutionError creates a new ExecutionError with the given task name, command, and underlying error.
func NewExecutionError(taskName, command string, err error) error {
	return &ExecutionError{
		TaskName: taskName,
		Command:  command,
		Err:      err,
	}
}

// ParseError represents an error that occurred while parsing a Babfile.
type ParseError struct {
	Path   string
	Line   int
	Column int
	Err    error
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("parse error in '%s' at line %d, column %d: %v", e.Path, e.Line, e.Column, e.Err)
	}
	return fmt.Sprintf("parse error in '%s': %v", e.Path, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// Is checks if the target error is ErrParseError.
func (e *ParseError) Is(target error) bool {
	return target == ErrParseError
}

// NewParseError creates a new ParseError with the given path and underlying error.
func NewParseError(path string, err error) error {
	return &ParseError{
		Path: path,
		Err:  err,
	}
}

// NewParseErrorWithPosition creates a new ParseError with the given path, position, and underlying error.
func NewParseErrorWithPosition(path string, line, column int, err error) error {
	return &ParseError{
		Path:   path,
		Line:   line,
		Column: column,
		Err:    err,
	}
}
