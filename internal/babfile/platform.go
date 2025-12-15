package babfile

import "github.com/invopop/jsonschema"

type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformDarwin  Platform = "darwin"
	PlatformWindows Platform = "windows"
)

var ValidPlatforms = []Platform{PlatformLinux, PlatformDarwin, PlatformWindows}

func (p Platform) Valid() bool {
	for _, v := range ValidPlatforms {
		if p == v {
			return true
		}
	}
	return false
}

func (p Platform) String() string {
	return string(p)
}

func (Platform) JSONSchema() *jsonschema.Schema {
	enumValues := make([]any, len(ValidPlatforms))
	for i, p := range ValidPlatforms {
		enumValues[i] = string(p)
	}
	return &jsonschema.Schema{
		Type:        "string",
		Enum:        enumValues,
		Description: "Target operating system",
	}
}

func PlatformsArraySchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "array",
		Description: "Run only on specified platforms",
		Items:       &jsonschema.Schema{Ref: "#/$defs/Platform"},
		UniqueItems: true,
	}
}
