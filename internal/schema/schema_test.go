package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/bab-sh/bab/internal/babfile"
)

func TestGenerateSchema(t *testing.T) {
	schema := babfile.GenerateSchema()

	if schema.Title != "Babfile Schema" {
		t.Errorf("expected title 'Babfile Schema', got %q", schema.Title)
	}

	if schema.ID != "https://bab.sh/schema/babfile.json" {
		t.Errorf("expected ID 'https://bab.sh/schema/babfile.json', got %q", schema.ID)
	}

	if schema.Description == "" {
		t.Error("schema should have a description")
	}
}

func TestSchemaIsValidJSON(t *testing.T) {
	schema := babfile.GenerateSchema()

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal schema to JSON: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("generated schema is not valid JSON: %v", err)
	}

	if _, ok := parsed["$schema"]; !ok {
		t.Error("schema should have $schema key")
	}
	if _, ok := parsed["$defs"]; !ok {
		t.Error("schema should have $defs key")
	}
}

func TestSchemaHasRequiredDefinitions(t *testing.T) {
	schema := babfile.GenerateSchema()

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	defs, ok := parsed["$defs"].(map[string]interface{})
	if !ok {
		t.Fatal("$defs should be an object")
	}

	requiredDefs := []string{
		"BabfileSchema",
		"TaskDef",
		"CommandDef",
		"Dependencies",
		"Platform",
		"IncludeConfig",
	}

	for _, def := range requiredDefs {
		if _, ok := defs[def]; !ok {
			t.Errorf("schema should have definition for %q", def)
		}
	}
}

func TestDependenciesSchemaHasOneOf(t *testing.T) {
	deps := babfile.Dependencies{}
	schema := deps.JSONSchema()

	if len(schema.OneOf) != 2 {
		t.Errorf("Dependencies schema should have 2 oneOf options, got %d", len(schema.OneOf))
	}

	if schema.OneOf[0].Type != "string" {
		t.Errorf("first oneOf option should be string, got %q", schema.OneOf[0].Type)
	}

	if schema.OneOf[1].Type != "array" {
		t.Errorf("second oneOf option should be array, got %q", schema.OneOf[1].Type)
	}
}

func TestPlatformSchemaHasEnum(t *testing.T) {
	platform := babfile.Platform("")
	schema := platform.JSONSchema()

	if schema.Type != "string" {
		t.Errorf("Platform schema type should be string, got %q", schema.Type)
	}

	if len(schema.Enum) != 3 {
		t.Errorf("Platform schema should have 3 enum values, got %d", len(schema.Enum))
	}

	expectedPlatforms := map[string]bool{
		"linux":   false,
		"darwin":  false,
		"windows": false,
	}

	for _, v := range schema.Enum {
		if s, ok := v.(string); ok {
			expectedPlatforms[s] = true
		}
	}

	for platform, found := range expectedPlatforms {
		if !found {
			t.Errorf("Platform enum should include %q", platform)
		}
	}
}

func TestTaskDefSchemaHasAdditionalProperties(t *testing.T) {
	schema := babfile.GenerateSchema()

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	defs := parsed["$defs"].(map[string]interface{})
	taskDef := defs["TaskDef"].(map[string]interface{})

	additionalProps, ok := taskDef["additionalProperties"]
	if !ok {
		t.Error("TaskDef should have additionalProperties for nested subtasks")
	}

	if ref, ok := additionalProps.(map[string]interface{}); ok {
		if ref["$ref"] != "#/$defs/TaskDef" {
			t.Errorf("additionalProperties should reference TaskDef, got %v", ref["$ref"])
		}
	}
}
