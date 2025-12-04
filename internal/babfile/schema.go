package babfile

import "github.com/invopop/jsonschema"

func GenerateSchema() *jsonschema.Schema {
	reflector := &jsonschema.Reflector{
		DoNotReference: false,
	}

	schema := reflector.Reflect(&BabfileSchema{})
	schema.ID = "https://bab.sh/schema/babfile.json"
	schema.Title = "Babfile Schema"
	schema.Description = "Schema for Babfile task runner configuration (https://bab.sh)"

	return schema
}
