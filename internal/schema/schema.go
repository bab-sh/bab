package schema

import (
	"github.com/bab-sh/bab/internal/babfile"
	"github.com/invopop/jsonschema"
)

const (
	ID          = "https://bab.sh/schema/babfile.json"
	Title       = "Babfile Schema"
	Description = "Schema for Babfile task runner configuration (https://bab.sh)"
)

func GenerateSchema() *jsonschema.Schema {
	r := &jsonschema.Reflector{}

	schema := r.Reflect(&babfile.Schema{})

	if schema.Definitions == nil {
		schema.Definitions = make(jsonschema.Definitions)
	}

	schema.Definitions["Task"] = babfile.Task{}.JSONSchema()

	schema.Definitions["Include"] = babfile.Include{}.JSONSchema()

	schema.Definitions["Platform"] = babfile.Platform("").JSONSchema()

	schema.Version = "http://json-schema.org/draft-07/schema#"
	schema.ID = jsonschema.ID(ID)
	schema.Title = Title
	schema.Description = Description

	return schema
}
