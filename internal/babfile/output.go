package babfile

import "github.com/invopop/jsonschema"

func OutputSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "boolean",
		Description: "Show process output (stdout/stderr). Defaults to true.",
	}
}
