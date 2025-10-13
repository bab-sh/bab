package parser

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	baberrors "github.com/bab-sh/bab/internal/errors"
	"github.com/bab-sh/bab/internal/registry"
	"gopkg.in/yaml.v3"
)

type Parser struct {
	registry registry.Registry
	ctx      context.Context
}

func New(reg registry.Registry) *Parser {
	return &Parser{
		registry: reg,
		ctx:      context.Background(),
	}
}

func NewWithContext(ctx context.Context, reg registry.Registry) *Parser {
	return &Parser{
		registry: reg,
		ctx:      ctx,
	}
}

func (p *Parser) ParseFile(filename string) error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
	}

	cleanPath := filepath.Clean(filename)
	file, err := os.Open(cleanPath)
	if err != nil {
		return baberrors.NewParseError(cleanPath, fmt.Errorf("failed to open file: %w", err))
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to close babfile: %v\n", closeErr)
		}
	}()

	return p.Parse(file)
}

func (p *Parser) Parse(reader io.Reader) error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
	}

	decoder := yaml.NewDecoder(reader)

	var root map[string]interface{}
	if err := decoder.Decode(&root); err != nil {
		return baberrors.NewParseError("", fmt.Errorf("failed to parse YAML: %w", err))
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
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	default:
	}

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
		return baberrors.NewTaskValidationError(name, "run", "task has no commands")
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
