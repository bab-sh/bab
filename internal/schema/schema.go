package schema

import (
	"github.com/bab-sh/bab/internal/babfile"
	"github.com/invopop/jsonschema"
)

const (
	ID          = "https://bab.sh/schema/babfile.json"
	Title       = "Babfile Schema"
	Description = "Schema for Babfile task runner configuration (https://bab.sh)"
)

func GenerateSchema() *jsonschema.Schema {
	schema := (&jsonschema.Reflector{}).Reflect(&babfile.Schema{})
	schema.ID = jsonschema.ID(ID)
	schema.Title = Title
	schema.Description = Description

	if schema.Definitions != nil {
		schema.Definitions["Task"] = buildTaskSchema()
	}

	return schema
}

func buildTaskSchema() *jsonschema.Schema {
	one := uint64(1)

	platformSchema := &jsonschema.Schema{
		Type:        "array",
		Items:       &jsonschema.Schema{Type: "string", Enum: []any{"linux", "darwin", "windows"}},
		Description: "Platforms to run on (if empty runs on all platforms)",
	}

	cmdSchema := &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: jsonschema.FalseSchema,
		Required:             []string{"cmd"},
		Properties:           jsonschema.NewProperties(),
	}
	cmdSchema.Properties.Set("cmd", &jsonschema.Schema{Type: "string", MinLength: &one, Description: "Shell command to execute"})
	cmdSchema.Properties.Set("platforms", platformSchema)

	taskRefSchema := &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: jsonschema.FalseSchema,
		Required:             []string{"task"},
		Properties:           jsonschema.NewProperties(),
	}
	taskRefSchema.Properties.Set("task", &jsonschema.Schema{Type: "string", MinLength: &one, Description: "Task name to execute"})
	taskRefSchema.Properties.Set("platforms", platformSchema)

	runSchema := &jsonschema.Schema{
		Type:        "array",
		Items:       &jsonschema.Schema{OneOf: []*jsonschema.Schema{cmdSchema, taskRefSchema}},
		Description: "List of run items to execute",
	}

	taskSchema := &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: jsonschema.FalseSchema,
		Properties:           jsonschema.NewProperties(),
	}
	taskSchema.Properties.Set("desc", &jsonschema.Schema{Type: "string", Description: "Human-readable description of the task"})
	taskSchema.Properties.Set("deps", &jsonschema.Schema{Type: "array", Items: &jsonschema.Schema{Type: "string"}, Description: "Task dependencies to run before this task"})
	taskSchema.Properties.Set("run", runSchema)

	return taskSchema
}
