package parser

import "github.com/bab-sh/bab/internal/babfile"

type Task struct {
	Name         string
	Description  string
	Commands     []Command
	Dependencies []string
}

func (t *Task) HasDependencies() bool {
	return len(t.Dependencies) > 0
}

type Command struct {
	Cmd       string
	Platforms []babfile.Platform
}

func (c Command) ShouldRunOnPlatform(platform string) bool {
	if len(c.Platforms) == 0 {
		return true
	}
	for _, p := range c.Platforms {
		if string(p) == platform {
			return true
		}
	}
	return false
}

type TaskMap map[string]*Task

func (tm TaskMap) Has(name string) bool {
	_, exists := tm[name]
	return exists
}

func (tm TaskMap) Names() []string {
	names := make([]string, 0, len(tm))
	for name := range tm {
		names = append(names, name)
	}
	return names
}
