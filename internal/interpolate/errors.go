package interpolate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	ErrVarNotFound = errors.New("variable not found")
	ErrVarCycle    = errors.New("circular variable reference")
)

type VarNotFoundError struct {
	Path      string
	Line      int
	Name      string
	Available []string
}

func (e *VarNotFoundError) Error() string {
	var suffix string
	if suggestion := findSimilar(e.Name, e.Available); suggestion != "" {
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
	path := relativePath(e.Path)
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
	path := relativePath(e.Path)
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d", path, e.Line)
	}
	return path
}

func (e *VarCycleError) Is(target error) bool {
	return target == ErrVarCycle
}

func relativePath(path string) string {
	if cwd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(cwd, path); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return path
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
