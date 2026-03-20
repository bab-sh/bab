package tui

import (
	"context"
	"errors"
	"image/color"
	"os"
	"strconv"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bab-sh/bab/internal/theme"
	"github.com/charmbracelet/x/ansi"
)

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

func RunParallel(model tea.Model) (*tea.Program, error) {
	program := tea.NewProgram(model, tea.WithOutput(os.Stderr))
	go func() {
		_, _ = program.Run()
	}()
	return program, nil
}

type baseModel struct {
	items     map[string]*itemState
	roots     []string
	width     int
	height    int
	maxLines  int
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

func newBaseModel(items []ParallelItem, cancel context.CancelFunc, maxLines int) baseModel {
	stateMap := make(map[string]*itemState, len(items))
	roots := make([]string, len(items))
	for i, item := range items {
		key := itemKey(i)
		stateMap[key] = &itemState{
			label: item.Label,
			color: item.Color,
		}
		roots[i] = key
	}
	return baseModel{
		items:    stateMap,
		roots:    roots,
		width:    80,
		maxLines: maxLines,
		cancel:   cancel,
	}
}

func itemKey(index int) string {
	return strconv.Itoa(index)
}

func (b *baseModel) handleMsg(msg tea.Msg) (handled bool, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height
		return true, nil

	case ItemRegisterMsg:
		b.items[msg.Key] = &itemState{
			label: msg.Label,
			color: msg.Color,
		}
		if msg.Parent != "" {
			if parent := b.items[msg.Parent]; parent != nil {
				parent.children = append(parent.children, msg.Key)
				parent.lines = nil
			}
		}
		return true, nil

	case ItemStartMsg:
		if item := b.items[msg.Key]; item != nil {
			item.started = true
		}
		return true, nil

	case ItemOutputMsg:
		if item := b.items[msg.Key]; item != nil {
			item.lines = append(item.lines, msg.Line)
			if b.maxLines > 0 && len(item.lines) > b.maxLines {
				item.lines = item.lines[len(item.lines)-b.maxLines:]
			}
		}
		return true, nil

	case ItemDoneMsg:
		if item := b.items[msg.Key]; item != nil {
			item.done = true
			item.err = msg.Err
		}
		return true, nil

	case ItemClearChildrenMsg:
		if item := b.items[msg.Key]; item != nil {
			for _, ck := range item.children {
				b.removeItemTree(ck)
			}
			item.children = nil
		}
		return true, nil

	case AllDoneMsg:
		b.done = true
		return true, tea.Quit

	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			if !b.cancelled {
				b.cancelled = true
				if b.cancel != nil {
					b.cancel()
				}
			}
			return true, nil
		}
	}

	return false, nil
}

func (b *baseModel) removeItemTree(key string) {
	item := b.items[key]
	if item != nil {
		for _, ck := range item.children {
			b.removeItemTree(ck)
		}
	}
	delete(b.items, key)
}

var (
	dimStyle     = lipgloss.NewStyle().Foreground(theme.Dim)
	successStyle = lipgloss.NewStyle().Foreground(theme.Cyan)
	failureStyle = lipgloss.NewStyle().Foreground(theme.Pink)
)

func truncateLine(line string, maxWidth int) string {
	if maxWidth <= 0 || ansi.StringWidth(line) <= maxWidth {
		return line
	}
	return ansi.Truncate(line, maxWidth, "")
}

func statusIcon(item *itemState, cancelled bool) string {
	switch {
	case item.done && item.err != nil && errors.Is(item.err, context.Canceled):
		return dimStyle.Render("⊘")
	case item.done && item.err != nil:
		return failureStyle.Render("✗")
	case item.done:
		return successStyle.Render("✓")
	case cancelled:
		return dimStyle.Render("⊘")
	case !item.started:
		return dimStyle.Render("∙")
	default:
		return dimStyle.Render("◦")
	}
}
