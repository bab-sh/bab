package parser

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrPathEmpty    = errors.New("path cannot be empty")
	ErrInvalidYAML  = errors.New("invalid YAML")
	ErrFileNotFound = errors.New("failed to read file")
	ErrCircularDep  = errors.New("circular dependency detected")
	ErrTaskNotFound = errors.New("task not found")
)

type ParseError struct {
	Path    string
	Message string
	Cause   error
}

func (e *ParseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Path, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

type CircularError struct {
	Type  string
	Chain []string
}

func (e *CircularError) Error() string {
	return fmt.Sprintf("circular %s detected: %s", e.Type, strings.Join(e.Chain, " → "))
}

func (e *CircularError) Is(target error) bool {
	return target == ErrCircularDep
}

type NotFoundError struct {
	TaskName     string
	ReferencedBy string
	Available    []string
}

func (e *NotFoundError) Error() string {
	if e.ReferencedBy != "" {
		return fmt.Sprintf("task %q not found (referenced by %q, available: %s)",
			e.TaskName, e.ReferencedBy, strings.Join(e.Available, ", "))
	}
	return fmt.Sprintf("task %q not found (available: %s)",
		e.TaskName, strings.Join(e.Available, ", "))
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrTaskNotFound
}
