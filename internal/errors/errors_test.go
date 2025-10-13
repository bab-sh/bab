package errors

import (
	"errors"
	"testing"
)

func TestTaskNotFoundError(t *testing.T) {
	err := NewTaskNotFoundError("mytask")
	if err.Error() != "task 'mytask' not found" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, ErrTaskNotFound) {
		t.Error("TaskNotFoundError should match ErrTaskNotFound")
	}
}

func TestInvalidBabfileError(t *testing.T) {
	err := NewInvalidBabfileError("/path/to/Babfile", "syntax error")
	if err.Error() != "invalid babfile '/path/to/Babfile': syntax error" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, ErrInvalidBabfile) {
		t.Error("InvalidBabfileError should match ErrInvalidBabfile")
	}
}

func TestTaskValidationError(t *testing.T) {
	err := NewTaskValidationError("mytask", "name", "cannot be empty")
	if err.Error() != "task 'mytask' validation failed: name - cannot be empty" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestExecutionError(t *testing.T) {
	innerErr := errors.New("command failed")
	err := NewExecutionError("build", "go build", innerErr)

	if !errors.Is(err, ErrExecutionFailed) {
		t.Error("ExecutionError should match ErrExecutionFailed")
	}

	var execErr *ExecutionError
	if !errors.As(err, &execErr) {
		t.Error("should be able to extract ExecutionError")
	}

	if execErr.TaskName != "build" {
		t.Errorf("unexpected task name: %s", execErr.TaskName)
	}
}

func TestParseError(t *testing.T) {
	innerErr := errors.New("yaml decode failed")
	err := NewParseError("/path/to/Babfile", innerErr)

	if !errors.Is(err, ErrParseError) {
		t.Error("ParseError should match ErrParseError")
	}

	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Error("should be able to extract ParseError")
	}
}

func TestParseErrorWithPosition(t *testing.T) {
	innerErr := errors.New("unexpected token")
	err := NewParseErrorWithPosition("/path/to/Babfile", 10, 5, innerErr)

	var parseErr *ParseError
	if !errors.As(err, &parseErr) {
		t.Fatal("expected ParseError type")
	}

	if parseErr.Line != 10 || parseErr.Column != 5 {
		t.Errorf("unexpected position: line %d, column %d", parseErr.Line, parseErr.Column)
	}

	expected := "parse error in '/path/to/Babfile' at line 10, column 5: unexpected token"
	if parseErr.Error() != expected {
		t.Errorf("unexpected error message: %s", parseErr.Error())
	}
}
