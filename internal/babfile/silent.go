package babfile

import "github.com/invopop/jsonschema"

func SilentSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "boolean",
		Description: "Suppress command execution output (e.g., '$ go mod download')",
	}
}
