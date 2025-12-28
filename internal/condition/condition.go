package condition

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bab-sh/bab/internal/interpolate"
)

var (
	comparisonPatternSQ = regexp.MustCompile(`^(.+?)\s*(==|!=)\s*'([^']*)'$`)
	comparisonPatternDQ = regexp.MustCompile(`^(.+?)\s*(==|!=)\s*"([^"]*)"$`)
)

type Result struct {
	ShouldRun bool
	Reason    string
}

func Evaluate(condition string, ctx *interpolate.Context) (Result, error) {
	if condition == "" {
		return Result{ShouldRun: true, Reason: "no condition"}, nil
	}

	interpolated, err := interpolate.Interpolate(condition, ctx)
	if err != nil {
		return Result{ShouldRun: false, Reason: fmt.Sprintf("variable error: %v", err)}, nil
	}

	interpolated = strings.TrimSpace(interpolated)

	if result, ok := evaluateComparison(interpolated); ok {
		return result, nil
	}

	return evaluateTruthy(interpolated), nil
}

func evaluateComparison(s string) (Result, bool) {
	if matches := comparisonPatternSQ.FindStringSubmatch(s); len(matches) == 4 {
		left := strings.TrimSpace(matches[1])
		op := matches[2]
		right := matches[3]
		left = strings.Trim(left, "'\"")
		return evaluateOp(left, op, right), true
	}

	if matches := comparisonPatternDQ.FindStringSubmatch(s); len(matches) == 4 {
		left := strings.TrimSpace(matches[1])
		op := matches[2]
		right := matches[3]
		left = strings.Trim(left, "'\"")
		return evaluateOp(left, op, right), true
	}

	return Result{}, false
}

func evaluateOp(left, op, right string) Result {
	switch op {
	case "==":
		return Result{
			ShouldRun: left == right,
			Reason:    fmt.Sprintf("%q == %q is %v", left, right, left == right),
		}
	case "!=":
		return Result{
			ShouldRun: left != right,
			Reason:    fmt.Sprintf("%q != %q is %v", left, right, left != right),
		}
	default:
		return Result{ShouldRun: false, Reason: fmt.Sprintf("unknown operator: %s", op)}
	}
}

func evaluateTruthy(s string) Result {
	s = strings.TrimSpace(s)

	if s == "" {
		return Result{ShouldRun: false, Reason: "empty value is falsy"}
	}

	if strings.EqualFold(s, "false") {
		return Result{ShouldRun: false, Reason: "value 'false' is falsy"}
	}

	return Result{ShouldRun: true, Reason: fmt.Sprintf("truthy value: %q", s)}
}
