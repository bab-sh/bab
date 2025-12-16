package parser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bab-sh/bab/internal/errs"
)

func TestParseError_Error(t *testing.T) {
	cwd, _ := os.Getwd()

	tests := []struct {
		name     string
		err      *errs.ParseError
		contains string
	}{
		{
			name:     "with line and column",
			err:      &errs.ParseError{Path: filepath.Join(cwd, "test.yml"), Line: 10, Column: 5, Message: "syntax error"},
			contains: "test.yml:10:5: syntax error",
		},
		{
			name:     "with line only",
			err:      &errs.ParseError{Path: filepath.Join(cwd, "test.yml"), Line: 10, Message: "syntax error"},
			contains: "test.yml:10: syntax error",
		},
		{
			name:     "path only",
			err:      &errs.ParseError{Path: filepath.Join(cwd, "test.yml"), Message: "file error"},
			contains: "test.yml: file error",
		},
		{
			name:     "with cause",
			err:      &errs.ParseError{Path: filepath.Join(cwd, "test.yml"), Cause: fmt.Errorf("wrapped error")},
			contains: "wrapped error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got == "" || !strings.Contains(got, tt.contains) {
				t.Errorf("ParseError.Error() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestParseError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    *errs.ParseError
		target error
		want   bool
	}{
		{
			name:   "invalid YAML",
			err:    &errs.ParseError{Message: "invalid YAML syntax"},
			target: errs.ErrInvalidYAML,
			want:   true,
		},
		{
			name:   "file not found",
			err:    &errs.ParseError{Message: "file not found"},
			target: errs.ErrFileNotFound,
			want:   true,
		},
		{
			name:   "path empty",
			err:    &errs.ParseError{Message: "path cannot be empty"},
			target: errs.ErrPathEmpty,
			want:   true,
		},
		{
			name:   "no match",
			err:    &errs.ParseError{Message: "some other error"},
			target: errs.ErrInvalidYAML,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("underlying cause")
	err := &errs.ParseError{Cause: cause}

	if !errors.Is(err.Unwrap(), cause) {
		t.Error("Unwrap should return the cause")
	}

	errNoCause := &errs.ParseError{}
	if errNoCause.Unwrap() != nil {
		t.Error("Unwrap should return nil when no cause")
	}
}

func TestCircularDepError_Error(t *testing.T) {
	cwd, _ := os.Getwd()

	tests := []struct {
		name     string
		err      *errs.CircularDepError
		contains string
	}{
		{
			name:     "with path",
			err:      &errs.CircularDepError{Path: filepath.Join(cwd, "test.yml"), Type: "dependency", Chain: []string{"a", "b", "a"}},
			contains: "test.yml: circular dependency: a → b → a",
		},
		{
			name:     "without path",
			err:      &errs.CircularDepError{Type: "include", Chain: []string{"x", "y", "x"}},
			contains: "circular include: x → y → x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("CircularDepError.Error() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestCircularDepError_Is(t *testing.T) {
	err := &errs.CircularDepError{Type: "dependency", Chain: []string{"a", "b"}}

	if !errors.Is(err, errs.ErrCircularDep) {
		t.Error("CircularDepError should match ErrCircularDep")
	}

	if errors.Is(err, errs.ErrTaskNotFound) {
		t.Error("CircularDepError should not match ErrTaskNotFound")
	}
}

func TestTaskNotFoundError_Error(t *testing.T) {
	cwd, _ := os.Getwd()

	tests := []struct {
		name     string
		err      *errs.TaskNotFoundError
		contains string
	}{
		{
			name:     "with path and line",
			err:      &errs.TaskNotFoundError{Path: filepath.Join(cwd, "test.yml"), Line: 5, TaskName: "build"},
			contains: "test.yml:5: task \"build\" not found",
		},
		{
			name:     "with path only",
			err:      &errs.TaskNotFoundError{Path: filepath.Join(cwd, "test.yml"), TaskName: "build"},
			contains: "test.yml: task \"build\" not found",
		},
		{
			name:     "no path",
			err:      &errs.TaskNotFoundError{TaskName: "build"},
			contains: "task \"build\" not found",
		},
		{
			name:     "with suggestion",
			err:      &errs.TaskNotFoundError{TaskName: "biuld", Available: []string{"build", "test"}},
			contains: "did you mean \"build\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("TaskNotFoundError.Error() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestTaskNotFoundError_Is(t *testing.T) {
	err := &errs.TaskNotFoundError{TaskName: "build"}

	if !errors.Is(err, errs.ErrTaskNotFound) {
		t.Error("TaskNotFoundError should match ErrTaskNotFound")
	}

	if errors.Is(err, errs.ErrCircularDep) {
		t.Error("TaskNotFoundError should not match ErrCircularDep")
	}
}

func TestDuplicateTaskError_Error(t *testing.T) {
	cwd, _ := os.Getwd()
	err := &errs.DuplicateTaskError{
		Path:         filepath.Join(cwd, "test.yml"),
		Line:         20,
		TaskName:     "build",
		OriginalLine: 5,
	}

	got := err.Error()
	if !strings.Contains(got, "duplicate task \"build\"") {
		t.Errorf("expected 'duplicate task' in error, got: %s", got)
	}
	if !strings.Contains(got, "line 5") {
		t.Errorf("expected 'line 5' in error, got: %s", got)
	}
}

func TestDuplicateTaskError_Is(t *testing.T) {
	err := &errs.DuplicateTaskError{TaskName: "build"}

	if !errors.Is(err, errs.ErrDuplicateTask) {
		t.Error("DuplicateTaskError should match ErrDuplicateTask")
	}

	if errors.Is(err, errs.ErrCircularDep) {
		t.Error("DuplicateTaskError should not match ErrCircularDep")
	}
}

func TestValidationErrors_Error(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		verrs := &errs.ValidationErrors{}
		if verrs.Error() != "" {
			t.Errorf("empty ValidationErrors should return empty string")
		}
	})

	t.Run("single error", func(t *testing.T) {
		verrs := &errs.ValidationErrors{}
		verrs.Add(fmt.Errorf("error one"))

		got := verrs.Error()
		if got != "error one" {
			t.Errorf("got %q, want %q", got, "error one")
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		verrs := &errs.ValidationErrors{}
		verrs.Add(fmt.Errorf("error one"))
		verrs.Add(fmt.Errorf("error two"))

		got := verrs.Error()
		if !strings.Contains(got, "error one") || !strings.Contains(got, "error two") {
			t.Errorf("expected both errors in output, got: %s", got)
		}
	})
}

func TestValidationErrors_Add(t *testing.T) {
	verrs := &errs.ValidationErrors{}

	verrs.Add(nil)
	if len(verrs.Errors) != 0 {
		t.Error("Add(nil) should not add to errors")
	}

	verrs.Add(fmt.Errorf("test error"))
	if len(verrs.Errors) != 1 {
		t.Error("Add should add non-nil error")
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	verrs := &errs.ValidationErrors{}

	if verrs.HasErrors() {
		t.Error("empty ValidationErrors should not have errors")
	}

	verrs.Add(fmt.Errorf("test"))
	if !verrs.HasErrors() {
		t.Error("ValidationErrors with error should have errors")
	}
}

func TestValidationErrors_OrNil(t *testing.T) {
	verrs := &errs.ValidationErrors{}

	if verrs.OrNil() != nil {
		t.Error("OrNil should return nil when no errors")
	}

	verrs.Add(fmt.Errorf("test"))
	if verrs.OrNil() == nil {
		t.Error("OrNil should return self when has errors")
	}
}

func TestValidationErrors_Unwrap(t *testing.T) {
	verrs := &errs.ValidationErrors{}
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")
	verrs.Add(err1)
	verrs.Add(err2)

	unwrapped := verrs.Unwrap()
	if len(unwrapped) != 2 {
		t.Errorf("expected 2 unwrapped errors, got %d", len(unwrapped))
	}
}

func TestFindSimilar(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		candidates []string
		want       string
	}{
		{
			name:       "exact match",
			target:     "build",
			candidates: []string{"build", "test", "lint"},
			want:       "build",
		},
		{
			name:       "typo",
			target:     "biuld",
			candidates: []string{"build", "test", "lint"},
			want:       "build",
		},
		{
			name:       "substring",
			target:     "bui",
			candidates: []string{"build", "test", "lint"},
			want:       "build",
		},
		{
			name:       "case insensitive",
			target:     "BUILD",
			candidates: []string{"build", "test"},
			want:       "build",
		},
		{
			name:       "no match",
			target:     "xyz",
			candidates: []string{"build", "test", "lint"},
			want:       "",
		},
		{
			name:       "empty candidates",
			target:     "build",
			candidates: []string{},
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errs.FindSimilar(tt.target, tt.candidates)
			if got != tt.want {
				t.Errorf("FindSimilar(%q, %v) = %q, want %q", tt.target, tt.candidates, got, tt.want)
			}
		})
	}
}

func TestExtractYAMLLocation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil error", nil, 0},
		{"with line", fmt.Errorf("yaml: line 10: error"), 10},
		{"with colon", fmt.Errorf("line 5: something"), 5},
		{"no line", fmt.Errorf("some error"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errs.ExtractYAMLLocation(tt.err); got != tt.want {
				t.Errorf("ExtractYAMLLocation() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCleanYAMLError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"nil error", nil, ""},
		{"yaml prefix", fmt.Errorf("yaml: mapping error"), "mapping error"},
		{"line number", fmt.Errorf("line 10: syntax error"), "syntax error"},
		{"unmarshal prefix", fmt.Errorf("unmarshal errors:\n  line 5: error"), "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errs.CleanYAMLError(tt.err)
			if !strings.Contains(got, tt.want) && tt.want != "" {
				t.Errorf("CleanYAMLError() = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

func TestRelativePath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("within cwd", func(t *testing.T) {
		path := filepath.Join(cwd, "test.yml")
		got := errs.RelativePath(path)
		if got != "test.yml" {
			t.Errorf("RelativePath() = %q, want %q", got, "test.yml")
		}
	})

	t.Run("outside cwd", func(t *testing.T) {
		path := "/some/absolute/path.yml"
		got := errs.RelativePath(path)
		if got != path && !strings.Contains(got, "..") {
			t.Errorf("RelativePath() = %q, unexpected", got)
		}
	})
}
