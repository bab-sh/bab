package picker

import (
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/bab-sh/bab/internal/babfile"
	"github.com/sahilm/fuzzy"
)

const (
	scoreExactAlias = 1000
	scoreFuzzyAlias = 500
)

type Match struct {
	Task         *babfile.Task
	NameIndexes  []int
	AliasIndexes []int
	MatchedAlias string
	Score        int
}

func search(query string, tasks []*babfile.Task) []Match {
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
	nameMatchMap := make(map[int][]int)
	nameScoreMap := make(map[int]int)
	for _, m := range fuzzy.Find(query, names) {
		nameMatchMap[m.Index] = m.MatchedIndexes
		nameScoreMap[m.Index] = m.Score
	}

	results := make([]Match, 0, len(tasks))
	matched := make(map[int]bool)

	for i, task := range tasks {
		alias, aliasIndexes, exact := searchAliases(query, task)
		if alias != "" {
			score := scoreFuzzyAlias
			if exact {
				score = scoreExactAlias
			} else if len(aliasIndexes) > 0 {
				score = scoreFuzzyAlias + len(aliasIndexes)
			}
			results = append(results, Match{
				Task:         task,
				MatchedAlias: alias,
				AliasIndexes: aliasIndexes,
				NameIndexes:  nameMatchMap[i],
				Score:        score,
			})
			matched[i] = true
		}
	}

	for i, task := range tasks {
		if matched[i] {
			continue
		}
		if nameIndexes, ok := nameMatchMap[i]; ok {
			results = append(results, Match{
				Task:        task,
				NameIndexes: nameIndexes,
				Score:       nameScoreMap[i],
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Task.Name < results[j].Task.Name
	})

	return results
}

func searchAliases(query string, task *babfile.Task) (alias string, indexes []int, exact bool) {
	aliases := task.GetAllAliases()
	if len(aliases) == 0 {
		return "", nil, false
	}

	for _, a := range aliases {
		if a == query {
			return a, nil, true
		}
	}

	matches := fuzzy.Find(query, aliases)
	if len(matches) > 0 {
		best := matches[0]
		return aliases[best.Index], best.MatchedIndexes, false
	}

	return "", nil, false
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
	var run strings.Builder
	prevHL := false
	first := true
	for i, r := range text {
		_, isHL := set[i]
		if !first && isHL != prevHL {
			if prevHL {
				b.WriteString(hl.Render(run.String()))
			} else {
				b.WriteString(base.Render(run.String()))
			}
			run.Reset()
		}
		first = false
		prevHL = isHL
		run.WriteRune(r)
	}
	if run.Len() > 0 {
		if prevHL {
			b.WriteString(hl.Render(run.String()))
		} else {
			b.WriteString(base.Render(run.String()))
		}
	}
	return b.String()
}

func highlightAlias(text, matchedAlias string, allAliases []string) string {
	aliasStart := -1
	for _, a := range allAliases {
		if a == matchedAlias {
			idx := strings.Index(text, a)
			if idx != -1 {
				aliasStart = idx
				break
			}
		}
	}

	if aliasStart == -1 {
		return aliasStyle.Render(text)
	}

	aliasEnd := aliasStart + len(matchedAlias)

	var b strings.Builder
	if aliasStart > 0 {
		b.WriteString(aliasStyle.Render(text[:aliasStart]))
	}
	b.WriteString(aliasMatchStyle.Render(text[aliasStart:aliasEnd]))
	if aliasEnd < len(text) {
		b.WriteString(aliasStyle.Render(text[aliasEnd:]))
	}
	return b.String()
}
