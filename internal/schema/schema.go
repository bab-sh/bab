package schema

import (
	"github.com/bab-sh/bab/internal/babfile"
	"github.com/invopop/jsonschema"
)

const (
	SchemaID          = "https://bab.sh/schema/babfile.json"
	SchemaTitle       = "Babfile Schema"
	SchemaDescription = "Schema for Babfile task runner configuration (https://bab.sh)"
)

func GenerateSchema() *jsonschema.Schema {
	schema := (&jsonschema.Reflector{}).Reflect(&babfile.Schema{})
	schema.ID = jsonschema.ID(SchemaID)
	schema.Title = SchemaTitle
	schema.Description = SchemaDescription
	return schema
}
