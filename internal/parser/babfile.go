// Package parser provides functionality for parsing Babfiles and registering tasks.
package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bab-sh/bab/internal/registry"
	"gopkg.in/yaml.v3"
)

// Parser parses Babfiles and registers tasks.
type Parser struct {
	registry registry.Registry
}

// New creates a new Parser with the given registry.
func New(reg registry.Registry) *Parser {
	return &Parser{
		registry: reg,
	}
}

// ParseFile parses a Babfile from the filesystem.
func (p *Parser) ParseFile(filename string) error {
	cleanPath := filepath.Clean(filename)
	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to open babfile: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to close babfile: %v\n", closeErr)
		}
	}()

	return p.Parse(file)
}

// Parse parses a Babfile from an io.Reader.
func (p *Parser) Parse(reader io.Reader) error {
	decoder := yaml.NewDecoder(reader)

	var root map[string]interface{}
	if err := decoder.Decode(&root); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	return p.parseNode(root, "")
}

func (p *Parser) parseNode(node map[string]interface{}, prefix string) error {
	for key, value := range node {
		fullName := key
		if prefix != "" {
			fullName = prefix + ":" + key
		}

		var taskMap map[string]interface{}
		switch v := value.(type) {
		case map[interface{}]interface{}:
			taskMap = convertMap(v)
		case map[string]interface{}:
			taskMap = v
		default:
			continue
		}

		if isTask(taskMap) {
			if err := p.registerTask(fullName, taskMap); err != nil {
				return err
			}
		} else {
			if err := p.parseNode(taskMap, fullName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Parser) registerTask(name string, taskMap map[string]interface{}) error {
	task := registry.NewTask(name)

	if desc, ok := taskMap["desc"].(string); ok {
		task.Description = desc
	}

	if run, ok := taskMap["run"]; ok {
		switch r := run.(type) {
		case string:
			task.Commands = []string{r}
		case []interface{}:
			for _, cmd := range r {
				if cmdStr, ok := cmd.(string); ok {
					task.Commands = append(task.Commands, cmdStr)
				}
			}
		}
	}

	if len(task.Commands) == 0 {
		return fmt.Errorf("task %s has no 'run' commands", name)
	}

	return p.registry.Register(task)
}

func isTask(node map[string]interface{}) bool {
	_, hasDesc := node["desc"]
	runValue, hasRun := node["run"]

	if hasRun {
		switch runValue.(type) {
		case string, []interface{}:
			return true
		}
	}

	return hasDesc
}

func convertMap(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		if key, ok := k.(string); ok {
			switch value := v.(type) {
			case map[interface{}]interface{}:
				result[key] = convertMap(value)
			default:
				result[key] = value
			}
		}
	}
	return result
}
