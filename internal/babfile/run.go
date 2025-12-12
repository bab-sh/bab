package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type RunItem interface {
	isRunItem()
	ShouldRunOnPlatform(platform string) bool
}

type CommandRun struct {
	Line      int               `json:"-" yaml:"-"`
	Cmd       string            `json:"cmd" yaml:"cmd"`
	Env       map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Platforms []Platform        `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}

func (CommandRun) isRunItem() {}

func (c CommandRun) ShouldRunOnPlatform(platform string) bool {
	return matchesPlatform(c.Platforms, platform)
}

func (CommandRun) JSONSchema() *jsonschema.Schema {
	minLen := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("cmd", &jsonschema.Schema{
		Type:        "string",
		MinLength:   &minLen,
		Description: "Shell command to execute",
	})
	props.Set("env", EnvSchema())
	props.Set("platforms", PlatformsArraySchema())

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Shell command",
		Required:             []string{"cmd"},
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}

type TaskRun struct {
	Line      int        `json:"-" yaml:"-"`
	Task      string     `json:"task" yaml:"task"`
	Platforms []Platform `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}

func (TaskRun) isRunItem() {}

func (t TaskRun) ShouldRunOnPlatform(platform string) bool {
	return matchesPlatform(t.Platforms, platform)
}

func (TaskRun) JSONSchema() *jsonschema.Schema {
	minLen := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("task", &jsonschema.Schema{
		Type:        "string",
		MinLength:   &minLen,
		Pattern:     TaskNamePattern,
		Description: "Task reference",
	})
	props.Set("platforms", PlatformsArraySchema())

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Task reference",
		Required:             []string{"task"},
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}

func RunItemSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			CommandRun{}.JSONSchema(),
			TaskRun{}.JSONSchema(),
		},
	}
}

func matchesPlatform(platforms []Platform, platform string) bool {
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
