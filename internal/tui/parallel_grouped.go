package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const groupedMaxLines = 5

type groupedModel struct {
	baseModel
}

func NewGroupedModel(items []ParallelItem, cancel context.CancelFunc) tea.Model {
	return groupedModel{baseModel: newBaseModel(items, cancel, groupedMaxLines)}
}

func (m groupedModel) Init() tea.Cmd {
	return tea.RequestWindowSize
}

func (m groupedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if handled, cmd := m.handleMsg(msg); handled {
		return m, cmd
	}
	return m, nil
}

func (m groupedModel) View() tea.View {
	if m.done {
		return tea.NewView(" ")
	}
	var out []string
	for _, key := range m.roots {
		out = append(out, m.renderItem(key, "", 0)...)
	}
	return tea.NewView(strings.Join(out, "\n"))
}

func (m groupedModel) renderItem(key string, indent string, depth int) []string {
	item := m.items[key]
	if item == nil {
		return nil
	}

	if depth > 0 {
		var lines []string
		lines = append(lines, m.renderCompactLine(item, indent))
		if len(item.children) > 0 {
			childIndent := indent + "  "
			for _, ck := range item.children {
				lines = append(lines, m.renderItem(ck, childIndent, depth+1)...)
			}
		}
		return lines
	}

	titleStyle := lipgloss.NewStyle().Foreground(item.color).Bold(true)
	status := statusIcon(item, m.cancelled)

	lines := make([]string, 0, groupedMaxLines+2)
	lines = append(lines, truncateLine(
		indent+dimStyle.Render("┌─")+" "+titleStyle.Render(item.label)+" "+status,
		m.width,
	))

	if len(item.children) > 0 {
		childIndent := indent + dimStyle.Render("│") + "  "
		for _, ck := range item.children {
			lines = append(lines, m.renderItem(ck, childIndent, 1)...)
		}
	} else {
		linePrefix := indent + dimStyle.Render("│") + "  "
		linePrefixWidth := ansi.StringWidth(linePrefix)
		contentMax := m.width - linePrefixWidth
		for _, line := range item.lines {
			if contentMax > 0 && ansi.StringWidth(line) > contentMax {
				line = ansi.Truncate(line, contentMax, "")
			}
			lines = append(lines, truncateLine(linePrefix+line, m.width))
		}
		if !item.done && len(item.lines) > 0 && len(item.lines) < groupedMaxLines {
			for range groupedMaxLines - len(item.lines) {
				lines = append(lines, truncateLine(indent+dimStyle.Render("│"), m.width))
			}
		}
	}

	lines = append(lines, truncateLine(indent+dimStyle.Render("└"), m.width))
	return lines
}

func (m groupedModel) renderCompactLine(item *itemState, indent string) string {
	titleStyle := lipgloss.NewStyle().Foreground(item.color).Bold(true)
	status := statusIcon(item, m.cancelled)
	label := titleStyle.Render(item.label)

	prefix := status + " " + label
	prefixWidth := ansi.StringWidth(prefix)

	if item.done && item.err == nil {
		return truncateLine(indent+prefix, m.width)
	}

	snippet := ""
	for i := len(item.lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(item.lines[i]) != "" {
			snippet = item.lines[i]
			break
		}
	}

	if snippet != "" {
		available := m.width - ansi.StringWidth(indent) - prefixWidth - 2
		if available > 0 {
			if ansi.StringWidth(snippet) > available {
				snippet = ansi.Truncate(snippet, available, "…")
			}
			return truncateLine(indent+prefix+"  "+dimStyle.Render(snippet), m.width)
		}
	}

	return truncateLine(indent+prefix, m.width)
}
