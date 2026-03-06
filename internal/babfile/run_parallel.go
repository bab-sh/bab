package babfile

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type ParallelMode string

const (
	ParallelInterleaved ParallelMode = "interleaved"
	ParallelGrouped     ParallelMode = "grouped"
)

var ValidParallelModes = []ParallelMode{ParallelInterleaved, ParallelGrouped}

func (m ParallelMode) Valid() bool {
	for _, v := range ValidParallelModes {
		if m == v {
			return true
		}
	}
	return false
}

type ParallelRun struct {
	Line      int          `json:"-" yaml:"-"`
	Items     []RunItem    `json:"-" yaml:"-"`
	Labels    []string     `json:"-" yaml:"-"`
	Mode      ParallelMode `json:"mode,omitempty" yaml:"mode,omitempty"`
	Limit     int          `json:"limit,omitempty" yaml:"limit,omitempty"`
	Platforms []Platform   `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	When      string       `json:"when,omitempty" yaml:"when,omitempty"`
}

func (ParallelRun) isRunItem() {}

func (p ParallelRun) ShouldRunOnPlatform(platform string) bool {
	return matchesPlatform(p.Platforms, platform)
}

func (p ParallelRun) GetWhen() string {
	return p.When
}

func (p ParallelRun) ItemLabel(index int) string {
	if index < len(p.Labels) && p.Labels[index] != "" {
		return p.Labels[index]
	}
	if index >= len(p.Items) {
		return ""
	}
	switch v := p.Items[index].(type) {
	case TaskRun:
		return v.Task
	case CommandRun:
		cmd := v.Cmd
		if len(cmd) > 20 {
			cmd = cmd[:20]
		}
		return cmd
	case LogRun:
		msg := v.Log
		if len(msg) > 20 {
			msg = msg[:20]
		}
		return msg
	default:
		return ""
	}
}

func ParallelRunSchema() *jsonschema.Schema {
	minItems := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("parallel", &jsonschema.Schema{
		Type:        "array",
		Description: "Run items to execute in parallel",
		MinItems:    &minItems,
		Items:       &jsonschema.Schema{Ref: "#/$defs/RunItem"},
	})
	props.Set("mode", ParallelModeSchema())
	props.Set("limit", &jsonschema.Schema{
		Type:        "integer",
		Description: "Maximum number of items to run concurrently (0 = unlimited)",
		Minimum:     json.Number("0"),
	})
	props.Set("platforms", PlatformsArraySchema())
	props.Set("when", WhenSchema())

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Run items in parallel",
		Required:             []string{"parallel"},
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}

func ParallelModeSchema() *jsonschema.Schema {
	enumValues := make([]any, len(ValidParallelModes))
	for i, m := range ValidParallelModes {
		enumValues[i] = string(m)
	}
	return &jsonschema.Schema{
		Type:        "string",
		Enum:        enumValues,
		Default:     string(ParallelInterleaved),
		Description: "Display mode for parallel output (interleaved, grouped)",
	}
}
