package parser

import (
	"fmt"
	"io"
	"os"

	"github.com/bab-sh/bab/internal/registry"
	"gopkg.in/yaml.v3"
)

type Parser struct {
	registry registry.Registry
}

func New(reg registry.Registry) *Parser {
	return &Parser{
		registry: reg,
	}
}

func (p *Parser) ParseFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open babfile: %w", err)
	}
	defer file.Close()

	return p.Parse(file)
}

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

		switch v := value.(type) {
		case map[interface{}]interface{}:
			converted := convertMap(v)
			if isTask(converted) {
				if err := p.registerTask(fullName, converted); err != nil {
					return err
				}
			} else {
				if err := p.parseNode(converted, fullName); err != nil {
					return err
				}
			}

		case map[string]interface{}:
			if isTask(v) {
				if err := p.registerTask(fullName, v); err != nil {
					return err
				}
			} else {
				if err := p.parseNode(v, fullName); err != nil {
					return err
				}
			}

		default:
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
	_, hasRun := node["run"]
	return hasDesc || hasRun
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
