package picker

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.cursor < len(m.matches) {
				m.selected = m.matches[m.cursor].Task
			}
			m.quitting = true
			return m, tea.Quit
		case "up", "ctrl+p":
			m.moveCursor(-1)
			return m, nil
		case "down", "ctrl+n":
			m.moveCursor(1)
			return m, nil
		case "ctrl+u":
			m.input.SetValue("")
			m.updateMatches()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.Width = msg.Width - 4
		return m, nil
	}

	prev := m.input.Value()
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != prev {
		m.updateMatches()
	}
	return m, cmd
}
