package tui

import (
	"fmt"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/errs"
	"github.com/bab-sh/bab/internal/tui/picker"
	tea "github.com/charmbracelet/bubbletea"
)

func PickTask(tasks babfile.TaskMap) (*babfile.Task, error) {
	if len(tasks) == 0 {
		return nil, errs.ErrNoTasks
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
