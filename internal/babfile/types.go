package babfile

import "github.com/invopop/jsonschema"

type Platform string

const (
	PlatformLinux   Platform = "linux"
	PlatformDarwin  Platform = "darwin"
	PlatformWindows Platform = "windows"
)

var ValidPlatforms = []Platform{PlatformLinux, PlatformDarwin, PlatformWindows}

func ValidPlatformStrings() []string {
	result := make([]string, len(ValidPlatforms))
	for i, p := range ValidPlatforms {
		result[i] = string(p)
	}
	return result
}

func IsValidPlatform(s string) bool {
	for _, p := range ValidPlatforms {
		if string(p) == s {
			return true
		}
	}
	return false
}

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

type Command struct {
	Cmd       string     `json:"cmd" yaml:"cmd" jsonschema:"description=Shell command to execute,minLength=1"`
	Platforms []Platform `json:"platforms,omitempty" yaml:"platforms,omitempty" jsonschema:"description=Platforms to run this command on (if empty runs on all platforms)"`
}

func (c Command) ShouldRunOnPlatform(platform string) bool {
	if len(c.Platforms) == 0 {
		return true
	}
	for _, p := range c.Platforms {
		if string(p) == platform {
			return true
		}
	}
	return false
}

type Task struct {
	Name         string
	Description  string
	Commands     []Command
	Dependencies []string
}

type TaskMap map[string]*Task

type IncludeMap map[string]string

type ParseContext struct {
	Visited map[string]bool `json:"-"`
	Stack   []string        `json:"-"`
}

func NewParseContext() *ParseContext {
	return &ParseContext{
		Visited: make(map[string]bool),
		Stack:   make([]string, 0),
	}
}

type BabfileSchema struct {
	Includes map[string]IncludeConfig `json:"includes,omitempty" jsonschema:"description=External babfile imports with namespace prefixes"`
	Tasks    map[string]TaskDef       `json:"tasks" jsonschema:"description=Task definitions"`
}

type IncludeConfig struct {
	Babfile string `json:"babfile" jsonschema:"description=Relative or absolute path to the babfile to include"`
}

type TaskDef struct {
	Desc string       `json:"desc,omitempty" jsonschema:"description=Human-readable description of the task"`
	Deps Dependencies `json:"deps,omitempty" jsonschema:"description=Task dependencies to run before this task"`
	Run  []CommandDef `json:"run,omitempty" jsonschema:"description=List of commands to execute"`
}

func (TaskDef) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.AdditionalProperties = &jsonschema.Schema{
		Ref: "#/$defs/TaskDef",
	}
}

type Dependencies struct{}

func (Dependencies) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type:        "string",
				Description: "Single dependency task name",
			},
			{
				Type: "array",
				Items: &jsonschema.Schema{
					Type: "string",
				},
				Description: "List of dependency task names",
			},
		},
	}
}

type CommandDef struct {
	Cmd       string     `json:"cmd" jsonschema:"description=Shell command to execute,minLength=1"`
	Platforms []Platform `json:"platforms,omitempty" jsonschema:"description=Platforms to run this command on (if empty runs on all platforms)"`
}
