package tui

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/bab-sh/bab/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const frameHeight = 5

type ParallelItem struct {
	Label string
	Color lipgloss.Color
}

type ItemOutputMsg struct {
	Index int
	Line  string
}

type ItemDoneMsg struct {
	Index int
	Err   error
}

type AllDoneMsg struct{}

func RunParallel(items []ParallelItem, cancel context.CancelFunc) (*tea.Program, error) {
	states := make([]itemState, len(items))
	for i, item := range items {
		states[i] = itemState{
			label: item.Label,
			color: item.Color,
		}
	}

	model := parallelModel{
		items:  states,
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
	items     []itemState
	width     int
	done      bool
	cancelled bool
	cancel    context.CancelFunc
}

type itemState struct {
	label string
	color lipgloss.Color
	lines []string
	done  bool
	err   error
}

var (
	dimStyle     = lipgloss.NewStyle().Foreground(theme.Dim)
	successStyle = lipgloss.NewStyle().Foreground(theme.Cyan)
	failureStyle = lipgloss.NewStyle().Foreground(theme.Pink)
)

func (m parallelModel) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m parallelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case ItemOutputMsg:
		if msg.Index >= 0 && msg.Index < len(m.items) {
			item := &m.items[msg.Index]
			item.lines = append(item.lines, msg.Line)
			if len(item.lines) > frameHeight {
				item.lines = item.lines[len(item.lines)-frameHeight:]
			}
		}
		return m, nil

	case ItemDoneMsg:
		if msg.Index >= 0 && msg.Index < len(m.items) {
			m.items[msg.Index].done = true
			m.items[msg.Index].err = msg.Err
		}
		return m, nil

	case AllDoneMsg:
		m.done = true
		return m, tea.Quit

	case tea.KeyMsg:
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

func (m parallelModel) View() string {
	if m.done {
		return " "
	}

	var out []string
	for _, item := range m.items {
		out = append(out, m.renderFrameLines(item)...)
	}
	return strings.Join(out, "\n")
}

func (m parallelModel) renderFrameLines(item itemState) []string {
	titleStyle := lipgloss.NewStyle().Foreground(item.color).Bold(true)

	var status string
	switch {
	case item.done && item.err != nil && errors.Is(item.err, context.Canceled):
		status = dimStyle.Render("⊘")
	case item.done && item.err != nil:
		status = failureStyle.Render("✗")
	case item.done:
		status = successStyle.Render("✓")
	case m.cancelled:
		status = dimStyle.Render("⊘")
	default:
		status = dimStyle.Render("◦")
	}

	lines := make([]string, 0, frameHeight+2)
	lines = append(lines, dimStyle.Render("┌─")+" "+titleStyle.Render(item.label)+" "+status)

	maxWidth := m.width - 4
	for _, line := range item.lines {
		if maxWidth > 0 && ansi.StringWidth(line) > maxWidth {
			line = ansi.Truncate(line, maxWidth, "")
		}
		lines = append(lines, dimStyle.Render("│")+"  "+line)
	}

	for range frameHeight - len(item.lines) {
		lines = append(lines, dimStyle.Render("│"))
	}

	lines = append(lines, dimStyle.Render("└"))
	return lines
}
