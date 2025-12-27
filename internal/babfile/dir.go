package babfile

import "github.com/invopop/jsonschema"

func DirSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Description: "Working directory for command execution. Relative paths are resolved from the Babfile location.",
	}
}
