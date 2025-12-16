package babfile

import (
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const TaskNamePattern = "^[a-zA-Z0-9_-]+(:[a-zA-Z0-9_-]+)*$"

type Task struct {
	Name     string            `json:"-" yaml:"-"`
	Line     int               `json:"-" yaml:"-"`
	DepsLine int               `json:"-" yaml:"-"`
	Desc     string            `json:"desc,omitempty" yaml:"desc,omitempty"`
	Vars     VarMap            `json:"vars,omitempty" yaml:"vars,omitempty"`
	Env      map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Silent   *bool             `json:"silent,omitempty" yaml:"silent,omitempty"`
	Output   *bool             `json:"output,omitempty" yaml:"output,omitempty"`
	Deps     []string          `json:"deps,omitempty" yaml:"deps,omitempty"`
	Run      []RunItem         `json:"-" yaml:"-"`
}

type TaskMap map[string]*Task

func (tm TaskMap) Has(name string) bool {
	_, exists := tm[name]
	return exists
}

func (tm TaskMap) Names() []string {
	names := make([]string, 0, len(tm))
	for name := range tm {
		names = append(names, name)
	}
	return names
}

func (Task) JSONSchema() *jsonschema.Schema {
	minRunItems := uint64(1)
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("desc", &jsonschema.Schema{
		Type:        "string",
		Description: "Task description",
	})
	props.Set("vars", VarsSchema())
	props.Set("env", EnvSchema())
	props.Set("silent", SilentSchema())
	props.Set("output", OutputSchema())
	props.Set("deps", DepsSchema())
	props.Set("run", &jsonschema.Schema{
		Type:        "array",
		Description: "Commands or tasks to execute",
		MinItems:    &minRunItems,
		Items:       RunItemSchema(),
	})

	return &jsonschema.Schema{
		Type:                 "object",
		Description:          "Task definition",
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           props,
	}
}

func DepsSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "array",
		Description: "Tasks to run first",
		Items: &jsonschema.Schema{
			Type:        "string",
			Pattern:     TaskNamePattern,
			Description: "Task name to depend on",
		},
		UniqueItems: true,
	}
}

func TaskNameSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Pattern:     TaskNamePattern,
		Description: "Task name (alphanumeric, hyphens, underscores, colons for namespacing)",
	}
}
