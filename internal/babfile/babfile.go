package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Schema struct {
	Vars     VarMap             `json:"vars,omitempty" yaml:"vars,omitempty"`
	Env      map[string]string  `json:"env,omitempty" yaml:"env,omitempty"`
	Silent   *bool              `json:"silent,omitempty" yaml:"silent,omitempty"`
	Includes map[string]Include `json:"includes,omitempty" yaml:"includes,omitempty"`
	Tasks    map[string]Task    `json:"tasks" yaml:"tasks"`
}

func (Schema) JSONSchema() *jsonschema.Schema {
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("vars", VarsSchema())
	props.Set("env", EnvSchema())
	props.Set("silent", SilentSchema())
	props.Set("includes", &jsonschema.Schema{
		Type:        "object",
		Description: "External babfiles to import",
		PropertyNames: &jsonschema.Schema{
			Pattern: TaskNamePattern,
		},
		AdditionalProperties: &jsonschema.Schema{Ref: "#/$defs/Include"},
	})

	minTasks := uint64(1)
	props.Set("tasks", &jsonschema.Schema{
		Type:          "object",
		Description:   "Task definitions",
		MinProperties: &minTasks,
		PropertyNames: &jsonschema.Schema{
			Pattern: TaskNamePattern,
		},
		AdditionalProperties: &jsonschema.Schema{Ref: "#/$defs/Task"},
	})

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Babfile configuration",
		Required:             []string{"tasks"},
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
		Definitions: jsonschema.Definitions{
			"Task":     Task{}.JSONSchema(),
			"TaskName": TaskNameSchema(),
			"Include":  Include{}.JSONSchema(),
			"Platform": Platform("").JSONSchema(),
		},
	}
}
