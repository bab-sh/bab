package picker

import (
	"strings"

	"github.com/bab-sh/bab/internal/parser"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

type Match struct {
	Task    *parser.Task
	Indexes []int
}

func search(query string, tasks []*parser.Task) []Match {
	if query == "" {
		results := make([]Match, len(tasks))
		for i, t := range tasks {
			results[i] = Match{Task: t}
		}
		return results
	}

	names := make([]string, len(tasks))
	for i, t := range tasks {
		names[i] = t.Name
	}

	matches := fuzzy.Find(query, names)
	results := make([]Match, len(matches))
	for i, m := range matches {
		results[i] = Match{Task: tasks[m.Index], Indexes: m.MatchedIndexes}
	}
	return results
}

func highlight(text string, indexes []int, base, hl lipgloss.Style) string {
	if len(indexes) == 0 {
		return base.Render(text)
	}

	set := make(map[int]struct{}, len(indexes))
	for _, i := range indexes {
		set[i] = struct{}{}
	}

	var b strings.Builder
	for i, r := range text {
		if _, ok := set[i]; ok {
			b.WriteString(hl.Render(string(r)))
		} else {
			b.WriteString(base.Render(string(r)))
		}
	}
	return b.String()
}
