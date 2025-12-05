package babfile

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type RunItem interface {
	isRunItem()
}

type CommandRun struct {
	Cmd       string     `json:"cmd" yaml:"cmd"`
	Platforms []Platform `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}

func (CommandRun) isRunItem() {}

type TaskRun struct {
	Task      string     `json:"task" yaml:"task"`
	Platforms []Platform `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}

func (TaskRun) isRunItem() {}

type rawRunItem struct {
	Cmd       string     `yaml:"cmd"`
	Task      string     `yaml:"task"`
	Platforms []Platform `yaml:"platforms"`
}

func ParseRunItems(node *yaml.Node) ([]RunItem, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("run must be a sequence")
	}

	items := make([]RunItem, 0, len(node.Content))
	for i, itemNode := range node.Content {
		if itemNode.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("run item %d must be a mapping", i)
		}

		var raw rawRunItem
		if err := itemNode.Decode(&raw); err != nil {
			return nil, fmt.Errorf("run item %d: %w", i, err)
		}

		item, err := convertRawRunItem(raw)
		if err != nil {
			return nil, fmt.Errorf("run item %d: %w", i, err)
		}
		items = append(items, item)
	}

	return items, nil
}

func convertRawRunItem(raw rawRunItem) (RunItem, error) {
	hasCmd := raw.Cmd != ""
	hasTask := raw.Task != ""

	switch {
	case hasCmd && hasTask:
		return nil, fmt.Errorf("cannot have both 'cmd' and 'task'")
	case hasCmd:
		return CommandRun{Cmd: raw.Cmd, Platforms: raw.Platforms}, nil
	case hasTask:
		return TaskRun{Task: raw.Task, Platforms: raw.Platforms}, nil
	default:
		return nil, fmt.Errorf("must have either 'cmd' or 'task'")
	}
}
