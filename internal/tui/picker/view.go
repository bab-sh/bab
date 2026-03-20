package picker

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(m.input.View() + "\n")
	b.WriteString(m.renderSeparator() + "\n")

	if len(m.matches) == 0 {
		b.WriteString(noResultsStyle.Render("  No matching tasks :(") + "\n")
		v := tea.NewView(b.String())
		v.AltScreen = true
		return v
	}

	end := min(m.offset+m.visibleLines(), len(m.matches))
	for i := m.offset; i < end; i++ {
		b.WriteString(m.renderLine(i) + "\n")
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m Model) renderSeparator() string {
	brand := " bab.sh "
	count := fmt.Sprintf(" %d/%d ", len(m.matches), len(m.tasks))
	lineWidth := max(0, m.width-len(brand)-len(count)-4)
	return separatorStyle.Render(strings.Repeat("─", lineWidth)) +
		countStyle.Render(count) +
		separatorStyle.Render("──") +
		countStyle.Render(brand) +
		separatorStyle.Render("──")
}

func (m Model) renderLine(i int) string {
	match := m.matches[i]
	selected := i == m.cursor

	var b strings.Builder
	if selected {
		b.WriteString(selectedIndicator.Render("│ "))
	} else {
		b.WriteString("  ")
	}

	style := taskNameStyle
	if selected {
		style = taskNameSelectedStyle
	}
	b.WriteString(highlight(match.Task.Name, match.NameIndexes, style, matchStyle))

	aliases := match.Task.GetAllAliases()
	if len(aliases) > 0 {
		aliasText := " (" + strings.Join(aliases, ", ") + ")"
		if match.MatchedAlias != "" {
			b.WriteString(highlightAlias(aliasText, match.MatchedAlias, aliases))
		} else {
			b.WriteString(aliasStyle.Render(aliasText))
		}
	}

	if match.Task.Desc != "" {
		b.WriteString("  " + descStyle.Render(match.Task.Desc))
	}
	return b.String()
}
