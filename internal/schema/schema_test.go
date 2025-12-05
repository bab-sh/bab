package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/bab-sh/bab/internal/babfile"
	"github.com/bab-sh/bab/internal/schema"
)

func TestGenerateSchema(t *testing.T) {
	s := schema.GenerateSchema()

	if s.Title != "Babfile Schema" {
		t.Errorf("expected title 'Babfile Schema', got %q", s.Title)
	}

	if s.ID != "https://bab.sh/schema/babfile.json" {
		t.Errorf("expected ID 'https://bab.sh/schema/babfile.json', got %q", s.ID)
	}

	if s.Description == "" {
		t.Error("schema should have a description")
	}
}

func TestSchemaIsValidJSON(t *testing.T) {
	s := schema.GenerateSchema()

	data, err := json.MarshalIndent(s, "", "  ")
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
	s := schema.GenerateSchema()

	data, err := json.Marshal(s)
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
		"Schema",
		"Task",
		"Include",
	}

	for _, def := range requiredDefs {
		if _, ok := defs[def]; !ok {
			t.Errorf("schema should have definition for %q", def)
		}
	}
}

func TestPlatformSchemaHasEnum(t *testing.T) {
	platform := babfile.Platform("")
	s := platform.JSONSchema()

	if s.Type != "string" {
		t.Errorf("Platform schema type should be string, got %q", s.Type)
	}

	if len(s.Enum) != 3 {
		t.Errorf("Platform schema should have 3 enum values, got %d", len(s.Enum))
	}

	expectedPlatforms := map[string]bool{
		"linux":   false,
		"darwin":  false,
		"windows": false,
	}

	for _, v := range s.Enum {
		if str, ok := v.(string); ok {
			expectedPlatforms[str] = true
		}
	}

	for platform, found := range expectedPlatforms {
		if !found {
			t.Errorf("Platform enum should include %q", platform)
		}
	}
}

func TestTaskSchemaNoNestedTasks(t *testing.T) {
	s := schema.GenerateSchema()

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	defs := parsed["$defs"].(map[string]interface{})
	taskDef := defs["Task"].(map[string]interface{})

	additionalProps, ok := taskDef["additionalProperties"]
	if !ok {
		t.Error("Task should have additionalProperties set")
		return
	}

	if additionalProps != false {
		t.Errorf("Task additionalProperties should be false (no nested tasks), got %v", additionalProps)
	}
}
