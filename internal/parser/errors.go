package parser

import (
	"fmt"
	"strings"
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
	return fmt.Sprintf("circular %s detected: %s", e.Type, strings.Join(e.Chain, " -> "))
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
