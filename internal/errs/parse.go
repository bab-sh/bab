package errs

import (
	"fmt"
	"strings"
)

type ParseError struct {
	Path    string
	Line    int
	Column  int
	Message string
	Cause   error
}

func (e *ParseError) Error() string {
	loc := FormatLocation(e.Path, e.Line, e.Column)
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", loc, e.Cause)
	}
	return fmt.Sprintf("%s: %s", loc, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

func (e *ParseError) Is(target error) bool {
	switch target {
	case ErrInvalidYAML:
		return strings.Contains(e.Message, "YAML")
	case ErrFileNotFound:
		return e.Message == "file not found"
	case ErrPathEmpty:
		return e.Message == "path cannot be empty"
	default:
		return false
	}
}

type CircularDepError struct {
	Path  string
	Type  string
	Chain []string
}

func (e *CircularDepError) Error() string {
	chainStr := strings.Join(e.Chain, " → ")
	if e.Path != "" {
		return fmt.Sprintf("%s: circular %s: %s", RelativePath(e.Path), e.Type, chainStr)
	}
	return fmt.Sprintf("circular %s: %s", e.Type, chainStr)
}

func (e *CircularDepError) Is(target error) bool {
	return target == ErrCircularDep
}

type TaskNotFoundError struct {
	Path         string
	Line         int
	TaskName     string
	ReferencedBy string
	Available    []string
}

func (e *TaskNotFoundError) Error() string {
	var loc string
	if e.Path != "" {
		path := RelativePath(e.Path)
		if e.Line > 0 {
			loc = fmt.Sprintf("%s:%d: ", path, e.Line)
		} else {
			loc = path + ": "
		}
	}

	var suffix string
	if suggestion := FindSimilar(e.TaskName, e.Available); suggestion != "" {
		suffix = fmt.Sprintf(" (did you mean %q?)", suggestion)
	}

	return fmt.Sprintf("%stask %q not found%s", loc, e.TaskName, suffix)
}

func (e *TaskNotFoundError) Is(target error) bool {
	return target == ErrTaskNotFound
}

type DuplicateTaskError struct {
	Path         string
	Line         int
	TaskName     string
	OriginalLine int
}

func (e *DuplicateTaskError) Error() string {
	path := RelativePath(e.Path)
	return fmt.Sprintf("%s:%d: duplicate task %q (first defined at line %d)", path, e.Line, e.TaskName, e.OriginalLine)
}

func (e *DuplicateTaskError) Is(target error) bool {
	return target == ErrDuplicateTask
}

type AliasConflictError struct {
	Path     string
	Line     int
	Alias    string
	TaskName string
}

func (e *AliasConflictError) Error() string {
	path := RelativePath(e.Path)
	return fmt.Sprintf("%s:%d: alias %q for task %q conflicts with existing task name", path, e.Line, e.Alias, e.TaskName)
}

func (e *AliasConflictError) Is(target error) bool {
	return target == ErrAliasConflict
}

type DuplicateAliasError struct {
	Path         string
	Line         int
	Alias        string
	TaskName     string
	OriginalTask string
	OriginalLine int
}

func (e *DuplicateAliasError) Error() string {
	path := RelativePath(e.Path)
	return fmt.Sprintf("%s:%d: alias %q for task %q already defined by task %q at line %d",
		path, e.Line, e.Alias, e.TaskName, e.OriginalTask, e.OriginalLine)
}

func (e *DuplicateAliasError) Is(target error) bool {
	return target == ErrDuplicateAlias
}
