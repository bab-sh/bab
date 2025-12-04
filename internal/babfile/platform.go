package babfile

import "github.com/invopop/jsonschema"

type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformDarwin  Platform = "darwin"
	PlatformWindows Platform = "windows"
)

var ValidPlatforms = []Platform{PlatformLinux, PlatformDarwin, PlatformWindows}

func (Platform) JSONSchema() *jsonschema.Schema {
	enumValues := make([]any, len(ValidPlatforms))
	for i, p := range ValidPlatforms {
		enumValues[i] = string(p)
	}
	return &jsonschema.Schema{
		Type:        "string",
		Enum:        enumValues,
		Description: "Target platform for the command",
	}
}
