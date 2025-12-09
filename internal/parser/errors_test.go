package parser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParseError_Error(t *testing.T) {
	cwd, _ := os.Getwd()

	tests := []struct {
		name     string
		err      *ParseError
		contains string
	}{
		{
			name:     "with line and column",
			err:      &ParseError{Path: filepath.Join(cwd, "test.yml"), Line: 10, Column: 5, Message: "syntax error"},
			contains: "test.yml:10:5: syntax error",
		},
		{
			name:     "with line only",
			err:      &ParseError{Path: filepath.Join(cwd, "test.yml"), Line: 10, Message: "syntax error"},
			contains: "test.yml:10: syntax error",
		},
		{
			name:     "path only",
			err:      &ParseError{Path: filepath.Join(cwd, "test.yml"), Message: "file error"},
			contains: "test.yml: file error",
		},
		{
			name:     "with cause",
			err:      &ParseError{Path: filepath.Join(cwd, "test.yml"), Cause: fmt.Errorf("wrapped error")},
			contains: "wrapped error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got == "" || !containsString(got, tt.contains) {
				t.Errorf("ParseError.Error() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestParseError_Is(t *testing.T) {
	tests := []struct {
		name   string
		err    *ParseError
		target error
		want   bool
	}{
		{
			name:   "invalid YAML",
			err:    &ParseError{Message: "invalid YAML syntax"},
			target: ErrInvalidYAML,
			want:   true,
		},
		{
			name:   "file not found",
			err:    &ParseError{Message: "file not found"},
			target: ErrFileNotFound,
			want:   true,
		},
		{
			name:   "path empty",
			err:    &ParseError{Message: "path cannot be empty"},
			target: ErrPathEmpty,
			want:   true,
		},
		{
			name:   "no match",
			err:    &ParseError{Message: "some other error"},
			target: ErrInvalidYAML,
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
	err := &ParseError{Cause: cause}

	if !errors.Is(err.Unwrap(), cause) {
		t.Error("Unwrap should return the cause")
	}

	errNoCause := &ParseError{}
	if errNoCause.Unwrap() != nil {
		t.Error("Unwrap should return nil when no cause")
	}
}

func TestCircularError_Error(t *testing.T) {
	cwd, _ := os.Getwd()

	tests := []struct {
		name     string
		err      *CircularError
		contains string
	}{
		{
			name:     "with path",
			err:      &CircularError{Path: filepath.Join(cwd, "test.yml"), Type: "dependency", Chain: []string{"a", "b", "a"}},
			contains: "test.yml: circular dependency: a → b → a",
		},
		{
			name:     "without path",
			err:      &CircularError{Type: "include", Chain: []string{"x", "y", "x"}},
			contains: "circular include: x → y → x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !containsString(got, tt.contains) {
				t.Errorf("CircularError.Error() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestCircularError_Is(t *testing.T) {
	err := &CircularError{Type: "dependency", Chain: []string{"a", "b"}}

	if !errors.Is(err, ErrCircularDep) {
		t.Error("CircularError should match ErrCircularDep")
	}

	if errors.Is(err, ErrTaskNotFound) {
		t.Error("CircularError should not match ErrTaskNotFound")
	}
}

func TestNotFoundError_Error(t *testing.T) {
	cwd, _ := os.Getwd()

	tests := []struct {
		name     string
		err      *NotFoundError
		contains string
	}{
		{
			name:     "with path and line",
			err:      &NotFoundError{Path: filepath.Join(cwd, "test.yml"), Line: 5, TaskName: "build"},
			contains: "test.yml:5: task \"build\" not found",
		},
		{
			name:     "with path only",
			err:      &NotFoundError{Path: filepath.Join(cwd, "test.yml"), TaskName: "build"},
			contains: "test.yml: task \"build\" not found",
		},
		{
			name:     "no path",
			err:      &NotFoundError{TaskName: "build"},
			contains: "task \"build\" not found",
		},
		{
			name:     "with suggestion",
			err:      &NotFoundError{TaskName: "biuld", Available: []string{"build", "test"}},
			contains: "did you mean \"build\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if !containsString(got, tt.contains) {
				t.Errorf("NotFoundError.Error() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestNotFoundError_Is(t *testing.T) {
	err := &NotFoundError{TaskName: "build"}

	if !errors.Is(err, ErrTaskNotFound) {
		t.Error("NotFoundError should match ErrTaskNotFound")
	}

	if errors.Is(err, ErrCircularDep) {
		t.Error("NotFoundError should not match ErrCircularDep")
	}
}

func TestDuplicateError_Error(t *testing.T) {
	cwd, _ := os.Getwd()
	err := &DuplicateError{
		Path:         filepath.Join(cwd, "test.yml"),
		Line:         20,
		TaskName:     "build",
		OriginalLine: 5,
	}

	got := err.Error()
	if !containsString(got, "duplicate task \"build\"") {
		t.Errorf("expected 'duplicate task' in error, got: %s", got)
	}
	if !containsString(got, "line 5") {
		t.Errorf("expected 'line 5' in error, got: %s", got)
	}
}

func TestDuplicateError_Is(t *testing.T) {
	err := &DuplicateError{TaskName: "build"}

	if !errors.Is(err, ErrDuplicateTask) {
		t.Error("DuplicateError should match ErrDuplicateTask")
	}

	if errors.Is(err, ErrCircularDep) {
		t.Error("DuplicateError should not match ErrCircularDep")
	}
}

func TestValidationErrors_Error(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		errs := &ValidationErrors{}
		if errs.Error() != "" {
			t.Errorf("empty ValidationErrors should return empty string")
		}
	})

	t.Run("single error", func(t *testing.T) {
		errs := &ValidationErrors{}
		errs.Add(fmt.Errorf("error one"))

		got := errs.Error()
		if got != "error one" {
			t.Errorf("got %q, want %q", got, "error one")
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := &ValidationErrors{}
		errs.Add(fmt.Errorf("error one"))
		errs.Add(fmt.Errorf("error two"))

		got := errs.Error()
		if !containsString(got, "error one") || !containsString(got, "error two") {
			t.Errorf("expected both errors in output, got: %s", got)
		}
	})
}

func TestValidationErrors_Add(t *testing.T) {
	errs := &ValidationErrors{}

	errs.Add(nil)
	if len(errs.Errors) != 0 {
		t.Error("Add(nil) should not add to errors")
	}

	errs.Add(fmt.Errorf("test error"))
	if len(errs.Errors) != 1 {
		t.Error("Add should add non-nil error")
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	errs := &ValidationErrors{}

	if errs.HasErrors() {
		t.Error("empty ValidationErrors should not have errors")
	}

	errs.Add(fmt.Errorf("test"))
	if !errs.HasErrors() {
		t.Error("ValidationErrors with error should have errors")
	}
}

func TestValidationErrors_OrNil(t *testing.T) {
	errs := &ValidationErrors{}

	if errs.OrNil() != nil {
		t.Error("OrNil should return nil when no errors")
	}

	errs.Add(fmt.Errorf("test"))
	if errs.OrNil() == nil {
		t.Error("OrNil should return self when has errors")
	}
}

func TestValidationErrors_Unwrap(t *testing.T) {
	errs := &ValidationErrors{}
	err1 := fmt.Errorf("error 1")
	err2 := fmt.Errorf("error 2")
	errs.Add(err1)
	errs.Add(err2)

	unwrapped := errs.Unwrap()
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
			got := findSimilar(tt.target, tt.candidates)
			if got != tt.want {
				t.Errorf("findSimilar(%q, %v) = %q, want %q", tt.target, tt.candidates, got, tt.want)
			}
		})
	}
}

func TestSimilarity(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		min  int
	}{
		{"exact match", "build", "build", 100},
		{"substring", "bui", "build", 80},
		{"reverse substring", "build", "bui", 80},
		{"partial match", "build", "built", 50},
		{"no match", "xyz", "abc", 0},
		{"empty strings", "", "", 100},
		{"one empty", "build", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := similarity(tt.a, tt.b)
			if got < tt.min {
				t.Errorf("similarity(%q, %q) = %d, want >= %d", tt.a, tt.b, got, tt.min)
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
			if got := extractYAMLLocation(tt.err); got != tt.want {
				t.Errorf("extractYAMLLocation() = %d, want %d", got, tt.want)
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
			got := cleanYAMLError(tt.err)
			if !containsString(got, tt.want) && tt.want != "" {
				t.Errorf("cleanYAMLError() = %q, want to contain %q", got, tt.want)
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
		got := relativePath(path)
		if got != "test.yml" {
			t.Errorf("relativePath() = %q, want %q", got, "test.yml")
		}
	})

	t.Run("outside cwd", func(t *testing.T) {
		path := "/some/absolute/path.yml"
		got := relativePath(path)
		if got != path && !containsString(got, "..") {
			t.Errorf("relativePath() = %q, unexpected", got)
		}
	})
}

func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (s == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
