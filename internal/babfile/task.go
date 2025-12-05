package babfile

import (
	"gopkg.in/yaml.v3"
)

type Task struct {
	Desc string    `json:"desc,omitempty" yaml:"desc,omitempty"`
	Deps []string  `json:"deps,omitempty" yaml:"deps,omitempty"`
	Run  []RunItem `json:"-" yaml:"-"`
}

type rawTask struct {
	Desc string    `yaml:"desc"`
	Deps []string  `yaml:"deps"`
	Run  yaml.Node `yaml:"run"`
}

func (t *Task) UnmarshalYAML(node *yaml.Node) error {
	var raw rawTask
	if err := node.Decode(&raw); err != nil {
		return err
	}

	t.Desc = raw.Desc
	t.Deps = raw.Deps

	if raw.Run.Kind != 0 {
		items, err := ParseRunItems(&raw.Run)
		if err != nil {
			return err
		}
		t.Run = items
	}

	return nil
}
