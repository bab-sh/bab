package tui

import (
	"context"
	"errors"
	"image/color"
	"os"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bab-sh/bab/internal/theme"
	"github.com/charmbracelet/x/ansi"
)

const frameHeight = 5

type ParallelItem struct {
	Label string
	Color color.Color
}

type ItemRegisterMsg struct {
	Key    string
	Parent string
	Label  string
	Color  color.Color
}

type ItemStartMsg struct {
	Key string
}

type ItemOutputMsg struct {
	Key  string
	Line string
}

type ItemDoneMsg struct {
	Key string
	Err error
}

type ItemClearChildrenMsg struct {
	Key string
}

type AllDoneMsg struct{}

func RunParallel(items []ParallelItem, cancel context.CancelFunc) (*tea.Program, error) {
	stateMap := make(map[string]*itemState, len(items))
	roots := make([]string, len(items))
	for i, item := range items {
		key := strconv.Itoa(i)
		stateMap[key] = &itemState{
			label: item.Label,
			color: item.Color,
		}
		roots[i] = key
	}

	model := parallelModel{
		items:  stateMap,
		roots:  roots,
		width:  80,
		cancel: cancel,
	}

	program := tea.NewProgram(model, tea.WithOutput(os.Stderr))

	go func() {
		_, _ = program.Run()
	}()

	return program, nil
}

type parallelModel struct {
	items     map[string]*itemState
	roots     []string
	width     int
	done      bool
	cancelled bool
	cancel    context.CancelFunc
}

type itemState struct {
	label    string
	color    color.Color
	lines    []string
	started  bool
	done     bool
	err      error
	children []string
}

var (
	dimStyle     = lipgloss.NewStyle().Foreground(theme.Dim)
	successStyle = lipgloss.NewStyle().Foreground(theme.Cyan)
	failureStyle = lipgloss.NewStyle().Foreground(theme.Pink)
)

func (m parallelModel) Init() tea.Cmd {
	return tea.RequestWindowSize
}

func (m parallelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case ItemRegisterMsg:
		m.items[msg.Key] = &itemState{
			label: msg.Label,
			color: msg.Color,
		}
		if msg.Parent != "" {
			if parent := m.items[msg.Parent]; parent != nil {
				parent.children = append(parent.children, msg.Key)
				parent.lines = nil
			}
		}
		return m, nil

	case ItemStartMsg:
		if item := m.items[msg.Key]; item != nil {
			item.started = true
		}
		return m, nil

	case ItemOutputMsg:
		if item := m.items[msg.Key]; item != nil {
			item.lines = append(item.lines, msg.Line)
			if len(item.lines) > frameHeight {
				item.lines = item.lines[len(item.lines)-frameHeight:]
			}
		}
		return m, nil

	case ItemDoneMsg:
		if item := m.items[msg.Key]; item != nil {
			item.done = true
			item.err = msg.Err
		}
		return m, nil

	case ItemClearChildrenMsg:
		if item := m.items[msg.Key]; item != nil {
			for _, ck := range item.children {
				m.removeItemTree(ck)
			}
			item.children = nil
		}
		return m, nil

	case AllDoneMsg:
		m.done = true
		return m, tea.Quit

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			if !m.cancelled {
				m.cancelled = true
				if m.cancel != nil {
					m.cancel()
				}
			}
			return m, nil
		}
	}

	return m, nil
}

func (m parallelModel) removeItemTree(key string) {
	item := m.items[key]
	if item != nil {
		for _, ck := range item.children {
			m.removeItemTree(ck)
		}
	}
	delete(m.items, key)
}

func (m parallelModel) View() tea.View {
	if m.done {
		return tea.NewView(" ")
	}

	var out []string
	for _, key := range m.roots {
		out = append(out, m.renderItem(key, "", m.width, 0)...)
	}
	return tea.NewView(strings.Join(out, "\n"))
}

func (m parallelModel) renderItem(key string, indent string, width int, depth int) []string {
	item := m.items[key]
	if item == nil {
		return nil
	}

	if depth > 0 {
		var lines []string
		lines = append(lines, m.renderCompactLine(item, indent, width))
		if len(item.children) > 0 {
			childIndent := indent + "  "
			childWidth := width - 2
			if childWidth < 20 {
				childWidth = 20
			}
			for _, ck := range item.children {
				lines = append(lines, m.renderItem(ck, childIndent, childWidth, depth+1)...)
			}
		}
		return lines
	}

	titleStyle := lipgloss.NewStyle().Foreground(item.color).Bold(true)
	status := m.statusIcon(item)

	lines := make([]string, 0, frameHeight+2)
	lines = append(lines, indent+dimStyle.Render("┌─")+" "+titleStyle.Render(item.label)+" "+status)

	if len(item.children) > 0 {
		childIndent := indent + dimStyle.Render("│") + "  "
		childWidth := width - 3
		if childWidth < 20 {
			childWidth = 20
		}
		for _, ck := range item.children {
			lines = append(lines, m.renderItem(ck, childIndent, childWidth, 1)...)
		}
	} else {
		maxWidth := width - 4
		for _, line := range item.lines {
			if maxWidth > 0 && ansi.StringWidth(line) > maxWidth {
				line = ansi.Truncate(line, maxWidth, "")
			}
			lines = append(lines, indent+dimStyle.Render("│")+"  "+line)
		}
		if !item.done && len(item.lines) > 0 {
			for range frameHeight - len(item.lines) {
				lines = append(lines, indent+dimStyle.Render("│"))
			}
		}
	}

	lines = append(lines, indent+dimStyle.Render("└"))
	return lines
}

func (m parallelModel) renderCompactLine(item *itemState, indent string, width int) string {
	titleStyle := lipgloss.NewStyle().Foreground(item.color).Bold(true)
	status := m.statusIcon(item)
	label := titleStyle.Render(item.label)

	prefix := status + " " + label
	prefixWidth := ansi.StringWidth(status) + 1 + ansi.StringWidth(item.label)

	if item.done && item.err == nil {
		return indent + prefix
	}

	snippet := ""
	for i := len(item.lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(item.lines[i]) != "" {
			snippet = item.lines[i]
			break
		}
	}

	if snippet != "" {
		available := width - ansi.StringWidth(indent) - prefixWidth - 2
		if available > 0 {
			if ansi.StringWidth(snippet) > available {
				snippet = ansi.Truncate(snippet, available, "…")
			}
			return indent + prefix + "  " + dimStyle.Render(snippet)
		}
	}

	return indent + prefix
}

func (m parallelModel) statusIcon(item *itemState) string {
	switch {
	case item.done && item.err != nil && errors.Is(item.err, context.Canceled):
		return dimStyle.Render("⊘")
	case item.done && item.err != nil:
		return failureStyle.Render("✗")
	case item.done:
		return successStyle.Render("✓")
	case m.cancelled:
		return dimStyle.Render("⊘")
	case !item.started:
		return dimStyle.Render("∙")
	default:
		return dimStyle.Render("◦")
	}
}
