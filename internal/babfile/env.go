package babfile

import "github.com/invopop/jsonschema"

func EnvSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "object",
		Description: "Environment variables as key-value pairs",
		AdditionalProperties: &jsonschema.Schema{
			Type: "string",
		},
	}
}

func MergeEnv(envMaps ...map[string]string) []string {
	merged := MergeEnvMaps(envMaps...)
	result := make([]string, 0, len(merged))
	for k, v := range merged {
		result = append(result, k+"="+v)
	}
	return result
}

func MergeEnvMaps(envMaps ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, env := range envMaps {
		for k, v := range env {
			merged[k] = v
		}
	}
	return merged
}
