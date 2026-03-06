package tui

import (
	"context"
	"os"
	"strings"

	"github.com/bab-sh/bab/internal/theme"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func RunParallel(ctx context.Context, items []ParallelItem) (*tea.Program, error) {
	states := make([]itemState, len(items))
	for i, item := range items {
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(item.Color)
		states[i] = itemState{
			label:   item.Label,
			color:   item.Color,
			spinner: s,
		}
	}

	model := parallelModel{
		items: states,
		width: 80,
	}

	program := tea.NewProgram(model, tea.WithOutput(os.Stderr), tea.WithContext(ctx))

	go func() {
		_, _ = program.Run()
	}()

	return program, nil
}

type parallelModel struct {
	items []itemState
	width int
	done  bool
}

type itemState struct {
	label   string
	color   lipgloss.Color
	lines   []string
	spinner spinner.Model
	done    bool
	err     error
}

var (
	dimBorder    = lipgloss.NewStyle().Foreground(theme.Dim)
	successStyle = lipgloss.NewStyle().Foreground(theme.Cyan)
	failureStyle = lipgloss.NewStyle().Foreground(theme.Pink)
)

func (m parallelModel) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.WindowSize()}
	for i := range m.items {
		cmds = append(cmds, m.items[i].spinner.Tick)
	}
	return tea.Batch(cmds...)
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
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmds []tea.Cmd
		for i := range m.items {
			if !m.items[i].done {
				var cmd tea.Cmd
				m.items[i].spinner, cmd = m.items[i].spinner.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m parallelModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder
	for _, item := range m.items {
		b.WriteString(m.renderFrame(item))
	}
	return b.String()
}

func (m parallelModel) renderFrame(item itemState) string {
	titleStyle := lipgloss.NewStyle().Foreground(item.color).Bold(true)

	var b strings.Builder

	status := item.spinner.View()
	if item.done {
		if item.err != nil {
			status = failureStyle.Render("✗")
		} else {
			status = successStyle.Render("✓")
		}
	}
	b.WriteString(dimBorder.Render("┌─") + " " + titleStyle.Render(item.label) + " " + status + "\n")

	lines := item.lines
	maxWidth := m.width - 4

	for _, line := range lines {
		if maxWidth > 0 && len(line) > maxWidth {
			line = line[:maxWidth]
		}
		b.WriteString(dimBorder.Render("│") + "  " + line + "\n")
	}

	remaining := frameHeight - len(lines)
	for j := 0; j < remaining; j++ {
		b.WriteString(dimBorder.Render("│") + "\n")
	}

	b.WriteString(dimBorder.Render("└") + "\n")

	return b.String()
}
