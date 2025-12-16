package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type CommandRun struct {
	Line      int               `json:"-" yaml:"-"`
	Cmd       string            `json:"cmd" yaml:"cmd"`
	Env       map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Silent    *bool             `json:"silent,omitempty" yaml:"silent,omitempty"`
	Platforms []Platform        `json:"platforms,omitempty" yaml:"platforms,omitempty"`
}

func (CommandRun) isRunItem() {}

func (c CommandRun) ShouldRunOnPlatform(platform string) bool {
	return matchesPlatform(c.Platforms, platform)
}

func CommandRunSchema() *jsonschema.Schema {
	minLen := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("cmd", &jsonschema.Schema{
		Type:        "string",
		MinLength:   &minLen,
		Description: "Shell command to execute",
	})
	props.Set("env", EnvSchema())
	props.Set("silent", SilentSchema())
	props.Set("platforms", PlatformsArraySchema())

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Shell command",
		Required:             []string{"cmd"},
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}
