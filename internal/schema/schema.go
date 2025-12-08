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
	schema := babfile.Schema{}.JSONSchema()
	schema.Version = "http://json-schema.org/draft-07/schema#"
	schema.ID = jsonschema.ID(ID)
	schema.Title = Title
	schema.Description = Description
	return schema
}
