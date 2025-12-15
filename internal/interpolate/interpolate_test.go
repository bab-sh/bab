package interpolate

import (
	"errors"
	"strings"
	"testing"
)

func TestInterpolate_BasicVariable(t *testing.T) {
	ctx := NewContext(map[string]string{
		"name":    "bab",
		"version": "1.0.0",
	})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single var", "Hello ${{ name }}", "Hello bab"},
		{"multiple vars", "${{ name }} v${{ version }}", "bab v1.0.0"},
		{"no whitespace", "${{name}}", "bab"},
		{"extra whitespace", "${{  name  }}", "bab"},
		{"no vars", "hello world", "hello world"},
		{"empty string", "", ""},
		{"var in middle", "app-${{ name }}-test", "app-bab-test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Interpolate(tt.input, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInterpolate_EnvAccess(t *testing.T) {
	t.Setenv("BAB_TEST_VAR", "test_value")

	ctx := NewContext(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"env var", "${{ env.BAB_TEST_VAR }}", "test_value"},
		{"unset env var", "${{ env.BAB_NONEXISTENT }}", ""},
		{"mixed", "${{ env.BAB_TEST_VAR }} rocks", "test_value rocks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Interpolate(tt.input, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInterpolate_MixedVarsAndEnv(t *testing.T) {
	t.Setenv("BAB_ENV", "production")

	ctx := NewContext(map[string]string{
		"app": "myapp",
	})

	input := "${{ app }} running in ${{ env.BAB_ENV }}"
	expected := "myapp running in production"

	result, err := Interpolate(input, ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestInterpolate_Escaping(t *testing.T) {
	ctx := NewContext(map[string]string{
		"name": "bab",
	})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"escaped braces", "Use $${{ var }} syntax", "Use ${{ var }} syntax"},
		{"escaped with real", "$${{ escaped }} and ${{ name }}", "${{ escaped }} and bab"},
		{"multiple escaped", "$${{ a }} $${{ b }}", "${{ a }} ${{ b }}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Interpolate(tt.input, ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInterpolate_UndefinedVar(t *testing.T) {
	ctx := NewContext(map[string]string{
		"name": "bab",
	})

	input := "Hello ${{ undefined }}"
	_, err := Interpolate(input, ctx)
	if err == nil {
		t.Error("expected error for undefined variable")
	}

	if !errors.Is(err, ErrVarNotFound) {
		t.Errorf("expected errors.Is(err, ErrVarNotFound), got %T", err)
	}

	var varErr *VarNotFoundError
	if !errors.As(err, &varErr) {
		t.Errorf("expected VarNotFoundError, got %T", err)
	}
	if varErr.Name != "undefined" {
		t.Errorf("expected var name 'undefined', got %q", varErr.Name)
	}
}

func TestInterpolate_UndefinedVarSuggestion(t *testing.T) {
	ctx := NewContext(map[string]string{
		"app_name":    "myapp",
		"app_version": "1.0.0",
	})

	input := "${{ app_nam }}"
	_, err := Interpolate(input, ctx)
	if err == nil {
		t.Error("expected error for undefined variable")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "did you mean") {
		t.Errorf("expected suggestion in error message, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "app_name") {
		t.Errorf("expected 'app_name' suggestion, got: %s", errMsg)
	}
}

func TestInterpolate_NilContext(t *testing.T) {
	result, err := Interpolate("hello world", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestResolveVars_Basic(t *testing.T) {
	vars := map[string]string{
		"name":    "bab",
		"version": "1.0.0",
	}

	resolved, err := ResolveVars(vars, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resolved["name"] != "bab" {
		t.Errorf("expected name='bab', got %q", resolved["name"])
	}
	if resolved["version"] != "1.0.0" {
		t.Errorf("expected version='1.0.0', got %q", resolved["version"])
	}
}

func TestResolveVars_References(t *testing.T) {
	vars := map[string]string{
		"base":   "/app",
		"build":  "${{ base }}/build",
		"output": "${{ build }}/bin",
	}

	resolved, err := ResolveVars(vars, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resolved["base"] != "/app" {
		t.Errorf("expected base='/app', got %q", resolved["base"])
	}
	if resolved["build"] != "/app/build" {
		t.Errorf("expected build='/app/build', got %q", resolved["build"])
	}
	if resolved["output"] != "/app/build/bin" {
		t.Errorf("expected output='/app/build/bin', got %q", resolved["output"])
	}
}

func TestResolveVars_WithParent(t *testing.T) {
	parent := map[string]string{
		"global": "parent_value",
	}

	vars := map[string]string{
		"local": "${{ global }}_child",
	}

	resolved, err := ResolveVars(vars, parent)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resolved["global"] != "parent_value" {
		t.Errorf("expected global='parent_value', got %q", resolved["global"])
	}
	if resolved["local"] != "parent_value_child" {
		t.Errorf("expected local='parent_value_child', got %q", resolved["local"])
	}
}

func TestResolveVars_EnvReference(t *testing.T) {
	t.Setenv("BAB_HOME", "/home/bab")

	vars := map[string]string{
		"home": "${{ env.BAB_HOME }}",
		"data": "${{ home }}/data",
	}

	resolved, err := ResolveVars(vars, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resolved["home"] != "/home/bab" {
		t.Errorf("expected home='/home/bab', got %q", resolved["home"])
	}
	if resolved["data"] != "/home/bab/data" {
		t.Errorf("expected data='/home/bab/data', got %q", resolved["data"])
	}
}

func TestResolveVars_CycleDetection(t *testing.T) {
	vars := map[string]string{
		"a": "${{ b }}",
		"b": "${{ a }}",
	}

	_, err := ResolveVars(vars, nil)
	if err == nil {
		t.Error("expected error for circular reference")
	}

	if !errors.Is(err, ErrVarCycle) {
		t.Errorf("expected errors.Is(err, ErrVarCycle), got %T", err)
	}

	var cycleErr *VarCycleError
	if !errors.As(err, &cycleErr) {
		t.Errorf("expected VarCycleError, got %T: %v", err, err)
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "→") {
		t.Errorf("expected cycle chain with arrows in error message, got: %s", errMsg)
	}
}

func TestResolveVars_SelfReference(t *testing.T) {
	vars := map[string]string{
		"a": "${{ a }}",
	}

	_, err := ResolveVars(vars, nil)
	if err == nil {
		t.Error("expected error for self reference")
	}

	if !errors.Is(err, ErrVarCycle) {
		t.Errorf("expected errors.Is(err, ErrVarCycle), got %T", err)
	}
}

func TestContainsVarRef(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"${{ var }}", true},
		{"hello ${{ name }}", true},
		{"no vars here", false},
		{"$var", false},
		{"${var}", false},
		{"$${{ escaped }}", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if ContainsVarRef(tt.input) != tt.expected {
				t.Errorf("ContainsVarRef(%q) = %v, expected %v", tt.input, !tt.expected, tt.expected)
			}
		})
	}
}

func TestExtractVarRefs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"${{ var }}", []string{"var"}},
		{"${{ a }} ${{ b }}", []string{"a", "b"}},
		{"${{ env.HOME }}", []string{"env.HOME"}},
		{"no vars", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			refs := ExtractVarRefs(tt.input)
			if len(refs) != len(tt.expected) {
				t.Errorf("expected %d refs, got %d", len(tt.expected), len(refs))
				return
			}
			for i, ref := range refs {
				if ref != tt.expected[i] {
					t.Errorf("expected ref[%d]=%q, got %q", i, tt.expected[i], ref)
				}
			}
		})
	}
}

func TestVarNotFoundError_Is(t *testing.T) {
	err := &VarNotFoundError{Name: "test"}
	if !errors.Is(err, ErrVarNotFound) {
		t.Error("VarNotFoundError should match ErrVarNotFound")
	}
	if errors.Is(err, ErrVarCycle) {
		t.Error("VarNotFoundError should not match ErrVarCycle")
	}
}

func TestVarCycleError_Is(t *testing.T) {
	err := &VarCycleError{Name: "test", Chain: []string{"a", "b", "a"}}
	if !errors.Is(err, ErrVarCycle) {
		t.Error("VarCycleError should match ErrVarCycle")
	}
	if errors.Is(err, ErrVarNotFound) {
		t.Error("VarCycleError should not match ErrVarNotFound")
	}
}

func TestVarCycleError_ChainFormat(t *testing.T) {
	err := &VarCycleError{Name: "c", Chain: []string{"a", "b", "c"}}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "a → b → c") {
		t.Errorf("expected chain format 'a → b → c', got: %s", errMsg)
	}
}

func TestFindSimilar(t *testing.T) {
	candidates := []string{"app_name", "app_version", "build_dir"}

	tests := []struct {
		target   string
		expected string
	}{
		{"app_nam", "app_name"},
		{"app_versio", "app_version"},
		{"build", "build_dir"},
		{"xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			result := findSimilar(tt.target, candidates)
			if result != tt.expected {
				t.Errorf("findSimilar(%q) = %q, expected %q", tt.target, result, tt.expected)
			}
		})
	}
}
