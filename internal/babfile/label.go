package babfile

import "github.com/invopop/jsonschema"

func LabelSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Description: "Display label for this item when running inside a parallel block",
	}
}
