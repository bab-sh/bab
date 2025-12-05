package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Task struct {
	Desc string    `json:"desc,omitempty" yaml:"desc,omitempty"`
	Deps []string  `json:"deps,omitempty" yaml:"deps,omitempty"`
	Run  []RunItem `json:"-" yaml:"-"`
}

func (Task) JSONSchema() *jsonschema.Schema {
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("desc", &jsonschema.Schema{
		Type:        "string",
		Description: "Task description",
	})
	props.Set("deps", DepsSchema())
	props.Set("run", &jsonschema.Schema{
		Type:        "array",
		Description: "Commands or tasks to execute",
		Items:       RunItemSchema(),
	})

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Task definition",
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}

func DepsSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "array",
		Description: "Tasks to run first",
		Items: &jsonschema.Schema{
			Type:    "string",
			Pattern: "^[a-zA-Z0-9_-]+(:[a-zA-Z0-9_-]+)?$",
		},
		UniqueItems: true,
	}
}
