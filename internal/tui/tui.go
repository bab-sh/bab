package tui

import (
	"errors"
	"fmt"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/tui/picker"
	tea "github.com/charmbracelet/bubbletea"
)

var ErrNoTasks = errors.New("no tasks available")

func PickTask(tasks babfile.TaskMap) (*babfile.Task, error) {
	if len(tasks) == 0 {
		return nil, ErrNoTasks
	}

	result, err := tea.NewProgram(picker.New(tasks), tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}

	model, ok := result.(picker.Model)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
	return model.Selected(), nil
}
