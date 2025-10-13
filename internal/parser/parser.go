package parser

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

func Parse(path string) (TaskMap, error) {
	log.Debug("Starting to parse Babfile", "path", path)

	data, err := os.ReadFile(path)
	if err != nil {
		log.Debug("Failed to read Babfile", "path", path, "error", err)
		return nil, err
	}
	log.Debug("Successfully read Babfile", "size", len(data))

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		log.Debug("Failed to unmarshal YAML", "error", err)
		return nil, err
	}
	log.Debug("Successfully unmarshaled YAML", "top-level-keys", len(raw))

	tasks := make(TaskMap)
	if err := flatten(raw, "", tasks); err != nil {
		log.Debug("Failed to flatten tasks", "error", err)
		return nil, err
	}

	log.Debug("Successfully parsed Babfile", "task-count", len(tasks))
	return tasks, nil
}

func flatten(data map[string]interface{}, prefix string, tasks TaskMap) error {
	for key, val := range data {
		name := prefix + ":" + key
		if prefix == "" {
			name = key
		}
		log.Debug("Processing task node", "name", name)

		// FixMe - when a key is a number it gets interpreted into a interface instead of string
		// FixMe - Correctly handle uncompleted tasks that contains partial keys like desc or run, these keys shouldn't be interpreted as new tasks
		m, ok := val.(map[string]interface{})
		if !ok {
			log.Debug("Expected map but got different type", "name", name, "type", fmt.Sprintf("%T", val))
			return fmt.Errorf("expected map at %s", name)
		}

		if run, hasRun := m["run"]; hasRun {
			log.Debug("Found executable task", "name", name)
			task := &Task{
				Name: name,
			}

			if desc, ok := m["desc"]; ok {
				task.Description = fmt.Sprint(desc)
				log.Debug("Task has description", "name", name, "desc", task.Description)
			}

			switch v := run.(type) {
			case string:
				task.Commands = []string{v}
				log.Debug("Task has single command", "name", name)
			case []interface{}:
				task.Commands = make([]string, len(v))
				for i, cmd := range v {
					task.Commands[i] = fmt.Sprint(cmd)
				}
				log.Debug("Task has multiple commands", "name", name, "count", len(v))
			default:
				task.Commands = []string{fmt.Sprint(run)}
				log.Debug("Task has command of unknown type, converted to string", "name", name)
			}

			tasks[name] = task
			log.Debug("Registered task", "name", name, "commands", len(task.Commands))
		} else {
			log.Debug("Node is a namespace, recursing", "name", name)
			if err := flatten(m, name, tasks); err != nil {
				return err
			}
		}
	}

	return nil
}
