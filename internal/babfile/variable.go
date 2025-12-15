package babfile

import "github.com/invopop/jsonschema"

const VarNamePattern = "^[a-zA-Z_][a-zA-Z0-9_]*$"

type VarMap map[string]string

type Vars map[string]string

func MergeVarMaps(varMaps ...VarMap) VarMap {
	merged := make(VarMap)
	for _, vm := range varMaps {
		for k, v := range vm {
			merged[k] = v
		}
	}
	return merged
}

func VarsSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "Variables as key-value pairs. Values can reference other variables or environment variables using ${{ var }} or ${{ env.VAR }} syntax.",
		PropertyNames: &jsonschema.Schema{
			Pattern:     VarNamePattern,
			Description: "Variable name (alphanumeric and underscores, must start with letter or underscore)",
		},
		AdditionalProperties: &jsonschema.Schema{
			Type: "string",
		},
	}
}
