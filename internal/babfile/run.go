package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type RunItem interface {
	isRunItem()
}

type CommandRun struct {
	Cmd       string     `json:"cmd" yaml:"cmd"`
	Platforms []Platform `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}

func (CommandRun) isRunItem() {}

func (CommandRun) JSONSchema() *jsonschema.Schema {
	minLen := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("cmd", &jsonschema.Schema{
		Type:        "string",
		MinLength:   &minLen,
		Description: "Shell command to execute",
	})
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
	Task      string     `json:"task" yaml:"task"`
	Platforms []Platform `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}

func (TaskRun) isRunItem() {}

func (TaskRun) JSONSchema() *jsonschema.Schema {
	minLen := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("task", &jsonschema.Schema{
		Type:        "string",
		MinLength:   &minLen,
		Pattern:     "^[a-zA-Z0-9_-]+(:[a-zA-Z0-9_-]+)*$",
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
