package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

func Parse(path string) (TaskMap, error) {
	ctx := NewParseContext()
	return parseWithContext(path, ctx)
}

func parseWithContext(path string, ctx *ParseContext) (TaskMap, error) {
	log.Debug("Starting to parse Babfile", "path", path)

	if err := validatePath(path); err != nil {
		return nil, fmt.Errorf("invalid babfile path: %w", err)
	}

	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		log.Debug("Failed to read Babfile", "path", path, "error", err)
		return nil, fmt.Errorf("failed to read Babfile: %w", err)
	}
	log.Debug("Successfully read Babfile", "size", len(data))

	var raw interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		log.Debug("Failed to unmarshal YAML", "error", err)
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	normalized := normalizeMap(raw)
	rootMap, ok := normalized.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("root of Babfile must be a map")
	}
	log.Debug("Successfully unmarshaled YAML", "top-level-keys", len(rootMap))

	includes, err := parseIncludes(rootMap, absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse includes: %w", err)
	}

	delete(rootMap, keyIncludes)

	tasksRaw, exists := rootMap[keyTasks]
	if !exists {
		return nil, fmt.Errorf("babfile must contain a 'tasks' key")
	}

	tasksSection, ok := safeMapCast(tasksRaw)
	if !ok {
		return nil, fmt.Errorf("'tasks' must be a map, got %T", tasksRaw)
	}

	tasks := make(TaskMap)
	if err := flatten(tasksSection, "", tasks); err != nil {
		log.Debug("Failed to flatten tasks", "error", err)
		return nil, err
	}

	for namespace, path := range includes {
		if err := resolveInclude(namespace, path, tasks, ctx); err != nil {
			return nil, err
		}
	}

	if err := ValidateDependencies(tasks); err != nil {
		log.Debug("Failed to validate dependencies", "error", err)
		return nil, fmt.Errorf("dependency validation failed: %w", err)
	}

	log.Debug("Successfully parsed Babfile", "task-count", len(tasks))
	return tasks, nil
}
