package parser

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
)

func processTaskNode(taskMap map[string]interface{}, taskName string, tasks TaskMap) error {
	runCmd, hasRun := taskMap[keyRun]

	if hasRun {
		task, err := buildTask(taskName, taskMap, runCmd)
		if err != nil {
			return err
		}
		tasks[taskName] = task
		log.Debug("Registered task", "name", taskName, "commands", len(task.Commands))
	}

	nestedKeys := getNestedKeys(taskMap)
	if len(nestedKeys) > 0 {
		log.Debug("Node has nested tasks, recursing", "name", taskName)
		for _, key := range nestedKeys {
			nestedMap := map[string]interface{}{key: taskMap[key]}
			if err := flatten(nestedMap, taskName, tasks); err != nil {
				return err
			}
		}
	}

	return nil
}

func buildTask(name string, taskMap map[string]interface{}, runCmd interface{}) (*Task, error) {
	task := &Task{Name: name}

	if desc, ok := taskMap[keyDesc]; ok {
		if descStr, ok := desc.(string); ok {
			task.Description = descStr
		} else {
			task.Description = fmt.Sprint(desc)
		}
		log.Debug("Task has description", "name", name, "desc", task.Description)
	}

	commands, err := parseCommands(name, runCmd)
	if err != nil {
		return nil, err
	}
	task.Commands = commands

	log.Debug("Found executable task", "name", name, "commands", len(commands))
	return task, nil
}

func parseCommands(taskName string, runCmd interface{}) ([]string, error) {
	if runCmd == nil {
		return nil, fmt.Errorf("task %q has nil 'run' command", taskName)
	}

	switch v := runCmd.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, fmt.Errorf("task %q has empty 'run' command", taskName)
		}
		log.Debug("Task has single command", "name", taskName)
		return []string{v}, nil

	case []interface{}:
		if len(v) == 0 {
			return nil, fmt.Errorf("task %q has empty 'run' command list", taskName)
		}
		commands := make([]string, 0, len(v))
		for i, cmd := range v {
			cmdStr := fmt.Sprint(cmd)
			if strings.TrimSpace(cmdStr) == "" {
				return nil, fmt.Errorf("task %q has empty command at index %d", taskName, i)
			}
			commands = append(commands, cmdStr)
		}
		log.Debug("Task has multiple commands", "name", taskName, "count", len(v))
		return commands, nil

	default:
		cmdStr := fmt.Sprint(runCmd)
		if strings.TrimSpace(cmdStr) == "" {
			return nil, fmt.Errorf("task %q has empty 'run' command", taskName)
		}
		log.Debug("Task has command of unknown type, converted to string", "name", taskName)
		return []string{cmdStr}, nil
	}
}
