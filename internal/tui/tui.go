package tui

import (
	"context"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/errs"
	"github.com/bab-sh/bab/internal/tui/picker"
)

func PickTask(ctx context.Context, tasks babfile.TaskMap) (*babfile.Task, error) {
	if len(tasks) == 0 {
		return nil, errs.ErrNoTasks
	}

	result, err := tea.NewProgram(picker.New(tasks), tea.WithContext(ctx)).Run()
	if err != nil {
		if errors.Is(err, tea.ErrProgramKilled) {
			return nil, context.Canceled
		}
		return nil, err
	}

	model, ok := result.(picker.Model)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
	return model.Selected(), nil
}
