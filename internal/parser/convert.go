package parser

import "github.com/bab-sh/bab/internal/babfile"

func convertTask(name string, st babfile.Task) *Task {
	commands := make([]Command, len(st.Run))
	for i, cmd := range st.Run {
		commands[i] = Command{
			Cmd:       cmd.Cmd,
			Platforms: cmd.Platforms,
		}
	}

	return &Task{
		Name:         name,
		Description:  st.Desc,
		Commands:     commands,
		Dependencies: st.Deps,
	}
}
