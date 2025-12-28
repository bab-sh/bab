package babfile

import "github.com/invopop/jsonschema"

func WhenSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Description: "Condition expression. Supports: ${{ var }}, ${{ var == 'value' }}, ${{ var != 'value' }}",
	}
}
