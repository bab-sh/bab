package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

var ValidLogLevels = []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError}

func (l LogLevel) Valid() bool {
	for _, v := range ValidLogLevels {
		if l == v {
			return true
		}
	}
	return false
}

func (l LogLevel) String() string {
	return string(l)
}

func LogLevelSchema() *jsonschema.Schema {
	enumValues := make([]any, len(ValidLogLevels))
	for i, l := range ValidLogLevels {
		enumValues[i] = string(l)
	}
	return &jsonschema.Schema{
		Type:        "string",
		Enum:        enumValues,
		Default:     "info",
		Description: "Log level (debug, info, warn, error)",
	}
}

type LogRun struct {
	Line      int        `json:"-" yaml:"-"`
	Log       string     `json:"log" yaml:"log"`
	Level     LogLevel   `json:"level,omitempty" yaml:"level,omitempty"`
	Platforms []Platform `json:"platforms,omitempty" yaml:"platforms,omitempty"`
	When      string     `json:"when,omitempty" yaml:"when,omitempty"`
}

func (LogRun) isRunItem() {}

func (l LogRun) ShouldRunOnPlatform(platform string) bool {
	return matchesPlatform(l.Platforms, platform)
}

func (l LogRun) GetWhen() string {
	return l.When
}

func LogRunSchema() *jsonschema.Schema {
	minLen := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("log", &jsonschema.Schema{
		Type:        "string",
		MinLength:   &minLen,
		Description: "Log message to display",
	})
	props.Set("level", LogLevelSchema())
	props.Set("platforms", PlatformsArraySchema())
	props.Set("when", WhenSchema())

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Log message",
		Required:             []string{"log"},
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}
