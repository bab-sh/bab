package errs

import (
	"fmt"
	"strings"
)

type VarNotFoundError struct {
	Path      string
	Line      int
	Name      string
	Available []string
}

func (e *VarNotFoundError) Error() string {
	var suffix string
	if suggestion := FindSimilar(e.Name, e.Available); suggestion != "" {
		suffix = fmt.Sprintf(" (did you mean %q?)", suggestion)
	}

	loc := e.location()
	if loc != "" {
		return fmt.Sprintf("%s: undefined variable: %s%s", loc, e.Name, suffix)
	}
	return fmt.Sprintf("undefined variable: %s%s", e.Name, suffix)
}

func (e *VarNotFoundError) location() string {
	if e.Path == "" {
		return ""
	}
	path := RelativePath(e.Path)
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d", path, e.Line)
	}
	return path
}

func (e *VarNotFoundError) Is(target error) bool {
	return target == ErrVarNotFound
}

type VarCycleError struct {
	Path  string
	Line  int
	Name  string
	Chain []string
}

func (e *VarCycleError) Error() string {
	var msg string
	if len(e.Chain) > 0 {
		msg = fmt.Sprintf("circular variable reference: %s", strings.Join(e.Chain, " → "))
	} else {
		msg = fmt.Sprintf("circular variable reference: %s", e.Name)
	}

	loc := e.location()
	if loc != "" {
		return fmt.Sprintf("%s: %s", loc, msg)
	}
	return msg
}

func (e *VarCycleError) location() string {
	if e.Path == "" {
		return ""
	}
	path := RelativePath(e.Path)
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d", path, e.Line)
	}
	return path
}

func (e *VarCycleError) Is(target error) bool {
	return target == ErrVarCycle
}
