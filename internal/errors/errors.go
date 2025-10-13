package errors

import (
	"errors"
	"fmt"
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrInvalidBabfile    = errors.New("invalid babfile")
	ErrNoBabfile         = errors.New("no babfile found")
	ErrEmptyTaskName     = errors.New("task name cannot be empty")
	ErrNoCommands        = errors.New("task has no commands")
	ErrTaskAlreadyExists = errors.New("task already exists")
	ErrExecutionFailed   = errors.New("task execution failed")
	ErrParseError        = errors.New("parse error")
)

type TaskNotFoundError struct {
	TaskName string
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("task '%s' not found", e.TaskName)
}

func (e *TaskNotFoundError) Is(target error) bool {
	return target == ErrTaskNotFound
}

func NewTaskNotFoundError(taskName string) error {
	return &TaskNotFoundError{TaskName: taskName}
}

type InvalidBabfileError struct {
	Path   string
	Reason string
}

func (e *InvalidBabfileError) Error() string {
	return fmt.Sprintf("invalid babfile '%s': %s", e.Path, e.Reason)
}

func (e *InvalidBabfileError) Is(target error) bool {
	return target == ErrInvalidBabfile
}

func NewInvalidBabfileError(path, reason string) error {
	return &InvalidBabfileError{Path: path, Reason: reason}
}

type TaskValidationError struct {
	TaskName string
	Field    string
	Reason   string
}

func (e *TaskValidationError) Error() string {
	return fmt.Sprintf("task '%s' validation failed: %s - %s", e.TaskName, e.Field, e.Reason)
}

func NewTaskValidationError(taskName, field, reason string) error {
	return &TaskValidationError{
		TaskName: taskName,
		Field:    field,
		Reason:   reason,
	}
}

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

func (e *ExecutionError) Is(target error) bool {
	return target == ErrExecutionFailed
}

func NewExecutionError(taskName, command string, err error) error {
	return &ExecutionError{
		TaskName: taskName,
		Command:  command,
		Err:      err,
	}
}

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

func (e *ParseError) Is(target error) bool {
	return target == ErrParseError
}

func NewParseError(path string, err error) error {
	return &ParseError{
		Path: path,
		Err:  err,
	}
}

func NewParseErrorWithPosition(path string, line, column int, err error) error {
	return &ParseError{
		Path:   path,
		Line:   line,
		Column: column,
		Err:    err,
	}
}
