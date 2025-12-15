package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/bab-sh/bab/internal/schema"
)

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
	s := schema.GenerateSchema()

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	defs := parsed["$defs"].(map[string]any)
	taskDef := defs["Task"].(map[string]any)
	props := taskDef["properties"].(map[string]any)
	runProp := props["run"].(map[string]any)
	items := runProp["items"].(map[string]any)
	oneOf := items["oneOf"].([]any)

	if len(oneOf) != 3 {
		t.Errorf("run items should have 3 oneOf options, got %d", len(oneOf))
	}

	cmdOption := oneOf[0].(map[string]any)
	cmdProps := cmdOption["properties"].(map[string]any)
	if _, ok := cmdProps["cmd"]; !ok {
		t.Error("first oneOf should have 'cmd' property")
	}

	taskOption := oneOf[1].(map[string]any)
	taskProps := taskOption["properties"].(map[string]any)
	if _, ok := taskProps["task"]; !ok {
		t.Error("second oneOf should have 'task' property")
	}

	logOption := oneOf[2].(map[string]any)
	logProps := logOption["properties"].(map[string]any)
	if _, ok := logProps["log"]; !ok {
		t.Error("third oneOf should have 'log' property")
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

	requiredDefs := []string{"Task", "TaskName", "Include", "Platform"}
	for _, def := range requiredDefs {
		if _, ok := defs[def]; !ok {
			t.Errorf("schema should have definition for %q", def)
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
		t.Errorf("Task additionalProperties should be false, got %v", additionalProps)
	}
}

func TestTaskRunRequiresAtLeastOneItem(t *testing.T) {
	s := schema.GenerateSchema()

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	defs := parsed["$defs"].(map[string]any)
	taskDef := defs["Task"].(map[string]any)
	props := taskDef["properties"].(map[string]any)
	runProp := props["run"].(map[string]any)

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
	s := schema.GenerateSchema()

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	defs := parsed["$defs"].(map[string]any)
	taskNameDef, ok := defs["TaskName"].(map[string]any)
	if !ok {
		t.Fatal("schema should have TaskName definition")
	}

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
