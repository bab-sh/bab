package condition

import (
	"testing"

	"github.com/bab-sh/bab/internal/interpolate"
)

func TestEvaluate_EmptyCondition(t *testing.T) {
	result, err := Evaluate("", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.ShouldRun {
		t.Error("empty condition should run")
	}
	if result.Reason != "no condition" {
		t.Errorf("expected reason 'no condition', got %q", result.Reason)
	}
}

func TestEvaluate_Equality(t *testing.T) {
	ctx := interpolate.NewContext(map[string]string{"env": "prod", "name": "test"})

	tests := []struct {
		name      string
		condition string
		shouldRun bool
	}{
		{"equals match", "${{ env }} == 'prod'", true},
		{"equals no match", "${{ env }} == 'dev'", false},
		{"not equals match", "${{ env }} != 'dev'", true},
		{"not equals no match", "${{ env }} != 'prod'", false},
		{"double quotes equals", "${{ env }} == \"prod\"", true},
		{"double quotes not equals", "${{ env }} != \"prod\"", false},
		{"empty string equals", "${{ missing }} == ''", false},
		{"whitespace in condition", "${{ env }}  ==  'prod'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.condition, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ShouldRun != tt.shouldRun {
				t.Errorf("expected ShouldRun=%v, got %v (reason: %s)", tt.shouldRun, result.ShouldRun, result.Reason)
			}
		})
	}
}

func TestEvaluate_Truthy(t *testing.T) {
	ctx := interpolate.NewContext(map[string]string{
		"hasValue": "yes",
		"isEmpty":  "",
		"isFalse":  "false",
		"isFALSE":  "FALSE",
		"isTrue":   "true",
		"isZero":   "0",
		"isOne":    "1",
	})

	tests := []struct {
		name      string
		condition string
		shouldRun bool
	}{
		{"non-empty value is truthy", "${{ hasValue }}", true},
		{"empty value is falsy", "${{ isEmpty }}", false},
		{"false string is falsy", "${{ isFalse }}", false},
		{"FALSE string is falsy", "${{ isFALSE }}", false},
		{"true string is truthy", "${{ isTrue }}", true},
		{"zero is truthy", "${{ isZero }}", true},
		{"one is truthy", "${{ isOne }}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.condition, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ShouldRun != tt.shouldRun {
				t.Errorf("expected ShouldRun=%v, got %v (reason: %s)", tt.shouldRun, result.ShouldRun, result.Reason)
			}
		})
	}
}

func TestEvaluate_UndefinedVariable(t *testing.T) {
	ctx := interpolate.NewContext(nil)

	result, err := Evaluate("${{ undefined }}", ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShouldRun {
		t.Error("undefined variable should be falsy")
	}
}

func TestEvaluate_ConfirmPromptPattern(t *testing.T) {
	tests := []struct {
		name      string
		confirm   string
		condition string
		shouldRun bool
	}{
		{"confirm true, check true", "true", "${{ confirm }} == 'true'", true},
		{"confirm true, check false", "true", "${{ confirm }} == 'false'", false},
		{"confirm false, check true", "false", "${{ confirm }} == 'true'", false},
		{"confirm false, check false", "false", "${{ confirm }} == 'false'", true},
		{"confirm true, truthy check", "true", "${{ confirm }}", true},
		{"confirm false, truthy check", "false", "${{ confirm }}", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := interpolate.NewContext(map[string]string{"confirm": tt.confirm})
			result, err := Evaluate(tt.condition, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ShouldRun != tt.shouldRun {
				t.Errorf("expected ShouldRun=%v, got %v (reason: %s)", tt.shouldRun, result.ShouldRun, result.Reason)
			}
		})
	}
}

func TestEvaluate_SelectPromptPattern(t *testing.T) {
	ctx := interpolate.NewContext(map[string]string{"environment": "staging"})

	tests := []struct {
		name      string
		condition string
		shouldRun bool
	}{
		{"equals staging", "${{ environment }} == 'staging'", true},
		{"equals prod", "${{ environment }} == 'prod'", false},
		{"not equals dev", "${{ environment }} != 'dev'", true},
		{"not equals staging", "${{ environment }} != 'staging'", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Evaluate(tt.condition, ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ShouldRun != tt.shouldRun {
				t.Errorf("expected ShouldRun=%v, got %v (reason: %s)", tt.shouldRun, result.ShouldRun, result.Reason)
			}
		})
	}
}

func TestEvaluateTruthy(t *testing.T) {
	tests := []struct {
		value     string
		shouldRun bool
	}{
		{"", false},
		{"false", false},
		{"FALSE", false},
		{"False", false},
		{"true", true},
		{"TRUE", true},
		{"yes", true},
		{"no", true},
		{"0", true},
		{"1", true},
		{"hello", true},
		{"  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := evaluateTruthy(tt.value)
			if result.ShouldRun != tt.shouldRun {
				t.Errorf("evaluateTruthy(%q): expected %v, got %v", tt.value, tt.shouldRun, result.ShouldRun)
			}
		})
	}
}

func TestEvaluateComparison(t *testing.T) {
	tests := []struct {
		input     string
		shouldRun bool
		found     bool
	}{
		{"prod == 'prod'", true, true},
		{"prod == 'dev'", false, true},
		{"prod != 'dev'", true, true},
		{"prod != 'prod'", false, true},
		{"'prod' == 'prod'", true, true},
		{"\"prod\" == \"prod\"", true, true},
		{"just a value", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, found := evaluateComparison(tt.input)
			if found != tt.found {
				t.Errorf("evaluateComparison(%q): expected found=%v, got %v", tt.input, tt.found, found)
			}
			if found && result.ShouldRun != tt.shouldRun {
				t.Errorf("evaluateComparison(%q): expected shouldRun=%v, got %v", tt.input, tt.shouldRun, result.ShouldRun)
			}
		})
	}
}
