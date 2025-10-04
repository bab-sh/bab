// Package tui provides an interactive fuzzy search interface for selecting tasks.
package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bab-sh/bab/internal/registry"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// Model represents the Bubble Tea model for the interactive task selector.
type Model struct {
	textInput     textinput.Model
	allTasks      []TaskItem
	filteredTasks []TaskItem
	cursor        int
	selectedTask  *registry.Task
	quitting      bool
	width         int
	height        int
}

// NewModel creates a new Model with all tasks from the registry.
func NewModel(reg registry.Registry) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to search..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	tasks := reg.List()
	items := make([]TaskItem, len(tasks))
	for i, task := range tasks {
		items[i] = NewTaskItem(task)
	}

	sort.Slice(items, func(i, j int) bool {
		return len(items[i].Title()) < len(items[j].Title())
	})

	return Model{
		textInput:     ti,
		allTasks:      items,
		filteredTasks: items,
		cursor:        0,
		width:         80,
		height:        24,
	}
}

// Init initializes the model and returns the initial command.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = msg.Width - 3
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			if len(m.filteredTasks) > 0 && m.cursor < len(m.filteredTasks) {
				m.selectedTask = m.filteredTasks[m.cursor].Task()
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil

		case tea.KeyUp, tea.KeyCtrlK:
			if len(m.filteredTasks) > 0 && m.cursor < len(m.filteredTasks)-1 {
				m.cursor++
			}
			return m, nil

		case tea.KeyDown, tea.KeyCtrlJ:
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case tea.KeyTab:
			completion := m.getCompletion()
			if completion != "" {
				m.textInput.SetValue(completion)
				m.textInput.SetCursor(len(completion))
				m = m.filterTasks(completion)
				m.cursor = 0
			}
			return m, nil
		}
	}

	oldValue := m.textInput.Value()
	m.textInput, cmd = m.textInput.Update(msg)
	newValue := m.textInput.Value()

	if oldValue != newValue {
		m = m.filterTasks(newValue)
		if len(m.filteredTasks) > 0 {
			m.cursor = 0
		}
	}

	return m, cmd
}

func (m Model) filterTasks(query string) Model {
	if query == "" {
		m.filteredTasks = m.allTasks
		return m
	}

	searchStrings := make([]string, len(m.allTasks))
	for i, task := range m.allTasks {
		searchStrings[i] = task.FilterValue()
	}

	matches := fuzzy.Find(query, searchStrings)

	type match struct {
		task  TaskItem
		score int
	}

	matched := make([]match, len(matches))
	for i, fm := range matches {
		matched[i] = match{
			task:  m.allTasks[fm.Index],
			score: fm.Score,
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		if matched[i].score != matched[j].score {
			return matched[i].score > matched[j].score
		}
		return len(matched[i].task.Title()) < len(matched[j].task.Title())
	})

	m.filteredTasks = make([]TaskItem, len(matched))
	for i, mt := range matched {
		m.filteredTasks[i] = mt.task
	}

	return m
}

func (m Model) getCompletion() string {
	input := m.textInput.Value()
	if input == "" || len(m.filteredTasks) == 0 {
		return ""
	}

	var matches []string
	inputLower := strings.ToLower(input)

	for _, task := range m.filteredTasks {
		name := task.Title()
		if strings.HasPrefix(strings.ToLower(name), inputLower) {
			matches = append(matches, name)
		}
	}

	if len(matches) == 0 {
		return ""
	}

	if len(matches) == 1 {
		return matches[0]
	}

	prefix := matches[0]
	for _, match := range matches[1:] {
		prefix = commonPrefix(prefix, match)
		if prefix == "" {
			return ""
		}
	}

	if len(prefix) > len(input) {
		return prefix
	}

	return ""
}

func commonPrefix(a, b string) string {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		ac, bc := a[i], b[i]
		if ac >= 'A' && ac <= 'Z' {
			ac += 32
		}
		if bc >= 'A' && bc <= 'Z' {
			bc += 32
		}
		if ac != bc {
			return a[:i]
		}
	}

	return a[:minLen]
}

