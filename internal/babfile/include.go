package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type Include struct {
	Babfile string `json:"babfile" yaml:"babfile"`
}

func (Include) JSONSchema() *jsonschema.Schema {
	minLen := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("babfile", &jsonschema.Schema{
		Type:        "string",
		MinLength:   &minLen,
		Pattern:     ".*babfile(\\..*)?\\.(ya?ml)$",
		Description: "Path to babfile",
	})

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "External babfile reference",
		Required:             []string{"babfile"},
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}
