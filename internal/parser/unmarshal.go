package parser

import (
	"fmt"

	"github.com/bab-sh/bab/internal/babfile"
	"gopkg.in/yaml.v3"
)

type rawSchema struct {
	Includes map[string]babfile.Include `yaml:"includes"`
	Tasks    map[string]rawTask         `yaml:"tasks"`
}

type rawTask struct {
	Desc string       `yaml:"desc"`
	Deps []string     `yaml:"deps"`
	Run  []rawRunItem `yaml:"run"`
}

type rawRunItem struct {
	Cmd       string             `yaml:"cmd"`
	Task      string             `yaml:"task"`
	Platforms []babfile.Platform `yaml:"platforms"`
}

func unmarshalBabfile(data []byte) (*babfile.Schema, error) {
	var raw rawSchema
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	schema := &babfile.Schema{
		Includes: raw.Includes,
		Tasks:    make(map[string]babfile.Task, len(raw.Tasks)),
	}

	for name, rt := range raw.Tasks {
		task, err := convertRawTask(rt)
		if err != nil {
			return nil, fmt.Errorf("task %q: %w", name, err)
		}
		schema.Tasks[name] = task
	}

	return schema, nil
}

func convertRawTask(rt rawTask) (babfile.Task, error) {
	runItems := make([]babfile.RunItem, 0, len(rt.Run))
	for i, item := range rt.Run {
		runItem, err := convertRawRunItem(item)
		if err != nil {
			return babfile.Task{}, fmt.Errorf("run[%d]: %w", i, err)
		}
		runItems = append(runItems, runItem)
	}

	return babfile.Task{
		Desc: rt.Desc,
		Deps: rt.Deps,
		Run:  runItems,
	}, nil
}

func convertRawRunItem(raw rawRunItem) (babfile.RunItem, error) {
	hasCmd := raw.Cmd != ""
	hasTask := raw.Task != ""

	switch {
	case hasCmd && hasTask:
		return nil, fmt.Errorf("cannot have both 'cmd' and 'task'")
	case hasCmd:
		return babfile.CommandRun{Cmd: raw.Cmd, Platforms: raw.Platforms}, nil
	case hasTask:
		return babfile.TaskRun{Task: raw.Task, Platforms: raw.Platforms}, nil
	default:
		return nil, fmt.Errorf("must have either 'cmd' or 'task'")
	}
}
