package picker

import (
	"sort"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/charmbracelet/bubbles/textinput"
)

const (
	scrollMargin = 5
	headerLines  = 4
)

type Model struct {
	input   textinput.Model
	tasks   []*babfile.Task
	matches []Match

	cursor int
	offset int

	selected *babfile.Task
	quitting bool

	width  int
	height int
}

func New(tasks babfile.TaskMap) Model {
	list := make([]*babfile.Task, 0, len(tasks))
	for _, t := range tasks {
		list = append(list, t)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50
	ti.Prompt = "> "
	ti.PromptStyle = promptStyle
	ti.TextStyle = inputStyle

	m := Model{input: ti, tasks: list, width: 80, height: 20}
	m.updateMatches()
	return m
}

func (m *Model) updateMatches() {
	m.matches = search(m.input.Value(), m.tasks)
	m.cursor = 0
	m.offset = 0
}

func (m *Model) moveCursor(delta int) {
	m.cursor = max(0, min(m.cursor+delta, len(m.matches)-1))
	m.adjustScroll()
}

func (m *Model) adjustScroll() {
	visible := m.visibleLines()
	if m.cursor < m.offset+scrollMargin {
		m.offset = max(0, m.cursor-scrollMargin)
	}
	if m.cursor >= m.offset+visible-scrollMargin {
		m.offset = m.cursor - visible + scrollMargin + 1
	}
	m.offset = min(m.offset, max(0, len(m.matches)-visible))
}

func (m *Model) visibleLines() int {
	return max(1, m.height-headerLines)
}

func (m Model) Selected() *babfile.Task { return m.selected }
