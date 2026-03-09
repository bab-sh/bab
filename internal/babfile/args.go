package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type ArgDef struct {
	Default *string
	Line    int
}

type ArgMap map[string]ArgDef

func ArgsSchema() *jsonschema.Schema {
	defaultProps := orderedmap.New[string, *jsonschema.Schema]()
	defaultProps.Set("default", &jsonschema.Schema{
		Type:        "string",
		Description: "Default value for the argument",
	})

	return &jsonschema.Schema{
		Type:        "object",
		Description: "Named arguments the task accepts. Null value = required, object with default = optional.",
		PropertyNames: &jsonschema.Schema{
			Pattern:     VarNamePattern,
			Description: "Argument name",
		},
		AdditionalProperties: &jsonschema.Schema{
			OneOf: []*jsonschema.Schema{
				{Type: "null", Description: "Required argument (no default)"},
				{
					Type:                 "object",
					Description:          "Optional argument with default value",
					Properties:           defaultProps,
					Required:             []string{"default"},
					AdditionalProperties: jsonschema.FalseSchema,
				},
			},
		},
	}
}

func TaskRunArgsSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "Arguments to pass to the referenced task",
		PropertyNames: &jsonschema.Schema{
			Pattern: VarNamePattern,
		},
		AdditionalProperties: &jsonschema.Schema{
			Type: "string",
		},
	}
}
