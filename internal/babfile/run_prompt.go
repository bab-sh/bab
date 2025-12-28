package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type PromptType string

const (
	PromptTypeConfirm     PromptType = "confirm"
	PromptTypeInput       PromptType = "input"
	PromptTypeSelect      PromptType = "select"
	PromptTypeMultiselect PromptType = "multiselect"
	PromptTypePassword    PromptType = "password"
	PromptTypeNumber      PromptType = "number"
)

var ValidPromptTypes = []PromptType{
	PromptTypeConfirm,
	PromptTypeInput,
	PromptTypeSelect,
	PromptTypeMultiselect,
	PromptTypePassword,
	PromptTypeNumber,
}

func (p PromptType) Valid() bool {
	for _, v := range ValidPromptTypes {
		if p == v {
			return true
		}
	}
	return false
}

func (p PromptType) String() string {
	return string(p)
}

func (p PromptType) RequiresOptions() bool {
	return p == PromptTypeSelect || p == PromptTypeMultiselect
}

func PromptTypeSchema() *jsonschema.Schema {
	enumValues := make([]any, len(ValidPromptTypes))
	for i, t := range ValidPromptTypes {
		enumValues[i] = string(t)
	}
	return &jsonschema.Schema{
		Type:        "string",
		Enum:        enumValues,
		Description: "Prompt type (confirm, input, select, multiselect, password, number)",
	}
}

type PromptRun struct {
	Line        int        `json:"-" yaml:"-"`
	Prompt      string     `json:"prompt" yaml:"prompt"`
	Type        PromptType `json:"type" yaml:"type"`
	Message     string     `json:"message" yaml:"message"`
	Platforms   []Platform `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	Default     string     `json:"default,omitempty" yaml:"default,omitempty"`
	Defaults    []string   `json:"defaults,omitempty" yaml:"defaults,omitempty"`
	Options     []string   `json:"options,omitempty" yaml:"options,omitempty"`
	Placeholder string     `json:"placeholder,omitempty" yaml:"placeholder,omitempty"`
	Validate    string     `json:"validate,omitempty" yaml:"validate,omitempty"`
	Min         *int       `json:"min,omitempty" yaml:"min,omitempty"`
	Max         *int       `json:"max,omitempty" yaml:"max,omitempty"`
	Confirm     *bool      `json:"confirm,omitempty" yaml:"confirm,omitempty"`
	When        string     `json:"when,omitempty" yaml:"when,omitempty"`
}

func (PromptRun) isRunItem() {}

func (p PromptRun) ShouldRunOnPlatform(platform string) bool {
	return matchesPlatform(p.Platforms, platform)
}

func (p PromptRun) GetWhen() string {
	return p.When
}

func PromptRunSchema() *jsonschema.Schema {
	minLen := uint64(1)
	minItems := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()

	props.Set("prompt", &jsonschema.Schema{
		Type:        "string",
		MinLength:   &minLen,
		Pattern:     VarNamePattern,
		Description: "Variable name to store the prompt result",
	})
	props.Set("type", PromptTypeSchema())
	props.Set("message", &jsonschema.Schema{
		Type:        "string",
		MinLength:   &minLen,
		Description: "Message to display to the user",
	})
	props.Set("default", &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{Type: "string"},
			{Type: "boolean"},
		},
		Description: "Default value for the prompt",
	})
	props.Set("defaults", &jsonschema.Schema{
		Type:        "array",
		Items:       &jsonschema.Schema{Type: "string"},
		Description: "Default selections for multiselect",
	})
	props.Set("options", &jsonschema.Schema{
		Type:        "array",
		MinItems:    &minItems,
		Items:       &jsonschema.Schema{Type: "string", MinLength: &minLen},
		Description: "Available options for select/multiselect",
	})
	props.Set("placeholder", &jsonschema.Schema{
		Type:        "string",
		Description: "Placeholder text for input prompts",
	})
	props.Set("validate", &jsonschema.Schema{
		Type:        "string",
		Description: "Regex pattern to validate input",
	})
	props.Set("min", &jsonschema.Schema{
		Type:        "integer",
		Description: "Minimum value for number, or minimum selections for multiselect",
	})
	props.Set("max", &jsonschema.Schema{
		Type:        "integer",
		Description: "Maximum value for number, or maximum selections for multiselect",
	})
	props.Set("confirm", &jsonschema.Schema{
		Type:        "boolean",
		Description: "Require password confirmation (re-entry)",
	})
	props.Set("platforms", PlatformsArraySchema())
	props.Set("when", WhenSchema())

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Interactive user prompt",
		Required:             []string{"prompt", "type", "message"},
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}
