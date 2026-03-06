package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/bab-sh/bab/internal/schema"
)

func parseSchema(t *testing.T) map[string]any {
	t.Helper()
	s := schema.GenerateSchema()
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}
	return parsed
}

func getMap(t *testing.T, m map[string]any, key string) map[string]any {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Fatalf("key %q not found", key)
	}
	result, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("key %q is not a map, got %T", key, v)
	}
	return result
}

func getSlice(t *testing.T, m map[string]any, key string) []any {
	t.Helper()
	v, ok := m[key]
	if !ok {
		t.Fatalf("key %q not found", key)
	}
	result, ok := v.([]any)
	if !ok {
		t.Fatalf("key %q is not a slice, got %T", key, v)
	}
	return result
}

func TestGenerateSchema(t *testing.T) {
	s := schema.GenerateSchema()

	if s.Title != "Babfile Schema" {
		t.Errorf("expected title 'Babfile Schema', got %q", s.Title)
	}

	if s.ID != "https://bab.sh/schema/babfile.schema.json" {
		t.Errorf("expected ID 'https://bab.sh/schema/babfile.schema.json', got %q", s.ID)
	}

	if s.Description == "" {
		t.Error("schema should have a description")
	}
}

func TestSchemaUsesDraft07(t *testing.T) {
	s := schema.GenerateSchema()

	expectedVersion := "http://json-schema.org/draft-07/schema#"
	if s.Version != expectedVersion {
		t.Errorf("expected schema version %q for IDE compatibility, got %q", expectedVersion, s.Version)
	}
}

func TestTaskSchemaHasRunWithOneOf(t *testing.T) {
	parsed := parseSchema(t)
	defs := getMap(t, parsed, "$defs")
	taskDef := getMap(t, defs, "Task")
	props := getMap(t, taskDef, "properties")
	runProp := getMap(t, props, "run")
	items := getMap(t, runProp, "items")
	oneOf := getSlice(t, items, "oneOf")

	if len(oneOf) != 4 {
		t.Fatalf("run items should have 4 oneOf options, got %d", len(oneOf))
	}

	expected := []string{"cmd", "task", "log", "prompt"}
	for i, key := range expected {
		option, ok := oneOf[i].(map[string]any)
		if !ok {
			t.Fatalf("oneOf[%d] is not a map", i)
		}
		optProps, ok := option["properties"].(map[string]any)
		if !ok {
			t.Fatalf("oneOf[%d] has no properties", i)
		}
		if _, ok := optProps[key]; !ok {
			t.Errorf("oneOf[%d] should have %q property", i, key)
		}
	}
}

func TestSchemaHasRequiredDefinitions(t *testing.T) {
	parsed := parseSchema(t)
	defs := getMap(t, parsed, "$defs")

	requiredDefs := []string{"Task", "TaskName", "Include", "Platform"}
	for _, def := range requiredDefs {
		if _, ok := defs[def]; !ok {
			t.Errorf("schema should have definition for %q", def)
		}
	}
}

func TestTaskSchemaNoNestedTasks(t *testing.T) {
	parsed := parseSchema(t)
	defs := getMap(t, parsed, "$defs")
	taskDef := getMap(t, defs, "Task")

	additionalProps, ok := taskDef["additionalProperties"]
	if !ok {
		t.Error("Task should have additionalProperties set")
		return
	}

	if additionalProps != false {
		t.Errorf("Task additionalProperties should be false, got %v", additionalProps)
	}
}

func TestTaskRunRequiresAtLeastOneItem(t *testing.T) {
	parsed := parseSchema(t)
	defs := getMap(t, parsed, "$defs")
	taskDef := getMap(t, defs, "Task")
	props := getMap(t, taskDef, "properties")
	runProp := getMap(t, props, "run")

	minItems, ok := runProp["minItems"]
	if !ok {
		t.Error("run should have minItems constraint")
		return
	}

	if minItems != float64(1) {
		t.Errorf("run minItems should be 1, got %v", minItems)
	}
}

func TestSchemaHasTaskNameDefinition(t *testing.T) {
	parsed := parseSchema(t)
	defs := getMap(t, parsed, "$defs")
	taskNameDef := getMap(t, defs, "TaskName")

	if taskNameDef["type"] != "string" {
		t.Errorf("TaskName type should be string, got %v", taskNameDef["type"])
	}

	if taskNameDef["pattern"] == nil {
		t.Error("TaskName should have pattern")
	}

	if taskNameDef["description"] == nil {
		t.Error("TaskName should have description")
	}
}
