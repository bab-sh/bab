package parser

import (
	"fmt"

	"github.com/charmbracelet/log"
)

func flatten(data map[string]interface{}, prefix string, tasks TaskMap) error {
	for key, val := range data {
		taskName := buildTaskName(prefix, key)
		log.Debug("Processing task node", "name", taskName)

		taskMap, ok := safeMapCast(val)
		if !ok {
			log.Debug("Expected map but got different type", "name", taskName, "type", fmt.Sprintf("%T", val))
			return fmt.Errorf("task %q must be a map, got %T", taskName, val)
		}

		if err := processTaskNode(taskMap, taskName, tasks); err != nil {
			return err
		}
	}

	return nil
}

func buildTaskName(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + ":" + key
}

func getNestedKeys(taskMap map[string]interface{}) []string {
	var keys []string
	for k := range taskMap {
		if k != keyRun && k != keyDesc {
			keys = append(keys, k)
		}
	}
	return keys
}
