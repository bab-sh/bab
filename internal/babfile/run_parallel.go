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
	Color     *bool        `json:"color,omitempty" yaml:"color,omitempty"`
	Platforms []Platform   `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	When      string       `json:"when,omitempty" yaml:"when,omitempty"`
}

func (p ParallelRun) UseColor() bool {
	return p.Color == nil || *p.Color
}

func (ParallelRun) isRunItem() {}

func (p ParallelRun) GetLine() int { return p.Line }

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
		return truncateRunes(v.Cmd, 20)
	case LogRun:
		return truncateRunes(v.Log, 20)
	default:
		return ""
	}
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max])
	}
	return s
}

func ParallelRunSchema() *jsonschema.Schema {
	minItems := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("parallel", &jsonschema.Schema{
		Type:        "array",
		Description: "Run items to execute in parallel",
		MinItems:    &minItems,
		Items:       &jsonschema.Schema{Ref: "#/$defs/ParallelChildItem"},
	})
	props.Set("mode", ParallelModeSchema())
	props.Set("limit", &jsonschema.Schema{
		Type:        "integer",
		Description: "Maximum number of items to run concurrently (0 = unlimited)",
		Minimum:     json.Number("0"),
	})
	props.Set("color", &jsonschema.Schema{
		Type:        "boolean",
		Default:     true,
		Description: "Preserve colors in child process output (default: true). When false, strips ANSI codes and sets NO_COLOR=1",
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
