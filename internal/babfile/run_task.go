package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type TaskRun struct {
	Line      int        `json:"-" yaml:"-"`
	Task      string     `json:"task" yaml:"task"`
	Silent    *bool      `json:"silent,omitempty" yaml:"silent,omitempty"`
	Output    *bool      `json:"output,omitempty" yaml:"output,omitempty"`
	Platforms []Platform `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}

func (TaskRun) isRunItem() {}

func (t TaskRun) ShouldRunOnPlatform(platform string) bool {
	return matchesPlatform(t.Platforms, platform)
}

func TaskRunSchema() *jsonschema.Schema {
	minLen := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("task", &jsonschema.Schema{
		Type:        "string",
		MinLength:   &minLen,
		Pattern:     TaskNamePattern,
		Description: "Task reference",
	})
	props.Set("silent", SilentSchema())
	props.Set("output", OutputSchema())
	props.Set("platforms", PlatformsArraySchema())

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Task reference",
		Required:             []string{"task"},
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}
