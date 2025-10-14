package parser

import (
	"fmt"

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
		descStr, err := safeStringCast(desc)
		if err != nil {
			return nil, fmt.Errorf("invalid description for task %q: %w", name, err)
		}
		task.Description = descStr
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

	if cmdStr, ok := runCmd.(string); ok {
		if err := validateCommand(cmdStr); err != nil {
			return nil, fmt.Errorf("task %q has invalid 'run' command: %w", taskName, err)
		}
		log.Debug("Task has single command", "name", taskName)
		return []string{cmdStr}, nil
	}

	if cmdSlice, ok := safeSliceCast(runCmd); ok {
		if len(cmdSlice) == 0 {
			return nil, fmt.Errorf("task %q has empty 'run' command list", taskName)
		}
		commands := make([]string, 0, len(cmdSlice))
		for i, cmd := range cmdSlice {
			cmdStr, err := safeStringCast(cmd)
			if err != nil {
				return nil, fmt.Errorf("task %q has invalid command at index %d: %w", taskName, i, err)
			}
			if err := validateCommand(cmdStr); err != nil {
				return nil, fmt.Errorf("task %q has invalid command at index %d: %w", taskName, i, err)
			}
			commands = append(commands, cmdStr)
		}
		log.Debug("Task has multiple commands", "name", taskName, "count", len(cmdSlice))
		return commands, nil
	}

	cmdStr, err := safeStringCast(runCmd)
	if err != nil {
		return nil, fmt.Errorf("task %q has invalid 'run' command: %w", taskName, err)
	}
	if err := validateCommand(cmdStr); err != nil {
		return nil, fmt.Errorf("task %q has invalid 'run' command: %w", taskName, err)
	}
	log.Debug("Task has command of unknown type, converted to string", "name", taskName)
	return []string{cmdStr}, nil
}
