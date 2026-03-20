package tui

import (
	"context"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bab-sh/bab/internal/theme"
	"github.com/charmbracelet/x/ansi"
)

type tabsModel struct {
	baseModel
	activeTab int
	viewport  viewport.Model
}

func NewTabsModel(items []ParallelItem, cancel context.CancelFunc) tea.Model {
	return &tabsModel{
		baseModel: newBaseModel(items, cancel, 0),
		viewport:  viewport.New(),
	}
}

func (m *tabsModel) Init() tea.Cmd {
	return tea.RequestWindowSize
}

func (m *tabsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if handled, cmd := m.handleMsg(msg); handled {
		switch msg.(type) {
		case ItemOutputMsg, ItemStartMsg, ItemDoneMsg, ItemRegisterMsg,
			ItemClearChildrenMsg, tea.WindowSizeMsg:
			m.updateViewport()
		}
		return m, cmd
	}

	if km, ok := msg.(tea.KeyPressMsg); ok && len(m.roots) > 0 {
		n := len(m.roots)
		switch km.String() {
		case "right", "l", "n", "tab":
			m.activeTab = (m.activeTab + 1) % n
			m.updateViewport()
			m.viewport.GotoBottom()
			return m, nil
		case "left", "h", "p", "shift+tab":
			m.activeTab = (m.activeTab - 1 + n) % n
			m.updateViewport()
			m.viewport.GotoBottom()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *tabsModel) View() tea.View {
	if m.done {
		return tea.NewView(" ")
	}
	v := tea.NewView(m.renderView())
	v.AltScreen = true
	return v
}

func (m *tabsModel) updateViewport() {
	if len(m.roots) == 0 {
		return
	}
	if m.activeTab < 0 {
		m.activeTab = 0
	}
	if m.activeTab >= len(m.roots) {
		m.activeTab = len(m.roots) - 1
	}

	activeKey := m.roots[m.activeTab]
	activeItem := m.items[activeKey]
	if activeItem == nil {
		m.viewport.SetContent("")
		return
	}

	contentWidth := m.width
	if contentWidth < 1 {
		contentWidth = 1
	}
	contentHeight := m.height - 4
	if contentHeight < 1 {
		contentHeight = 1
	}
	m.viewport.SetWidth(contentWidth)
	m.viewport.SetHeight(contentHeight)

	content := m.buildContent(activeItem, contentWidth)

	atBottom := m.viewport.AtBottom()
	m.viewport.SetContent(content)
	if atBottom {
		m.viewport.GotoBottom()
	}
}

func (m *tabsModel) renderView() string {
	if len(m.roots) == 0 {
		return ""
	}

	borderColor := theme.Dim
	activeKey := m.roots[m.activeTab]
	if activeItem := m.items[activeKey]; activeItem != nil {
		borderColor = activeItem.color
	}

	inactiveTabBorder := tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder := tabBorderWithBottom("┘", " ", "└")

	inactiveTab := lipgloss.NewStyle().
		Border(inactiveTabBorder, true).
		BorderForeground(borderColor).
		Padding(0, 1)
	activeTab := inactiveTab.
		Border(activeTabBorder, true)

	renderedTabs := make([]string, 0, len(m.roots))
	for i, key := range m.roots {
		item := m.items[key]
		if item == nil {
			continue
		}

		isFirst, isActive := i == 0, i == m.activeTab

		var style lipgloss.Style
		if isActive {
			style = activeTab.Foreground(item.color).Bold(true)
		} else {
			style = inactiveTab.Foreground(dimStyle.GetForeground())
		}

		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		}
		style = style.Border(border)

		label := statusIcon(item, m.cancelled) + " " + item.label
		renderedTabs = append(renderedTabs, style.Render(label))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	windowStyle := lipgloss.NewStyle().
		BorderForeground(borderColor).
		Border(lipgloss.NormalBorder()).
		UnsetBorderTop().
		Width(m.width)
	window := windowStyle.Render(m.viewport.View())

	targetWidth := lipgloss.Width(window)
	lines := strings.Split(row, "\n")
	lastLine := lines[len(lines)-1]
	lastLineWidth := ansi.StringWidth(lastLine)
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	if lastLineWidth < targetWidth {
		gap := targetWidth - lastLineWidth - 1
		if gap < 0 {
			gap = 0
		}
		lines[len(lines)-1] = lastLine + borderStyle.Render(strings.Repeat("─", gap)+"┐")
	} else if lastLineWidth >= targetWidth {
		lines[len(lines)-1] = truncateLine(lastLine, targetWidth-1) + borderStyle.Render("┐")
	}
	row = strings.Join(lines, "\n")

	var b strings.Builder
	b.WriteString(row)
	b.WriteString("\n")
	b.WriteString(window)
	return b.String()
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func (m *tabsModel) buildContent(item *itemState, width int) string {
	if len(item.children) > 0 {
		var lines []string
		for _, ck := range item.children {
			child := m.items[ck]
			if child == nil {
				continue
			}
			status := statusIcon(child, m.cancelled)
			titleStyle := lipgloss.NewStyle().Foreground(child.color).Bold(true)
			line := status + " " + titleStyle.Render(child.label)
			if child.done && child.err != nil {
				line += "  " + failureStyle.Render(child.err.Error())
			} else if len(child.lines) > 0 {
				last := child.lines[len(child.lines)-1]
				if strings.TrimSpace(last) != "" {
					line += "  " + dimStyle.Render(last)
				}
			}
			lines = append(lines, truncateLine(line, width))
		}
		return strings.Join(lines, "\n")
	}

	if len(item.lines) == 0 {
		if !item.started {
			return dimStyle.Render("Waiting…")
		}
		return dimStyle.Render("Running…")
	}

	lines := make([]string, 0, len(item.lines))
	for _, line := range item.lines {
		if width > 0 && ansi.StringWidth(line) > width {
			line = ansi.Truncate(line, width, "")
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
