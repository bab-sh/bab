package parser

import "github.com/bab-sh/bab/internal/babfile"

func convertTask(name string, st babfile.Task) *Task {
	runItems := make([]RunItem, 0, len(st.Run))
	for _, item := range st.Run {
		switch v := item.(type) {
		case babfile.CommandRun:
			runItems = append(runItems, CommandRun{
				Cmd:       v.Cmd,
				Platforms: v.Platforms,
			})
		case babfile.TaskRun:
			runItems = append(runItems, TaskRun{
				TaskRef:   v.Task,
				Platforms: v.Platforms,
			})
		}
	}

	return &Task{
		Name:         name,
		Description:  st.Desc,
		RunItems:     runItems,
		Dependencies: st.Deps,
	}
}
