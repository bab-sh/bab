package parser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrPathEmpty     = errors.New("path cannot be empty")
	ErrInvalidYAML   = errors.New("invalid YAML")
	ErrFileNotFound  = errors.New("file not found")
	ErrCircularDep   = errors.New("circular dependency")
	ErrTaskNotFound  = errors.New("task not found")
	ErrDuplicateTask = errors.New("duplicate task")
)

type ParseError struct {
	Path    string
	Line    int
	Column  int
	Message string
	Cause   error
}

func (e *ParseError) Error() string {
	loc := e.location()
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", loc, e.Cause)
	}
	return fmt.Sprintf("%s: %s", loc, e.Message)
}

func (e *ParseError) location() string {
	path := relativePath(e.Path)
	if e.Line > 0 {
		if e.Column > 0 {
			return fmt.Sprintf("%s:%d:%d", path, e.Line, e.Column)
		}
		return fmt.Sprintf("%s:%d", path, e.Line)
	}
	return path
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

type CircularError struct {
	Path  string
	Type  string
	Chain []string
}

func (e *CircularError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: circular %s: %s", relativePath(e.Path), e.Type, strings.Join(e.Chain, " → "))
	}
	return fmt.Sprintf("circular %s: %s", e.Type, strings.Join(e.Chain, " → "))
}

func (e *CircularError) Is(target error) bool {
	return target == ErrCircularDep
}

type NotFoundError struct {
	Path         string
	Line         int
	TaskName     string
	ReferencedBy string
	Available    []string
}

func (e *NotFoundError) Error() string {
	var loc string
	if e.Path != "" {
		path := relativePath(e.Path)
		if e.Line > 0 {
			loc = fmt.Sprintf("%s:%d: ", path, e.Line)
		} else {
			loc = path + ": "
		}
	}

	var suffix string
	if suggestion := findSimilar(e.TaskName, e.Available); suggestion != "" {
		suffix = fmt.Sprintf(" (did you mean %q?)", suggestion)
	}

	return fmt.Sprintf("%stask %q not found%s", loc, e.TaskName, suffix)
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrTaskNotFound
}

type DuplicateError struct {
	Path         string
	Line         int
	TaskName     string
	OriginalLine int
}

func (e *DuplicateError) Error() string {
	path := relativePath(e.Path)
	return fmt.Sprintf("%s:%d: duplicate task %q (first defined at line %d)", path, e.Line, e.TaskName, e.OriginalLine)
}

func (e *DuplicateError) Is(target error) bool {
	return target == ErrDuplicateTask
}

type ValidationErrors struct {
	Errors []error
}

func (e *ValidationErrors) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "\n")
}

func (e *ValidationErrors) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *ValidationErrors) OrNil() error {
	if e.HasErrors() {
		return e
	}
	return nil
}

func (e *ValidationErrors) Unwrap() []error {
	return e.Errors
}

func relativePath(path string) string {
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, path); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return path
}

var yamlLineRegex = regexp.MustCompile(`line (\d+):?`)

func extractYAMLLocation(err error) int {
	if err == nil {
		return 0
	}
	matches := yamlLineRegex.FindStringSubmatch(err.Error())
	if len(matches) >= 2 {
		if n, e := strconv.Atoi(matches[1]); e == nil {
			return n
		}
	}
	return 0
}

func cleanYAMLError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	msg = strings.TrimPrefix(msg, "yaml: ")
	msg = strings.TrimPrefix(msg, "unmarshal errors:\n  ")
	msg = yamlLineRegex.ReplaceAllString(msg, "")
	msg = strings.TrimSpace(msg)
	return msg
}

func findSimilar(target string, candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}

	sorted := make([]string, len(candidates))
	copy(sorted, candidates)
	sort.Strings(sorted)

	target = strings.ToLower(target)
	var best string
	bestScore := 0

	for _, c := range sorted {
		score := similarity(target, strings.ToLower(c))
		if score >= 50 && (score > bestScore || (score == bestScore && len(c) < len(best))) {
			bestScore = score
			best = c
		}
	}
	return best
}

func similarity(a, b string) int {
	if a == b {
		return 100
	}
	if strings.Contains(b, a) || strings.Contains(a, b) {
		return 80
	}
	matches := 0
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] == b[i] {
			matches++
		}
	}
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	return (matches * 100) / max(len(a), len(b))
}
