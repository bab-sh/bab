package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

func Parse(path string) (TaskMap, error) {
	log.Debug("Starting to parse Babfile", "path", path)

	if path == "" {
		return nil, fmt.Errorf("babfile path cannot be empty")
	}

	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath)
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

	tasks := make(TaskMap)
	if err := flatten(rootMap, "", tasks); err != nil {
		log.Debug("Failed to flatten tasks", "error", err)
		return nil, err
	}

	log.Debug("Successfully parsed Babfile", "task-count", len(tasks))
	return tasks, nil
}
