package errs

import (
	"sort"
	"strings"
)

func FindSimilar(target string, candidates []string) string {
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
