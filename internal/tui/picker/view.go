package picker

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(m.input.View() + "\n")
	b.WriteString(m.renderSeparator() + "\n")

	if len(m.matches) == 0 {
		b.WriteString(noResultsStyle.Render("  No matching tasks :(") + "\n")
		return b.String()
	}

	end := min(m.offset+m.visibleLines(), len(m.matches))
	for i := m.offset; i < end; i++ {
		b.WriteString(m.renderLine(i) + "\n")
	}
	return b.String()
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
	b.WriteString(highlight(match.Task.Name, match.Indexes, style, matchStyle))

	if match.Task.Desc != "" {
		b.WriteString("  " + descStyle.Render(match.Task.Desc))
	}
	return b.String()
}
