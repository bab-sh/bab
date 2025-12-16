package errs

import (
	"errors"
	"strings"
)

var (
	ErrPathEmpty     = errors.New("path cannot be empty")
	ErrInvalidYAML   = errors.New("invalid YAML")
	ErrFileNotFound  = errors.New("file not found")
	ErrCircularDep   = errors.New("circular dependency")
	ErrTaskNotFound  = errors.New("task not found")
	ErrDuplicateTask = errors.New("duplicate task")

	ErrVarNotFound = errors.New("variable not found")
	ErrVarCycle    = errors.New("circular variable reference")

	ErrBabfileNotFound = errors.New("no Babfile found")

	ErrNoTasks = errors.New("no tasks available")
)

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
