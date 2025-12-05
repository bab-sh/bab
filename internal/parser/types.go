package parser

import "github.com/bab-sh/bab/internal/babfile"

type Task struct {
	Name         string
	Description  string
	RunItems     []RunItem
	Dependencies []string
}

func (t *Task) HasDependencies() bool {
	return len(t.Dependencies) > 0
}

type RunItem interface {
	isRunItem()
	ShouldRunOnPlatform(platform string) bool
}

type CommandRun struct {
	Cmd       string
	Platforms []babfile.Platform
}

func (CommandRun) isRunItem() {}

func (c CommandRun) ShouldRunOnPlatform(platform string) bool {
	return matchesPlatform(c.Platforms, platform)
}

type TaskRun struct {
	TaskRef   string
	Platforms []babfile.Platform
}

func (TaskRun) isRunItem() {}

func (t TaskRun) ShouldRunOnPlatform(platform string) bool {
	return matchesPlatform(t.Platforms, platform)
}

func matchesPlatform(platforms []babfile.Platform, platform string) bool {
	if len(platforms) == 0 {
		return true
	}
	for _, p := range platforms {
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