// View renders the model to a string for display.
func (m Model) View() string {
	if m.quitting && m.selectedTask == nil {
		return ""
	}

	var b strings.Builder

	availableHeight := m.height - 2
	if availableHeight < 1 {
		availableHeight = 1
	}

	b.WriteString(m.renderStatusBar())
	b.WriteString("\n")

	visibleTasks, startIdx := m.getVisibleWindow()

	paddingLines := availableHeight - len(visibleTasks)
	for i := 0; i < paddingLines; i++ {
		b.WriteString("\n")
	}

	for i := len(visibleTasks) - 1; i >= 0; i-- {
		taskIdx := startIdx + i
		b.WriteString(m.renderTask(visibleTasks[i], taskIdx == m.cursor))
		if i > 0 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.renderPrompt())

	return b.String()
}

func (m Model) renderStatusBar() string {
	status := fmt.Sprintf("  %d/%d", len(m.filteredTasks), len(m.allTasks))

	separatorLen := m.width - len(status) - 2
	if separatorLen < 0 {
		separatorLen = 0
	}
	separator := strings.Repeat("─", separatorLen)

	return StatusStyle.Render(status + " " + separator)
}

func (m Model) renderTask(task TaskItem, isSelected bool) string {
	var b strings.Builder

	if isSelected {
		b.WriteString(CursorStyle.Render("❯ "))
	} else {
		b.WriteString("  ")
	}

	taskName := task.Title()
	input := m.textInput.Value()
	currentSegment := m.getCurrentSegmentLevel()

	matchStart := -1
	matchEnd := -1
	if input != "" {
		inputLower := strings.ToLower(input)
		taskLower := strings.ToLower(taskName)
		matchStart = strings.Index(taskLower, inputLower)
		if matchStart >= 0 {
			matchEnd = matchStart + len(input)
		}
	}

	// Calculate segment boundaries
	segmentBounds := m.getSegmentBoundaries(taskName)

	// Render character by character with appropriate styling
	for i, char := range taskName {
		var style lipgloss.Style

		// check if this character is part of exact match
		if matchStart >= 0 && i >= matchStart && i < matchEnd {
			style = ExactMatchStyle
		} else {
			// determine which segment this character belongs to
			segmentIdx := m.getSegmentIndex(i, segmentBounds)
			if segmentIdx == currentSegment {
				style = ActiveSegmentStyle
			} else {
				style = InactiveSegmentStyle
			}
		}

		b.WriteString(style.Render(string(char)))
	}

	if task.Description() != "" {
		b.WriteString(" ")
		b.WriteString(DescriptionStyle.Render("- " + task.Description()))
	}

	return b.String()
}

// getCurrentSegmentLevel returns the segment level to highlight based on the current input.
func (m Model) getCurrentSegmentLevel() int {
	input := m.textInput.Value()
	if input == "" {
		return 0
	}

	// Count colons to determine segment level
	colonCount := strings.Count(input, ":")

	return colonCount
}

// getSegmentBoundaries returns the start and end positions of each segment in the task name.
func (m Model) getSegmentBoundaries(taskName string) [][]int {
	segments := strings.Split(taskName, ":")
	bounds := make([][]int, 0, len(segments))
	pos := 0

	for i, seg := range segments {
		start := pos
		end := pos + len(seg)
		bounds = append(bounds, []int{start, end})
		pos = end + 1 // +1 for the colon separator

		// no add extra for last segment
		if i == len(segments)-1 {
			break
		}
	}

	return bounds
}

// getSegmentIndex returns which segment index a character position belongs to.
func (m Model) getSegmentIndex(charPos int, segmentBounds [][]int) int {
	for i, bounds := range segmentBounds {
		if charPos >= bounds[0] && charPos < bounds[1] {
			return i
		}
	}
	// If its a colon separator, return the segment before it
	if len(segmentBounds) > 0 {
		for i := 0; i < len(segmentBounds)-1; i++ {
			if charPos == segmentBounds[i][1] {
				return i
			}
		}
	}
	return 0
}

func (m Model) renderPrompt() string {
	return m.textInput.View()
}

func (m Model) getVisibleWindow() ([]TaskItem, int) {
	if len(m.filteredTasks) == 0 {
		return []TaskItem{}, 0
	}

	maxVisible := m.height - 2
	if maxVisible < 1 {
		maxVisible = 1
	}

	if len(m.filteredTasks) <= maxVisible {
		return m.filteredTasks, 0
	}

	start := m.cursor - maxVisible/2
	if start < 0 {
		start = 0
	}

	end := start + maxVisible
	if end > len(m.filteredTasks) {
		end = len(m.filteredTasks)
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}

	return m.filteredTasks[start:end], start
}

// SelectedTask returns the task selected by the user, or nil if no task was selected.
func (m Model) SelectedTask() *registry.Task {
	return m.selectedTask
}
